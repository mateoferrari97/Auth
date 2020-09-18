package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mateoferrari97/auth/cmd/server"
	"github.com/mateoferrari97/auth/internal"
	"gopkg.in/go-playground/validator.v9"
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

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrWeakPassword        = errors.New("weak password")
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
	Firstname string `json:"firstname" validate:"required"`
	Lastname  string `json:"lastname" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type RegisterHandler func(req RegisterRequest) error

func (h *Handler) RouteRegister(handler RegisterHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return handleError(ErrUnprocessableEntity)
		}

		if err := validateRegisterInformation(req); err != nil {
			return handleError(err)
		}

		if err := handler(req); err != nil {
			return handleError(err)
		}

		return internal.RespondJSON(w, nil, http.StatusCreated)
	}

	h.Wrap(http.MethodPost, postUsers, wrapH)
}

func validateRegisterInformation(userInformation RegisterRequest) error {
	if err := _v.Struct(userInformation); err != nil {
		return ErrUnprocessableEntity
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
			return ErrWeakPassword
		}
	}

	return nil
}

type LoginWithGoogleHandler func() (string, error)

func (h *Handler) RouteLoginWithGoogle(handler LoginWithGoogleHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		resp, err := handler()
		if err != nil {
			return handleError(err)
		}

		http.Redirect(w, r, resp, http.StatusTemporaryRedirect)

		return nil
	}

	h.Wrap(http.MethodGet, getLoginWithGoogle, wrapH)
}

type LoginWithGoogleCallbackHandler func(code string) (string, error)

func (h *Handler) RouteLoginWithGoogleCallback(handler LoginWithGoogleCallbackHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) error {
		code := r.FormValue("code")
		if code == "" {
			return handleError(ErrBadRequest)
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
		c := &http.Cookie{
			Name:     "authorization",
			Expires:  time.Now().Add(-1 * time.Hour),
			MaxAge:   -1,
		}

		http.SetCookie(w, c)

		return internal.RespondJSON(w, nil, http.StatusOK)
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
	case ErrBadRequest:
		return internal.NewError(message, http.StatusBadRequest)
	case ErrWeakPassword:
		return internal.NewError(message, http.StatusBadRequest)
	case ErrUnprocessableEntity:
		return internal.NewError(message, http.StatusUnprocessableEntity)
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
