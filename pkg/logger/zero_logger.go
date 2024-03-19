package logger

//nolint:depguard
import (
	"context"
	"io"
	"os"
	"time"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	logger zerolog.Logger
	ctx    context.Context
}

func NewZeroLogger() Logger {
	var writer io.Writer = os.Stdout
	if options.GetEnvironment() == types.EnvDev {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    !options.GetLogColor(),
			TimeFormat: time.RFC3339,
		}
	}

	return &ZeroLogger{logger: zerolog.New(writer).With().Timestamp().Logger(), ctx: context.Background()}
}

func (z *ZeroLogger) Trace(msg string) {
	z.logger.Trace().Msg(msg)
}

func (z *ZeroLogger) Tracef(msg string, args ...interface{}) {
	z.logger.Trace().Msgf(msg, args...)
}

func (z *ZeroLogger) Debug(msg string) {
	z.logger.Debug().Msg(msg)
}

func (z *ZeroLogger) Debugf(msg string, args ...interface{}) {
	z.logger.Debug().Msgf(msg, args...)
}

func (z *ZeroLogger) Info(msg string) {
	z.logger.Info().Msg(msg)
}

func (z *ZeroLogger) Infof(msg string, args ...interface{}) {
	z.logger.Info().Msgf(msg, args...)
}

func (z *ZeroLogger) Warn(msg string) {
	z.logger.Warn().Msg(msg)
}

func (z *ZeroLogger) Warnf(msg string, args ...interface{}) {
	z.logger.Warn().Msgf(msg, args...)
}

func (z *ZeroLogger) Error(msg string) {
	z.logger.Error().Msg(msg)
}

func (z *ZeroLogger) Errorf(msg string, args ...interface{}) {
	z.logger.Error().Msgf(msg, args...)
}

func (z *ZeroLogger) Fatal(msg string) {
	z.logger.Panic().Msg(msg)
}

func (z *ZeroLogger) Fatalf(msg string, args ...interface{}) {
	z.logger.Panic().Msgf(msg, args...)
}

func (z *ZeroLogger) HandleError(err error) {
	z.logger.Error().Msg(err.Error())
}

func (z *ZeroLogger) Str(key string, value string) Logger {
	return &ZeroLogger{logger: z.logger.With().Str(key, value).Logger(), ctx: z.ctx}
}

func (z *ZeroLogger) Int(key string, value int) Logger {
	return &ZeroLogger{logger: z.logger.With().Int(key, value).Logger(), ctx: z.ctx}
}

func (z *ZeroLogger) Bool(key string, value bool) Logger {
	return &ZeroLogger{logger: z.logger.With().Bool(key, value).Logger(), ctx: z.ctx}
}

func (z *ZeroLogger) Any(key string, value interface{}) Logger {
	return &ZeroLogger{logger: z.logger.With().Any(key, value).Logger(), ctx: z.ctx}
}

func (z *ZeroLogger) With(key string, value interface{}) Logger {
	z.logger = z.logger.With().Any(key, value).Logger()
	return z
}

func (z *ZeroLogger) WithStr(key string, value string) Logger {
	z.logger = z.logger.With().Str(key, value).Logger()
	return z
}

func (z *ZeroLogger) WithInt(key string, value int) Logger {
	z.logger = z.logger.With().Int(key, value).Logger()
	return z
}

func (z *ZeroLogger) WithBool(key string, value bool) Logger {
	z.logger = z.logger.With().Bool(key, value).Logger()
	return z
}

func (z *ZeroLogger) GetLogLevel() string {
	return zerolog.GlobalLevel().String()
}

// Hook wrapper implementation.
type zerologHook struct {
	hookFunc func(level string, msg string)
}

func (z *zerologHook) Run(_ *zerolog.Event, level zerolog.Level, msg string) {
	z.hookFunc(level.String(), msg)
}

func (z *ZeroLogger) Hook(hookFunc func(level string, msg string)) Logger {
	z.logger = z.logger.Hook(&zerologHook{hookFunc: hookFunc})
	return z
}

func (z *ZeroLogger) SetLogLevel(level string) {
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		break
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		break
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		break
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		break
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		break
	default:
		return
	}
}
