package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(isDevelopment bool) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if isDevelopment {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		log.Logger = zerolog.New(os.Stderr).
			With().
			Timestamp().
			Caller().
			Logger()
	}
}
