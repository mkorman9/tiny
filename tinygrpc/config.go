package tinygrpc

import (
	"google.golang.org/grpc"
)

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
	grpcOptions        []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// ServerOpt is an option to be specified to NewServer.
type ServerOpt = func(*ServerConfig)

// ServerOptions allows to specify custom grpc.ServerOption options.
func ServerOptions(opts ...grpc.ServerOption) ServerOpt {
	return func(serverConfig *ServerConfig) {
		for _, opt := range opts {
			serverConfig.grpcOptions = append(serverConfig.grpcOptions, opt)
		}
	}
}

// UnaryInterceptor adds specified interceptor to the tail of unary interceptors chains.
func UnaryInterceptor(interceptor grpc.UnaryServerInterceptor) ServerOpt {
	return func(serverConfig *ServerConfig) {
		serverConfig.unaryInterceptors = append(
			serverConfig.unaryInterceptors,
			interceptor,
		)
	}
}

// StreamInterceptor adds specified interceptor to the tail of stream interceptors chains.
func StreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerOpt {
	return func(serverConfig *ServerConfig) {
		serverConfig.streamInterceptors = append(
			serverConfig.streamInterceptors,
			interceptor,
		)
	}
}

// EnableAuthMiddlewareFunc makes server use token-based authorization based on passed TokenVerifierFunc.
func EnableAuthMiddlewareFunc[T any](verifierFunc TokenVerifierFunc[T]) ServerOpt {
	return func(serverConfig *ServerConfig) {
		UnaryInterceptor(authUnaryInterceptor(verifierFunc))(serverConfig)
		StreamInterceptor(authStreamInterceptor(verifierFunc))(serverConfig)
	}
}

// EnableAuthMiddleware makes server use token-based authorization based on passed TokenVerifier.
func EnableAuthMiddleware[T any](verifier TokenVerifier[T]) ServerOpt {
	return EnableAuthMiddlewareFunc(verifier.Verify)
}
