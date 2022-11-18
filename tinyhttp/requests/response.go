package requests

import (
	"encoding/json"
	"io"
	"net/http"
)

// ReadResponseBody extracts the whole request body from the HTTP response.
func ReadResponseBody(response *http.Response) ([]byte, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	_ = response.Body.Close()

	return body, nil
}

// ReadResponseJSON extracts the whole request body from the HTTP response and converts it from JSON to the given value.
func ReadResponseJSON(response *http.Response, v any) error {
	body, err := ReadResponseBody(response)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
