package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofrs/uuid"
	"github.com/mateoferrari97/auth/internal"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var mySigningKey = os.Getenv("PRIVATE_KEY")

var config = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8081/login/google/callback",
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
	},
}

const state = "random"

type Client interface {
	GetUserEmailFromAccessToken(accessToken string) (string, error)
}

type Repository interface {
	SaveUser(newUser NewUser) error
	GetUserByEmail(email string) (User, error)
	FindUserByEmail(email string) error
}

type Service struct {
	UserRepository Repository
	Client         Client
}

type NewUser struct {
	ID        string
	Firstname string
	Lastname  string
	Email     string
	Password  string
}

type User struct {
	ID        string `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
}

func NewService(repository Repository, client Client) *Service {
	return &Service{
		UserRepository: repository,
		Client:         client,
	}
}

func (s *Service) Register(newUser RegisterRequest) error {
	err := s.UserRepository.FindUserByEmail(newUser.Email)
	if err == nil {
		return fmt.Errorf("%w: user already exists", internal.ErrResourceAlreadyExists)
	}

	if !errors.Is(err, internal.ErrResourceNotFound) {
		return err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("creating user: %v", err)
	}

	b, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), 10)
	if err != nil {
		return fmt.Errorf("generating password: %v", err)
	}

	user := NewUser{
		ID:        id.String(),
		Firstname: newUser.Firstname,
		Lastname:  newUser.Lastname,
		Email:     newUser.Email,
		Password:  string(b),
	}

	return s.UserRepository.SaveUser(user)
}

func (s *Service) Authorize(token string) (User, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(mySigningKey), nil
	})

	if err != nil {
		return User{}, fmt.Errorf("parsing token: %v", err)
	}

	if !t.Valid {
		return User{}, internal.ErrInvalidToken
	}

	c, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, internal.ErrAlteredTokenClaims
	}

	subject, ok := c["sub"].(string)
	if !ok {
		return User{}, errors.New("something wrong happened while parsing subject")
	}

	var u User
	if err := json.Unmarshal([]byte(subject), &u); err != nil {
		return User{}, fmt.Errorf("decoding claims: %v", err)
	}

	user, err := s.UserRepository.GetUserByEmail(u.Email)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Service) LoginWithGoogle() (string, error) {
	return config.AuthCodeURL(state), nil
}

func (s *Service) LoginWithGoogleCallback(code string) (string, error) {
	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", fmt.Errorf("getting token from google: %v", err)
	}

	email, err := s.Client.GetUserEmailFromAccessToken(token.AccessToken)
	if err != nil {
		return "", fmt.Errorf("getting user email from google: %v", err)
	}

	user, err := s.UserRepository.GetUserByEmail(email)
	if err != nil && !errors.Is(err, internal.ErrResourceAlreadyExists) {
		return "", err
	}

	if errors.Is(err, internal.ErrResourceAlreadyExists) {
		return "", internal.ErrResourceNotFound
	}

	t, err := newJWT(user)
	if err != nil {
		return "", fmt.Errorf("authorizing user: %v", err)
	}

	return t, nil
}

func newJWT(user User) (string, error) {
	u, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("marshaling user: %v", err)
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		Subject:   string(u),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(mySigningKey))
	if err != nil {
		return "", fmt.Errorf("creating token: %v", err)
	}

	return t, nil
}
