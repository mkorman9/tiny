package httpauth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// VerifyCookieFunc is a user-provided function that is called in able to validate given cookie value.
type VerifyCookieFunc = func(c *fiber.Ctx, cookie string) (*VerificationResult, error)

// NewSessionCookieMiddleware creates new cookie-based Middleware.
// This middleware reads a cookie specified by cookieName argument and calls verifyCookie with its value.
func NewSessionCookieMiddleware(cookieName string, verifyCookie VerifyCookieFunc, opts ...MiddlewareOpt) *Middleware {
	config := MiddlewareConfig{}
	for _, opt := range opts {
		opt(&config)
	}

	return newMiddleware(
		func(c *fiber.Ctx) (*VerificationResult, error) {
			cookie := extractCookie(c, cookieName)
			return verifyCookie(c, cookie)
		},
		&config,
	)
}

func extractCookie(c *fiber.Ctx, cookieName string) string {
	return utils.CopyString(c.Cookies(cookieName))
}
