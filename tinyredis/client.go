package tinyredis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

// Dial creates a connection to Redis and returns *redis.Client instance.
func Dial(address string, opts ...Opt) (*redis.Client, error) {
	config := Config{
		ConnectionTimeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if address == "" {
		return nil, errors.New("address cannot be empty")
	}

	client := redis.NewClient(&redis.Options{
		Addr:      address,
		Username:  config.Username,
		Password:  config.Password,
		DB:        config.DB,
		TLSConfig: config.TLSConfig,
	})

	if !config.NoPing {
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}
	}

	return client, nil
}
