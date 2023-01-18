package httpauth

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

// VerifyTokenFunc is a user-provided function that is called in able to validate given API token.
type VerifyTokenFunc = func(c *fiber.Ctx, token string) (*VerificationResult, error)

// NewBearerTokenMiddleware creates new bearer-token based Middleware.
// This middleware reads Authorization header and expects it to begin with "Bearer" string.
func NewBearerTokenMiddleware(verifyToken VerifyTokenFunc, config ...*MiddlewareConfig) *Middleware {
	c := &MiddlewareConfig{}
	if config != nil {
		c = config[0]
	}

	return newMiddleware(
		func(c *fiber.Ctx) (*VerificationResult, error) {
			token := extractToken(c)
			return verifyToken(c, token)
		},
		c,
	)
}

func extractToken(c *fiber.Ctx) string {
	authorizationHeader := c.Get("Authorization")
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
