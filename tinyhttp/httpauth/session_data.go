package httpauth

import "github.com/gin-gonic/gin"

const sessionDataContextKey = "httpauth/sessionData"

// GetSessionData tries to extract session data set by middleware from the request's context.
func GetSessionData(c *gin.Context) (any, bool) {
	return c.Get(sessionDataContextKey)
}

// MustGetSessionData extracts session data set by middleware from the request's context, or panics.
func MustGetSessionData(c *gin.Context) any {
	if v, ok := GetSessionData(c); ok {
		return v
	}

	panic("MustGetSessionData() expected session data to be present")
}

func setSessionData(c *gin.Context, sessionData any) {
	c.Set(sessionDataContextKey, sessionData)
}
