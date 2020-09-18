package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mateoferrari97/auth/cmd/server"
	"github.com/stretchr/testify/require"
)

func TestHandler_RouteRegister(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteRegister(func(_ RegisterRequest) error {
		return nil
	})

	b := []byte(`{
		"firstname": "Mateo",
		"lastname": "Ferrari Coronel",
		"email": "mateo.ferrari97@gmail.com",
		"password": "KeepImproving1!"
	}`)

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Post(fmt.Sprintf("%s/users", ts.URL), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	// Then
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_RouteRegister_UnprocessableEntityError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteRegister(func(_ RegisterRequest) error {
		return nil
	})

	b := []byte(`{
		"firstname": "Other fields are missing",
	}`)

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Post(fmt.Sprintf("%s/users", ts.URL), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "422: unprocessable entity", m)
}

func TestHandler_RouteRegister_PasswordMinLengthError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteRegister(func(_ RegisterRequest) error {
		return nil
	})

	b := []byte(`{
		"firstname": "Mateo",
		"lastname": "Ferrari Coronel",
		"email": "mateo.ferrari97@gmail.com",
		"password": "short"
	}`)

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Post(fmt.Sprintf("%s/users", ts.URL), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	require.Equal(t, "422: unprocessable entity", m)
}

func TestHandler_RouteRegister_WeakPasswordError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteRegister(func(_ RegisterRequest) error {
		return nil
	})

	b := []byte(`{
		"firstname": "Mateo",
		"lastname": "Ferrari Coronel",
		"email": "mateo.ferrari97@gmail.com",
		"password": "John weak"
	}`)

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Post(fmt.Sprintf("%s/users", ts.URL), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "400: weak password", m)
}

func TestHandler_RouteRegister_HandlerError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteRegister(func(_ RegisterRequest) error {
		return errors.New("internal server error")
	})

	b := []byte(`{
		"firstname": "Mateo",
		"lastname": "Ferrari Coronel",
		"email": "mateo.ferrari97@gmail.com",
		"password": "KeepImproving1!"
	}`)

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Post(fmt.Sprintf("%s/users", ts.URL), "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "500: internal server error", m)
}

func TestHandler_RouteLoginWithGoogle(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteLoginWithGoogle(func() (string, error) {
		return "http://login.com", nil
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/login/google", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	// Then
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RouteLoginWithGoogle_HandlerError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteLoginWithGoogle(func() (string, error) {
		return "", errors.New("internal server error")
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/login/google", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "500: internal server error", m)
}

func TestHandler_RouteLoginWithGoogleCallback(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteLoginWithGoogleCallback(func(code string) (string, error) {
		return "token", nil
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/login/google/callback?code=308", ts.URL), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	cookies := resp.Cookies()

	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "authorization" {
			cookie = c
		}
	}

	// Then
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "authorization", cookie.Name)
	require.Equal(t, "token", cookie.Value)
}

func TestHandler_RouteLoginWithGoogleCallback_MissingCodeError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteLoginWithGoogleCallback(func(code string) (string, error) {
		return "token", nil
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/login/google/callback", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "400: bad request", m)
}

func TestHandler_RouteMe(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteMe(func(token string) (User, error) {
		return User{
			ID:        "id",
			Firstname: "luken",
			Lastname:  "straka",
			Email:     "mateo.ferrari97@gmail.com",
		}, nil
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users/me", ts.URL), nil)
	req.AddCookie(&http.Cookie{
		Name:  "authorization",
		Value: "token",
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	var r struct {
		ID        string `json:"id"`
		Firstname string `json:"firstname"`
		Lastname  string `json:"lastname"`
		Email     string `json:"email"`
	}

	_ = json.NewDecoder(resp.Body).Decode(&r)

	defer resp.Body.Close()

	// Then
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "id", r.ID)
	require.Equal(t, "luken", r.Firstname)
	require.Equal(t, "straka", r.Lastname)
	require.Equal(t, "mateo.ferrari97@gmail.com", r.Email)

}

func TestHandler_RouteMe_MissingTokenError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteMe(func(token string) (User, error) {
		return User{
			ID:        "id",
			Firstname: "luken",
			Lastname:  "straka",
			Email:     "mateo.ferrari97@gmail.com",
		}, nil
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users/me", ts.URL), nil)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	require.Equal(t, "403: can't access to the resource. invalid token", m)
}

func TestHandler_RouteMe_HandlerError(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteMe(func(token string) (User, error) {
		return User{}, errors.New("internal server error")
	})

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users/me", ts.URL), nil)
	req.AddCookie(&http.Cookie{
		Name:  "authorization",
		Value: "token",
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	m := decodeErrorMessageFromBody(resp.Body)

	// Then
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.Equal(t, "500: internal server error", m)
}

func TestHandler_RouteLogout(t *testing.T) {
	// Given
	w := server.NewServer()
	h := NewHandler(w)

	h.RouteLogout()

	// When
	ts := httptest.NewServer(w.Router)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/logout", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	cookies := resp.Cookies()

	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "authorization" {
			cookie = c
		}
	}

	// Then
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "", cookie.Value)
	require.True(t, cookie.Expires.Before(time.Now()))
	require.Less(t, cookie.MaxAge, 0)
}

func decodeErrorMessageFromBody(body io.ReadCloser) string {
	var r struct {
		Message string `json:"message"`
	}

	_ = json.NewDecoder(body).Decode(&r)

	return r.Message
}
