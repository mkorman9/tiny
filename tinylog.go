package tiny

import "github.com/mkorman9/tiny/tinylog"

// SetupLogger configures the global instance of zerolog.Logger.
// Default configuration can be overwritten by providing custom options as arguments.
func SetupLogger(opts ...tinylog.Opt) {
	tinylog.SetupLogger(opts...)
}
