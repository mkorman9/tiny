package httpauth

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// VerifyCookieFunc is a user-provided function that is called in able to validate given cookie value.
type VerifyCookieFunc = func(c *gin.Context, cookie string) (*VerificationResult, error)

// NewSessionCookieMiddleware creates new cookie-based Middleware.
// This middleware reads a cookie specified by cookieName argument and calls verifyCookie with its value.
func NewSessionCookieMiddleware(cookieName string, verifyCookie VerifyCookieFunc, opts ...MiddlewareOpt) Middleware {
	handlers := middlewareHandlers{}
	for _, opt := range opts {
		opt(&handlers)
	}

	return newMiddleware(
		func(rolesCheckingFunc rolesCheckingFunc) gin.HandlerFunc {
			return func(c *gin.Context) {
				cookie, err := c.Cookie(cookieName)
				if err != nil {
					cookie = ""
				}

				verificationResult, err := verifyCookie(c, cookie)
				if err != nil {
					if handlers.errorHandler != nil {
						handlers.errorHandler(c, err)
						return
					}

					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				if !verificationResult.Verified {
					if handlers.unverifiedHandler != nil {
						handlers.unverifiedHandler(c)
						return
					}

					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}

				rolesCheckingResult := rolesCheckingFunc(verificationResult.Roles)

				if !rolesCheckingResult {
					if handlers.accessDeniedHandler != nil {
						handlers.accessDeniedHandler(c)
						return
					}

					c.AbortWithStatus(http.StatusForbidden)
					return
				}

				setSessionData(c, verificationResult.SessionData)
				c.Next()
			}
		},
	)
}
