package requests

import "time"

type clientConfig struct {
	timeout          time.Duration
	maxRetries       int
	retryDelayFactor time.Duration
}

type ClientOpt = func(*clientConfig)

func Timeout(timeout time.Duration) ClientOpt {
	return func(config *clientConfig) {
		config.timeout = timeout
	}
}

func MaxRetries(maxRetries int) ClientOpt {
	return func(config *clientConfig) {
		if maxRetries < 0 {
			maxRetries = 0
		}
		config.maxRetries = maxRetries
	}
}

func RetryDelayFactor(retryDelayFactor time.Duration) ClientOpt {
	return func(config *clientConfig) {
		config.retryDelayFactor = retryDelayFactor
	}
}
