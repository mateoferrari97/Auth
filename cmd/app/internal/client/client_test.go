package client

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type client struct {
	mock.Mock
}

func (c *client) Get(path string) (*http.Response, error) {
	args := c.Called(path)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestClient_GetUserEmailFromAccessToken(t *testing.T) {
	// Given
	w := httptest.NewRecorder()
	w.Code = http.StatusOK
	w.Body = bytes.NewBuffer([]byte(`{"email": "luken@gmail.com"}`))

	c := &client{}
	c.On("Get", "https://www.googleapis.com/oauth2/v2/userinfo?access_token=ble").Return(w.Result(), nil)

	token := "ble"
	nc := NewClient(c)

	// When
	resp, err := nc.GetUserEmailFromAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, "luken@gmail.com", resp)
}

func TestClient_GetUserEmailFromAccessToken_GetError(t *testing.T) {
	// Given
	var r *http.Response

	c := &client{}
	c.On("Get", "https://www.googleapis.com/oauth2/v2/userinfo?access_token=ble").
		Return(r, errors.New("internal server error"))

	token := "ble"
	nc := NewClient(c)

	// When
	_, err := nc.GetUserEmailFromAccessToken(token)

	// Then
	require.EqualError(t, err, "getting user information: internal server error")
}

func TestClient_GetUserEmailFromAccessToken_DecodeError(t *testing.T) {
	// Given
	w := httptest.NewRecorder()
	w.Code = http.StatusOK
	w.Body = bytes.NewBuffer([]byte(`{"email": error}`))

	c := &client{}
	c.On("Get", "https://www.googleapis.com/oauth2/v2/userinfo?access_token=ble").Return(w.Result(), nil)

	token := "ble"
	nc := NewClient(c)

	// When
	_, err := nc.GetUserEmailFromAccessToken(token)

	// Then
	require.EqualError(t, err, "decoding user information from google: invalid character 'e' looking for beginning of value")
}
