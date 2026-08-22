package log

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	level  string
	format string
}

type Fields map[string]interface{}

func NewLogger(level, format string) *Logger {
	return &Logger{level: level, format: format}
}

func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{"debug": 0, "info": 1, "warn": 2, "error": 3}
	return levels[level] >= levels[l.level]
}

func (l *Logger) log(level string, msg string, fields ...Fields) {
	if !l.shouldLog(level) {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level, msg)

	if len(fields) > 0 {
		for k, v := range fields[0] {
			logLine += fmt.Sprintf(" %s=%v", k, v)
		}
	}

	fmt.Fprintln(os.Stdout, logLine)
}

func (l *Logger) Debug(msg string, fields ...Fields) {
	l.log("debug", msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Fields) {
	l.log("info", msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Fields) {
	l.log("warn", msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Fields) {
	l.log("error", msg, fields...)
}

func (l *Logger) WithFields(fields Fields) *Logger {
	return l
}
