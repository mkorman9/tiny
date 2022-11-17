package httpauth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// VerifyTokenFunc is a user-provided function that is called in able to validate given API token.
type VerifyTokenFunc = func(c *gin.Context, token string) (*VerificationResult, error)

// NewBearerTokenMiddleware creates new bearer-token based Middleware.
// This middleware reads Authorization header and expects it to begin with "Bearer" string.
func NewBearerTokenMiddleware(verifyToken VerifyTokenFunc, opts ...MiddlewareOpt) Middleware {
	handlers := middlewareHandlers{}
	for _, opt := range opts {
		opt(&handlers)
	}

	return newMiddleware(
		func(rolesCheckingFunc rolesCheckingFunc) gin.HandlerFunc {
			return func(c *gin.Context) {
				token := extractToken(c)

				verificationResult, err := verifyToken(c, token)
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
