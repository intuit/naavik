package logger

import (
	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
)

type LogLevel string

const (
	TraceLevel LogLevel = "trace"
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

func (l LogLevel) String() string {
	return string(l)
}

// Log is a global logger for global context logging.
var Log = NewLogger()

type Logger interface {
	Trace(msg string)

	Tracef(msg string, args ...interface{})

	Debug(msg string)

	Debugf(msg string, args ...interface{})

	Info(msg string)

	Infof(msg string, args ...interface{})

	Warn(msg string)

	Warnf(msg string, args ...interface{})

	Error(msg string)

	Errorf(msg string, args ...interface{})

	Fatal(msg string)

	Fatalf(msg string, args ...interface{})

	Str(key string, value string) Logger

	Int(key string, value int) Logger

	Any(key string, value interface{}) Logger

	Bool(key string, value bool) Logger

	WithStr(key string, value string) Logger

	WithBool(key string, value bool) Logger

	WithInt(key string, value int) Logger

	With(key string, value interface{}) Logger

	// SetLogLevel sets the log level for the logger at global level
	SetLogLevel(level string)

	GetLogLevel() string

	Hook(hook func(level string, msg string)) Logger

	// Error Handler
	HandleError(err error)
}

func NewLogger() Logger {
	if options.GetEnvironment() == types.EnvDev {
		return NewZeroLogger()
	}
	return NewZapLogger()
}
