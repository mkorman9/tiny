package tinylog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultOutput = os.Stderr

// SetupLogger configures the global instance of zerolog.Logger.
func SetupLogger(config ...*Config) {
	var providedConfig *Config
	if config != nil {
		providedConfig = config[0]
	}
	c := mergeConfig(providedConfig)

	configureSettings(c)
	_ = configureWriters(c)
	configureFields(c)
}

// SetLevel sets global log level.
func SetLevel(level string) error {
	levelValue, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}

	zerolog.SetGlobalLevel(levelValue)

	return nil
}

func configureSettings(config *Config) {
	if err := SetLevel(config.Level); err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

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

	if !config.Console.Disabled {
		writer, err := createFormattedWriter(
			config.Console.Output,
			config.Console.Format,
			config.Console.ColorsDisabled,
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

		writer, err := createFormattedWriter(fileWriter, config.File.Format, true, config.TimeFormat)
		if err != nil {
			_, _ = fmt.Fprintf(config.Console.Output, "Failed to configure file logger: %v\n", err)
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
	if len(config.Fields) != 0 {
		ctx := log.Logger.With()

		for name, value := range config.Fields {
			ctx = ctx.Str(name, value)
		}

		log.Logger = ctx.Logger()
	}
}

func createFormattedWriter(output io.Writer, format string, noColors bool, timeFormat string) (io.Writer, error) {
	if format == LogText {
		formattedOutput := zerolog.ConsoleWriter{
			Out:        output,
			NoColor:    noColors,
			TimeFormat: timeFormat,
		}

		return &formattedOutput, nil
	} else if format == LogJSON {
		return output, nil
	} else {
		return nil, fmt.Errorf("unknown logging format: %v", format)
	}
}

func stackTraceMarshaller(_ error) interface{} {
	var stackTrace []map[string]string

	for i := 3; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)

		stackTrace = append(stackTrace, map[string]string{
			"src":  fmt.Sprintf("%v:%v", file, line),
			"func": fn.Name(),
		})
	}

	return stackTrace
}
