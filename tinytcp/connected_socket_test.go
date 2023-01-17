package tinytcp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func TestConnectedSocketInput(t *testing.T) {
	// given
	payload := []byte("Hello world!")
	payloadSize := len(payload)

	in := bytes.NewBuffer(payload)
	socket := MockConnectedSocket(in, io.Discard)

	// when
	buffer := make([]byte, payloadSize)
	n, err := socket.Read(buffer)

	// then
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, payloadSize, n, "n should equal to bytes read")
	assert.Equal(t, payload, buffer, "payloads should match")
}

func TestConnectedSocketInputEOF(t *testing.T) {
	// given
	socket := MockConnectedSocket(&eofReader{}, io.Discard)

	var closeHandlerCalled bool
	socket.OnClose(func() {
		closeHandlerCalled = true
	})

	// when
	_, err := socket.Read(nil)

	// then
	assert.Error(t, io.EOF, err, "err should be equal to io.EOF")
	assert.Truef(t, closeHandlerCalled, "close handler should be called")
}

func TestConnectedSocketOutput(t *testing.T) {
	// given
	payload := []byte("Hello world")
	payloadSize := len(payload)

	var out bytes.Buffer
	socket := MockConnectedSocket(nil, &out)

	// when
	n, err := socket.Write(payload)

	// then
	assert.Nil(t, err, "err should be nil")
	assert.Equal(t, payloadSize, n, "n should equal to bytes read")
	assert.Equal(t, payload, out.Bytes(), "payloads should match")
}

func TestConnectedSocketOutputEOF(t *testing.T) {
	// given
	socket := MockConnectedSocket(nil, &eofWriter{})

	var closeHandlerCalled bool
	socket.OnClose(func() {
		closeHandlerCalled = true
	})

	// when
	_, err := socket.Write(nil)

	// then
	assert.Error(t, io.EOF, err, "err should be equal to io.EOF")
	assert.Truef(t, closeHandlerCalled, "close handler should be called")
}

type eofReader struct {
}

func (er *eofReader) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

type eofWriter struct {
}

func (ew *eofWriter) Write(_ []byte) (int, error) {
	return 0, io.EOF
}
