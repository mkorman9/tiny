package tinyhttp

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

// Server is an object representing fiber.App and implementing the tiny.Service interface.
type Server struct {
	*fiber.App

	config         *ServerConfig
	noRouteHandler func(c *fiber.Ctx)
	panicHandler   func(c *fiber.Ctx, recovered any)
}

// NewServer creates new Server instance.
func NewServer(address string, opts ...ServerOpt) *Server {
	config := ServerConfig{
		address:         address,
		Network:         "tcp",
		SecurityHeaders: true,
		ShutdownTimeout: 5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     2 * time.Minute,
		TrustedProxies: []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"fc00::/7",
			"::1/128",
		},
		RemoteIPHeader: "X-Forwarded-For",
	}

	for _, opt := range opts {
		opt(&config)
	}

	server := &Server{config: &config}
	server.App = server.createApp()

	return server
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	log.Info().Msgf("HTTP server started (%s)", s.config.address)

	var err error
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		err = s.ListenTLS(s.config.address, s.config.TLSCert, s.config.TLSKey)
	} else {
		err = s.Listen(s.config.address)
	}

	return err
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	if err := s.ShutdownWithTimeout(s.config.ShutdownTimeout); err != nil {
		log.Error().Err(err).Msgf("Error shutting down HTTP server (%s)", s.config.address)
	} else {
		log.Info().Msgf("HTTP server stopped (%s)", s.config.address)
	}
}

// OnPanic sets a handler for requests that resulted in panic and would end up as 500s.
func (s *Server) OnPanic(handler func(c *fiber.Ctx, recovered any)) {
	s.panicHandler = handler
}

func (s *Server) createApp() *fiber.App {
	appConfig := fiber.Config{
		ErrorHandler: s.errorHandler,
		Network:      s.config.Network,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	if len(s.config.TrustedProxies) > 0 {
		appConfig.EnableTrustedProxyCheck = true
		appConfig.TrustedProxies = s.config.TrustedProxies
	} else {
		appConfig.EnableTrustedProxyCheck = false
	}
	appConfig.ProxyHeader = s.config.RemoteIPHeader

	app := fiber.New(appConfig)

	app.Use(recover.New(recover.Config{
		StackTraceHandler: s.recoveryFunction,
	}))

	if s.config.SecurityHeaders {
		app.Use(s.securityHeadersFunction)
	}

	return app
}

func (s *Server) errorHandler(ctx *fiber.Ctx, err error) error {
	code := http.StatusInternalServerError

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}

	ctx.Status(code)
	return nil
}

func (s *Server) recoveryFunction(c *fiber.Ctx, recovered any) {
	log.Error().
		Stack().
		Err(fmt.Errorf("%v", recovered)).
		Msg("Panic inside an HTTP handler function")

	if s.panicHandler != nil {
		s.panicHandler(c, recovered)
		return
	}

	c.Status(http.StatusInternalServerError)
}

func (s *Server) securityHeadersFunction(c *fiber.Ctx) error {
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-XSS-Protection", "0")

	if c.Protocol() == "https" {
		c.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	}

	return c.Next()
}
