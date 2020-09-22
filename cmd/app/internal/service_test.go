package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mateoferrari97/auth/internal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type repository struct {
	mock.Mock
}

func (r *repository) SaveUser(newUser NewUser) error {
	return r.Called(newUser).Error(0)
}

func (r *repository) GetUserByEmail(email string) (User, error) {
	args := r.Called(email)
	return args.Get(0).(User), args.Error(1)
}

func (r *repository) FindUserByEmail(email string) error {
	return r.Called(email).Error(0)
}

func TestRegister(t *testing.T) {
	// Given
	u := RegisterRequest{
		Firstname: "Mateo",
		Lastname:  "Ferrari Coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "luk1n",
	}

	r := &repository{}
	r.On("FindUserByEmail", u.Email).Return(internal.ErrResourceNotFound)
	r.On("SaveUser", mock.AnythingOfType("NewUser")).Return(nil)

	s := NewService(r, nil)

	// When
	err := s.Register(u)

	// Then
	require.NoError(t, err)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	// Given
	u := RegisterRequest{
		Firstname: "Mateo",
		Lastname:  "Ferrari Coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "luk1n",
	}

	r := &repository{}
	r.On("FindUserByEmail", u.Email).Return(nil)

	s := NewService(r, nil)

	// When
	err := s.Register(u)

	// Then
	require.EqualError(t, err, "resource already exists: user already exists")
}

func TestRegister_FindingUserByEmail_InternalServerError(t *testing.T) {
	// Given
	u := RegisterRequest{
		Firstname: "Mateo",
		Lastname:  "Ferrari Coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "luk1n",
	}

	r := &repository{}
	r.On("FindUserByEmail", u.Email).Return(errors.New("repository error"))

	s := NewService(r, nil)

	// When
	err := s.Register(u)

	// Then
	require.EqualError(t, err, "repository error")
}

func TestRegister_SavingUser_InternalServerError(t *testing.T) {
	// Given
	u := RegisterRequest{
		Firstname: "Mateo",
		Lastname:  "Ferrari Coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "luk1n",
	}

	r := &repository{}
	r.On("FindUserByEmail", u.Email).Return(internal.ErrResourceNotFound)
	r.On("SaveUser", mock.AnythingOfType("NewUser")).Return(errors.New("repository error"))

	s := NewService(r, nil)

	// When
	err := s.Register(u)

	// Then
	require.EqualError(t, err, "repository error")
}

func TestAuthorize(t *testing.T) {
	// Given
	u := User{
		ID:        "id",
		Firstname: "luken",
		Lastname:  "straka",
		Email:     "mateo.ferrari97@gmail.com",
	}

	token, _ := _newJWT(u)

	r := &repository{}
	r.On("GetUserByEmail", u.Email).Return(u, nil)

	s := NewService(r, nil)

	// When
	resp, err := s.Authorize(token)
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, "id", resp.ID)
	require.Equal(t, "luken", resp.Firstname)
	require.Equal(t, "straka", resp.Lastname)
	require.Equal(t, "mateo.ferrari97@gmail.com", resp.Email)
}

func TestAuthorize_ParsingTokenError(t *testing.T) {
	// Given
	token := "invalid token"
	s := NewService(&repository{}, nil)

	// When
	_, err := s.Authorize(token)

	// Then
	require.EqualError(t, err, "parsing token: token contains an invalid number of segments")
}

func TestAuthorize_RepositoryInternalServerError(t *testing.T) {
	// Given
	u := User{
		ID:        "id",
		Firstname: "luken",
		Lastname:  "straka",
		Email:     "mateo.ferrari97@gmail.com",
	}

	token, _ := _newJWT(u)

	r := &repository{}
	r.On("GetUserByEmail", u.Email).Return(User{}, errors.New("internal server error"))

	s := NewService(r, nil)

	// When
	_, err := s.Authorize(token)

	// Then
	require.EqualError(t, err, "internal server error")
}

func TestAuthorize_RepositoryNotFoundError(t *testing.T) {
	// Given
	u := User{
		ID:        "id",
		Firstname: "luken",
		Lastname:  "straka",
		Email:     "mateo.ferrari97@gmail.com",
	}

	token, _ := _newJWT(u)

	r := &repository{}
	r.On("GetUserByEmail", u.Email).Return(User{}, internal.ErrResourceNotFound)

	s := NewService(r, nil)

	// When
	_, err := s.Authorize(token)

	// Then
	require.EqualError(t, err, "resource not found")
}

func TestLoginWithGoogle(t *testing.T) {
	// Given
	s := NewService(&repository{}, nil)
	expectedURL := "https://accounts.google.com/o/oauth2/auth?client_id=176380119677-5r99e6b9jqho14cvfpc0inmeb1m48gkr.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8081%2Flogin%2Fgoogle%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email&state=random"

	// When
	resp, err := s.LoginWithGoogle()
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, expectedURL, resp)
}

func TestLoginWithGoogle_Error(t *testing.T) {
	// Given
	s := NewService(&repository{}, nil)
	expectedURL := "https://accounts.google.com/o/oauth2/auth?client_id=176380119677-5r99e6b9jqho14cvfpc0inmeb1m48gkr.apps.googleusercontent.com&redirect_uri=http%3A%2F%2Flocalhost%3A8081%2Flogin%2Fgoogle%2Fcallback&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email&state=random"

	// When
	resp, err := s.LoginWithGoogle()
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, expectedURL, resp)
}

func _newJWT(user User) (string, error) {
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
