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

// ConnectedSocket represents a dedicated socket for given TCP client.
type ConnectedSocket struct {
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

// ConnectedSocketHandler represents a signature of function used by Server to handle new connections.
type ConnectedSocketHandler func(*ConnectedSocket)

func (cs *ConnectedSocket) reset() {
	cs.remoteAddress = ""
	cs.connection = nil
	cs.reader = nil
	cs.writer = nil
	cs.byteCountingReader = nil
	cs.byteCountingWriter = nil
	cs.isClosed = 0
	cs.closeOnce = sync.Once{}
	cs.closeHandlers = nil
	cs.closeHandlersMutex = sync.RWMutex{}
}

// RemoteAddress returns a remote address of the socket.
func (cs *ConnectedSocket) RemoteAddress() string {
	return cs.remoteAddress
}

// ConnectedAt returns an exact time the socket has connected.
func (cs *ConnectedSocket) ConnectedAt() time.Time {
	return cs.connectedAt
}

// IsClosed check whether this connection has been closed, either by the server or the client.
func (cs *ConnectedSocket) IsClosed() bool {
	return atomic.LoadUint32(&cs.isClosed) == 1
}

// Close closes TCP connection.
func (cs *ConnectedSocket) Close() {
	cs.closeOnce.Do(func() {
		log.Debug().
			Msgf("Closing TCP client connection: %s", cs.connection.RemoteAddr().String())

		atomic.StoreUint32(&cs.isClosed, 1)

		if err := cs.connection.Close(); err != nil {
			log.Error().
				Err(err).
				Msgf("Error while closing TCP client connection: %s", cs.connection.RemoteAddr().String())
		}

		cs.closeHandlersMutex.RLock()
		defer cs.closeHandlersMutex.RUnlock()

		for i := len(cs.closeHandlers) - 1; i >= 0; i-- {
			handler := cs.closeHandlers[i]
			handler()
		}
	})
}

// Read conforms to the io.Reader interface.
func (cs *ConnectedSocket) Read(b []byte) (int, error) {
	n, err := cs.reader.Read(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", cs.connection.RemoteAddr().String())
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while reading from TCP client connection: %s", cs.connection.RemoteAddr().String())
		}

		return n, err
	}

	return n, nil
}

// Write conforms to the io.Writer interface.
func (cs *ConnectedSocket) Write(b []byte) (int, error) {
	n, err := cs.writer.Write(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", cs.connection.RemoteAddr().String())
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while writing to TCP client connection: %s", cs.connection.RemoteAddr().String())
		}

		return n, err
	}

	return n, nil
}

// SetReadDeadline sets read deadline for underlying socket.
func (cs *ConnectedSocket) SetReadDeadline(deadline time.Time) error {
	err := cs.connection.SetReadDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", cs.connection.RemoteAddr().String())
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while setting read deadline for TCP client connection: %s", cs.connection.RemoteAddr().String())
		}

		return err
	}

	return nil
}

// SetWriteDeadline sets read deadline for underlying socket.
func (cs *ConnectedSocket) SetWriteDeadline(deadline time.Time) error {
	err := cs.connection.SetWriteDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().
				Msgf("Connection closed by TCP client: %s", cs.connection.RemoteAddr().String())
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().
				Err(err).
				Msgf("Error while setting write deadline for TCP client connection: %s", cs.connection.RemoteAddr().String())
		}

		return err
	}

	return nil
}

// OnClose registers a handler that is called when underlying TCP connection is being closed.
func (cs *ConnectedSocket) OnClose(handler func()) {
	cs.closeHandlersMutex.Lock()
	defer cs.closeHandlersMutex.Unlock()

	cs.closeHandlers = append(cs.closeHandlers, handler)
}

// Unwrap returns underlying net.Conn instance from ConnectedSocket.
func (cs *ConnectedSocket) Unwrap() net.Conn {
	return cs.connection
}

// UnwrapTLS tries to return underlying tls.Conn instance from ConnectedSocket.
func (cs *ConnectedSocket) UnwrapTLS() (*tls.Conn, bool) {
	if conn, ok := cs.connection.(*tls.Conn); ok {
		return conn, true
	}

	return nil, false
}

// WrapReader allows to wrap reader object into user defined wrapper.
func (cs *ConnectedSocket) WrapReader(wrapper func(io.Reader) io.Reader) {
	cs.reader = wrapper(cs.reader)
}

// WrapWriter allows to wrap writer object into user defined wrapper.
func (cs *ConnectedSocket) WrapWriter(wrapper func(io.Writer) io.Writer) {
	cs.writer = wrapper(cs.writer)
}

// TotalRead returns a total number of bytes read through this socket.
func (cs *ConnectedSocket) TotalRead() uint64 {
	return cs.byteCountingReader.Total()
}

// ReadsPerSecond returns a total number of bytes read through socket this second.
func (cs *ConnectedSocket) ReadsPerSecond() uint64 {
	return cs.byteCountingReader.Current()
}

// TotalWritten returns a total number of bytes written through this socket.
func (cs *ConnectedSocket) TotalWritten() uint64 {
	return cs.byteCountingWriter.Total()
}

// WritesPerSecond returns a total number of bytes written through socket this second.
func (cs *ConnectedSocket) WritesPerSecond() uint64 {
	return cs.byteCountingWriter.Current()
}

func (cs *ConnectedSocket) resetMetrics() {
	cs.byteCountingReader.reset()
	cs.byteCountingWriter.reset()
}
