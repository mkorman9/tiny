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

// ClientSocket represents a dedicated socket for given TCP client.
type ClientSocket struct {
	id                 int64
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

// ClientSocketHandler represents a signature of function used by Server to handle new connections.
type ClientSocketHandler func(*ClientSocket)

func (cs *ClientSocket) reset() {
	cs.id = 0
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

// Id returns a unique id associated with this socket.
func (cs *ClientSocket) Id() int64 {
	return cs.id
}

// RemoteAddress returns a remote address of the client.
func (cs *ClientSocket) RemoteAddress() string {
	return cs.remoteAddress
}

// ConnectedAt returns an exact time the client has connected.
func (cs *ClientSocket) ConnectedAt() time.Time {
	return cs.connectedAt
}

// IsClosed check whether this connection has been closed, either by the server or the client.
func (cs *ClientSocket) IsClosed() bool {
	return atomic.LoadUint32(&cs.isClosed) == 1
}

// Close closes TCP connection.
func (cs *ClientSocket) Close() {
	cs.closeOnce.Do(func() {
		log.Debug().Msgf("Closing TCP client connection #%d", cs.id)

		atomic.StoreUint32(&cs.isClosed, 1)

		if err := cs.connection.Close(); err != nil {
			log.Error().Err(err).Msgf("Error while closing TCP client connection #%d", cs.id)
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
func (cs *ClientSocket) Read(b []byte) (int, error) {
	n, err := cs.reader.Read(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().Msgf("Connection closed by TCP client #%d", cs.id)
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().Err(err).Msgf("Error while reading from TCP client connection #%d", cs.id)
		}

		return n, err
	}

	return n, nil
}

// Write conforms to the io.Writer interface.
func (cs *ClientSocket) Write(b []byte) (int, error) {
	n, err := cs.writer.Write(b)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().Msgf("Connection closed by TCP client #%d", cs.id)
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().Err(err).Msgf("Error while writing to TCP client connection #%d", cs.id)
		}

		return n, err
	}

	return n, nil
}

// SetReadDeadline sets read deadline for underlying socket.
func (cs *ClientSocket) SetReadDeadline(deadline time.Time) error {
	err := cs.connection.SetReadDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().Msgf("Connection closed by TCP client #%d", cs.id)
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().Err(err).Msgf("Error while setting read deadline for TCP client connection #%d", cs.id)
		}

		return err
	}

	return nil
}

// SetWriteDeadline sets read deadline for underlying socket.
func (cs *ClientSocket) SetWriteDeadline(deadline time.Time) error {
	err := cs.connection.SetWriteDeadline(deadline)
	if err != nil {
		if isBrokenPipe(err) {
			log.Debug().Msgf("Connection closed by TCP client #%d", cs.id)
			cs.Close()
		} else if isTimeout(err) {
			// ignore
		} else {
			log.Error().Err(err).Msgf("Error while setting write deadline for TCP client connection #%d", cs.id)
		}

		return err
	}

	return nil
}

// OnClose registers a handler that is called when underlying TCP connection is being closed.
func (cs *ClientSocket) OnClose(handler func()) {
	cs.closeHandlersMutex.Lock()
	defer cs.closeHandlersMutex.Unlock()

	cs.closeHandlers = append(cs.closeHandlers, handler)
}

// Unwrap returns underlying net.Conn instance from ClientSocket.
func (cs *ClientSocket) Unwrap() net.Conn {
	return cs.connection
}

// UnwrapTLS tries to return underlying tls.Conn instance from ClientSocket.
func (cs *ClientSocket) UnwrapTLS() (*tls.Conn, bool) {
	if conn, ok := cs.connection.(*tls.Conn); ok {
		return conn, true
	}

	return nil, false
}

// WrapReader allows to wrap reader object into user defined wrapper.
func (cs *ClientSocket) WrapReader(wrapper func(io.Reader) io.Reader) {
	cs.reader = wrapper(cs.reader)
}

// WrapWriter allows to wrap writer object into user defined wrapper.
func (cs *ClientSocket) WrapWriter(wrapper func(io.Writer) io.Writer) {
	cs.writer = wrapper(cs.writer)
}

// TotalRead returns a total number of bytes read through this socket.
func (cs *ClientSocket) TotalRead() uint64 {
	return cs.byteCountingReader.Total()
}

// ReadsPerSecond returns a total number of bytes read through socket this second.
func (cs *ClientSocket) ReadsPerSecond() uint64 {
	return cs.byteCountingReader.Current()
}

// TotalWritten returns a total number of bytes written through this socket.
func (cs *ClientSocket) TotalWritten() uint64 {
	return cs.byteCountingWriter.Total()
}

// WritesPerSecond returns a total number of bytes written through socket this second.
func (cs *ClientSocket) WritesPerSecond() uint64 {
	return cs.byteCountingWriter.Current()
}

func (cs *ClientSocket) resetMetrics() {
	cs.byteCountingReader.reset()
	cs.byteCountingWriter.reset()
}
