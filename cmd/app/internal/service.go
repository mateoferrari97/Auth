package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

const mySigningKey = "secret"

const (
	ErrUnauthorized = errors.New("")
)

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