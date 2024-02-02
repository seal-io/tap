package logging

import (
	"io"
	"log"

	hclog "github.com/hashicorp/go-hclog"
)

func Log(level hclog.Level, msg string, args ...any) {
	lg.Log(level, msg, args...)
}

func Trace(msg string, args ...any) {
	lg.Trace(msg, args...)
}

func Debug(msg string, args ...any) {
	lg.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	lg.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	lg.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	lg.Error(msg, args...)
}

func IsTrace() bool {
	return lg.IsTrace()
}

func IsDebug() bool {
	return lg.IsDebug()
}

func IsInfo() bool {
	return lg.IsInfo()
}

func IsWarn() bool {
	return lg.IsWarn()
}

func IsError() bool {
	return lg.IsError()
}

func With(args ...any) hclog.Logger {
	return lg.With(args...)
}

func Named(name string) hclog.Logger {
	return lg.Named(name)
}

func SetLevel(level hclog.Level) {
	lg.SetLevel(level)
}

func GetLevel() hclog.Level {
	return lg.GetLevel()
}

func StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return lg.StandardLogger(opts)
}

func StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return lg.StandardWriter(opts)
}
