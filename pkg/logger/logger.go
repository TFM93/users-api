package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

var ExitFn = os.Exit

type Interface interface {
	// Debug logs a debug message, it can receive a message or an error
	Debug(message interface{}, args ...interface{})
	// Info logs an info message
	Info(message string, args ...interface{})
	// Warn warn an info message
	Warn(message string, args ...interface{})
	// Error logs an err message, it can receive a message or an error
	Error(message interface{}, args ...interface{})
	// GinLoggerFn returns a gin.HandlerFunc that configures the gin to log using this logger
	GinLoggerFn() gin.HandlerFunc
}

type Logger struct {
	logger zerolog.Logger
}

var _ Interface = (*Logger)(nil)

func New(level string) *Logger {
	var l zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		l = zerolog.ErrorLevel
	case "warn", "warning":
		l = zerolog.WarnLevel
	case "info":
		l = zerolog.InfoLevel
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}

	logger := zerolog.New(os.Stdout).Level(l).With().Timestamp().Logger()
	return &Logger{
		logger: logger,
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(message interface{}, args ...interface{}) {
	l.msg(l.logger.Debug(), message, args...)
}

// Info logs an info message.
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(l.logger.Info(), message, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(l.logger.Warn(), message, args...)
}

// Error logs an error message.
func (l *Logger) Error(message interface{}, args ...interface{}) {
	l.msg(l.logger.Error(), message, args...)
}

// log handles the message formatting for standard log entries.
func (l *Logger) log(loggerEvent *zerolog.Event, message string, args ...interface{}) {
	if len(args) == 0 {
		loggerEvent.Msg(message)
	} else {
		loggerEvent.Msgf(message, args...)
	}
}

// msg handles the message formatting depending on message type.
func (l *Logger) msg(loggerEvent *zerolog.Event, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		l.log(loggerEvent, msg.Error(), args...)
	case string:
		l.log(loggerEvent, msg, args...)
	default:
		l.log(loggerEvent, fmt.Sprintf("message %v has unknown type %v", message, msg), args...)
	}
}

// GinLoggerFn logs a gin HTTP request in JSON format.
// adapted from https://learninggolang.com/it5-gin-structured-logging.html
func (l *Logger) GinLoggerFn() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		param := gin.LogFormatterParams{
			TimeStamp:  time.Now(),
			ClientIP:   c.ClientIP(),
			Method:     c.Request.Method,
			StatusCode: c.Writer.Status(),
		}

		param.Latency = param.TimeStamp.Sub(start)
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		// Log using the params
		var logEvent *zerolog.Event
		if c.Writer.Status() >= 500 {
			logEvent = l.logger.Error()
		} else if c.Writer.Status() >= 300 {
			logEvent = l.logger.Info()
		} else {
			logEvent = l.logger.Debug()
		}

		logEvent.Str("ip", param.ClientIP).
			Str("method", param.Method).
			Int("code", param.StatusCode).
			Str("path", param.Path).
			Str("latency", param.Latency.String()).
			Msg(param.ErrorMessage)
	}
}
