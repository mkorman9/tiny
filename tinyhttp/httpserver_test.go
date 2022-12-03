package tinyhttp

import (
	"github.com/gin-gonic/gin"
	"github.com/mkorman9/tiny"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	tiny.SetupLogger()
}

func TestHTTPServer(t *testing.T) {
	// given
	responsePayload := "payload"

	engine := NewServer("address").Engine
	engine.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, responsePayload)
	})

	// when
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(recorder, req)

	// then
	assert.Equal(t, http.StatusOK, recorder.Code, "response code should be 200")
	assert.Equal(t, []byte(responsePayload), recorder.Body.Bytes(), "response payload should match")
}
