package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

var _v = validator.New()

type Wrapper interface {
	Wrap(method string, pattern string, handler http.HandlerFunc)
}

type Handler struct {
	Wrapper
}

func NewHandler(wrapper Wrapper) *Handler {
	return &Handler{wrapper}
}

func (h *Handler) Ping() {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	}

	h.Wrap(http.MethodGet, "/ping", wrapH)
}

type AuthorizeHandler func(token string) (User, error)

func (h *Handler) RouteHome(handler AuthorizeHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("authorization")
		if err != nil {
			http.Error(w, fmt.Sprintf("unauthorized: %v", err), http.StatusForbidden)
			return
		}

		user, err := handler(c.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		fmt.Fprint(w, fmt.Sprintf("Welcome, %s.\n Email: %s\n ID: %s\n Password: %s", user.Name, user.Email, user.ID, user.Password))
	}

	h.Wrap(http.MethodGet, "/", wrapH)
}

type RegisterRequest struct {
	Name string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterHandler func(req RegisterRequest) (string, error)

func (h *Handler) RouteRegister(handler RegisterHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("can't process body from client: %v", err), http.StatusUnprocessableEntity)
			return
		}

		if err := validateRegisterInformation(req); err != nil {
			http.Error(w, fmt.Sprintf("validating user information for his creation: %v", err), http.StatusBadRequest)
			return
		}

		token, err := handler(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("creating user: %v", err), http.StatusInternalServerError)
			return
		}

		c := &http.Cookie{
			Name:       "authorization",
			Value:      token,
			Path: "/",
		}

		http.SetCookie(w, c)
	}

	h.Wrap(http.MethodPost, "/register", wrapH)
}

func validateRegisterInformation(registerInformation RegisterRequest) error {
	if err := _v.Struct(registerInformation); err != nil {
		return err
	}

	return validatePassword(registerInformation.Password)
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
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, handler(), http.StatusTemporaryRedirect)
	}

	h.Wrap(http.MethodGet, "/login/google", wrapH)
}

type LoginWithGoogleCallbackHandler func(code string) (string, error)

func (h *Handler) RouteLoginWithGoogleCallback(handler LoginWithGoogleCallbackHandler) {
	wrapH := func(w http.ResponseWriter, r *http.Request) {
		code := r.FormValue("code")
		if code == "" {
			http.Error(w, "code is required", http.StatusBadRequest)
			return
		}

		token, err := handler(code)
		if err != nil {
			http.Error(w, fmt.Sprintf("logging user with google: %v", err), http.StatusInternalServerError)
			return
		}

		c := &http.Cookie{
			Name:       "authorization",
			Value:      token,
			Path: "/",
		}

		http.SetCookie(w, c)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}

	h.Wrap(http.MethodGet, "/login/google/callback", wrapH)
}