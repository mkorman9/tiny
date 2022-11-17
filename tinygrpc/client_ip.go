package tinygrpc

import (
	"context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
	"strings"
)

// GetClientIP resolves the IP address (either v4 or v6) of the client.
// By default, function returns a remote address associated with the socket.
// In case the "x-forwarded-for" header is specified and parseable - the value of this header is returned.
func GetClientIP(ctx context.Context) net.IP {
	p, _ := peer.FromContext(ctx)
	address := p.Addr.(*net.TCPAddr).IP

	md, _ := metadata.FromIncomingContext(ctx)
	if values := md.Get("x-forwarded-for"); values != nil && address.IsPrivate() {
		raw := values[0]
		parts := strings.Split(raw, ",")
		value := strings.TrimSpace(parts[len(parts)-1])

		ip := net.ParseIP(value)
		if ip != nil {
			address = ip
		}
	}

	return address
}
