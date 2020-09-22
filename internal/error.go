package internal

import (
	"errors"
	"fmt"
)

var (
	ErrBadRequest            = errors.New("bad request")
	ErrUnprocessableEntity   = errors.New("unprocessable entity")
	ErrWeakPassword          = errors.New("weak password")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrInvalidToken          = errors.New("can't access to the resource. invalid token")
	ErrAlteredTokenClaims    = errors.New("can't access to the resource. claims don't match from original token")
	ErrResourceNotFound      = errors.New("resource not found")
)

type Error struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func NewError(message string, statusCode int) *Error {
	return &Error{
		StatusCode: statusCode,
		Message:    message,
	}
}

func (e *Error) Error() string {
	if e == nil || e.StatusCode == 0 || e.Message == "" {
		return "unexpected error"
	}

	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}
