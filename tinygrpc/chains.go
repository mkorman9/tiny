package tinygrpc

import (
	"context"
	"google.golang.org/grpc"
)

func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		currHandler := handler

		for i := len(interceptors) - 1; i > 0; i-- {
			innerHandler, i := currHandler, i
			currHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptors[i](currentCtx, currentReq, info, innerHandler)
			}
		}

		return interceptors[0](ctx, req, info, currHandler)
	}
}

func chainStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		currHandler := handler

		for i := len(interceptors) - 1; i > 0; i-- {
			innerHandler, i := currHandler, i
			currHandler = func(currentSrv interface{}, currentStream grpc.ServerStream) error {
				return interceptors[i](currentSrv, currentStream, info, innerHandler)
			}
		}

		return interceptors[0](srv, stream, info, currHandler)
	}
}
