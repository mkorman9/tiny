package tinylog

import (
	"github.com/mattn/go-isatty"
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
	// Level is a log level to enable (default: info).
	Level string

	// TimeFormat specifies time format to use (default: "2006-01-02 15:04:05")
	TimeFormat string

	// Console is an instance of ConsoleConfig.
	Console *ConsoleConfig

	// File is an instance of FileConfig.
	File *FileConfig

	// Fields is a set of fields to include in each log line.
	Fields map[string]string
}

// ConsoleConfig represents a configuration for console output. This output is emitted to os.Stderr.
type ConsoleConfig struct {
	// Disabled decides whether this output should be disabled or not (default: false).
	Disabled bool

	// Output is a writer to write logs to (default: os.Stderr).
	Output io.Writer

	// ColorsDisabled decides whether logging output should be colored or not.
	// (default: false for interactive terminals, true for others).
	ColorsDisabled bool

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

func mergeConfig(provided *Config) *Config {
	config := &Config{
		Level:      "info",
		TimeFormat: "2006-01-02 15:04:05",
		Console: &ConsoleConfig{
			Disabled:       false,
			Output:         defaultOutput,
			ColorsDisabled: !isatty.IsTerminal(os.Stdout.Fd()),
			Format:         LogText,
		},
		File: &FileConfig{
			Enabled:   false,
			Location:  "log.txt",
			Format:    LogText,
			FileFlags: os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			FileMode:  0666,
		},
	}

	if provided == nil {
		return config
	}

	if provided.Level != "" {
		config.Level = provided.Level
	}
	if provided.TimeFormat != "" {
		config.TimeFormat = provided.TimeFormat
	}
	if provided.Console != nil {
		if provided.Console.Disabled {
			config.Console.Disabled = true
		}
		if provided.Console.Output != nil {
			config.Console.Output = provided.Console.Output
		}
		if provided.Console.ColorsDisabled {
			config.Console.ColorsDisabled = true
		}
		if provided.Console.Format != "" {
			config.Console.Format = provided.Console.Format
		}
	}
	if provided.File != nil {
		if provided.File.Enabled {
			config.File.Enabled = true
		}
		if provided.File.Location != "" {
			config.File.Location = provided.File.Location
		}
		if provided.File.Format != "" {
			config.File.Format = provided.File.Format
		}
		if provided.File.FileFlags != 0 {
			config.File.FileFlags = provided.File.FileFlags
		}
		if provided.File.FileMode != 0 {
			config.File.FileMode = provided.File.FileMode
		}
	}

	return config
}
