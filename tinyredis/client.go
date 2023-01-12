package tinyredis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

// Dial creates a connection to Redis and returns *redis.Client instance.
func Dial(url string, opts ...Opt) (*redis.Client, error) {
	config := Config{
		ConnectionTimeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(&config)
	}

	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if config.redisOpts != nil {
		config.redisOpts(options)
	}

	client := redis.NewClient(options)

	if !config.NoPing {
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}
	}

	return client, nil
}
