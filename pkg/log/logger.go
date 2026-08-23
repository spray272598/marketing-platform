package log

import (
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	slog *slog.Logger
}

type Fields map[string]interface{}

func NewLogger(level, format string) *Logger {
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	})

	return &Logger{
		slog: slog.New(handler),
	}
}

func (l *Logger) toSlogArgs(fields ...Fields) []any {
	if len(fields) > 0 {
		args := make([]any, 0)
		for k, v := range fields[0] {
			args = append(args, k, v)
		}
		return args
	}
	return nil
}

func (l *Logger) Debug(msg string, fields ...Fields) {
	args := l.toSlogArgs(fields...)
	l.slog.Debug(msg, args...)
}

func (l *Logger) Info(msg string, fields ...Fields) {
	args := l.toSlogArgs(fields...)
	l.slog.Info(msg, args...)
}

func (l *Logger) Warn(msg string, fields ...Fields) {
	args := l.toSlogArgs(fields...)
	l.slog.Warn(msg, args...)
}

func (l *Logger) Error(msg string, fields ...Fields) {
	args := l.toSlogArgs(fields...)
	l.slog.Error(msg, args...)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	args := make([]any, 0)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{slog: l.slog.With(args...)}
}

func (l *Logger) Slog() *slog.Logger {
	return l.slog
}
