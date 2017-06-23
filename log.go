package log

import (
	"io"
	"log"
	"os"
)

var (
	DefaultOutput = os.Stderr
	DefaultLevel  = Info
)

type Level int

const (
	// This is a special level used to indicate that no level has been
	// set and allow for a default to be used.
	NoLevel Level = 0

	Trace Level = 1
	Debug Level = 2
	Info  Level = 3
	Warn  Level = 4
	Error Level = 5
)

type Logger interface {
	// Args are alternating key, val pairs
	// keys must be strings
	// vals can be any type, but display is implementation specific
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})

	IsTrace() bool
	IsDebug() bool
	IsInfo() bool
	IsWarn() bool
	IsError() bool

	// Creates a sublogger that will always have the given key/value pairs
	With(args ...interface{}) Logger

	// Create a logger that will prepend the name string on the front of all messages.
	// If the logger already has a name, the new value will be appended to the current
	// name. That way, a major subsystem can use this to decorate all it's own logs
	// without losing context.
	Named(name string) Logger

	// Create a logger that will prepend the name string on the front of all messages.
	// This sets the name of the logger to the value directly, unlike Named which honor
	// the current name as well.
	ResetNamed(name string) Logger

	// Return a value that conforms to the stdlib log.Logger interface
	StandardLogger(opts *StandardLoggerOptions) *log.Logger
}

type StandardLoggerOptions struct {
	// Indicate that some minimal parsing should be done on strings to try
	// and detect their level and re-emit them.
	// This supports the strings like [ERROR], [ERR] [TRACE], [WARN], [INFO],
	// [DEBUG] and strip it off before reapplying it.
	InferLevels bool
}

type LoggerOptions struct {
	// Name of the subsystem to prefix logs with
	Name string

	// The threshold for the logger. Anything less severe is supressed
	Level Level

	// Where to write the logs to. Defaults to os.Stdout if nil
	Output io.Writer

	// Control if the output should be in JSON.
	JSONFormat bool

	// Intclude file and line information in each log line
	IncludeLocation bool
}
