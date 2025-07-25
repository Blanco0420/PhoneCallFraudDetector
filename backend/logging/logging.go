package logging

import (
	"fmt"
	"os"
	"time"

	"github.com/Blanco0420/Phone-Number-Check/backend/config"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func getLogLevel() zerolog.Level {
	logLevelEnvVar, exists := config.GetEnvVariable("LOG_LEVEL")
	if !exists {
		return zerolog.InfoLevel
	}
	switch logLevelEnvVar {
	case "TRACE":
		return zerolog.TraceLevel
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	default:
		fmt.Fprintf(os.Stderr, "Unknown LOG_LEVEL: %s, defaulting to INFO\n", logLevelEnvVar)
		return zerolog.InfoLevel
	}
}

func init() {
	logger = zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC1123},
	).Level(getLogLevel()).With().Timestamp().Caller().Logger()
}

func Trace() *zerolog.Event {
	return logger.Trace()
}

func Debug() *zerolog.Event {
	return logger.Debug()
}

func Info() *zerolog.Event {
	return logger.Info()
}

func Warn() *zerolog.Event {
	return logger.Warn()
}

func Error() *zerolog.Event {
	return logger.Error()
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func Panic() *zerolog.Event {
	return logger.Panic()
}
