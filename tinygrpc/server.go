package tinygrpc

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
)

// Server is an object representing grpc.Server and implementing the tiny.Service interface.
type Server struct {
	*grpc.Server

	address string
}

// NewServer create new Server using global configuration and provided options.
func NewServer(opts ...ServerOpt) *Server {
	serverConfig := serverConfig{
		address: "0.0.0.0:9000",
	}

	for _, opt := range opts {
		opt(&serverConfig)
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{recoveryUnaryInterceptor}
	unaryInterceptors = append(unaryInterceptors, serverConfig.unaryInterceptors...)

	streamInterceptors := []grpc.StreamServerInterceptor{recoveryStreamInterceptor}
	streamInterceptors = append(streamInterceptors, serverConfig.streamInterceptors...)

	grpcOptions := serverConfig.grpcOptions
	grpcOptions = append(grpcOptions, grpc.UnaryInterceptor(chainUnaryInterceptors(unaryInterceptors...)))
	grpcOptions = append(grpcOptions, grpc.StreamInterceptor(chainStreamInterceptors(streamInterceptors...)))

	return &Server{
		Server:  grpc.NewServer(grpcOptions...),
		address: serverConfig.address,
	}
}

// Start implements the interface of tiny.Service.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	log.Info().Msgf("gRPC server started (%s)", s.address)

	return s.Serve(listener)
}

// Stop implements the interface of tiny.Service.
func (s *Server) Stop() {
	s.GracefulStop()
	log.Info().Msgf("gRPC server stopped (%s)", s.address)
}
