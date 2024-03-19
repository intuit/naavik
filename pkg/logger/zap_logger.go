package logger

//nolint:depguard
import (
	"context"

	"github.com/intuit/naavik/cmd/options"
	"github.com/intuit/naavik/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
	ctx    context.Context
}

var atomicLevel = zap.NewAtomicLevelAt(zap.InfoLevel)

func GetEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		NameKey:        "logger",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	}
	if options.GetLogColor() {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if options.GetEnvironment() == types.EnvDev {
		encoderConfig.LevelKey = "level"
		return encoderConfig
	}
	if atomicLevel.Level().String() > TraceLevel.String() {
		encoderConfig.StacktraceKey = "stacktrace"
		return encoderConfig
	}
	return encoderConfig
}

func IsDevelopment() bool {
	return options.GetEnvironment() == types.EnvDev
}

func EncoderType() string {
	if options.GetEnvironment() == types.EnvDev {
		return "console"
	}
	return "json"
}

func NewZapLogger() Logger {
	config := zap.Config{
		Level:            atomicLevel,
		Development:      IsDevelopment(),
		Encoding:         EncoderType(),
		EncoderConfig:    GetEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := config.Build()
	if err != nil {
		panic("initializing logger failed. " + err.Error())
	}

	return &ZapLogger{logger: logger, ctx: context.Background()}
}

func (l *ZapLogger) Trace(msg string) {
	l.logger.With(zap.String(LevelKey, TraceLevel.String())).Debug(msg)
}

func (l *ZapLogger) Tracef(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, TraceLevel.String())).Debugf(msg, args...)
}

func (l *ZapLogger) Debug(msg string) {
	l.logger.With(zap.String(LevelKey, DebugLevel.String())).Debug(msg)
}

func (l *ZapLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, DebugLevel.String())).Debugf(msg, args)
}

func (l *ZapLogger) Info(msg string) {
	l.logger.With(zap.String(LevelKey, InfoLevel.String())).Info(msg)
}

func (l *ZapLogger) Infof(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, InfoLevel.String())).Infof(msg, args...)
}

func (l *ZapLogger) Warn(msg string) {
	l.logger.With(zap.String(LevelKey, WarnLevel.String())).Warn(msg)
}

func (l *ZapLogger) Warnf(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, WarnLevel.String())).Warnf(msg, args...)
}

func (l *ZapLogger) Error(msg string) {
	l.logger.With(zap.String(LevelKey, ErrorLevel.String())).Error(msg)
}

func (l *ZapLogger) Errorf(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, ErrorLevel.String())).Errorf(msg, args...)
}

func (l *ZapLogger) Fatal(msg string) {
	l.logger.With(zap.String(LevelKey, FatalLevel.String())).Panic(msg)
}

func (l *ZapLogger) Fatalf(msg string, args ...interface{}) {
	l.logger.Sugar().With(zap.String(LevelKey, FatalLevel.String())).Panicf(msg, args...)
}

func (l *ZapLogger) HandleError(err error) {
	l.logger.Error(err.Error())
}

func (l *ZapLogger) Str(key string, value string) Logger {
	return &ZapLogger{logger: l.logger.WithLazy(zap.String(key, value)), ctx: l.ctx}
}

func (l *ZapLogger) Int(key string, value int) Logger {
	return &ZapLogger{logger: l.logger.WithLazy(zap.Int(key, value)), ctx: l.ctx}
}

func (l *ZapLogger) Bool(key string, value bool) Logger {
	return &ZapLogger{logger: l.logger.WithLazy(zap.Bool(key, value)), ctx: l.ctx}
}

func (l *ZapLogger) Any(key string, value interface{}) Logger {
	return &ZapLogger{logger: l.logger.WithLazy(zap.Any(key, value)), ctx: l.ctx}
}

func (l *ZapLogger) WithStr(key string, value string) Logger {
	l.logger = l.logger.With(zap.String(key, value))
	return l
}

func (l *ZapLogger) WithInt(key string, value int) Logger {
	l.logger = l.logger.With(zap.Int(key, value))
	return l
}

func (l *ZapLogger) WithBool(key string, value bool) Logger {
	l.logger = l.logger.With(zap.Bool(key, value))
	return l
}

func (l *ZapLogger) With(key string, value interface{}) Logger {
	l.logger = l.logger.With(zap.Any(key, value))
	return l
}

func (l *ZapLogger) GetLogLevel() string {
	return atomicLevel.Level().String()
}

func (l *ZapLogger) Hook(hookFunc func(level string, msg string)) Logger {
	l.logger = l.logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		hookFunc(entry.Level.String(), entry.Message)
		return nil
	}))
	return l
}

func (ZapLogger) SetLogLevel(level string) {
	switch level {
	case "trace":
		atomicLevel.SetLevel(zapcore.DebugLevel)
		break
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
		break
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
		break
	case "warn":
		atomicLevel.SetLevel(zapcore.WarnLevel)
		break
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
		break
	default:
		return
	}
}
