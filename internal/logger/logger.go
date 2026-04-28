package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configures zerolog for JSON output and context-awareness.
func Setup() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.With().Caller().Logger()

	// JSON output in production, pretty in dev
	if os.Getenv("LOG_FORMAT") == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

// Logger returns the global logger instance.
func Logger() zerolog.Logger {
	return log.Logger
}
