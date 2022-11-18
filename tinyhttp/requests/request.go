package requests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// RequestConfig holds a configuration for request while it's constructed.
type RequestConfig struct {
	method  string
	url     string
	body    io.Reader
	headers map[string]string
	host    string
	cookies []*http.Cookie
}

// RequestOpt is an option to be specified to NewRequest.
type RequestOpt = func(*RequestConfig) error

// RequestPart represents a single part of the HTTP multipart form.
type RequestPart struct {
	fieldName string
	fileName  string
	data      any
	diskPath  string
}

// NewRequest constructs a request using given options.
func NewRequest(opts ...RequestOpt) (*http.Request, error) {
	config := &RequestConfig{
		method:  "GET",
		headers: map[string]string{},
	}

	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	request, err := http.NewRequest(config.method, config.url, config.body)
	if err != nil {
		return nil, err
	}

	for header, value := range config.headers {
		request.Header.Set(header, value)
	}

	if config.host != "" {
		request.Host = config.host
	}

	for _, cookie := range config.cookies {
		request.AddCookie(cookie)
	}

	return request, nil
}

var (
	GET     = Method("GET")
	POST    = Method("POST")
	PUT     = Method("PUT")
	DELETE  = Method("DELETE")
	HEAD    = Method("HEAD")
	TRACE   = Method("TRACE")
	OPTIONS = Method("OPTIONS")
	CONNECT = Method("CONNECT")
)

// Method is a method of the HTTP request (default: "GET").
func Method(method string) RequestOpt {
	return func(config *RequestConfig) error {
		config.method = method
		return nil
	}
}

// URL is a target URL of the HTTP request.
func URL(url string) RequestOpt {
	return func(config *RequestConfig) error {
		config.url = url
		return nil
	}
}

// Body is an optional body to be included in the request.
func Body(body io.Reader) RequestOpt {
	return func(config *RequestConfig) error {
		config.body = body
		return nil
	}
}

// JSONBody is an optional body to be included in the request.
// Given value is first converted to JSON and then appended to the request.
func JSONBody(body interface{}) RequestOpt {
	return func(config *RequestConfig) error {
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)

		err := encoder.Encode(body)
		if err != nil {
			return err
		}

		config.body = buffer
		config.headers["Content-Type"] = "application/json"
		return nil
	}
}

// FormBody is an optional form data to be included in the request.
func FormBody(form *url.Values) RequestOpt {
	return func(config *RequestConfig) error {
		config.body = strings.NewReader(form.Encode())
		config.headers["Content-Type"] = "application/x-www-form-urlencoded"
		return nil
	}
}

// MultipartForm is an optional multipart form data to be included in the request.
func MultipartForm(parts ...*RequestPart) RequestOpt {
	return func(config *RequestConfig) error {
		var filesToClose []*os.File

		var buffer bytes.Buffer
		w := multipart.NewWriter(&buffer)

		for _, part := range parts {
			var data io.Reader

			switch {
			case part.data != nil:
				if reader, ok := part.data.(io.Reader); ok {
					data = reader
				} else if b, ok := part.data.([]byte); ok {
					data = bytes.NewReader(b)
				} else if s, ok := part.data.(string); ok {
					data = strings.NewReader(s)
				} else {
					return errors.New("invalid type of data field in multipart form")
				}
			case part.diskPath != "":
				file, err := os.Open(part.diskPath)
				if err != nil {
					return err
				}

				data = file
				filesToClose = append(filesToClose, file)
			default:
				return errors.New("no data/diskPath specified for mutlipart form")
			}

			fileWriter, err := w.CreateFormFile(part.fieldName, part.fileName)
			if err != nil {
				return err
			}

			_, err = io.Copy(fileWriter, data)
			if err != nil {
				return err
			}
		}

		if err := w.Close(); err != nil {
			return err
		}

		for _, file := range filesToClose {
			_ = file.Close()
		}

		config.body = &buffer
		config.headers["Content-Type"] = w.FormDataContentType()
		return nil
	}
}

// Header sets a request header specified by the given key.
func Header(key, value string) RequestOpt {
	return func(config *RequestConfig) error {
		config.headers[key] = value
		return nil
	}
}

// BearerToken sets Authorization request header to "Bearer %token%".
func BearerToken(token string) RequestOpt {
	return Header("Authorization", fmt.Sprintf("Bearer %s", token))
}

// BasicAuth sets Authorization request header to "Basic base64(%username%:%password%)".
func BasicAuth(username, password string) RequestOpt {
	authString := fmt.Sprintf("%s:%s", username, password)
	base64AuthString := base64.StdEncoding.EncodeToString([]byte(authString))
	return Header("Authorization", fmt.Sprintf("Basic %s", base64AuthString))
}

// UserAgent sets User-Agent request header.
func UserAgent(userAgent string) RequestOpt {
	return Header("User-Agent", userAgent)
}

// ContentType sets Content-Type request header.
func ContentType(contentType string) RequestOpt {
	return Header("Content-Type", contentType)
}

// Host overwrites the Host header value. If not provided, the host is extracted from the URL.
func Host(host string) RequestOpt {
	return func(config *RequestConfig) error {
		config.host = host
		return nil
	}
}

// Cookie adds HTTP request cookie.
func Cookie(cookie *http.Cookie) RequestOpt {
	return func(config *RequestConfig) error {
		config.cookies = append(config.cookies, cookie)
		return nil
	}
}

// PartFromData creates a part of multipart form from the in-memory buffer.
func PartFromData(fieldName, fileName string, data any) *RequestPart {
	return &RequestPart{
		fieldName: fieldName,
		fileName:  fileName,
		data:      data,
	}
}

// PartFromDiskFile creates a part of multipart form from the disk file.
func PartFromDiskFile(fieldName, fileName, diskPath string) *RequestPart {
	return &RequestPart{
		fieldName: fieldName,
		fileName:  fileName,
		diskPath:  diskPath,
	}
}
