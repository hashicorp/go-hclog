package hclog

import (
	"errors"
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"
)

// TimeFormat to use for logging. This is a version of RFC3339 that contains
// contains millisecond precision
const TimeFormat = "2006-01-02T15:04:05.000Z0700"

// errJsonUnsupportedTypeMsg is included in log json entries, if an arg cannot be serialized to json
const errJsonUnsupportedTypeMsg = "logging contained values that don't serialize to json"

// Make sure that intLogger is a Logger
var _ Logger = &intLogger{}

// intLogger is an internal logger implementation. Internal in that it is
// defined entirely by this package.
type intLogger struct {
	formatter formatter
	name      string

	// This is an interface so that it's shared by any derived loggers, since
	// those derived loggers share the bufio.Writer as well.
	mutex  Locker
	writer *writer
	level  *int32

	implied []interface{}

	exclude func(level Level, msg string, args ...interface{}) bool
}

// New returns a configured logger.
func New(opts *LoggerOptions) Logger {
	return newLogger(opts)
}

// NewSinkAdapter returns a SinkAdapter with configured settings
// defined by LoggerOptions
func NewSinkAdapter(opts *LoggerOptions) SinkAdapter {
	return newLogger(opts)
}

func newLogger(opts *LoggerOptions) *intLogger {
	if opts == nil {
		opts = &LoggerOptions{}
	}

	output := opts.Output
	if output == nil {
		output = DefaultOutput
	}

	level := opts.Level
	if level == NoLevel {
		level = DefaultLevel
	}

	mutex := opts.Mutex
	if mutex == nil {
		mutex = new(sync.Mutex)
	}

	l := &intLogger{
		formatter: newFormatter(opts),
		name:      opts.Name,
		mutex:     mutex,
		writer:    newWriter(output, opts.Color),
		level:     new(int32),
		exclude:   opts.Exclude,
	}
	atomic.StoreInt32(l.level, int32(level))
	return l
}

