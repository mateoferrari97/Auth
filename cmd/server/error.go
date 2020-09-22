package server

import (
	"errors"
	"net/http"

	"github.com/mateoferrari97/auth/internal"
)

func handleError(w http.ResponseWriter, err error) {
	message := err.Error()

	var e *internal.Error
	switch errors.Unwrap(err) {
	case internal.ErrBadRequest:
		e = internal.NewError(message, http.StatusBadRequest)
	case internal.ErrWeakPassword:
		e = internal.NewError(message, http.StatusBadRequest)
	case internal.ErrUnprocessableEntity:
		e = internal.NewError(message, http.StatusUnprocessableEntity)
	case internal.ErrResourceNotFound:
		e = internal.NewError(message, http.StatusNotFound)
	case internal.ErrInvalidToken:
		e = internal.NewError(message, http.StatusForbidden)
	case internal.ErrAlteredTokenClaims:
		e = internal.NewError(message, http.StatusForbidden)
	case internal.ErrResourceAlreadyExists:
		e = internal.NewError(message, http.StatusConflict)
	default:
		e = internal.NewError(message, http.StatusInternalServerError)
	}

	_ = internal.RespondJSON(w, e, e.StatusCode)
}
