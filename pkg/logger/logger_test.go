package logger

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{name: "default level", args: args{level: ""}, want: zerolog.InfoLevel},
		{name: "unknown level", args: args{level: "unknown"}, want: zerolog.InfoLevel},
		{name: "info level", args: args{level: "info"}, want: zerolog.InfoLevel},
		{name: "error level", args: args{level: "error"}, want: zerolog.ErrorLevel},
		{name: "warning level", args: args{level: "warning"}, want: zerolog.WarnLevel},
		{name: "warn level", args: args{level: "warn"}, want: zerolog.WarnLevel},
		{name: "debug level", args: args{level: "debug"}, want: zerolog.DebugLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.args.level)
			assert.NotNil(t, l)
			assert.Equal(t, tt.want, l.logger.GetLevel())
		})
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{Out: &buf}
	lg := zerolog.New(consoleWriter).With().Logger()
	l := &Logger{
		logger: lg,
	}
	l.Debug("debug message: %d", 1)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "DBG")
	assert.Contains(t, logOutput, "debug message: 1")
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{Out: &buf}
	lg := zerolog.New(consoleWriter).With().Logger()
	l := &Logger{
		logger: lg,
	}
	l.Error("error message: %d", 1)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "ERR")
	assert.Contains(t, logOutput, "error message: 1")
	// clear buffer
	buf.Truncate(0)
	l.Error(fmt.Errorf("error message: %d", 1))
	logOutput = buf.String()
	assert.Contains(t, logOutput, "ERR")
	assert.Contains(t, logOutput, "error message: 1")
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{Out: &buf}
	lg := zerolog.New(consoleWriter).With().Logger()
	l := &Logger{
		logger: lg,
	}
	l.Info("info message: %d", 1)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "INF")
	assert.Contains(t, logOutput, "info message: 1")
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{Out: &buf}
	lg := zerolog.New(consoleWriter).With().Logger()
	l := &Logger{
		logger: lg,
	}
	l.Warn("warn message: %d", 1)
	logOutput := buf.String()
	assert.Contains(t, logOutput, "WRN")
	assert.Contains(t, logOutput, "warn message: 1")
}

func TestMsg(t *testing.T) {
	var buf bytes.Buffer
	consoleWriter := zerolog.ConsoleWriter{Out: &buf}
	lg := zerolog.New(consoleWriter).With().Logger()
	l := &Logger{
		logger: lg,
	}
	l.msg(l.logger.Debug(), struct{ A string }{A: "test"})
	logOutput := buf.String()
	assert.Contains(t, logOutput, "DBG")
	assert.Contains(t, logOutput, "message {test} has unknown type {test}")
}
