package internal

import "fmt"

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
