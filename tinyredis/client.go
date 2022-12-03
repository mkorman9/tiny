package tinyredis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
	"time"
)

// Client is a wrapper for *redis.Client providing a handy Close() function.
type Client struct {
	*redis.Client
}

// DialRedis creates a connection to Redis and returns Client instance.
func DialRedis(address string, opts ...Opt) (*Client, error) {
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

	log.Debug().Msg("Establishing Redis connection")

	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectionTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	} else {
		log.Info().Msg("Successfully connected to Redis")
	}

	return &Client{Client: client}, nil
}

// Close closes a connection to Redis.
func (c *Client) Close() {
	log.Debug().Msg("Closing Redis connection")

	if err := c.Client.Close(); err != nil {
		log.Error().Err(err).Msg("Error when closing Redis connection")
	} else {
		log.Info().Msg("Redis connection closed successfully")
	}
}
