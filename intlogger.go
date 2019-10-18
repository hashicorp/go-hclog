package hclog

import (
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Make sure that intLogger is a Logger
var _ Logger = &intLogger{}

// intLogger is an internal logger implementation. Internal in that it is
// defined entirely by this package.
type intLogger struct {
	json       bool
	caller     bool
	name       string
	timeFormat string

	// This is a pointer so that it's shared by any derived loggers, since
	// those derived loggers share the bufio.Writer as well.
	mutex  *sync.Mutex
	writer *writer
	level  *int32

	implied []interface{}
}

// New returns a configured logger.
func New(opts *LoggerOptions) Logger {
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
		json:       opts.JSONFormat,
		caller:     opts.IncludeLocation,
		name:       opts.Name,
		timeFormat: TimeFormat,
		mutex:      mutex,
		writer:     newWriter(output),
		level:      new(int32),
	}

	if opts.TimeFormat != "" {
		l.timeFormat = opts.TimeFormat
	}

	atomic.StoreInt32(l.level, int32(level))

	return l
}

// Log a message and a set of key/value pairs if the given level is at
// or more severe that the threshold configured in the Logger.
func (l *intLogger) Log(level Level, msg string, args ...interface{}) {
	if level < Level(atomic.LoadInt32(l.level)) {
		return
	}

	t := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.json {
		l.logJSON(t, level, msg, args...)
	} else {
		l.log(t, level, msg, args...)
	}

	l.writer.Flush(level)
}

// Non-JSON logging format function
func (l *intLogger) log(t time.Time, level Level, msg string, args ...interface{}) {
	ld := &lineDetails{
		t:       t,
		tfmt:    l.timeFormat,
		name:    l.name,
		caller:  l.caller,
		implied: l.implied,
	}
	l.writer.Write(ld.build(level, msg, args...))
}

// JSON logging function
func (l *intLogger) logJSON(t time.Time, level Level, msg string, args ...interface{}) {
	ld := &lineDetails{
		t:       t,
		tfmt:    l.timeFormat,
		name:    l.name,
		caller:  l.caller,
		implied: l.implied,
	}
	l.writer.Write(ld.buildJSON(level, msg, args...))
}

// Emit the message and args at DEBUG level
func (l *intLogger) Debug(msg string, args ...interface{}) {
	l.Log(Debug, msg, args...)
}

// Emit the message and args at TRACE level
func (l *intLogger) Trace(msg string, args ...interface{}) {
	l.Log(Trace, msg, args...)
}

// Emit the message and args at INFO level
func (l *intLogger) Info(msg string, args ...interface{}) {
	l.Log(Info, msg, args...)
}

// Emit the message and args at WARN level
func (l *intLogger) Warn(msg string, args ...interface{}) {
	l.Log(Warn, msg, args...)
}

// Emit the message and args at ERROR level
func (l *intLogger) Error(msg string, args ...interface{}) {
	l.Log(Error, msg, args...)
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

// Return a sub-Logger for which every emitted log message will contain
// the given key/value pairs. This is used to create a context specific
// Logger.
func (l *intLogger) With(args ...interface{}) Logger {
	if len(args)%2 != 0 {
		panic("With() call requires paired arguments")
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
