package tinyhttp

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"time"
)

// Server is an object representing fiber.App and implementing the tiny.Service interface.
type Server struct {
	*fiber.App

	config       *ServerConfig
	errorHandler func(c *fiber.Ctx, err error) error
	panicHandler func(c *fiber.Ctx, recovered any)
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
		Concurrency:    256 * 1024,
		BodyLimit:      4 * 1024 * 1024,
		ReadBufferSize: 4096,
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

// OnPanic sets a handler for requests that resulted in panic.
func (s *Server) OnPanic(handler func(c *fiber.Ctx, recovered any)) {
	s.panicHandler = handler
}

// OnError sets a handler for requests that resulted in error.
func (s *Server) OnError(handler func(c *fiber.Ctx, err error) error) {
	s.errorHandler = handler
}

func (s *Server) createApp() *fiber.App {
	appConfig := fiber.Config{
		ErrorHandler:          s.errorFunction,
		Network:               s.config.Network,
		ReadTimeout:           s.config.ReadTimeout,
		WriteTimeout:          s.config.WriteTimeout,
		IdleTimeout:           s.config.IdleTimeout,
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Views:                 s.config.ViewEngine,
		ViewsLayout:           s.config.ViewLayout,
		Concurrency:           s.config.Concurrency,
		BodyLimit:             s.config.BodyLimit,
		ReadBufferSize:        s.config.ReadBufferSize,
	}

	if len(s.config.TrustedProxies) > 0 {
		appConfig.EnableTrustedProxyCheck = true
		appConfig.TrustedProxies = s.config.TrustedProxies
	} else {
		appConfig.EnableTrustedProxyCheck = false
	}
	appConfig.ProxyHeader = s.config.RemoteIPHeader

	if s.config.fiberOption != nil {
		s.config.fiberOption(&appConfig)
	}

	app := fiber.New(appConfig)

	app.Use(recover.New(recover.Config{
		StackTraceHandler: s.recoveryFunction,
	}))

	if s.config.SecurityHeaders {
		app.Use(s.securityHeadersFunction)
	}

	return app
}

func (s *Server) errorFunction(c *fiber.Ctx, err error) error {
	if s.errorHandler != nil {
		return s.errorHandler(c, err)
	}

	code := fiber.StatusInternalServerError

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}

	c.Status(code)
	return nil
}

func (s *Server) recoveryFunction(c *fiber.Ctx, recovered any) {
	if s.panicHandler != nil {
		s.panicHandler(c, recovered)
		return
	}

	log.Error().
		Stack().
		Err(fmt.Errorf("%v", recovered)).
		Msg("Panic inside an HTTP handler function")
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
