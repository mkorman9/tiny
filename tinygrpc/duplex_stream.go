package tinygrpc

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DuplexStream simplifies operation on bidirectional gRPC streams.
type DuplexStream[R any, S any] struct {
	stream grpc.ServerStream

	receiveChannel chan *R
	sendChannel    chan *S
	errorChannel   chan error
	exitChannel    chan struct{}
	endHandler     func(error)
}

// DuplexStreamConfig provides a configuration for DuplexStream.
type DuplexStreamConfig struct {
	receiveChannelCapacity int64
	sendChannelCapacity    int64
}

// DuplexStreamOpt is an option to be passed to NewDuplexStream.
type DuplexStreamOpt = func(config *DuplexStreamConfig)

// ReceiveChannelCapacity sets a total capacity (in number of messages) of receive channel's buffer.
func ReceiveChannelCapacity(capacity int64) DuplexStreamOpt {
	return func(config *DuplexStreamConfig) {
		config.receiveChannelCapacity = capacity
	}
}

// SendChannelCapacity sets a total capacity (in number of messages) of send channel's buffer.
func SendChannelCapacity(capacity int64) DuplexStreamOpt {
	return func(config *DuplexStreamConfig) {
		config.sendChannelCapacity = capacity
	}
}

// NewDuplexStream creates new DuplexStream.
func NewDuplexStream[R any, S any](stream grpc.ServerStream, opts ...DuplexStreamOpt) *DuplexStream[R, S] {
	config := DuplexStreamConfig{
		receiveChannelCapacity: 1024,
		sendChannelCapacity:    1024,
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &DuplexStream[R, S]{
		stream:         stream,
		receiveChannel: make(chan *R, config.receiveChannelCapacity),
		sendChannel:    make(chan *S, config.sendChannelCapacity),
		errorChannel:   make(chan error),
		exitChannel:    make(chan struct{}, 4),
	}
}

// Start bootstraps goroutines responsible for handling receive and send channels and blocks until either the server
// (with Stop), or the client interrupts connection.
func (ds *DuplexStream[R, S]) Start() (err error) {
	sendCancelChannel := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				ds.errorChannel <- err
				log.Error().Stack().Err(err).Msg("Panic while receiving gRPC message")
			}
		}()

		for {
			var msg R

			err := ds.stream.RecvMsg(&msg)
			if err != nil {
				ds.errorChannel <- err
				break
			}

			ds.receiveChannel <- &msg
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				ds.errorChannel <- err
				log.Error().Stack().Err(err).Msg("Panic while sending gRPC message")
			}
		}()

		for {
			select {
			case msg := <-ds.sendChannel:
				_ = ds.stream.SendMsg(msg)
			case _ = <-sendCancelChannel:
				return
			}
		}
	}()

	defer func() {
		sendCancelChannel <- struct{}{}
		close(ds.receiveChannel)

		if ds.endHandler != nil {
			ds.endHandler(err)
		}
	}()

	for {
		select {
		case _ = <-ds.errorChannel:
			err = status.Errorf(codes.Canceled, "call cancelled")
			return
		case _ = <-ds.exitChannel:
			return
		}
	}
}

// Stop cancels goroutines responsible for handling receive and send channels and unblocks Start.
func (ds *DuplexStream[R, S]) Stop() {
	ds.exitChannel <- struct{}{}
}

// Send sends a new message to the client.
func (ds *DuplexStream[R, S]) Send(msg *S) {
	ds.sendChannel <- msg
}

// OnReceive specifies a handler for incoming messages.
// The function will call the handler for all incoming messages sequentially, using the same goroutine for each call.
func (ds *DuplexStream[R, S]) OnReceive(handler func(msg *R)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("%v", r)
				ds.errorChannel <- err
				log.Error().Stack().Err(err).Msg("Panic in gRPC DuplexStream handler")
			}
		}()

		for msg := range ds.receiveChannel {
			handler(msg)
		}
	}()
}

// OnEnd specifies a handler for stream end event.
// The handler is called either on stream error or after you call Stop on given stream.
func (ds *DuplexStream[R, S]) OnEnd(handler func(reason error)) {
	ds.endHandler = handler
}
