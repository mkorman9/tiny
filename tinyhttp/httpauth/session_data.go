package httpauth

import (
	"github.com/gofiber/fiber/v2"
)

const sessionDataContextKey = "httpauth/sessionData"

// GetSessionData tries to extract session data set by middleware from the request's context.
func GetSessionData(c *fiber.Ctx) any {
	return c.Context().UserValue(sessionDataContextKey)
}

// MustGetSessionData extracts session data set by middleware from the request's context, or panics.
func MustGetSessionData(c *fiber.Ctx) any {
	if v := GetSessionData(c); v != nil {
		return v
	}

	panic("MustGetSessionData() expected session data to be present")
}

func setSessionData(c *fiber.Ctx, sessionData any) {
	c.Context().SetUserValue(sessionDataContextKey, sessionData)
}
