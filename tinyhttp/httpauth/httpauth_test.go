package httpauth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinyhttp"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

func init() {
	tiny.Init()
}

func TestMissingToken(t *testing.T) {
	// given
	payload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	app := tinyhttp.NewServer("address").App
	app.Get(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).
				SendString(payload)
		},
	)

	// when
	req, _ := http.NewRequest("GET", "/secured", nil)

	response, err := app.Test(req, -1)
	if err != nil {
		assert.Error(t, err)
		return
	}

	// then
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode, "response code should be 401")
}

func TestInvalidToken(t *testing.T) {
	// given
	payload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	app := tinyhttp.NewServer("address").App
	app.Get(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).
				SendString(payload)
		},
	)

	// when
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer incorrectToken")

	response, err := app.Test(req, -1)
	if err != nil {
		assert.Error(t, err)
		return
	}

	// then
	assert.Equal(t, fiber.StatusUnauthorized, response.StatusCode, "response code should be 401")
}

func TestValidToken(t *testing.T) {
	// given
	payload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	app := tinyhttp.NewServer("address").App
	app.Get(
		"/secured",
		middleware.AnyOfRoles("ADMIN"),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).
				SendString(payload)
		},
	)

	// when
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer "+correctToken)

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

func TestInvalidRoles(t *testing.T) {
	// given
	payload := "payload"
	correctToken := "token"

	middleware := createBearerTokenMiddleware(correctToken)

	app := tinyhttp.NewServer("address").App
	app.Get(
		"/secured",
		middleware.AnyOfRoles("SUPERUSER"),
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusOK).
				SendString(payload)
		},
	)

	// when
	req, _ := http.NewRequest("GET", "/secured", nil)
	req.Header.Set("Authorization", "Bearer "+correctToken)

	response, err := app.Test(req, -1)
	if err != nil {
		assert.Error(t, err)
		return
	}

	// then
	assert.Equal(t, fiber.StatusForbidden, response.StatusCode, "response code should be 403")
}

func createBearerTokenMiddleware(correctToken string) *Middleware {
	return NewBearerTokenMiddleware(func(c *fiber.Ctx, token string) (*VerificationResult, error) {
		if token == correctToken {
			return &VerificationResult{Verified: true, Roles: []string{"ADMIN"}}, nil
		} else {
			return &VerificationResult{}, nil
		}
	})
}
