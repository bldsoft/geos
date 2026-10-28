package middleware

import (
	"fmt"

	"github.com/bldsoft/gost/log"
	"github.com/rs/zerolog"
)

var Logger = NewLoggerV2(&log.Logger)

type LoggerV2 struct {
	logger *log.ServiceLogger
}

func NewLoggerV2(logger *log.ServiceLogger) *LoggerV2 {
	return &LoggerV2{logger: logger}
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (l *LoggerV2) Info(args ...any) {
	l.logger.Infof(fmt.Sprint(args...))
}

// Infoln logs to INFO log. Arguments are handled in the manner of fmt.Println.
func (l *LoggerV2) Infoln(args ...any) {
	l.Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (l *LoggerV2) Infof(format string, args ...any) {
	l.logger.Infof(format, args...)
}

// Warning logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (l *LoggerV2) Warning(args ...any) {
	l.logger.Warn(fmt.Sprint(args...))
}

// Warningln logs to WARNING log. Arguments are handled in the manner of fmt.Println.
func (l *LoggerV2) Warningln(args ...any) {
	l.Warning(args...)
}

// Warningf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (l *LoggerV2) Warningf(format string, args ...any) {
	l.logger.Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (l *LoggerV2) Error(args ...any) {
	l.logger.Error(fmt.Sprint(args...))
}

// Errorln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
func (l *LoggerV2) Errorln(args ...any) {
	l.Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (l *LoggerV2) Errorf(format string, args ...any) {
	l.logger.Errorf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *LoggerV2) Fatal(args ...any) {
	l.logger.Fatal(fmt.Sprint(args...))
}

// Fatalln logs to ERROR log. Arguments are handled in the manner of fmt.Println.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *LoggerV2) Fatalln(args ...any) {
	l.Fatal(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
// gRPC ensures that all Fatal logs will exit with os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func (l *LoggerV2) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}

// V reports whether verbosity level l is at least the requested verbose level.
func (l *LoggerV2) V(level int) bool {
	switch level {
	case 0:
		return zerolog.InfoLevel <= zerolog.GlobalLevel()
	case 1:
		return zerolog.WarnLevel <= zerolog.GlobalLevel()
	case 2:
		return zerolog.ErrorLevel <= zerolog.GlobalLevel()
	case 3:
		return zerolog.FatalLevel <= zerolog.GlobalLevel()
	default:
		panic("unhandled gRPC logger level")
	}
}
