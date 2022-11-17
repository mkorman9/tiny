package tinygrpc

import (
	"google.golang.org/grpc"
)

type serverConfig struct {
	address            string
	grpcOptions        []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
}

// ServerOpt is an option to be specified to NewServer.
type ServerOpt = func(*serverConfig)

// Address is an address to bind to (default: "0.0.0.0:9000").
func Address(address string) ServerOpt {
	return func(serverConfig *serverConfig) {
		serverConfig.address = address
	}
}

// ServerOptions allows to specify custom grpc.ServerOption options.
func ServerOptions(opts ...grpc.ServerOption) ServerOpt {
	return func(serverConfig *serverConfig) {
		for _, opt := range opts {
			serverConfig.grpcOptions = append(serverConfig.grpcOptions, opt)
		}
	}
}

// UnaryInterceptor adds specified interceptor to the tail of unary interceptors chains.
func UnaryInterceptor(interceptor grpc.UnaryServerInterceptor) ServerOpt {
	return func(serverConfig *serverConfig) {
		serverConfig.unaryInterceptors = append(
			serverConfig.unaryInterceptors,
			interceptor,
		)
	}
}

// StreamInterceptor adds specified interceptor to the tail of stream interceptors chains.
func StreamInterceptor(interceptor grpc.StreamServerInterceptor) ServerOpt {
	return func(serverConfig *serverConfig) {
		serverConfig.streamInterceptors = append(
			serverConfig.streamInterceptors,
			interceptor,
		)
	}
}

// EnableAuthMiddlewareFunc makes server use token-based authorization based on passed TokenVerifierFunc.
func EnableAuthMiddlewareFunc[T any](verifierFunc TokenVerifierFunc[T]) ServerOpt {
	return func(serverConfig *serverConfig) {
		UnaryInterceptor(authUnaryInterceptor(verifierFunc))(serverConfig)
		StreamInterceptor(authStreamInterceptor(verifierFunc))(serverConfig)
	}
}

// EnableAuthMiddleware makes server use token-based authorization based on passed TokenVerifier.
func EnableAuthMiddleware[T any](verifier TokenVerifier[T]) ServerOpt {
	return EnableAuthMiddlewareFunc(verifier.Verify)
}
