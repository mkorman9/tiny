package httpauth

import (
	"github.com/gofiber/fiber/v2"
)

// MiddlewareConfig holds a configuration for the Middleware.
type MiddlewareConfig struct {
	// OnError sets an error handler for the middleware.
	OnError func(c *fiber.Ctx, err error) error

	// OnUnverified sets an unverified handler for the middleware.
	OnUnverified func(c *fiber.Ctx, result *VerificationResult) error

	// OnAccessDenied sets an access denied handler for the middleware.
	OnAccessDenied func(c *fiber.Ctx, result *VerificationResult) error
}
