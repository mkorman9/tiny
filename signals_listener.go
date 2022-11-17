package tiny

import (
	"os"
	"os/signal"
)

// SignalHandler represents a handler function for OS signals.
type SignalHandler func(signal os.Signal)

// SignalsListener is a listener of OS signals that implements the Service interface.
type SignalsListener struct {
	signals     []os.Signal
	handler     SignalHandler
	stopChannel chan struct{}
}

// NewSignalsListener creates new SignalsListener.
func NewSignalsListener(signals []os.Signal, handler SignalHandler) *SignalsListener {
	return &SignalsListener{
		signals:     signals,
		handler:     handler,
		stopChannel: make(chan struct{}),
	}
}

// Start implements the interface of Service.
func (s *SignalsListener) Start() error {
	signalsChannel := make(chan os.Signal)
	signal.Notify(signalsChannel, s.signals...)

	for {
		select {
		case <-s.stopChannel:
			return nil
		case receivedSignal := <-signalsChannel:
			if s.handler != nil {
				s.handler(receivedSignal)
			}
		}
	}
}

// Stop implements the interface of Service.
func (s *SignalsListener) Stop() {
	select {
	case s.stopChannel <- struct{}{}:
	default:
	}
}
