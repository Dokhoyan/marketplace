package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger = zap.NewNop()

func Init(core zapcore.Core, options ...zap.Option) {
	globalLogger = zap.New(core, options...)
}

func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}

func Debugf(format string, args ...interface{}) {
	globalLogger.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...interface{}) {
	globalLogger.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}) {
	globalLogger.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}) {
	globalLogger.Error(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatal(fmt.Sprintf(format, args...))
}
