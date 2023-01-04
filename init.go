package tiny

import "github.com/mkorman9/tiny/tinylog"

// Init initializes global configuration and logger with the default values.
func Init(opts ...tinylog.Opt) {
	LoadConfig()
	tinylog.SetupLogger(opts...)
}
