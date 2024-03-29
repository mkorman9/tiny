package tinyhttp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog/log"
	"net"
)

// Server is an object representing fiber.App and implementing the tiny.Service interface.
type Server struct {
	*fiber.App

	config       *ServerConfig
	address      string
	errorHandler func(c *fiber.Ctx, err error) error
	panicHandler func(c *fiber.Ctx, recovered any)
}

// NewServer creates new Server instance.
func NewServer(address string, config ...*ServerConfig) *Server {
	var providedConfig *ServerConfig
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeServerConfig(providedConfig)

	server := &Server{
		config:  c,
		address: address,
	}
	server.App = server.createApp()

	return server
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	log.Info().Msgf("HTTP server started (%s)", s.address)

	var listener net.Listener

	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return err
		}

		tlsHandler := &fiber.TLSHandler{}
		s.config.TLSConfig.Certificates = []tls.Certificate{cert}
		s.config.TLSConfig.GetCertificate = tlsHandler.GetClientInfo
		s.SetTLSHandler(tlsHandler)

		socket, err := tls.Listen(s.config.Network, s.address, s.config.TLSConfig)
		if err != nil {
			return err
		}

		listener = socket
	} else {
		socket, err := net.Listen(s.config.Network, s.address)
		if err != nil {
			return err
		}

		listener = socket
	}

	return s.Listener(listener)
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	if err := s.ShutdownWithTimeout(s.config.ShutdownTimeout); err != nil {
		log.Error().Err(err).Msgf("Error shutting down HTTP server (%s)", s.address)
	} else {
		log.Info().Msgf("HTTP server stopped (%s)", s.address)
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
		ReadTimeout:           s.config.ReadTimeout,
		WriteTimeout:          s.config.WriteTimeout,
		IdleTimeout:           s.config.IdleTimeout,
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		EnableIPValidation:    false,
		JSONEncoder:           json.Marshal,
		JSONDecoder:           json.Unmarshal,
		Views:                 s.config.ViewEngine,
		ViewsLayout:           s.config.ViewLayout,
		Concurrency:           s.config.Concurrency,
		BodyLimit:             s.config.BodyLimit,
		ReadBufferSize:        s.config.ReadBufferSize,
		WriteBufferSize:       s.config.WriteBufferSize,
	}

	if len(s.config.TrustedProxies) > 0 {
		appConfig.EnableTrustedProxyCheck = true
		appConfig.TrustedProxies = s.config.TrustedProxies
	} else {
		appConfig.EnableTrustedProxyCheck = false
	}
	appConfig.ProxyHeader = s.config.RemoteIPHeader

	if s.config.FiberOpt != nil {
		s.config.FiberOpt(&appConfig)
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
