package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
	"time"
)

const mySigningKey = "secret"

var config = &oauth2.Config{
	ClientID:     "176380119677-5r99e6b9jqho14cvfpc0inmeb1m48gkr.apps.googleusercontent.com",
	ClientSecret: "h-D4PY8U_-uu-JInbPwDQ_Es",
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8081/login/google/callback",
	Scopes:       []string{
		"https://www.googleapis.com/auth/userinfo.email",
	},
}

type Service struct {
	DB map[string]User
}

type User struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewService() *Service{
	return &Service{
		DB: make(map[string]User),
	}
}

func (s *Service) Register(newUser RegisterRequest) (string, error){
	if _, exist := s.DB[newUser.Email]; exist {
		return "", fmt.Errorf("user already exists")
	}

	user := User{
		Name:      newUser.Name,
		Email:     newUser.Email,
		Password:  newUser.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.DB[newUser.Email] = user

	token, err := newJWT(user)
	if err != nil {
		return "", fmt.Errorf("authorizing user: %v", err)
	}

	return token, nil
}

func newJWT(user User) (string, error){
	u, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("marshaling user: %v", err)
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		Subject:    string(u),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(mySigningKey))
	if err != nil {
		return "", fmt.Errorf("creating token: %v", err)
	}

	return t, nil
}

func (s *Service) Authorize(token string) (User, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (i interface{}, err error) {
		return []byte(mySigningKey), nil
	})

	if err != nil {
		return User{}, fmt.Errorf("parsing token: %v", err)
	}

	if !t.Valid {
		return User{}, errors.New("invalid token")
	}

	c, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, fmt.Errorf("invalid claims type: %v", err)
	}

	subject, ok := c["sub"].(string)
	if !ok {
		return User{}, fmt.Errorf("something wrong happend while parsing subject")
	}

	var user User
	if err := json.Unmarshal([]byte(subject), &user); err != nil {
		return User{}, fmt.Errorf("decoding claims: %v", err)
	}

	if _, exist := s.DB[user.Email]; !exist {
		return User{}, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Service) LoginWithGoogle() string {
	return config.AuthCodeURL("random")
}

func (s *Service) LoginWithGoogleCallback(code string) (string, error) {
	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		return "", fmt.Errorf("getting token from google: %v", err)
	}

	path := fmt.Sprintf("https://www.googleapis.com/oauth2/v2/userinfo?access_token=%s", token.AccessToken)
	resp, err := http.Get(path)
	if err != nil {
		return "", fmt.Errorf("getting user information: %v", err)
	}

	defer resp.Body.Close()

	var userInformation struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInformation); err != nil {
		return "", fmt.Errorf("decoding user information from google: %v", err)
	}

	user, exist := s.DB[userInformation.Email]
	if !exist {
		return "", fmt.Errorf("user with email %s: not found", userInformation.Email)
	}

	t, err := newJWT(user)
	if err != nil {
		return "", fmt.Errorf("authorizing user: %v", err)
	}

	return t, nil
}