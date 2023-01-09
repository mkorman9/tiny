package tinytcp

import (
	"crypto/rand"
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
	config                 *ServerConfig
	listener               net.Listener
	forkingStrategy        ForkingStrategy
	sockets                map[*ClientSocket]struct{}
	socketsMutex           sync.RWMutex
	ticker                 *time.Ticker
	metrics                ServerMetrics
	metricsUpdateHandler   func()
	clientSocketPool       sync.Pool
	byteCountingReaderPool sync.Pool
	byteCountingWriterPool sync.Pool
}

// NewServer returns new Server instance.
func NewServer(address string, opts ...ServerOpt) *Server {
	config := &ServerConfig{
		address:    address,
		Network:    "tcp",
		MaxClients: -1,
		TLSConfig: &tls.Config{
			Rand: rand.Reader,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return &Server{
		config:  config,
		sockets: make(map[*ClientSocket]struct{}),
		clientSocketPool: sync.Pool{
			New: func() any {
				return &ClientSocket{}
			},
		},
		byteCountingReaderPool: sync.Pool{
			New: func() any {
				return &byteCountingReader{}
			},
		},
		byteCountingWriterPool: sync.Pool{
			New: func() any {
				return &byteCountingWriter{}
			},
		},
	}
}

// ForkingStrategy sets forking strategy used by this server (see ForkingStrategy).
func (s *Server) ForkingStrategy(forkingStrategy ForkingStrategy) {
	s.forkingStrategy = forkingStrategy
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	if s.forkingStrategy == nil {
		log.Error().Msg(
			"Cannot start a TCP server with empty Forking ForkingStrategy. Call ForkingStrategy() before Start().",
		)

		return errors.New("empty forking strategy")
	}

	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		var err error

		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return err
		}

		s.config.TLSConfig.Certificates = []tls.Certificate{cert}

		var socket net.Listener
		socket, err = tls.Listen(s.config.Network, s.config.address, s.config.TLSConfig)
		if err != nil {
			return err
		}

		s.listener = socket
	} else {
		socket, err := net.Listen(s.config.Network, s.config.address)
		if err != nil {
			return err
		}

		s.listener = socket
	}

	go s.startBackgroundJob()
	s.forkingStrategy.OnStart()

	log.Info().Msgf("TCP server started (%s)", s.config.address)

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

	return nil
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	if s.listener == nil {
		return
	}

	if err := s.listener.Close(); err != nil {
		log.Error().Err(err).Msgf("Error shutting down TCP server (%s)", s.config.address)
	}

	if s.ticker != nil {
		s.ticker.Stop()
	}

	sockets := s.Sockets()
	for _, socket := range sockets {
		socket.Close()
	}

	s.forkingStrategy.OnStop()

	log.Info().Msgf("TCP server stopped (%s)", s.config.address)
}

// Sockets returns a list of all client sockets currently connected.
func (s *Server) Sockets() []*ClientSocket {
	s.socketsMutex.RLock()
	defer s.socketsMutex.RUnlock()

	var list []*ClientSocket
	for socket := range s.sockets {
		if !socket.IsClosed() {
			list = append(list, socket)
		}
	}

	return list
}

// Metrics returns aggregated server metrics.
func (s *Server) Metrics() ServerMetrics {
	return s.metrics
}

// OnMetricsUpdate sets a handler that is called everytime the server metrics are updated.
func (s *Server) OnMetricsUpdate(handler func()) {
	s.metricsUpdateHandler = handler
}

func (s *Server) handleNewConnection(connection net.Conn) {
	clientSocket := s.newClientSocket(connection)

	if registered := s.registerClientSocket(clientSocket); !registered {
		// instantly terminate the connection if it can't be added to the server pool
		_ = clientSocket.connection.Close()
		s.recycleClientSocket(clientSocket)
		return
	}

	log.Debug().Msgf("Opening TCP client connection: %s", clientSocket.connection.RemoteAddr().String())

	s.forkingStrategy.OnAccept(clientSocket)
}

func (s *Server) newClientSocket(connection net.Conn) *ClientSocket {
	reader := s.byteCountingReaderPool.Get().(*byteCountingReader)
	reader.reader = connection

	writer := s.byteCountingWriterPool.Get().(*byteCountingWriter)
	writer.writer = connection

	cs := s.clientSocketPool.Get().(*ClientSocket)
	cs.remoteAddress = parseRemoteAddress(connection)
	cs.connectedAt = time.Now()
	cs.connection = connection
	cs.reader = reader
	cs.writer = writer
	cs.byteCountingReader = reader
	cs.byteCountingWriter = writer
	return cs
}

func (s *Server) registerClientSocket(clientSocket *ClientSocket) bool {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	if s.config.MaxClients >= 0 && len(s.sockets) >= s.config.MaxClients {
		return false
	}

	s.sockets[clientSocket] = struct{}{}
	return true
}

func (s *Server) startBackgroundJob() {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Stack().
				Err(fmt.Errorf("%v", r)).
				Msg("Panic inside TCP server background job")
		}
	}()

	s.ticker = time.NewTicker(1 * time.Second)

	for {
		select {
		case <-s.ticker.C:
			s.updateMetrics()
			s.cleanupClientSockets()
		}
	}
}

func (s *Server) updateMetrics() {
	s.socketsMutex.RLock()
	s.metrics.Connections = len(s.sockets)
	s.socketsMutex.RUnlock()

	s.metrics.ReadsPerSecond = 0
	s.metrics.WritesPerSecond = 0
	if s.metrics.Connections > s.metrics.MaxConnections {
		s.metrics.MaxConnections = s.metrics.Connections
	}
	s.forkingStrategy.OnMetricsUpdate(&s.metrics)

	for socket := range s.sockets {
		reads := socket.ReadsPerSecond()
		writes := socket.WritesPerSecond()

		s.metrics.TotalRead += reads
		s.metrics.TotalWritten += writes
		s.metrics.ReadsPerSecond += reads
		s.metrics.WritesPerSecond += writes
	}

	if s.metricsUpdateHandler != nil {
		s.metricsUpdateHandler()
	}

	for socket := range s.sockets {
		socket.resetMetrics()
	}
}

func (s *Server) cleanupClientSockets() {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	for socket := range s.sockets {
		if socket.IsClosed() {
			delete(s.sockets, socket)
			s.recycleClientSocket(socket)
		}
	}
}

func (s *Server) recycleClientSocket(clientSocket *ClientSocket) {
	clientSocket.byteCountingReader.reader = nil
	clientSocket.byteCountingReader.totalBytes = 0
	clientSocket.byteCountingReader.currentBytes = 0
	s.byteCountingReaderPool.Put(clientSocket.byteCountingReader)

	clientSocket.byteCountingWriter.writer = nil
	clientSocket.byteCountingWriter.totalBytes = 0
	clientSocket.byteCountingWriter.currentBytes = 0
	s.byteCountingWriterPool.Put(clientSocket.byteCountingWriter)

	clientSocket.reset()
	s.clientSocketPool.Put(clientSocket)
}
