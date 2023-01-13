package httpauth

import (
	"github.com/gofiber/fiber/v2"
)

// MiddlewareConfig holds a configuration for the Middleware.
type MiddlewareConfig struct {
	errorHandler        func(c *fiber.Ctx, err error) error
	unverifiedHandler   func(c *fiber.Ctx) error
	accessDeniedHandler func(c *fiber.Ctx) error
}

// MiddlewareOpt is an option to be specified when creating new middleware.
type MiddlewareOpt func(*MiddlewareConfig)

// ErrorHandler sets an error handler for the middleware.
func ErrorHandler(handler func(c *fiber.Ctx, err error) error) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.errorHandler = handler
	}
}

// UnverifiedHandler sets an unverified handler for the middleware.
func UnverifiedHandler(handler func(c *fiber.Ctx) error) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.unverifiedHandler = handler
	}
}

// AccessDeniedHandler sets an access denied handler for the middleware.
func AccessDeniedHandler(handler func(c *fiber.Ctx) error) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.accessDeniedHandler = handler
	}
}
