package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogLevel represents log level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Logger wraps zerolog logger
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Pretty     bool
	TimeFormat string
	Output     io.Writer
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:      LevelInfo,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// DevelopmentConfig returns development logger configuration
func DevelopmentConfig() Config {
	return Config{
		Level:      LevelDebug,
		Pretty:     true,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// ProductionConfig returns production logger configuration
func ProductionConfig() Config {
	return Config{
		Level:      LevelInfo,
		Pretty:     false,
		TimeFormat: time.RFC3339,
		Output:     os.Stdout,
	}
}

// New creates a new logger instance
func New(config Config) *Logger {
	// Set time format
	zerolog.TimeFieldFormat = config.TimeFormat

	// Configure output
	var output io.Writer = config.Output
	if config.Pretty {
		output = zerolog.ConsoleWriter{
			Out:        config.Output,
			TimeFormat: time.RFC3339,
		}
	}

	// Set global log level
	level := parseLogLevel(config.Level)
	zerolog.SetGlobalLevel(level)

	// Create logger
	logger := zerolog.New(output).With().Timestamp().Caller().Logger()

	return &Logger{logger: logger}
}

// parseLogLevel converts string level to zerolog level
func parseLogLevel(level LogLevel) zerolog.Level {
	switch level {
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error with error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &Logger{logger: newLogger}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	newLogger := ctx.Logger()
	return &Logger{logger: newLogger}
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *Logger {
	newLogger := l.logger.With().Err(err).Logger()
	return &Logger{logger: newLogger}
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

// Global logger instance
var global *Logger

// Init initializes the global logger
func Init(config Config) {
	global = New(config)
	log.Logger = global.logger
}

// GetLogger returns the global logger
func GetLogger() *Logger {
	if global == nil {
		global = New(DefaultConfig())
	}
	return global
}

// Global convenience functions
func Debug(msg string) {
	GetLogger().Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Info(msg string) {
	GetLogger().Info(msg)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warn(msg string) {
	GetLogger().Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Error(msg string) {
	GetLogger().Error(msg)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func ErrorWithErr(err error, msg string) {
	GetLogger().ErrorWithErr(err, msg)
}

func Fatal(msg string) {
	GetLogger().Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}
