package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRespondJSON(t *testing.T) {
	tt := []struct {
		name       string
		v          interface{}
		statusCode int
	}{
		{
			name: "0 status code",
			v: struct {
				name string
			}{
				name: "luken",
			},
			statusCode: http.StatusOK,
		},
		{
			name:       "empty message",
			v:          []byte(`fail to improve`),
			statusCode: http.StatusOK,
		},
		{
			name:       "nil value",
			v:          nil,
			statusCode: http.StatusOK,
		},
		{
			name:       "status no content",
			v:          nil,
			statusCode: http.StatusNoContent,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			w := httptest.NewRecorder()

			// When
			err := RespondJSON(w, tc.v, tc.statusCode)

			// Then
			require.NoError(t, err)
			require.Equal(t, tc.statusCode, w.Code)
		})
	}
}

func TestRespondJSON_MarshalError(t *testing.T) {
	// Given
	v := make(chan int)
	w := httptest.NewRecorder()

	// When
	err := RespondJSON(w, v, http.StatusOK)

	// Then
	require.EqualError(t, err, "json: unsupported type: chan int")
}
