package tinyhttp

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Server is an object representing gin.Engine and implementing the tiny.Service interface.
type Server struct {
	*gin.Engine

	config         *ServerConfig
	httpServer     *http.Server
	httpServerLock sync.Mutex
	noRouteHandler func(c *gin.Context)
	panicHandler   func(c *gin.Context, recovered any)
}

// NewServer creates new Server instance.
func NewServer(address string, opts ...ServerOpt) *Server {
	config := ServerConfig{
		address:          address,
		Network:          "tcp",
		GinMode:          gin.ReleaseMode,
		SecurityHeaders:  true,
		MethodNotAllowed: false,
		ShutdownTimeout:  5 * time.Second,
		TLSConfig:        &tls.Config{},
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     10 * time.Second,
		IdleTimeout:      2 * time.Minute,
		TrustedProxies: []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
			"192.168.0.0/16",
			"127.0.0.0/8",
			"fc00::/7",
			"::1/128",
		},
		RemoteIPHeaders: []string{
			"X-Forwarded-For",
		},
	}

	for _, opt := range opts {
		opt(&config)
	}

	server := &Server{config: &config}
	server.Engine = server.createEngine()

	return server
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	log.Info().Msgf("HTTP server started (%s)", s.config.address)

	httpServer := &http.Server{
		Handler:           s.Engine,
		TLSConfig:         s.config.TLSConfig,
		ReadTimeout:       s.config.ReadTimeout,
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		WriteTimeout:      s.config.WriteTimeout,
		IdleTimeout:       s.config.IdleTimeout,
		MaxHeaderBytes:    s.config.MaxHeaderBytes,
	}

	s.httpServerLock.Lock()
	s.httpServer = httpServer
	s.httpServerLock.Unlock()

	l, err := net.Listen(s.config.Network, s.config.address)
	if err != nil {
		return err
	}

	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		err = httpServer.ServeTLS(l, s.config.TLSCert, s.config.TLSKey)
	} else {
		err = httpServer.Serve(l)
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	s.httpServerLock.Lock()
	defer func() {
		s.httpServer = nil
		s.httpServerLock.Unlock()
	}()

	if s.httpServer == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Error shutting down HTTP server (%s)", s.config.address)
	} else {
		log.Info().Msgf("HTTP server stopped (%s)", s.config.address)
	}
}

// OnNoRoute sets a handler for requests that cannot be routed and would end up as 404s.
func (s *Server) OnNoRoute(handler func(c *gin.Context)) {
	s.noRouteHandler = handler
}

// OnPanic sets a handler for requests that resulted in panic and would end up as 500s.
func (s *Server) OnPanic(handler func(c *gin.Context, recovered any)) {
	s.panicHandler = handler
}

func (s *Server) createEngine() *gin.Engine {
	gin.SetMode(s.config.GinMode)

	engine := gin.New()
	engine.Use(gin.CustomRecoveryWithWriter(io.Discard, s.recoveryFunction))

	if len(s.config.RemoteIPHeaders) > 0 && len(s.config.TrustedProxies) > 0 {
		engine.ForwardedByClientIP = true
		engine.RemoteIPHeaders = s.config.RemoteIPHeaders
		_ = engine.SetTrustedProxies(s.config.TrustedProxies)
	} else {
		engine.ForwardedByClientIP = false
	}

	if s.config.SecurityHeaders {
		engine.Use(s.securityHeadersFunction)
	}

	engine.HandleMethodNotAllowed = s.config.MethodNotAllowed
	engine.NoRoute(s.noRouteFunction)

	return engine
}

func (s *Server) noRouteFunction(c *gin.Context) {
	if s.noRouteHandler != nil {
		s.noRouteHandler(c)
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}

func (s *Server) securityHeadersFunction(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "0")

	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	}
}

func (s *Server) recoveryFunction(c *gin.Context, recovered any) {
	log.Error().
		Stack().
		Err(fmt.Errorf("%v", recovered)).
		Msg("Panic inside an HTTP handler function")

	if s.panicHandler != nil {
		s.panicHandler(c, recovered)
		return
	}

	c.AbortWithStatus(http.StatusInternalServerError)
}
