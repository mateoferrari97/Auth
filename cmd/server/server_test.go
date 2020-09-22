package server

import (
	"encoding/json"
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

func TestServer_Wrap_HandleError(t *testing.T) {
	tt := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "bad request",
			err:          fmt.Errorf("%w: %v", internal.ErrBadRequest, "some error"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "weak password",
			err:          fmt.Errorf("%w: %v", internal.ErrWeakPassword, "some error"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unprocessable entity",
			err:          fmt.Errorf("%w: %v", internal.ErrUnprocessableEntity, "some error"),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name:         "resource not found",
			err:          fmt.Errorf("%w: %v", internal.ErrResourceNotFound, "some error"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid token",
			err:          fmt.Errorf("%w: %v", internal.ErrInvalidToken, "some error"),
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "altered token",
			err:          fmt.Errorf("%w: %v", internal.ErrAlteredTokenClaims, "some error"),
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "resource already exists",
			err:          fmt.Errorf("%w: %v", internal.ErrResourceAlreadyExists, "some error"),
			expectedCode: http.StatusConflict,
		},
		{
			name:         "internal server error",
			err:          errors.New("internal server error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			s := NewServer()

			ts := httptest.NewServer(s.Router)
			defer ts.Close()

			s.Wrap(http.MethodGet, "/users/me", func(w http.ResponseWriter, r *http.Request) error {
				return tc.err
			})

			// When
			resp, err := http.Get(fmt.Sprintf("%s/users/me", ts.URL))
			if err != nil {
				t.Fatal(err)
			}

			var r struct {
				Message string `json:"message"`
			}

			_ = json.NewDecoder(resp.Body).Decode(&r)

			// Then
			require.Equal(t, tc.expectedCode, resp.StatusCode)
			require.Equal(t, tc.err.Error(), r.Message)
		})
	}
}