// Log a message and a set of key/value pairs if the given level is at
// or more severe that the threshold configured in the Logger.
func (l *intLogger) log(name string, level Level, msg string, args ...interface{}) {
	if level < Level(atomic.LoadInt32(l.level)) {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.exclude != nil && l.exclude(level, msg, args...) {
		return
	}

	a := make([]interface{}, 0, len(l.implied)+len(args))
	a = append(a, l.implied...)
	a = append(a, args...)
	raw := l.formatter.Format(name, level, msg, a...)

	// TODO: leveled writer, remove this extra writer
	l.writer.Write(raw)
	l.writer.Flush(level)
}

// Emit the message and args at the provided level
func (l *intLogger) Log(level Level, msg string, args ...interface{}) {
	l.log(l.Name(), level, msg, args...)
}

// Emit the message and args at DEBUG level
func (l *intLogger) Debug(msg string, args ...interface{}) {
	l.log(l.Name(), Debug, msg, args...)
}

// Emit the message and args at TRACE level
func (l *intLogger) Trace(msg string, args ...interface{}) {
	l.log(l.Name(), Trace, msg, args...)
}

// Emit the message and args at INFO level
func (l *intLogger) Info(msg string, args ...interface{}) {
	l.log(l.Name(), Info, msg, args...)
}

// Emit the message and args at WARN level
func (l *intLogger) Warn(msg string, args ...interface{}) {
	l.log(l.Name(), Warn, msg, args...)
}

// Emit the message and args at ERROR level
func (l *intLogger) Error(msg string, args ...interface{}) {
	l.log(l.Name(), Error, msg, args...)
}

// Indicate that the logger would emit TRACE level logs
func (l *intLogger) IsTrace() bool {
	return Level(atomic.LoadInt32(l.level)) == Trace
}

// Indicate that the logger would emit DEBUG level logs
func (l *intLogger) IsDebug() bool {
	return Level(atomic.LoadInt32(l.level)) <= Debug
}

// Indicate that the logger would emit INFO level logs
func (l *intLogger) IsInfo() bool {
	return Level(atomic.LoadInt32(l.level)) <= Info
}

// Indicate that the logger would emit WARN level logs
func (l *intLogger) IsWarn() bool {
	return Level(atomic.LoadInt32(l.level)) <= Warn
}

// Indicate that the logger would emit ERROR level logs
func (l *intLogger) IsError() bool {
	return Level(atomic.LoadInt32(l.level)) <= Error
}

const MissingKey = "EXTRA_VALUE_AT_END"

// Return a sub-Logger for which every emitted log message will contain
// the given key/value pairs. This is used to create a context specific
// Logger.
func (l *intLogger) With(args ...interface{}) Logger {
	var extra interface{}

	if len(args)%2 != 0 {
		extra = args[len(args)-1]
		args = args[:len(args)-1]
	}

	sl := *l

	result := make(map[string]interface{}, len(l.implied)+len(args))
	keys := make([]string, 0, len(l.implied)+len(args))

	// Read existing args, store map and key for consistent sorting
	for i := 0; i < len(l.implied); i += 2 {
		key := l.implied[i].(string)
		keys = append(keys, key)
		result[key] = l.implied[i+1]
	}
	// Read new args, store map and key for consistent sorting
	for i := 0; i < len(args); i += 2 {
		key := args[i].(string)
		_, exists := result[key]
		if !exists {
			keys = append(keys, key)
		}
		result[key] = args[i+1]
	}

	// Sort keys to be consistent
	sort.Strings(keys)

	sl.implied = make([]interface{}, 0, len(l.implied)+len(args))
	for _, k := range keys {
		sl.implied = append(sl.implied, k)
		sl.implied = append(sl.implied, result[k])
	}

	if extra != nil {
		sl.implied = append(sl.implied, MissingKey, extra)
	}

	return &sl
}

// Create a new sub-Logger that a name decending from the current name.
// This is used to create a subsystem specific Logger.
func (l *intLogger) Named(name string) Logger {
	sl := *l

	if sl.name != "" {
		sl.name = sl.name + "." + name
	} else {
		sl.name = name
	}

	return &sl
}

// Create a new sub-Logger with an explicit name. This ignores the current
// name. This is used to create a standalone logger that doesn't fall
// within the normal hierarchy.
func (l *intLogger) ResetNamed(name string) Logger {
	sl := *l

	sl.name = name

	return &sl
}

func (l *intLogger) ResetOutput(opts *LoggerOptions) error {
	if opts.Output == nil {
		return errors.New("given output is nil")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	return l.resetOutput(opts)
}

func (l *intLogger) ResetOutputWithFlush(opts *LoggerOptions, flushable Flushable) error {
	if opts.Output == nil {
		return errors.New("given output is nil")
	}
	if flushable == nil {
		return errors.New("flushable is nil")
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if err := flushable.Flush(); err != nil {
		return err
	}

	return l.resetOutput(opts)
}

func (l *intLogger) resetOutput(opts *LoggerOptions) error {
	l.writer = newWriter(opts.Output, opts.Color)
	return nil
}

// Update the logging level on-the-fly. This will affect all subloggers as
// well.
func (l *intLogger) SetLevel(level Level) {
	atomic.StoreInt32(l.level, int32(level))
}

// Create a *log.Logger that will send it's data through this Logger. This
// allows packages that expect to be using the standard library log to actually
// use this logger.
func (l *intLogger) StandardLogger(opts *StandardLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &StandardLoggerOptions{}
	}

	return log.New(l.StandardWriter(opts), "", 0)
}

func (l *intLogger) StandardWriter(opts *StandardLoggerOptions) io.Writer {
	return &stdlogAdapter{
		log:         l,
		inferLevels: opts.InferLevels,
		forceLevel:  opts.ForceLevel,
	}
}

// Accept implements the SinkAdapter interface
func (i *intLogger) Accept(name string, level Level, msg string, args ...interface{}) {
	i.log(name, level, msg, args...)
}

// ImpliedArgs returns the loggers implied args
func (i *intLogger) ImpliedArgs() []interface{} {
	return i.implied
}

// Name returns the loggers name
func (i *intLogger) Name() string {
	return i.name
}
