package httpauth

import "github.com/gin-gonic/gin"

// MiddlewareConfig holds a configuration for the Middleware.
type MiddlewareConfig struct {
	errorHandler        func(c *gin.Context, err error)
	unverifiedHandler   func(c *gin.Context)
	accessDeniedHandler func(c *gin.Context)
}

// MiddlewareOpt is an option to be specified when creating new middleware.
type MiddlewareOpt func(*MiddlewareConfig)

// ErrorHandler sets an error handler for the middleware.
func ErrorHandler(handler func(c *gin.Context, err error)) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.errorHandler = handler
	}
}

// UnverifiedHandler sets an unverified handler for the middleware.
func UnverifiedHandler(handler func(c *gin.Context)) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.unverifiedHandler = handler
	}
}

// AccessDeniedHandler sets an access denied handler for the middleware.
func AccessDeniedHandler(handler func(c *gin.Context)) MiddlewareOpt {
	return func(m *MiddlewareConfig) {
		m.accessDeniedHandler = handler
	}
}
