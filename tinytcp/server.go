package tinytcp

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	mathrand "math/rand"
	"net"
	"sync"
	"time"
)

// Server represents a TCP server, and conforms to the tiny.Service interface.
type Server struct {
	config                 *ServerConfig
	listener               net.Listener
	forkingStrategy        ForkingStrategy
	sockets                []*ClientSocket
	socketsMutex           sync.RWMutex
	ticker                 *time.Ticker
	metrics                ServerMetrics
	metricsUpdateHandler   func()
	idRand                 *mathrand.Rand
	clientSocketPool       sync.Pool
	byteCountingReaderPool sync.Pool
	byteCountingWriterPool sync.Pool
}

// NewServer returns new Server instance.
func NewServer(address string, opts ...ServerOpt) *Server {
	config := &ServerConfig{
		address:    address,
		Mode:       Both,
		MaxClients: -1,
		TLSConfig: &tls.Config{
			Rand: rand.Reader,
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	return &Server{
		config: config,
		idRand: mathrand.New(mathrand.NewSource(time.Now().Unix())),
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

	listenMode := "tcp"
	switch s.config.Mode {
	case IPv4_Only:
		listenMode = "tcp4"
	case IPv6_Only:
		listenMode = "tcp6"
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
		socket, err = tls.Listen(listenMode, s.config.address, s.config.TLSConfig)
		if err != nil {
			return err
		}

		s.listener = socket
	} else {
		socket, err := net.Listen(listenMode, s.config.address)
		if err != nil {
			return err
		}

		s.listener = socket
	}

	log.Info().Msgf("TCP server started (%s)", s.config.address)
	go s.startBackgroundJob()
	s.forkingStrategy.OnStart()

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
	for _, socket := range s.sockets {
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
	clientSocket := s.newClientSocket(connection, s.idRand.Int63())

	if added := s.addClientSocket(clientSocket); !added {
		// instantly terminate the connection if it can't be added to the server pool
		_ = clientSocket.connection.Close()
		s.recycleClientSocket(clientSocket)
		return
	}

	log.Debug().Msgf("Opening TCP client connection #%d (%s)", clientSocket.Id(), clientSocket.RemoteAddress())

	s.forkingStrategy.OnAccept(clientSocket)
}

func (s *Server) newClientSocket(connection net.Conn, id int64) *ClientSocket {
	reader := s.byteCountingReaderPool.Get().(*byteCountingReader)
	reader.reader = connection

	writer := s.byteCountingWriterPool.Get().(*byteCountingWriter)
	writer.writer = connection

	cs := s.clientSocketPool.Get().(*ClientSocket)
	cs.id = id
	cs.remoteAddress = parseRemoteAddress(connection)
	cs.connectedAt = time.Now()
	cs.connection = connection
	cs.reader = reader
	cs.writer = writer
	cs.byteCountingReader = reader
	cs.byteCountingWriter = writer
	return cs
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
	defer s.socketsMutex.RUnlock()

	s.metrics.ReadsPerSecond = 0
	s.metrics.WritesPerSecond = 0
	s.metrics.Connections = len(s.sockets)
	if s.metrics.Connections > s.metrics.MaxConnections {
		s.metrics.MaxConnections = s.metrics.Connections
	}
	s.forkingStrategy.OnMetricsUpdate(&s.metrics)

	for _, socket := range s.sockets {
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

	for _, socket := range s.sockets {
		socket.resetMetrics()
	}
}

func (s *Server) addClientSocket(clientSocket *ClientSocket) bool {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	if s.config.MaxClients >= 0 && len(s.sockets) >= s.config.MaxClients {
		return false
	}

	s.sockets = append(s.sockets, clientSocket)
	return true
}

func (s *Server) cleanupClientSockets() {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	var toRemove map[*ClientSocket]struct{}

	for _, socket := range s.sockets {
		if socket.IsClosed() {
			if toRemove == nil {
				toRemove = map[*ClientSocket]struct{}{}
			}

			toRemove[socket] = struct{}{}
		}
	}

	if len(toRemove) > 0 {
		var sockets []*ClientSocket
		for _, socket := range s.sockets {
			if _, toBeRemoved := toRemove[socket]; toBeRemoved {
				s.recycleClientSocket(socket)
			} else {
				sockets = append(sockets, socket)
			}
		}

		s.sockets = sockets
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
