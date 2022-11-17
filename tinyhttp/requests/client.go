package requests

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

type Client struct {
	config     *clientConfig
	httpClient *http.Client
}

func NewClient(opts ...ClientOpt) *Client {
	config := &clientConfig{
		timeout:          10 * time.Second,
		maxRetries:       0,
		retryDelayFactor: time.Second,
	}

	for _, opt := range opts {
		opt(config)
	}

	httpClient := &http.Client{
		Timeout: config.timeout,
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
	}
}

func (client *Client) Send(request *http.Request) (*http.Response, error) {
	for retry := 0; retry <= client.config.maxRetries; retry++ {
		response, err := client.httpClient.Do(request)

		shouldRetry := false

		if err != nil {
			urlError, isUrlError := err.(*url.Error)
			if !isUrlError {
				return nil, err
			}

			if _, isNetError := urlError.Err.(*net.OpError); isNetError {
				shouldRetry = true
			}
		} else {
			if response.StatusCode >= http.StatusInternalServerError { // 500, retry only for server-side errors
				shouldRetry = true
				err = fmt.Errorf("status %v", response.StatusCode)
			}
		}

		if !shouldRetry {
			return response, err
		} else {
			log.Debug().Err(err).Msgf("Request to '%v' has failed. Retry %v/%v", request.URL.Host, retry+1, client.config.maxRetries+1)
			if retry == client.config.maxRetries {
				return response, err
			}

			if client.config.retryDelayFactor != 0 {
				time.Sleep(time.Duration(retry+1) * client.config.retryDelayFactor)
			}
		}
	}

	return nil, errors.New("invalid state")
}
