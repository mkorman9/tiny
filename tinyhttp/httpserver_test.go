package tinyhttp

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mkorman9/tiny"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp/fasthttputil"
	"io"
	"net/http"
	"testing"
)

func init() {
	tiny.Init()
}

func TestHTTPServer(t *testing.T) {
	// given
	payload := "payload"

	app := NewServer("address").App
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).
			SendString(payload)
	})

	// when
	fasthttputil.NewInmemoryListener()
	req, _ := http.NewRequest("GET", "/test", nil)
	response, err := app.Test(req, -1)
	if err != nil {
		assert.Error(t, err)
		return
	}

	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		assert.Error(t, err)
		return
	}

	// then
	assert.Equal(t, fiber.StatusOK, response.StatusCode, "response code should be 200")
	assert.Equal(t, []byte(payload), responseBody, "response payload should match")
}
