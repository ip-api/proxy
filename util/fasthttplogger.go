package util

import (
	"github.com/rs/zerolog"
)

type FasthttpLogger struct {
	Logger zerolog.Logger
}

func (f FasthttpLogger) Printf(format string, args ...interface{}) {
	f.Logger.Info().Msgf(format, args...)
}
