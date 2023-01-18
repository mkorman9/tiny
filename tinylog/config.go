package tinylog

import (
	"github.com/rs/zerolog"
	"io"
	"os"
)

// LogFormat is a logging format (text or json).
type LogFormat = string

const (
	// LogText is a plaintext output.
	LogText LogFormat = "text"

	// LogJSON is a JSON-formatted output.
	LogJSON LogFormat = "json"
)

// Config represents a configuration of the global logger.
type Config struct {
	// Level is a log level to enable (default: InfoLevel).
	Level zerolog.Level

	// TimeFormat specifies time format to use (default: "2006-01-02 15:04:05")
	TimeFormat string

	// Console is an instance of ConsoleConfig.
	Console ConsoleConfig

	// File is an instance of FileConfig.
	File FileConfig

	// Fields is a set of fields to include in each log line.
	Fields map[string]string
}

// ConsoleConfig represents a configuration for console output. This output is emitted to os.Stderr.
type ConsoleConfig struct {
	// Enabled decides whether this output should be enabled or not (default: true).
	Enabled bool

	// Output is a writer to write logs to (default: os.Stderr).
	Output io.Writer

	// Colors decides whether logging output should be colored or not (default: true for interactive terminals).
	Colors bool

	// Format is a format of this output. It could be either LogText or LogJSON (default: LogText).
	Format LogFormat
}

// FileConfig represents a configuration for file output. This output is emitted to a file.
type FileConfig struct {
	// Enabled decides whether this output should be enabled or not (default: false).
	Enabled bool

	// Location is a path to the output file (default: "log.txt").
	Location string

	// FileFlags specifies what flags to use when opening file (default: os.O_WRONLY | os.O_CREATE | os.O_APPEND).
	FileFlags int

	// FileMode specifies what mode to use when opening file (default: 0666).
	FileMode os.FileMode

	// Format is a format of this output. It could be either LogText or LogJSON (default: LogText).
	Format LogFormat
}

// Opt is an option to be specified to SetupLogger.
type Opt func(*Config)

// Level sets logging level to enable.
func Level(level string) Opt {
	return func(config *Config) {
		levelParsed, err := zerolog.ParseLevel(level)
		if err == nil {
			config.Level = levelParsed
		}
	}
}

// LevelDebug sets logging level to DebugLevel
func LevelDebug() Opt {
	return func(config *Config) {
		config.Level = zerolog.DebugLevel
	}
}

// LevelInfo sets logging level to InfoLevel
func LevelInfo() Opt {
	return func(config *Config) {
		config.Level = zerolog.InfoLevel
	}
}

// LevelWarn sets logging level to WarnLevel
func LevelWarn() Opt {
	return func(config *Config) {
		config.Level = zerolog.WarnLevel
	}
}

// LevelError sets logging level to ErrorLevel
func LevelError() Opt {
	return func(config *Config) {
		config.Level = zerolog.ErrorLevel
	}
}

// LevelFatal sets logging level to FatalLevel
func LevelFatal() Opt {
	return func(config *Config) {
		config.Level = zerolog.FatalLevel
	}
}

// TimeFormat specifies time format to use
func TimeFormat(format string) Opt {
	return func(config *Config) {
		config.TimeFormat = format
	}
}

// Fields adds custom fields to the logger.
func Fields(fields map[string]string) Opt {
	return func(config *Config) {
		if config.Fields == nil {
			config.Fields = make(map[string]string)
		}

		for key, value := range fields {
			config.Fields[key] = value
		}
	}
}

// Field adds a custom field to the logger.
func Field(name, value string) Opt {
	return func(config *Config) {
		if config.Fields == nil {
			config.Fields = make(map[string]string)
		}

		config.Fields[name] = value
	}
}

// ConsoleEnabled sets Enabled parameter of the console output.
func ConsoleEnabled(enabled bool) Opt {
	return func(config *Config) {
		config.Console.Enabled = enabled
	}
}

// ConsoleOutput sets the writer to write logs to.
func ConsoleOutput(output io.Writer) Opt {
	return func(config *Config) {
		config.Console.Output = output
	}
}

// ConsoleColors sets Colors parameter of the console output.
func ConsoleColors(colors bool) Opt {
	return func(config *Config) {
		config.Console.Colors = colors
	}
}

// ConsoleFormat sets Format parameter of the console output.
func ConsoleFormat(format LogFormat) Opt {
	return func(config *Config) {
		config.Console.Format = format
	}
}

// FileEnabled sets Enabled parameter of the file output.
func FileEnabled(enabled bool) Opt {
	return func(config *Config) {
		config.File.Enabled = enabled
	}
}

// FileLocation sets Location parameter of the file output.
func FileLocation(location string) Opt {
	return func(config *Config) {
		config.File.Location = location
	}
}

// FileFlags specifies what flags to use when opening file.
func FileFlags(flags int) Opt {
	return func(config *Config) {
		config.File.FileFlags = flags
	}
}

// FileMode specifies what mode to use when opening file.
func FileMode(mode os.FileMode) Opt {
	return func(config *Config) {
		config.File.FileMode = mode
	}
}

// FileFormat sets Format parameter of the file output.
func FileFormat(format LogFormat) Opt {
	return func(config *Config) {
		config.File.Format = format
	}
}
