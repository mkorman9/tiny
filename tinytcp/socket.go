package tinytcp

import (
	"crypto/tls"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Socket represents a dedicated socket for given TCP client.
type Socket struct {
	remoteAddress      string
	connectedAt        time.Time
	connection         net.Conn
	reader             io.Reader
	writer             io.Writer
	byteCountingReader *byteCountingReader
	byteCountingWriter *byteCountingWriter
	isClosed           uint32
	closeOnce          sync.Once
	closeHandlers      []func()
	closeHandlersMutex sync.RWMutex
}

// SocketHandler represents a signature of function used by Server to handle new connections.
type SocketHandler func(*Socket)

func (s *Socket) reset() {
	s.remoteAddress = ""
	s.connection = nil
	s.reader = nil
	s.writer = nil
	s.byteCountingReader = nil
	s.byteCountingWriter = nil
	s.isClosed = 0
	s.closeOnce = sync.Once{}
	s.closeHandlers = nil
	s.closeHandlersMutex = sync.RWMutex{}
}

// RemoteAddress returns a remote address of the socket.
func (s *Socket) RemoteAddress() string {
	return s.remoteAddress
}

// ConnectedAt returns an exact time the socket has connected.
func (s *Socket) ConnectedAt() time.Time {
	return s.connectedAt
}

// IsClosed check whether this connection has been closed, either by the server or the client.
func (s *Socket) IsClosed() bool {
	return atomic.LoadUint32(&s.isClosed) == 1
}

// Close closes underlying TCP connection and executes all the registered close handlers.
// This method always returns nil, but its signature is meant to stick to the io.Closer interface.
func (s *Socket) Close() error {
	s.closeOnce.Do(func() {
		atomic.StoreUint32(&s.isClosed, 1)

		log.Debug().
			Msgf("Closing TCP client connection: %s", s.connection.RemoteAddr().String())

		if err := s.connection.Close(); err != nil {
			log.Error().
				Err(err).
				Msgf("Error while closing TCP client connection: %s", s.connection.RemoteAddr().String())
		}

		s.closeHandlersMutex.RLock()
		defer s.closeHandlersMutex.RUnlock()

		for i := len(s.closeHandlers) - 1; i >= 0; i-- {
			handler := s.closeHandlers[i]
			handler()
		}
	})

	return nil
}

// Read conforms to the io.Reader interface.
func (s *Socket) Read(b []byte) (int, error) {
	n, err := s.reader.Read(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", s.connection.RemoteAddr().String())
			_ = s.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while reading from TCP client connection: %s", s.connection.RemoteAddr().String())
		}

		return n, err
	}

	return n, nil
}

// Write conforms to the io.Writer interface.
func (s *Socket) Write(b []byte) (int, error) {
	n, err := s.writer.Write(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", s.connection.RemoteAddr().String())
			_ = s.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while writing to TCP client connection: %s", s.connection.RemoteAddr().String())
		}

		return n, err
	}

	return n, nil
}

// SetReadDeadline sets read deadline for underlying socket.
func (s *Socket) SetReadDeadline(deadline time.Time) error {
	err := s.connection.SetReadDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", s.connection.RemoteAddr().String())
			_ = s.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while setting read deadline for TCP client connection: %s", s.connection.RemoteAddr().String())
		}

		return err
	}

	return nil
}

// SetWriteDeadline sets read deadline for underlying socket.
func (s *Socket) SetWriteDeadline(deadline time.Time) error {
	err := s.connection.SetWriteDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", s.connection.RemoteAddr().String())
			_ = s.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while setting write deadline for TCP client connection: %s", s.connection.RemoteAddr().String())
		}

		return err
	}

	return nil
}

// OnClose registers a handler that is called when underlying TCP connection is being closed.
func (s *Socket) OnClose(handler func()) {
	s.closeHandlersMutex.Lock()
	defer s.closeHandlersMutex.Unlock()

	s.closeHandlers = append(s.closeHandlers, handler)
}

// Unwrap returns underlying net.Conn instance from Socket.
func (s *Socket) Unwrap() net.Conn {
	return s.connection
}

// UnwrapTLS tries to return underlying tls.Conn instance from Socket.
func (s *Socket) UnwrapTLS() (*tls.Conn, bool) {
	if conn, ok := s.connection.(*tls.Conn); ok {
		return conn, true
	}

	return nil, false
}

// WrapReader allows to wrap reader object into user defined wrapper.
func (s *Socket) WrapReader(wrapper func(io.Reader) io.Reader) {
	s.reader = wrapper(s.reader)
}

// WrapWriter allows to wrap writer object into user defined wrapper.
func (s *Socket) WrapWriter(wrapper func(io.Writer) io.Writer) {
	s.writer = wrapper(s.writer)
}

// TotalRead returns a total number of bytes read through this socket.
func (s *Socket) TotalRead() uint64 {
	return s.byteCountingReader.Total()
}

// ReadsPerSecond returns a total number of bytes read through socket this second.
func (s *Socket) ReadsPerSecond() uint64 {
	return s.byteCountingReader.Current()
}

// TotalWritten returns a total number of bytes written through this socket.
func (s *Socket) TotalWritten() uint64 {
	return s.byteCountingWriter.Total()
}

// WritesPerSecond returns a total number of bytes written through socket this second.
func (s *Socket) WritesPerSecond() uint64 {
	return s.byteCountingWriter.Current()
}

func (s *Socket) resetMetrics() {
	s.byteCountingReader.reset()
	s.byteCountingWriter.reset()
}
