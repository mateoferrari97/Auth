package server

import (
	"errors"
	"github.com/mateoferrari97/auth/internal"
	"net/http"
)

type handlerError struct {
	StatusCode int
	Error      string
}

func handleError(err error) handlerError {
	var e *internal.Error
	if !errors.As(err, &e) {
		e = internal.NewError(err.Error(), http.StatusInternalServerError)
	}

	return handlerError{
		StatusCode: e.StatusCode,
		Error:      e.Error(),
	}
}
