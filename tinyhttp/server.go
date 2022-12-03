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
		address:         address,
		Network:         "tcp",
		SecurityHeaders: true,
		ShutdownTimeout: 5 * time.Second,
		TLSConfig:       &tls.Config{},
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
func (server *Server) Start() error {
	log.Info().Msgf("HTTP server started (%s)", server.config.address)

	httpServer := &http.Server{
		Handler:           server.Engine,
		TLSConfig:         server.config.TLSConfig,
		ReadTimeout:       server.config.ReadTimeout,
		ReadHeaderTimeout: server.config.ReadHeaderTimeout,
		WriteTimeout:      server.config.WriteTimeout,
		IdleTimeout:       server.config.IdleTimeout,
		MaxHeaderBytes:    server.config.MaxHeaderBytes,
	}

	server.httpServerLock.Lock()
	server.httpServer = httpServer
	server.httpServerLock.Unlock()

	l, err := net.Listen(server.config.Network, server.config.address)
	if err != nil {
		return err
	}

	if server.config.TLSCert != "" && server.config.TLSKey != "" {
		err = httpServer.ServeTLS(l, server.config.TLSCert, server.config.TLSKey)
	} else {
		err = httpServer.Serve(l)
	}

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop implements the interface of tiny.Service.
func (server *Server) Stop() {
	server.httpServerLock.Lock()
	defer func() {
		server.httpServer = nil
		server.httpServerLock.Unlock()
	}()

	if server.httpServer == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), server.config.ShutdownTimeout)
	defer cancel()

	err := server.httpServer.Shutdown(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("Error shutting down HTTP server (%s)", server.config.address)
	} else {
		log.Info().Msgf("HTTP server stopped (%s)", server.config.address)
	}
}

// OnNoRoute sets a handler for requests that cannot be routed and would end up as 404s.
func (server *Server) OnNoRoute(handler func(c *gin.Context)) {
	server.noRouteHandler = handler
}

// OnPanic sets a handler for requests that resulted in panic and would end up as 500s.
func (server *Server) OnPanic(handler func(c *gin.Context, recovered any)) {
	server.panicHandler = handler
}

func (server *Server) createEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.CustomRecoveryWithWriter(io.Discard, server.recoveryFunction))

	if len(server.config.RemoteIPHeaders) > 0 && len(server.config.TrustedProxies) > 0 {
		engine.ForwardedByClientIP = true
		engine.RemoteIPHeaders = server.config.RemoteIPHeaders
		_ = engine.SetTrustedProxies(server.config.TrustedProxies)
	} else {
		engine.ForwardedByClientIP = false
	}

	if server.config.SecurityHeaders {
		engine.Use(server.securityHeadersFunction)
	}

	engine.HandleMethodNotAllowed = false

	engine.NoRoute(server.noRouteFunction)

	return engine
}

func (server *Server) noRouteFunction(c *gin.Context) {
	if server.noRouteHandler != nil {
		server.noRouteHandler(c)
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}

func (server *Server) securityHeadersFunction(c *gin.Context) {
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "0")

	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	}
}

func (server *Server) recoveryFunction(c *gin.Context, recovered any) {
	log.Error().
		Stack().
		Err(fmt.Errorf("%v", recovered)).
		Msg("Panic inside an HTTP handler function")

	if server.panicHandler != nil {
		server.panicHandler(c, recovered)
		return
	}

	c.AbortWithStatus(http.StatusInternalServerError)
}
