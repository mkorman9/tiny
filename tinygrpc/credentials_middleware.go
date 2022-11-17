package tinygrpc

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"strings"
)

const tokenVerificationResultKey = "tokenVerificationResult"

// TokenVerificationResult represents a result of token verification procedure.
type TokenVerificationResult[T any] struct {
	// IsAuthorized defines whether call has been authorized by the middleware.
	IsAuthorized bool

	// SessionData carries additional information about the session, such as account data.
	// It can be set by TokenVerifierFunc.
	SessionData T
}

// CallMetadata represents additional information about the gRPC call.
type CallMetadata struct {
	// IP is an IP address of the client.
	IP net.IP

	// MethodName is the full name of the method.
	MethodName string
}

// TokenVerifierFunc defines an application-specific procedure for token verification.
type TokenVerifierFunc[T any] func(token string, metadata *CallMetadata) (*TokenVerificationResult[T], error)

// TokenVerifier defines an interface for application-specific procedure for token verification.
type TokenVerifier[T any] interface {
	// Verify defines an application-specific procedure for token verification.
	Verify(token string, metadata *CallMetadata) (*TokenVerificationResult[T], error)
}

// GetTokenVerificationResult returns token verification result for a given call.
func GetTokenVerificationResult[T any](ctx context.Context) *TokenVerificationResult[T] {
	value := ctx.Value(tokenVerificationResultKey)
	if value != nil {
		if result, ok := value.(*TokenVerificationResult[T]); ok {
			return result
		}
	}

	return &TokenVerificationResult[T]{IsAuthorized: false}
}

func authUnaryInterceptor[T any](tokenVerifierFunc TokenVerifierFunc[T]) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		tokenVerificationResult := &TokenVerificationResult[T]{IsAuthorized: false}

		token := retrieveBearerToken(ctx)
		if token != "" {
			result, err := tokenVerifierFunc(token, &CallMetadata{
				IP:         GetClientIP(ctx),
				MethodName: info.FullMethod,
			})
			if err != nil {
				log.Error().Err(err).Msg("Error in token verification function")
				return nil, status.Error(codes.Internal, "internal server error")
			}

			tokenVerificationResult = result
		}

		ctx = context.WithValue(ctx, tokenVerificationResultKey, tokenVerificationResult)

		return handler(ctx, req)
	}
}

func authStreamInterceptor[T any](tokenVerifierFunc TokenVerifierFunc[T]) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		tokenVerificationResult := &TokenVerificationResult[T]{IsAuthorized: false}

		token := retrieveBearerToken(ss.Context())
		if token != "" {
			result, err := tokenVerifierFunc(token, &CallMetadata{
				IP:         GetClientIP(ss.Context()),
				MethodName: info.FullMethod,
			})
			if err != nil {
				log.Error().Err(err).Msg("Error in token verification function")
				return status.Error(codes.Internal, "internal server error")
			}

			tokenVerificationResult = result
		}

		wrappedStream := wrapServerStream(ss)
		wrappedStream.wrappedContext = context.WithValue(
			ss.Context(),
			tokenVerificationResultKey,
			tokenVerificationResult,
		)

		return handler(srv, wrappedStream)
	}
}

func retrieveBearerToken(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		headerValue := md.Get(authorizationHeader)
		if headerValue == nil {
			return ""
		}

		split := strings.SplitN(headerValue[0], " ", 2)
		if len(split) != 2 || !strings.EqualFold(split[0], "bearer") {
			return ""
		}

		return split[1]
	}

	return ""
}
