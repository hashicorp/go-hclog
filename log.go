package log

import (
	"io"
	"log"
)

type Level int

const (
	Debug Level = 0
	Trace Level = 1
	Info  Level = 2
	Warn  Level = 3
	Error Level = 4
)

type Logger interface {
	// Args are alternating key, val pairs
	// keys must be strings
	// vals can be any type, but display is implementation specific
	Debug(msg string, args ...interface{})
	Trace(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})

	IsDebug() bool
	IsTrace() bool
	IsInfo() bool
	IsWarn() bool
	IsError() bool

	// Creates a sublogger that will always have the given key/value pairs
	With(args ...interface{}) Logger

	// Create a logger that will prepend the name string on the front of all messages
	Named(name string) Logger

	// Log the arguments and then a stacktrace
	Stacktrace(args ...interface{})

	// Return a value that conforms to the stdlib log.Logger interface
	// if inferLevels is set, then check for strings like [ERROR], [ERR]
	// [TRACE], [WARN], [INFO], [DEBUG] and strip it off before reapplying it.
	StandardLogger(inferLevels bool) *log.Logger
}

type LoggerOptions struct {
	Name   string
	Level  Level
	Output io.Writer

	IncludeLocation bool
}
