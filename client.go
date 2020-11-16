package logger

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggingConfig struct {
	// Service name
	Service string `env:"LOGGER_SERVICE"`

	// Minimum log level to be sent.
	// E.g. if set to "warn" and .Info() called, log will be neither sent nor logged.
	Level string `env:"LOGGER_LEVEL"`

	// Namespace defines Elasticsearch index where logs will be stored.
	// Can be overwritten for each log using .With method.
	Namespace string `env:"LOGGER_NAMESPACE"`

	// Disables stdout if not needed.
	DisableStdout bool   `env:"LOGGER_DISABLE_STDOUT"`
	FormatStdout  string `env:"LOGGER_FORMAT_STDOUT"`

	// TCP connection settings. Only for development and testing, publishers should be used instead in production.
	LogstashURI      string `env:"LOGGER_LOGSTASH_URI"`
	LogstashProtocol string `env:"LOGGER_LOGSTASH_PROTOCOL"`
}

var DefaultConfig = LoggingConfig{
	Service:   "awesome-service",
	Level:     "debug",
	Namespace: "awesome-namespace",

	DisableStdout: false,
	FormatStdout:  FormatJSON,

	// Not used by default
	LogstashURI:      "",
	LogstashProtocol: "udp",
}

var (
	FormatJSON   = "json"
	FormatPretty = "pretty"
)

type Logger interface {
	Debug(message ...interface{})
	Debugf(format string, args ...interface{})

	Info(message ...interface{})
	Infof(format string, args ...interface{})

	Warn(message ...interface{})
	Warnf(format string, args ...interface{})

	Error(message ...interface{})
	Errorf(format string, args ...interface{})

	Panic(message ...interface{})
	Panicf(format string, args ...interface{})

	Fatal(message ...interface{})
	Fatalf(format string, args ...interface{})

	// Add extra fields to message
	With(fields Fields) Logger

	// Override namespace
	Namespace(namespace string) Logger

	// Logs call stack for error
	Trace(err error)

	// Tries to recover from panic. Logs trace of error if occurred and calls Panic with passed message
	// Like any recover should be deferred
	Recover(msg string)

	GetField(field string) (interface{}, bool)
}

type loggerImpl struct {
	base *zap.SugaredLogger

	// Extra fields
	fields Fields
}

func (l loggerImpl) prepare() *zap.SugaredLogger {
	flatten := l.fields.Flatten()

	prepared := l.base.With(flatten...)

	putFlatten(flatten)

	return prepared
}

func (l loggerImpl) Debug(message ...interface{}) {
	l.prepare().Debug(message...)
}

func (l loggerImpl) Debugf(format string, args ...interface{}) {
	l.prepare().Debugf(format, args...)
}

func (l loggerImpl) Info(message ...interface{}) {
	l.prepare().Info(message...)
}

func (l loggerImpl) Infof(format string, args ...interface{}) {
	l.prepare().Infof(format, args...)
}

func (l loggerImpl) Warn(message ...interface{}) {
	l.prepare().Warn(message...)
}

func (l loggerImpl) Warnf(format string, args ...interface{}) {
	l.prepare().Warnf(format, args...)
}

func (l loggerImpl) Error(message ...interface{}) {
	l.prepare().Error(message...)
}

func (l loggerImpl) Errorf(format string, args ...interface{}) {
	l.prepare().Errorf(format, args...)
}

func (l loggerImpl) Panic(message ...interface{}) {
	l.prepare().Panic(message...)
}

func (l loggerImpl) Panicf(format string, args ...interface{}) {
	l.prepare().Panicf(format, args...)
}

func (l loggerImpl) Fatal(message ...interface{}) {
	l.prepare().Fatal(message...)
}

func (l loggerImpl) Fatalf(format string, args ...interface{}) {
	l.prepare().Fatalf(format, args...)
}

func (l loggerImpl) With(fields Fields) Logger {
	l.fields = l.fields.Merge(fields)

	return l
}

