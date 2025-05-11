package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
}

type ConfigLogger struct {
	Level string
}

const (
	logCtxKey = "zerrologer-ctx"
)

var (
	logger zerolog.Logger
)

func init() {

	l := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	SetupLogger(l)
}

func NewZerologLogger(cfg ConfigLogger) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).Level(level).With().Timestamp().Logger()

	return logger
}

func SetupLogger(l zerolog.Logger) {
	logger = l
}

func ToContext(ctx context.Context, l zerolog.Logger) context.Context {
	return context.WithValue(ctx, logCtxKey, l)
}

func Infof(ctx context.Context, msg string, v ...interface{}) {
	if l, ok := ctx.Value(logCtxKey).(zerolog.Logger); ok {
		l.Info().Msgf(msg, v...)
		return
	}
	logger.Info().Msgf(msg, v...)
}

func Errorf(ctx context.Context, msg string, v ...interface{}) {
	if l, ok := ctx.Value(logCtxKey).(zerolog.Logger); ok {
		l.Error().Msgf(msg, v...)
		return
	}
	logger.Error().Msgf(msg, v...)
}

func Debugf(ctx context.Context, msg string, v ...interface{}) {
	if l, ok := ctx.Value(logCtxKey).(zerolog.Logger); ok {
		l.Debug().Msgf(msg, v...)
		return
	}
	logger.Debug().Msgf(msg, v...)
}
