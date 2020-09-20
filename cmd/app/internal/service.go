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
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrInvalidToken          = errors.New("can't access to the resource. invalid token")
	ErrAlteredTokenClaims    = errors.New("can't access to the resource. claims don't match from original token")
	ErrResourceNotFound      = errors.New("resource not found")
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

type GoogleClient interface {
	GetUserEmailFromAccessToken(accessToken string) (string, error)
}

type Repository interface {
	SaveUser(newUser NewUser) error
	GetUserByEmail(email string) (User, error)
	FindUserByEmail(email string) error
}

type Service struct {
	UserRepository Repository
	Cli            GoogleClient
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

func NewService(repository Repository) *Service {
	return &Service{
		UserRepository: repository,
	}
}

func (s *Service) Register(newUser RegisterRequest) error {
	err := s.UserRepository.FindUserByEmail(newUser.Email)
	if err == nil {
		return ErrResourceAlreadyExists
	}

	if !errors.Is(err, ErrNotFound) {
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
		return User{}, ErrInvalidToken
	}

	c, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, ErrAlteredTokenClaims
	}

	subject, ok := c["sub"].(string)
	if !ok {
		return User{}, fmt.Errorf("something wrong happend while parsing subject")
	}

	var u User
	if err := json.Unmarshal([]byte(subject), &u); err != nil {
		return User{}, fmt.Errorf("decoding claims: %v", err)
	}

	user, err := s.UserRepository.GetUserByEmail(u.Email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return User{}, err
	}

	if errors.Is(err, ErrNotFound) {
		return User{}, ErrResourceNotFound
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

	email, err := s.Cli.GetUserEmailFromAccessToken(token.AccessToken)
	if err != nil {
		return "", fmt.Errorf("getting user email from google: %v", err)
	}

	user, err := s.UserRepository.GetUserByEmail(email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", err
	}

	if errors.Is(err, ErrNotFound) {
		return "", ErrResourceNotFound
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
