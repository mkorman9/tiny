package httpauth

import (
	"github.com/gin-gonic/gin"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinyhttp"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	tiny.Init()
}

func TestMissingToken(t *testing.T) {
	// given
	responsePayload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	engine := tinyhttp.NewServer("address").Engine
	engine.GET(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *gin.Context) {
			c.String(http.StatusOK, responsePayload)
		},
	)

	// when
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secured", nil)
	engine.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "response code should be 401")
}

func TestInvalidToken(t *testing.T) {
	// given
	responsePayload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	engine := tinyhttp.NewServer("address").Engine
	engine.GET(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *gin.Context) {
			c.String(http.StatusOK, responsePayload)
		},
	)

	// when
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer incorrectToken")
	engine.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "response code should be 401")
}

func TestValidToken(t *testing.T) {
	// given
	responsePayload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	engine := tinyhttp.NewServer("address").Engine
	engine.GET(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *gin.Context) {
			c.String(http.StatusOK, responsePayload)
		},
	)

	// when
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer "+correctToken)
	engine.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code, "response code should be 200")
	assert.Equal(t, []byte(responsePayload), recorder.Body.Bytes(), "response payload should match")
}

func TestInvalidRoles(t *testing.T) {
	// given
	responsePayload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	engine := tinyhttp.NewServer("address").Engine
	engine.GET(
		"/secured",
		middleware.AnyOfRoles("SUPERUSER"),
		func(c *gin.Context) {
			c.String(http.StatusOK, responsePayload)
		},
	)

	// when
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer "+correctToken)
	engine.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusForbidden, recorder.Code, "response code should be 403")
}

func createBearerTokenMiddleware(correctToken string) *Middleware {
	return NewBearerTokenMiddleware(func(c *gin.Context, token string) (*VerificationResult, error) {
		if token == correctToken {
			return &VerificationResult{Verified: true, Roles: []string{"ADMIN"}}, nil
		} else {
			return &VerificationResult{}, nil
		}
	})
}
