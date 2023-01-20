package tinytcp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"sync"
	"time"
)

// Server represents a TCP server, and conforms to the tiny.Service interface.
type Server struct {
	config               *ServerConfig
	address              string
	listener             net.Listener
	listenerMutex        sync.RWMutex
	errorChannel         chan error
	forkingStrategy      ForkingStrategy
	sockets              *socketsList
	ticker               *time.Ticker
	abortOnce            sync.Once
	metrics              ServerMetrics
	metricsUpdateHandler func(*ServerMetrics)
	startHandler         func()
	stopHandler          func()
}

// NewServer returns new Server instance.
func NewServer(address string, config ...*ServerConfig) *Server {
	var providedConfig *ServerConfig
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeServerConfig(providedConfig)

	return &Server{
		config:       c,
		address:      address,
		errorChannel: make(chan error, 1),
		sockets:      newSocketsList(c.MaxClients),
	}
}

// Abort immediately stops the server with error.
func (s *Server) Abort(err error) {
	s.abortOnce.Do(func() {
		select {
		case s.errorChannel <- err:
		default:
		}

		s.Stop()
	})
}

// ForkingStrategy sets forking strategy used by this server (see ForkingStrategy).
func (s *Server) ForkingStrategy(forkingStrategy ForkingStrategy) {
	s.forkingStrategy = forkingStrategy
}

// Port returns a port number used by underlying listener. Only returns a valid value after Start().
func (s *Server) Port() int {
	s.listenerMutex.RLock()
	defer s.listenerMutex.RUnlock()

	if s.listener == nil {
		return -1
	}

	return resolveListenerPort(s.listener)
}

// Sockets returns a list of all client sockets currently connected.
func (s *Server) Sockets() []*Socket {
	return s.sockets.Copy()
}

// Metrics returns aggregated server metrics.
func (s *Server) Metrics() ServerMetrics {
	return s.metrics
}

// OnMetricsUpdate sets a handler that is called everytime the server metrics are updated.
func (s *Server) OnMetricsUpdate(handler func(*ServerMetrics)) {
	s.metricsUpdateHandler = handler
}

// OnStart sets a handler that is called when server starts.
func (s *Server) OnStart(handler func()) {
	s.startHandler = handler
}

// OnStop sets a handler that is called when server stops.
func (s *Server) OnStop(handler func()) {
	s.stopHandler = handler
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	if s.forkingStrategy == nil {
		log.Error().Msg(
			"Cannot start a TCP server with empty Forking ForkingStrategy. Call ForkingStrategy() before Start().",
		)

		return errors.New("empty forking strategy")
	}

	err := s.startServer()
	if err != nil {
		return err
	}

	if s.startHandler != nil {
		s.startHandler()
	}

	log.Info().Msgf("TCP server started (%s)", s.address)

	return s.acceptLoop()
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	s.listenerMutex.Lock()
	defer s.listenerMutex.Unlock()

	if s.listener == nil {
		return
	}

	if err := s.listener.Close(); err != nil {
		if !isBrokenPipe(err) {
			log.Error().Err(err).Msgf("Error shutting down TCP server (%s)", s.address)
		}
	}
	s.listener = nil

	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.ticker = nil

	sockets := s.Sockets()
	for _, socket := range sockets {
		_ = socket.Close()
	}
	s.sockets.Cleanup()

	s.forkingStrategy.OnStop()

	if s.stopHandler != nil {
		s.stopHandler()
	}

	log.Info().Msgf("TCP server stopped (%s)", s.address)
}

func (s *Server) startServer() error {
	s.listenerMutex.Lock()
	defer s.listenerMutex.Unlock()

	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return err
		}

		s.config.TLSConfig.Certificates = []tls.Certificate{cert}

		socket, err := tls.Listen(s.config.Network, s.address, s.config.TLSConfig)
		if err != nil {
			return err
		}

		s.listener = socket
	} else {
		socket, err := net.Listen(s.config.Network, s.address)
		if err != nil {
			return err
		}

		s.listener = socket
	}

	go s.startBackgroundJob()
	s.forkingStrategy.OnStart()

	return nil
}

func (s *Server) acceptLoop() error {
	for {
		connection, err := s.listener.Accept()
		if err != nil {
			if isBrokenPipe(err) {
				log.Debug().Msg("Broken pipe while calling Accept(), TCP server shutting down")
				break
			}

			log.Error().Err(err).Msg("Error while accepting TCP connection")
			continue
		}

		s.handleNewConnection(connection)
	}

	var err error
	select {
	case e := <-s.errorChannel:
		err = e
	default:
		err = nil
	}

	return err
}

func (s *Server) handleNewConnection(connection net.Conn) {
	socket := s.sockets.New(connection)
	if socket == nil {
		return
	}

	log.Debug().Msgf("Opening TCP client connection: %s", connection.RemoteAddr().String())

	s.forkingStrategy.OnAccept(socket)
}

func (s *Server) startBackgroundJob() {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Stack().
				Err(fmt.Errorf("%v", r)).
				Msg("Panic inside TCP server background job")

			s.Abort(errors.New("server background job restart loop"))
		}
	}()

	if s.ticker == nil {
		s.ticker = time.NewTicker(s.config.TickInterval)
	}

	for {
		select {
		case <-s.ticker.C:
			s.updateMetrics()
			s.sockets.Cleanup()
		}
	}
}

func (s *Server) updateMetrics() {
	s.sockets.ExecRead(func(head *Socket) {
		s.metrics.Connections = s.sockets.Len()
		if s.metrics.Connections > s.metrics.MaxConnections {
			s.metrics.MaxConnections = s.metrics.Connections
		}

		var (
			readsPerInterval  uint64
			writesPerInterval uint64
		)

		for socket := head; socket != nil; socket = socket.next {
			reads, writes := socket.updateMetrics(s.config.TickInterval)
			readsPerInterval += reads
			writesPerInterval += writes
		}

		s.metrics.TotalRead += readsPerInterval
		s.metrics.TotalWritten += writesPerInterval
		s.metrics.ReadsPerSecond = uint64(float64(readsPerInterval) / s.config.TickInterval.Seconds())
		s.metrics.WritesPerSecond = uint64(float64(writesPerInterval) / s.config.TickInterval.Seconds())

		s.forkingStrategy.OnMetricsUpdate(&s.metrics)

		if s.metricsUpdateHandler != nil {
			s.metricsUpdateHandler(&s.metrics)
		}
	})
}
