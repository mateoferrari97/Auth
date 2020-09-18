package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mateoferrari97/auth/internal"
	"github.com/stretchr/testify/require"
)

func TestServer_Wrap(t *testing.T) {
	// Given
	s := NewServer()

	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	s.Wrap(http.MethodGet, "/users/me", func(w http.ResponseWriter, r *http.Request) error {
		return nil
	})

	// When
	resp, err := http.Get(fmt.Sprintf("%s/users/me", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestServer_Wrap_HandlerUnknownError(t *testing.T) {
	// Given
	s := NewServer()

	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	s.Wrap(http.MethodGet, "/users/me", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("internal server error")
	})

	// When
	resp, err := http.Get(fmt.Sprintf("%s/users/me", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestServer_Wrap_HandlerInternalError(t *testing.T) {
	// Given
	s := NewServer()

	ts := httptest.NewServer(s.Router)
	defer ts.Close()

	s.Wrap(http.MethodGet, "/users/me", func(w http.ResponseWriter, r *http.Request) error {
		return internal.NewError("internal server error", http.StatusInternalServerError)
	})

	// When
	resp, err := http.Get(fmt.Sprintf("%s/users/me", ts.URL))
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
