package tinygrpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
)

const authorizationHeader = "authorization"

// NewTokenCredentials creates new credentials.PerRPCCredentials instance, appending given token to gRPC call.
func NewTokenCredentials(token string) credentials.PerRPCCredentials {
	return &tokenCredentials{token: token}
}

type tokenCredentials struct {
	token string
}

func (c *tokenCredentials) RequireTransportSecurity() bool {
	return false
}

func (c *tokenCredentials) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	headers := map[string]string{
		authorizationHeader: fmt.Sprintf("Bearer %s", c.token),
	}
	return headers, nil
}
