package internal

import (
	"encoding/json"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, v interface{}, statusCode int) error {
	if statusCode == http.StatusNoContent || v == nil {
		w.WriteHeader(statusCode)
		return nil
	}

	var jsonData []byte

	var err error
	switch v := v.(type) {
	case []byte:
		jsonData = v
	default:
		jsonData, err = json.Marshal(v)
	}

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
