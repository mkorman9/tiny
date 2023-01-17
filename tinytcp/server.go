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
	config                 *ServerConfig
	listener               net.Listener
	forkingStrategy        ForkingStrategy
	socketsListHead        *connectedSocketNode
	socketsCount           int
	socketsMutex           sync.RWMutex
	ticker                 *time.Ticker
	metrics                ServerMetrics
	metricsUpdateHandler   func()
	connectedSocketPool    sync.Pool
	byteCountingReaderPool sync.Pool
	byteCountingWriterPool sync.Pool
	socketNodesPool        sync.Pool
}

type connectedSocketNode struct {
	socket *ConnectedSocket
	prev   *connectedSocketNode
	next   *connectedSocketNode
}

// NewServer returns new Server instance.
func NewServer(address string, opts ...ServerOpt) *Server {
	config := &ServerConfig{
		address:    address,
		Network:    "tcp",
		MaxClients: -1,
		TLSConfig:  &tls.Config{},
	}

	for _, opt := range opts {
		opt(config)
	}

	return &Server{
		config:          config,
		socketsListHead: nil,
		socketsCount:    0,
		connectedSocketPool: sync.Pool{
			New: func() any {
				return &ConnectedSocket{}
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
		socketNodesPool: sync.Pool{
			New: func() any {
				return &connectedSocketNode{}
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

	err := s.startListener()
	if err != nil {
		return err
	}

	go s.startBackgroundJob()
	s.forkingStrategy.OnStart()

	log.Info().Msgf("TCP server started (%s)", s.config.address)

	return s.acceptLoop()
}

func (s *Server) startListener() error {
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return err
		}

		s.config.TLSConfig.Certificates = []tls.Certificate{cert}

		socket, err := tls.Listen(s.config.Network, s.config.address, s.config.TLSConfig)
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
		_ = socket.Close()
	}

	s.forkingStrategy.OnStop()

	log.Info().Msgf("TCP server stopped (%s)", s.config.address)
}

// Sockets returns a list of all client sockets currently connected.
func (s *Server) Sockets() []*ConnectedSocket {
	s.socketsMutex.RLock()
	defer s.socketsMutex.RUnlock()

	var list []*ConnectedSocket
	var node = s.socketsListHead

	for node != nil {
		if !node.socket.IsClosed() {
			list = append(list, node.socket)
		}

		node = node.next
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
	socket := s.newConnectedSocket(connection)

	if registered := s.registerConnectedSocket(socket); !registered {
		// instantly terminate the connection if it can't be added to the server pool
		_ = socket.connection.Close()
		s.recycleConnectedSocket(socket)
		return
	}

	log.Debug().Msgf("Opening TCP client connection: %s", socket.connection.RemoteAddr().String())

	s.forkingStrategy.OnAccept(socket)
}

func (s *Server) newConnectedSocket(connection net.Conn) *ConnectedSocket {
	reader := s.byteCountingReaderPool.Get().(*byteCountingReader)
	reader.reader = connection

	writer := s.byteCountingWriterPool.Get().(*byteCountingWriter)
	writer.writer = connection

	cs := s.connectedSocketPool.Get().(*ConnectedSocket)
	cs.remoteAddress = parseRemoteAddress(connection)
	cs.connectedAt = time.Now()
	cs.connection = connection
	cs.reader = reader
	cs.writer = writer
	cs.byteCountingReader = reader
	cs.byteCountingWriter = writer
	return cs
}

func (s *Server) registerConnectedSocket(socket *ConnectedSocket) bool {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	if s.config.MaxClients >= 0 && s.socketsCount >= s.config.MaxClients {
		return false
	}

	node := s.socketNodesPool.Get().(*connectedSocketNode)
	node.socket = socket

	if s.socketsListHead == nil {
		s.socketsListHead = node
	} else {
		s.socketsListHead.next = node
		node.prev = s.socketsListHead
	}

	s.socketsCount++
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
			s.cleanupConnectedSockets()
		}
	}
}

func (s *Server) updateMetrics() {
	s.socketsMutex.RLock()
	defer s.socketsMutex.RUnlock()

	s.metrics.Connections = s.socketsCount

	s.metrics.ReadsPerSecond = 0
	s.metrics.WritesPerSecond = 0
	if s.metrics.Connections > s.metrics.MaxConnections {
		s.metrics.MaxConnections = s.metrics.Connections
	}
	s.forkingStrategy.OnMetricsUpdate(&s.metrics)

	var node = s.socketsListHead
	for node != nil {
		socket := node.socket

		reads := socket.ReadsPerSecond()
		writes := socket.WritesPerSecond()

		s.metrics.TotalRead += reads
		s.metrics.TotalWritten += writes
		s.metrics.ReadsPerSecond += reads
		s.metrics.WritesPerSecond += writes

		node = node.next
	}

	if s.metricsUpdateHandler != nil {
		s.metricsUpdateHandler()
	}

	node = s.socketsListHead
	for node != nil {
		socket := node.socket
		socket.resetMetrics()
		node = node.next
	}
}

func (s *Server) cleanupConnectedSockets() {
	s.socketsMutex.Lock()
	defer s.socketsMutex.Unlock()

	var node = s.socketsListHead

	for node != nil {
		socket := node.socket

		if socket.IsClosed() {
			if node == s.socketsListHead {
				s.socketsListHead = node
			} else {
				node.prev.next = node.next
				node.next.prev = node.prev
			}

			next := node.next
			node.socket = nil
			node.next = nil
			node.prev = nil
			s.socketNodesPool.Put(node)
			node = next

			s.recycleConnectedSocket(socket)
		} else {
			node = node.next
		}
	}
}

func (s *Server) recycleConnectedSocket(socket *ConnectedSocket) {
	socket.byteCountingReader.reader = nil
	socket.byteCountingReader.totalBytes = 0
	socket.byteCountingReader.currentBytes = 0
	s.byteCountingReaderPool.Put(socket.byteCountingReader)

	socket.byteCountingWriter.writer = nil
	socket.byteCountingWriter.totalBytes = 0
	socket.byteCountingWriter.currentBytes = 0
	s.byteCountingWriterPool.Put(socket.byteCountingWriter)

	socket.reset()
	s.connectedSocketPool.Put(socket)
}
