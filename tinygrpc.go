package tiny

import "github.com/mkorman9/tiny/tinygrpc"

// NewGrpcServer create new tinygrpc.Server instance using provided options.
func NewGrpcServer(opts ...tinygrpc.ServerOpt) *tinygrpc.Server {
	return tinygrpc.NewServer(opts...)
}
