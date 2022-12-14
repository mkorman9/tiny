package tiny

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var shutdownSignals = []os.Signal{syscall.SIGINT, syscall.SIGTERM}

// Service represents concurrent job, that is expected to run in background for the whole lifetime of the process.
// Typical implementations of Service include network servers, such as HTTP or gRPC servers.
type Service interface {
	// Start is expected to start execution of the service and block.
	// If the execution cannot be started, or it fails abruptly, it should return a non-nil error.
	Start() error

	// Stop is expected to stop the running service gracefully and unblock the thread used by Start function.
	Stop()
}

// StartAndBlock starts all passed services in their designated goroutines and then blocks the current thread.
// Thread is unblocked when the process receives SIGINT or SIGTERM signals or one of the Start() functions returns an error.
// When exiting, StartAndBlock gracefully stops all the services by calling their Stop() functions and waiting for them to exit.
func StartAndBlock(services ...Service) {
	errorChannel := make(chan error)

	for _, service := range services {
		s := service

		go func() {
			defer func() {
				if r := recover(); r != nil {
					select {
					case errorChannel <- fmt.Errorf("%v", r):
					default:
					}
				}
			}()

			if err := s.Start(); err != nil {
				select {
				case errorChannel <- err:
				default:
				}
			}
		}()
	}

	defer func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(services))

		for _, service := range services {
			s := service

			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().
							Stack().
							Err(fmt.Errorf("%v", r)).
							Msg("Panic while stopping service")
					}

					wg.Done()
				}()

				s.Stop()
			}()
		}

		wg.Wait()
	}()

	blockThread(errorChannel)
}

func blockThread(errorChannel <-chan error) {
	shutdownSignalsChannel := make(chan os.Signal)
	signal.Notify(shutdownSignalsChannel, shutdownSignals...)

	for {
		select {
		case err := <-errorChannel:
			log.Error().Err(err).Msg("Unblocking thread due to an error")
			return
		case s := <-shutdownSignalsChannel:
			log.Info().Msgf("Unblocking thread due to a signal: %v", s)
			return
		}
	}
}
