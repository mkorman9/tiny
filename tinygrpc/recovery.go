package tinygrpc

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func recoveryUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (_ interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Stack().
				Err(fmt.Errorf("%v", r)).
				Msgf("Panic inside gRPC function %s", info.FullMethod)

			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	resp, err := handler(ctx, req)
	return resp, err
}

func recoveryStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Stack().
				Err(fmt.Errorf("%v", r)).
				Msgf("Panic inside gRPC function %s", info.FullMethod)

			err = status.Error(codes.Internal, "internal server error")
		}
	}()

	err = handler(srv, ss)
	return err
}
