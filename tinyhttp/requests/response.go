package requests

import (
	"encoding/json"
	"io"
	"net/http"
)

func ReadResponseBody(response *http.Response) ([]byte, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	return body, nil
}

func BindResponseJSON(response *http.Response, v interface{}) error {
	body, err := ReadResponseBody(response)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}
