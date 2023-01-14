package tinytcp

import "crypto/tls"

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
	address string

	// Network is a network parameter to pass to net.Listen (default: "tcp").
	Network string

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

// Network is a network parameter to pass to net.Listen.
func Network(network string) ServerOpt {
	return func(config *ServerConfig) {
		config.Network = network
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

		if tlsConfig != nil {
			config.TLSConfig = tlsConfig[0]
		}
	}
}
