package requests

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

// Client is an HTTP client, capable of executing HTTP requests and performing retries.
type Client struct {
	config     *ClientConfig
	httpClient *http.Client
}

// NewClient creates an instance of Client using given options.
func NewClient(opts ...ClientOpt) *Client {
	config := &ClientConfig{
		Network:          "tcp",
		Timeout:          10 * time.Second,
		MaxRetries:       0,
		RetryDelayFactor: 0,
		TLSConfig:        &tls.Config{},
	}

	for _, opt := range opts {
		opt(config)
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if config.Address != "" {
					addr = config.Address
				}

				d := net.Dialer{}
				return d.DialContext(ctx, config.Network, addr)
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if config.Address != "" {
					addr = config.Address
				}

				d := tls.Dialer{
					Config: config.TLSConfig,
				}
				return d.DialContext(ctx, config.Network, addr)
			},
			TLSClientConfig: config.TLSConfig,
		},
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
	}
}

// Send tries to send given HTTP request and return a response.
// Depending on the configuration specified, requests might be retried on error.
func (client *Client) Send(request *http.Request) (*http.Response, error) {
	for retry := 0; retry <= client.config.MaxRetries; retry++ {
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
			log.Debug().Err(err).Msgf(
				"Request to '%v' has failed. Retry %v/%v",
				request.URL.Host,
				retry+1,
				client.config.MaxRetries+1,
			)

			if retry >= client.config.MaxRetries {
				return response, err
			}

			if client.config.RetryDelayFactor != 0 {
				time.Sleep(time.Duration(retry+1) * client.config.RetryDelayFactor)
			}
		}
	}

	return nil, errors.New("invalid state")
}
