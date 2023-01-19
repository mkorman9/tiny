package tinytcp

import (
	"io"
	"sync/atomic"
	"time"
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
	reader  io.Reader
	total   uint64
	current uint64
	rate    uint64
}

func (r *byteCountingReader) Read(b []byte) (int, error) {
	n, err := r.reader.Read(b)

	if n > 0 {
		atomic.AddUint64(&r.current, uint64(n))
	}

	return n, err
}

func (r *byteCountingReader) Total() uint64 {
	return atomic.LoadUint64(&r.total)
}

func (r *byteCountingReader) PerSecond() uint64 {
	return atomic.LoadUint64(&r.rate)
}

func (r *byteCountingReader) update(interval time.Duration) uint64 {
	current := atomic.SwapUint64(&r.current, 0)

	atomic.StoreUint64(&r.rate, uint64(float64(current)/interval.Seconds()))
	atomic.AddUint64(&r.total, current)

	return current
}

func (r *byteCountingReader) reset() {
	r.reader = nil
	r.total = 0
	r.current = 0
	r.rate = 0
}

type byteCountingWriter struct {
	writer  io.Writer
	total   uint64
	current uint64
	rate    uint64
}

func (w *byteCountingWriter) Write(b []byte) (int, error) {
	n, err := w.writer.Write(b)

	if n > 0 {
		atomic.AddUint64(&w.current, uint64(n))
	}

	return n, err
}

func (w *byteCountingWriter) Total() uint64 {
	return atomic.LoadUint64(&w.total)
}

func (w *byteCountingWriter) PerSecond() uint64 {
	return atomic.LoadUint64(&w.rate)
}

func (w *byteCountingWriter) update(interval time.Duration) uint64 {
	current := atomic.SwapUint64(&w.current, 0)

	atomic.StoreUint64(&w.rate, uint64(float64(current)/interval.Seconds()))
	atomic.AddUint64(&w.total, current)

	return current
}

func (w *byteCountingWriter) reset() {
	w.writer = nil
	w.total = 0
	w.current = 0
	w.rate = 0
}
