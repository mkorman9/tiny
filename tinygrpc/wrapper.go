package tinygrpc

import (
	"context"
	"google.golang.org/grpc"
)

type wrappedServerStream struct {
	grpc.ServerStream

	wrappedContext context.Context
}

func (wss *wrappedServerStream) Context() context.Context {
	return wss.wrappedContext
}

func wrapServerStream(stream grpc.ServerStream) *wrappedServerStream {
	if existing, ok := stream.(*wrappedServerStream); ok {
		return existing
	}

	return &wrappedServerStream{ServerStream: stream, wrappedContext: stream.Context()}
}
