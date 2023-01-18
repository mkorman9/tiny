package httpauth

import (
	"github.com/gofiber/fiber/v2"
)

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

type middlewareAction = func(*fiber.Ctx) (*VerificationResult, error)
type rolesCheckingFunc = func(roles []string) bool

// Middleware is an interface that represents a generic authorization middleware.
// It provides user-friendly API that can be easily integrated with existing fiber request handlers.
// Underlying implementation might utilize Basic Auth, Bearer-Token or other mechanisms but this API is transparent.
type Middleware struct {
	action middlewareAction
	config *MiddlewareConfig
}

func newMiddleware(action middlewareAction, config *MiddlewareConfig) *Middleware {
	return &Middleware{
		action: action,
		config: config,
	}
}

// Authenticated enables access to all authenticated clients, no matter the roles.
func (m *Middleware) Authenticated() fiber.Handler {
	checkRoles := func(_ []string) bool {
		return true
	}

	return m.authorize(m.action, m.config, checkRoles)
}

// AnyOfRoles enables access to only those clients who have at least one of the given roles associated with them.
func (m *Middleware) AnyOfRoles(allowedRoles ...string) fiber.Handler {
	allowedRolesSet := make(map[string]struct{})
	for _, role := range allowedRoles {
		allowedRolesSet[role] = struct{}{}
	}

	checkRoles := func(providedRoles []string) bool {
		for _, role := range providedRoles {
			if _, ok := allowedRolesSet[role]; ok {
				return true
			}
		}

		return false
	}

	return m.authorize(m.action, m.config, checkRoles)
}

// AllOfRoles enables access to only those clients who have all specified roles associated with them.
func (m *Middleware) AllOfRoles(requiredRoles ...string) fiber.Handler {
	checkRoles := func(providedRoles []string) bool {
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
	}

	return m.authorize(m.action, m.config, checkRoles)
}

func (m *Middleware) authorize(
	action middlewareAction,
	config *MiddlewareConfig,
	checkRoles rolesCheckingFunc,
) fiber.Handler {
	return func(c *fiber.Ctx) error {
		verificationResult, err := action(c)
		if err != nil {
			if config.OnError != nil {
				return config.OnError(c, err)
			}

			c.Status(fiber.StatusInternalServerError)
			return nil
		}

		if !verificationResult.Verified {
			if config.OnUnverified != nil {
				return config.OnUnverified(c, verificationResult)
			}

			c.Status(fiber.StatusUnauthorized)
			return nil
		}

		rolesCheckingResult := checkRoles(verificationResult.Roles)
		if !rolesCheckingResult {
			if config.OnAccessDenied != nil {
				return config.OnAccessDenied(c, verificationResult)
			}

			c.Status(fiber.StatusForbidden)
			return nil
		}

		setSessionData(c, verificationResult.SessionData)
		return c.Next()
	}
}
