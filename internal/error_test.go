package internal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError(t *testing.T) {
	// Given
	e := NewError("internal server error", http.StatusInternalServerError)

	// When
	m := e.Error()

	// Then
	require.Equal(t, "500: internal server error", m)
}

func TestError_EmptyValues(t *testing.T) {
	tt := []struct {
		name string
		err  error
	}{
		{
			name: "0 status code",
			err:  NewError("0 status code", 0),
		},
		{
			name: "empty message",
			err:  NewError("", http.StatusInternalServerError),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			e := tc.err

			// When
			m := e.Error()

			// Then
			require.Equal(t, "unexpected error", m)
		})
	}
}
