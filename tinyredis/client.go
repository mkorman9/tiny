package tinyredis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// Dial creates a connection to Redis and returns *redis.Client instance.
func Dial(url string, config ...*Config) (*redis.Client, error) {
	var providedConfig *Config
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeConfig(providedConfig)

	options, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	if c.RedisOpt != nil {
		c.RedisOpt(options)
	}

	client := redis.NewClient(options)

	if !c.NoPing {
		ctx, cancel := context.WithTimeout(context.Background(), c.ConnectionTimeout)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			return nil, err
		}
	}

	return client, nil
}
