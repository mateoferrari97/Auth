package server

import (
	"errors"
	"github.com/mateoferrari97/auth/internal"
	"net/http"
)

type HandlerError struct {
	StatusCode int
	Error string
}

func HandleError(err error) HandlerError {
	var e *internal.Error
	if !errors.As(err, &e) {
		e = internal.NewError(err.Error(), http.StatusInternalServerError)
	}

	return HandlerError{
		StatusCode: e.StatusCode,
		Error:      e.Error(),
	}
}
