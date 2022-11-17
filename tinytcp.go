package tiny

import "github.com/mkorman9/tiny/tinytcp"

// NewTCPServer returns new tinytcp.Server instance.
func NewTCPServer(opts ...tinytcp.ServerOpt) *tinytcp.Server {
	return tinytcp.NewServer(opts...)
}