func (l loggerImpl) Namespace(namespace string) Logger {
	l.fields = l.fields.Merge(Fields{"namespace": namespace})

	return l
}

func (l loggerImpl) GetField(fieldName string) (value interface{}, ok bool) {
	value, ok = l.fields[fieldName]
	return value, ok
}

func New(config LoggingConfig) (logger Logger, err error) {
	level := config.Level
	if level == "" {
		log.Println("logging level not set, using 'info'")
		level = "info"
	}

	format, err := getFormat(config.FormatStdout)
	if err != nil {
		return nil, err
	}

	zapLevel, err := getLevel(level)
	if err != nil {
		return nil, err
	}

	zapLogger, err := newZapLogger(
		zapLevel,
		config.Service,
		config.LogstashProtocol, config.LogstashURI,
		config.DisableStdout,
		format,
	)
	if err != nil {
		return nil, err
	}

	logger = &loggerImpl{
		base:   zapLogger.Sugar(),
		fields: Fields{"namespace": config.Namespace},
	}

	return logger, nil
}

func newZapLogger(
	zapLevel zapcore.Level,
	service string,
	logstashProtocol, logstashURI string,
	disableStdout bool,
	formatStdout string,
) (*zap.Logger, error) {
	var cores []zapcore.Core

	if !disableStdout {
		cores = append(cores, newStdoutCore(zapLevel, formatStdout))
	}

	// Optional logstash connection
	if logstashURI != "" {
		log.Println("using logstash, should not be used in production")
		logstashCore, err := newLogstashCore(zapLevel, logstashProtocol, logstashURI)
		if err != nil {
			return nil, err
		}
		cores = append(cores, logstashCore)
	}

	core := zapcore.NewTee(
		cores...,
	)

	// Add general fields
	core = core.With(
		[]zap.Field{
			zap.String("service", service),
		},
	)

	zapLogger := zap.New(core)

	return zapLogger, nil
}

func newStdoutCore(zapLevel zapcore.Level, format string) zapcore.Core {
	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapLevel
	})

	console := zapcore.Lock(os.Stdout)

	var encoder zapcore.Encoder
	encoderConfig := newEncoderConfig()
	if format == FormatJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	stdoutCore := zapcore.NewCore(encoder, console, levelEnabler)

	return stdoutCore
}

func newLogstashCore(zapLevel zapcore.Level, protocol, addr string) (zapcore.Core, error) {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		return nil, err
	}

	levelEnabler := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapLevel
	})

	logstashEncoder := zapcore.NewJSONEncoder(newEncoderConfig())

	tcpWriter := zapcore.AddSync(conn)

	logstashCore := zapcore.
		NewCore(logstashEncoder, tcpWriter, levelEnabler).
		With([]zap.Field{
			// Extra fields from logrustash formatter, not sure if they are really needed
			zap.String("@version", "1"),
			zap.String("type", "log"),
		})

	return logstashCore, nil
}

func newEncoderConfig() zapcore.EncoderConfig {
	logstashEncoderConfig := zap.NewProductionEncoderConfig()
	logstashEncoderConfig.MessageKey = "message"
	logstashEncoderConfig.TimeKey = "@timestamp"
	logstashEncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
	return logstashEncoderConfig
}

func getFormat(format string) (string, error) {
	if format == "" {
		return FormatJSON, nil
	}

	if format != FormatJSON && format != FormatPretty {
		return "", fmt.Errorf("invalid FormatStdout %v, must be %v or %v",
			format, FormatJSON, FormatPretty)
	}

	return format, nil
}

func getLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	default:
		return 0, fmt.Errorf("bad logging level %v", level)
	}
}

func (l loggerImpl) Trace(err error) {
	if err != nil {
		l.Errorf("%+v", errors.WithStack(err))
	}
}

func (l loggerImpl) Recover(msg string) {
	if i := recover(); i != nil {
		switch v := i.(type) {
		case error:
			l.Trace(v)
		case string:
			l.Trace(errors.New(v))
		}
		l.Panicf("recovered %s from %v", msg, i)
	}
}
