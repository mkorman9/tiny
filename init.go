package tiny

import "github.com/mkorman9/tiny/tinylog"

// Config hold a configuration for Init().
// It allows the end-user to customize core functionalities, such as global logger or locations of config files.
type Config struct {
	// ConfigFiles specifies a list of files that should be loaded during initialization.
	ConfigFiles []string

	// Log specifies an optional configuration for the global logger.
	Log *tinylog.Config
}

// Init initializes global logger and loads configuration from env variables and specified files.
func Init(config ...*Config) {
	c := &Config{}
	if config != nil {
		c = config[0]
	}

	LoadConfig(c.ConfigFiles...)
	tinylog.SetupLogger(c.Log)
}
