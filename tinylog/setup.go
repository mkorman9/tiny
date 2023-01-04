package tinylog

import (
	"fmt"
	"github.com/mattn/go-isatty"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/mkorman9/tiny/tinylog/gelf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultOutput = os.Stderr

// SetupLogger configures the global instance of zerolog.Logger.
// Default configuration can be overwritten by providing custom options as arguments.
func SetupLogger(opts ...Opt) {
	config := &Config{
		Level:      zerolog.InfoLevel,
		TimeFormat: "2006-01-02 15:04:05",
		Console: ConsoleConfig{
			Enabled: true,
			Output:  defaultOutput,
			Colors:  isatty.IsTerminal(os.Stdout.Fd()),
			Format:  LogText,
		},
		File: FileConfig{
			Enabled:   false,
			Location:  "log.txt",
			Format:    LogText,
			FileFlags: os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			FileMode:  0666,
		},
		Gelf: GelfConfig{
			Enabled: false,
			Address: "",
		},
	}

	for _, opt := range opts {
		opt(config)
	}

	configureSettings(config)
	_ = configureWriters(config)
	configureFields(config)
}

func configureSettings(config *Config) {
	zerolog.SetGlobalLevel(config.Level)
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}
	zerolog.TimestampFieldName = "time"
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	zerolog.ErrorStackMarshaler = stackTraceMarshaller
}

func configureWriters(config *Config) error {
	var writers []io.Writer

	if config.Console.Enabled {
		writer, err := createFormattedWriter(
			config.Console.Output,
			config.Console.Format,
			config.Console.Colors,
			config.TimeFormat,
		)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to configure console logger: %v\n", err)
			return err
		}

		writers = append(writers, writer)
	}

	if config.File.Enabled {
		fileWriter, err := os.OpenFile(config.File.Location, config.File.FileFlags, config.File.FileMode)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to open file logger location: %v\n", err)
			return err
		}

		writer, err := createFormattedWriter(fileWriter, config.File.Format, false, config.TimeFormat)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to configure file logger: %v\n", err)
			return err
		}

		writers = append(writers, writer)
	}

	if config.Gelf.Enabled {
		gelfWriter, err := gelf.NewWriter(config.Gelf.Address)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to create gelf logger connection: %v\n", err)
			return err
		}

		writer, err := createFormattedWriter(gelfWriter, LogJSON, false, config.TimeFormat)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to configure gelf logger: %v\n", err)
			return err
		}

		writers = append(writers, writer)
	}

	if len(writers) != 0 {
		log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))
	}

	return nil
}

func configureFields(config *Config) {
	if len(config.fields) != 0 {
		ctx := log.Logger.With()

		for name, value := range config.fields {
			ctx = ctx.Str(name, value)
		}

		log.Logger = ctx.Logger()
	}
}

func createFormattedWriter(output io.Writer, format string, colors bool, timeFormat string) (io.Writer, error) {
	if format == LogText {
		formattedOutput := zerolog.ConsoleWriter{
			Out:        output,
			NoColor:    !colors,
			TimeFormat: timeFormat,
		}

		return &formattedOutput, nil
	} else if format == LogJSON {
		return output, nil
	} else {
		return nil, fmt.Errorf("unknown logging format: %v", format)
	}
}

func stackTraceMarshaller(err error) interface{} {
	var stackTrace []string

	for i := 3; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)

		stackTrace = append(stackTrace, fmt.Sprintf("%v:%v (%v)", file, line, fn.Name()))
	}

	return strings.Join(stackTrace, ", ")
}
