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

var (
	// ErrRedirect is return when client reaches its maximum number of redirects when performing HTTP request.
	ErrRedirect = errors.New("redirect limit exceeded")
)

// Client is an HTTP client, capable of executing HTTP requests and performing retries.
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates an instance of Client using given options.
func NewClient(config ...*Config) *Client {
	var providedConfig *Config
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeConfig(providedConfig)

	httpClient := &http.Client{
		Timeout: c.Timeout,
		Jar:     c.CookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= c.MaxRedirects {
				return ErrRedirect
			} else {
				return nil
			}
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if c.Address != "" {
					addr = c.Address
				}

				d := net.Dialer{}
				return d.DialContext(ctx, c.Network, addr)
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if c.Address != "" {
					addr = c.Address
				}

				d := tls.Dialer{
					Config: c.TLSConfig,
				}
				return d.DialContext(ctx, c.Network, addr)
			},
			TLSClientConfig: c.TLSConfig,
		},
	}

	return &Client{
		config:     c,
		httpClient: httpClient,
	}
}

// Send tries to send given HTTP request and return a response.
// Depending on the configuration specified, requests might be retried on error.
// If client reaches its maximum number of redirects - both the latest response and ErrRedirect are returned.
func (client *Client) Send(request *http.Request) (*http.Response, error) {
	for retry := 0; retry <= client.config.MaxRetries; retry++ {
		response, err := client.httpClient.Do(request)

		shouldRetry := false

		if err != nil {
			urlError, isUrlError := err.(*url.Error)
			if !isUrlError {
				if errors.Is(err, ErrRedirect) {
					return response, ErrRedirect
				}

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
