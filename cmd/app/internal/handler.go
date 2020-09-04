package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mateoferrari97/auth/cmd/server"
	"github.com/mateoferrari97/auth/internal"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
	"time"
)

const (
	getHome                    = "/"
	getPing                    = "/ping"
	getMe                      = "/users/me"
	postUsers                  = "/users"
	getLogout                  = "/logout"
	getLoginWithGoogle         = "/login/google"
	getLoginWithGoogleCallback = "/login/google/callback"
)

var _v = validator.New()

type Wrapper interface {
	Wrap(method string, pattern string, handler server.HandlerFunc)
}

type Handler struct {
	Wrapper
}

func NewHandler(wrapper Wrapper) *Handler {
	return &Handler{wrapper}
}

func (h *Handler) Ping() {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprintln(w, "pong")

		return nil
	}

	h.Wrap(http.MethodGet, getPing, wrapH)
}

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterHandler func(req RegisterRequest) (string, error)

func (h *Handler) RouteRegister(handler RegisterHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return handleError(err)
		}

		if err := validateRegisterInformation(req); err != nil {
			return handleError(err)
		}

		token, err := handler(req)
		if err != nil {
			return handleError(err)
		}

		c := &http.Cookie{
			Name:     "authorization",
			Value:    token,
			Path:     getHome,
			HttpOnly: true,
		}

		http.SetCookie(w, c)
		return internal.RespondJSON(w, nil, http.StatusOK)
	}

	h.Wrap(http.MethodPost, postUsers, wrapH)
}

func validateRegisterInformation(userInformation RegisterRequest) error {
	if err := _v.Struct(userInformation); err != nil {
		return err
	}

	return validatePassword(userInformation.Password)
}

func validatePassword(password string) error {
	var validSequences = []string{
		"abcdefghijklmñopqrstuvwxyz",
		"ABCDEFGHIJKLMÑOPQRSTUUVWXYZ",
		"123456789",
		"!#$%&*?@",
	}

	for _, sequence := range validSequences {
		hasCharacter := false
		for _, c := range sequence {
			if strings.ContainsRune(password, c) {
				hasCharacter = true
				break
			}
		}

		if !hasCharacter {
			return errors.New("password is weak")
		}
	}

	return nil
}

type LoginWithGoogleHandler func() string

func (h *Handler) RouteLoginWithGoogle(handler LoginWithGoogleHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, handler(), http.StatusTemporaryRedirect)

		return nil
	}

	h.Wrap(http.MethodGet, getLoginWithGoogle, wrapH)
}

type LoginWithGoogleCallbackHandler func(code string) (string, error)

func (h *Handler) RouteLoginWithGoogleCallback(handler LoginWithGoogleCallbackHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		code := r.FormValue("code")
		if code == "" {
			return handleError(errors.New("code is required"))
		}

		token, err := handler(code)
		if err != nil {
			return handleError(err)
		}

		c := &http.Cookie{
			Name:     "authorization",
			Value:    token,
			Path:     getHome,
			HttpOnly: true,
		}

		http.SetCookie(w, c)
		return internal.RespondJSON(w, nil, http.StatusOK)
	}

	h.Wrap(http.MethodGet, getLoginWithGoogleCallback, wrapH)
}

func (h *Handler) RouteLogout() {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("authorization")
		if err == nil {
			c = &http.Cookie{
				Name:     "authorization",
				Expires:  time.Now().Add(-1 * time.Hour),
				MaxAge:   -1,
				HttpOnly: true,
			}

			http.SetCookie(w, c)
		}

		internal.RespondJSON(w, nil, http.StatusOK)
		return nil
	}

	h.Wrap(http.MethodGet, getLogout, wrapH)
}

type AuthorizeMeHandler func(token string) (User, error)

func (h *Handler) RouteMe(handler AuthorizeMeHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		c, err := r.Cookie("authorization")
		if err != nil {
			return handleError(ErrInvalidToken)
		}

		user, err := handler(c.Value)
		if err != nil {
			return handleError(err)
		}

		return internal.RespondJSON(w, user, http.StatusOK)
	}

	h.Wrap(http.MethodGet, getMe, wrapH)
}

func handleError(err error) error {
	message := err.Error()

	switch err {
	case ErrResourceNotFound:
		return internal.NewError(message, http.StatusNotFound)
	case ErrInvalidToken:
		return internal.NewError(message, http.StatusForbidden)
	case ErrAlteredTokenClaims:
		return internal.NewError(message, http.StatusForbidden)
	case ErrResourceAlreadyExists:
		return internal.NewError(message, http.StatusConflict)
	default:
		return internal.NewError(message, http.StatusInternalServerError)
	}
}
