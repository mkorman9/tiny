package tinytcp

import "crypto/tls"

// ListenMode is a mode in which the server starts listening.
type ListenMode = int

const (
	// Both means listen in both IPv4, IPv6 modes.
	Both ListenMode = iota

	// IPv4_Only means listen in IPv4 mode only.
	IPv4_Only

	// IPv6_Only means listen in IPv6 mode only.
	IPv6_Only
)

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
	// Address is an address to bind server socket to (default: "0.0.0.0:7000").
	Address string

	// Mode is a mode in which the server starts listening - IPv4_Only, IPv6_Only or Both (default: Both).
	Mode ListenMode

	// Max clients denotes the maximum number of connection that can be accepted at once, -1 for no limit (default: -1).
	MaxClients int

	// TLSCert is a path to TLS certificate to use. When specified with TLSKey - enables TLS mode.
	TLSCert string

	// TLSKey is a path to TLS key to use. When specified with TLSCert - enables TLS mode.
	TLSKey string

	// TLSConfig is an optional TLS configuration to pass when using TLS mode.
	TLSConfig *tls.Config
}

// ServerOpt is an option to be specified to NewServer.
type ServerOpt func(*ServerConfig)

// Address is an address to bind server socket to.
func Address(address string) ServerOpt {
	return func(config *ServerConfig) {
		config.Address = address
	}
}

// Mode is a mode in which the server starts listening - IPv4_Only, IPv6_Only or Both.
func Mode(mode ListenMode) ServerOpt {
	return func(config *ServerConfig) {
		config.Mode = mode
	}
}

// MaxClients denotes the maximum number of connection that can be accepted at once, -1 for no limit.
func MaxClients(maxClients int) ServerOpt {
	return func(config *ServerConfig) {
		config.MaxClients = maxClients
	}
}

// TLS enables TLS mode if both cert and key point to valid TLS credentials. tlsConfig is optional.
func TLS(cert, key string, tlsConfig ...*tls.Config) ServerOpt {
	return func(config *ServerConfig) {
		config.TLSCert = cert
		config.TLSKey = key

		if len(tlsConfig) > 0 {
			config.TLSConfig = tlsConfig[0]
		}
	}
}
