package httpauth

import (
	"github.com/gin-gonic/gin"
)

// VerifyCookieFunc is a user-provided function that is called in able to validate given cookie value.
type VerifyCookieFunc = func(c *gin.Context, cookie string) (*VerificationResult, error)

// NewSessionCookieMiddleware creates new cookie-based Middleware.
// This middleware reads a cookie specified by cookieName argument and calls verifyCookie with its value.
func NewSessionCookieMiddleware(cookieName string, verifyCookie VerifyCookieFunc, opts ...MiddlewareOpt) *Middleware {
	config := MiddlewareConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return newMiddleware(
		func(c *gin.Context) (*VerificationResult, error) {
			cookie, err := c.Cookie(cookieName)
			if err != nil {
				cookie = ""
			}

			return verifyCookie(c, cookie)
		},
		&config,
	)
}
