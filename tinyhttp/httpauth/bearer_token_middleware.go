package httpauth

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// VerifyTokenFunc is a user-provided function that is called in able to validate given API token.
type VerifyTokenFunc = func(c *gin.Context, token string) (*VerificationResult, error)

// NewBearerTokenMiddleware creates new bearer-token based Middleware.
// This middleware reads Authorization header and expects it to begin with "Bearer" string.
func NewBearerTokenMiddleware(verifyToken VerifyTokenFunc, opts ...MiddlewareOpt) *Middleware {
	config := MiddlewareConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return newMiddleware(
		func(c *gin.Context) (*VerificationResult, error) {
			token := extractToken(c)
			return verifyToken(c, token)
		},
		&config,
	)
}

func extractToken(c *gin.Context) string {
	authorizationHeader := c.GetHeader("Authorization")
	if len(authorizationHeader) == 0 {
		return ""
	}

	fields := strings.Fields(authorizationHeader)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return ""
	}

	token := fields[1]
	if len(token) == 0 {
		return ""
	}

	return token
}
