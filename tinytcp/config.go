package tinytcp

import "crypto/tls"

// ServerConfig holds a configuration for NewServer.
type ServerConfig struct {
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

func mergeServerConfig(provided *ServerConfig) *ServerConfig {
	config := &ServerConfig{
		Network:    "tcp",
		MaxClients: -1,
		TLSConfig:  &tls.Config{},
	}

	if provided == nil {
		return config
	}

	if provided.Network != "" {
		config.Network = provided.Network
	}
	if provided.MaxClients > -1 {
		config.MaxClients = provided.MaxClients
	}
	if provided.TLSCert != "" {
		config.TLSCert = provided.TLSCert
	}
	if provided.TLSKey != "" {
		config.TLSKey = provided.TLSKey
	}
	if provided.TLSConfig != nil {
		config.TLSConfig = provided.TLSConfig
	}

	return config
}
