package tinytcp

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

const (
	defaultSendBufferSize = 32
	defaultRetryQueueSize = 32
)

// BulkBroadcaster is a high-performance data pump for broadcasting server packets to multiple clients.
// BulkBroadcaster uses an underlying workers pool. Workers are goroutines that constantly fetch data to broadcast
// and are responsible for sending fetched data to their subset of clients, specified as recipients.
// This approach is inefficient for small numbers of recipients but may be useful for some less-common scenarios,
// that involve interacting with huge number of clients.
type BulkBroadcaster struct {
	poolSize int
	workers  []*bulkBroadcasterWorker
}

type bulkBroadcasterWorker struct {
	writeQuantum   time.Duration
	stopChannel    chan struct{}
	messageChannel chan *broadcastMessage
	retryQueue     chan *broadcastMessage
}

type broadcastMessage struct {
	data    []byte
	targets []*ClientSocket
}

// StartBulkBroadcaster creates and starts BulkBroadcaster.
// workersCount is a number of workers to start.
// poolSize is a number of messages to schedule for a single worker in one iteration.
// writeQuantum is the maximum amount of time that should be allocated to Write() operation on a single socket.
func StartBulkBroadcaster(workersCount, poolSize int, writeQuantum time.Duration) *BulkBroadcaster {
	if workersCount < 1 {
		log.Error().Msgf("Invalid value for workersCount: %d", workersCount)
		return nil
	}
	if poolSize < 1 {
		log.Error().Msgf("Invalid value for poolSize: %d", poolSize)
		return nil
	}

	broadcaster := &BulkBroadcaster{
		poolSize: poolSize,
		workers:  []*bulkBroadcasterWorker{},
	}

	for i := 0; i < workersCount; i++ {
		broadcaster.startWorker(writeQuantum)
	}

	return broadcaster
}

// Stop stops all the workers owned by the BulkBroadcaster.
func (b *BulkBroadcaster) Stop() {
	for _, worker := range b.workers {
		worker.stop()
	}
}

// Broadcast schedules data to be written to all the sockets specified by targets array.
// The work will be split between workers for efficient processing.
func (b *BulkBroadcaster) Broadcast(data []byte, targets []*ClientSocket) error {
	if len(b.workers) == 0 {
		return errors.New("no workers in pool")
	}

	left, right := 0, b.poolSize
	currentWorker := 0
	targetsNumber := len(targets)

	for left < targetsNumber {
		if right > targetsNumber {
			right = targetsNumber
		}

		targetsSegment := targets[left:right]

		segmentSize := right - left
		left += segmentSize
		right += segmentSize

		b.workers[currentWorker].send(data, targetsSegment)

		currentWorker++
		if currentWorker >= len(b.workers) {
			currentWorker = 0
		}
	}

	return nil
}

func (b *BulkBroadcaster) startWorker(writeQuantum time.Duration) {
	worker := &bulkBroadcasterWorker{
		writeQuantum:   writeQuantum,
		stopChannel:    make(chan struct{}),
		messageChannel: make(chan *broadcastMessage, defaultSendBufferSize),
		retryQueue:     make(chan *broadcastMessage, defaultRetryQueueSize),
	}

	b.workers = append(b.workers, worker)

	worker.run()
}

func (w *bulkBroadcasterWorker) send(data []byte, targets []*ClientSocket) {
	message := &broadcastMessage{data: data, targets: targets}
	w.messageChannel <- message
}

func (w *bulkBroadcasterWorker) run() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().
					Stack().
					Err(fmt.Errorf("%v", r)).
					Msg("Panic inside BulkBroadcaster worker")
			}
		}()

		for {
			select {
			case <-w.stopChannel:
				return
			case message := <-w.messageChannel:
				w.onMessage(message)
			case message := <-w.retryQueue:
				w.onMessage(message)
			}
		}
	}()
}

func (w *bulkBroadcasterWorker) stop() {
	select {
	case w.stopChannel <- struct{}{}:
	default:
	}
}

func (w *bulkBroadcasterWorker) onMessage(message *broadcastMessage) {
	var err error

	for i, socket := range message.targets {
		err = socket.SetWriteDeadline(time.Now().Add(w.writeQuantum))
		if err != nil {
			continue
		}

		var bytesWritten int
		bytesWritten, err = socket.Write(message.data)

		if socket.IsClosed() {
			continue
		}

		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				w.retryQueue <- &broadcastMessage{
					data:    message.data[bytesWritten:],
					targets: message.targets[i : i+1],
				}

				continue
			}

			log.Error().Err(err).Msg("Error while broadcasting to TCP socket")
			continue
		}
	}
}
