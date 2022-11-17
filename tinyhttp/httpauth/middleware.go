package httpauth

import "github.com/gin-gonic/gin"

// VerificationResult is a structure returned by the user-provided token/cookie verification functions.
type VerificationResult struct {
	// Verified tells the middleware whether given request has been authorized or not.
	// Returning false value for this field will immediately finish request with 401.
	Verified bool

	// Roles is a set of permissions associated with the entity identified by the credentials.
	// Roles are expected to be returned only if Verified is equal to true.
	Roles []string

	// SessionData is an optional field that associates some arbitrary value with the current session.
	// Request handlers are able to extract this value later by calling GetSessionData() / MustGetSessionData().
	SessionData any
}

type rolesCheckingFunc = func(roles []string) bool
type middlewareHandler = func(rolesCheckingFunc rolesCheckingFunc) gin.HandlerFunc

// Middleware is an interface that represents a generic authorization middleware.
// It provides user-friendly API that can be easily integrated with existing Gin request handlers.
// Underlying implementation might utilize Basic Auth, Bearer-Token or other mechanisms but this API is transparent.
type Middleware interface {
	// Authenticated enables access to all authenticated clients, no matter the roles.
	Authenticated() gin.HandlerFunc

	// AnyOfRoles enables access to only those clients who have at least one of the given roles associated with them.
	AnyOfRoles(allowedRoles ...string) gin.HandlerFunc

	// AllOfRoles enables access to only those clients who have all specified roles associated with them.
	AllOfRoles(requiredRoles ...string) gin.HandlerFunc
}

type middleware struct {
	handler middlewareHandler
}

func newMiddleware(handler middlewareHandler) *middleware {
	return &middleware{handler}
}

func (m *middleware) Authenticated() gin.HandlerFunc {
	return m.handler(func(_ []string) bool {
		return true
	})
}

func (m *middleware) AnyOfRoles(allowedRoles ...string) gin.HandlerFunc {
	allowedRolesSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		allowedRolesSet[role] = struct{}{}
	}

	return m.handler(func(providedRoles []string) bool {
		for _, role := range providedRoles {
			if _, ok := allowedRolesSet[role]; ok {
				return true
			}
		}

		return false
	})
}

func (m *middleware) AllOfRoles(requiredRoles ...string) gin.HandlerFunc {
	return m.handler(func(providedRoles []string) bool {
		for _, role := range requiredRoles {
			hasRole := false
			for _, providedRole := range providedRoles {
				if role == providedRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return false
			}
		}

		return true
	})
}
