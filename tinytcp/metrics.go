package tinytcp

import (
	"io"
	"sync/atomic"
)

// ServerMetrics contains basic metrics gathered from TCP server.
type ServerMetrics struct {
	// TotalRead is total number of bytes read by server since start.
	TotalRead uint64

	// TotalRead is total number of bytes written by server since start.
	TotalWritten uint64

	// ReadsPerSecond is total number of bytes read by server last second.
	ReadsPerSecond uint64

	// ReadsPerSecond is total number of bytes written by server last second.
	WritesPerSecond uint64

	// Connections is total number of clients connected during last second.
	Connections int

	// MaxConnections is maximum number of clients connected at a single time.
	MaxConnections int

	// Goroutines is total number of goroutines active during last second.
	Goroutines int

	// MaxGoroutines is maximum number of goroutines active at a single time.
	MaxGoroutines int
}

type byteCountingReader struct {
	reader       io.Reader
	totalBytes   uint64
	currentBytes uint64
}

func (r *byteCountingReader) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)

	if n > 0 {
		atomic.AddUint64(&r.totalBytes, uint64(n))
		atomic.AddUint64(&r.currentBytes, uint64(n))
	}

	return n, err
}

func (r *byteCountingReader) Total() uint64 {
	return atomic.LoadUint64(&r.totalBytes)
}

func (r *byteCountingReader) Current() uint64 {
	return atomic.LoadUint64(&r.currentBytes)
}

func (r *byteCountingReader) reset() {
	atomic.StoreUint64(&r.currentBytes, 0)
}

type byteCountingWriter struct {
	writer       io.Writer
	totalBytes   uint64
	currentBytes uint64
}

func (w *byteCountingWriter) Write(b []byte) (int, error) {
	n, err := w.writer.Write(b)

	if n > 0 {
		atomic.AddUint64(&w.totalBytes, uint64(n))
		atomic.AddUint64(&w.currentBytes, uint64(n))
	}

	return n, err
}

func (w *byteCountingWriter) Total() uint64 {
	return atomic.LoadUint64(&w.totalBytes)
}

func (w *byteCountingWriter) Current() uint64 {
	return atomic.LoadUint64(&w.currentBytes)
}

func (w *byteCountingWriter) reset() {
	atomic.StoreUint64(&w.currentBytes, 0)
}
