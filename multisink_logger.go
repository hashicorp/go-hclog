package hclog

import (
	"io"
	"log"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Make sure that sinkLogger is a Logger
var _ Logger = &sinkLogger{}
var _ MultiSinkLogger = &sinkLogger{}

// sinkLogger is an internal logger implementation. Internal in that it is
// defined entirely by this package.
type sinkLogger struct {
	json       bool
	caller     bool
	name       string
	timeFormat string

	// This is a pointer so that it's shared by any derived loggers, since
	// those derived loggers share the bufio.Writer as well.
	mutex       *sync.Mutex
	writer      *writer
	level       Level
	lowestLevel *int32

	sinks map[*Sink]struct{}

	implied []interface{}
}

// New returns a configured logger.
func NewMultiSink(opts *LoggerOptions) MultiSinkLogger {
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

	l := &sinkLogger{
		json:        opts.JSONFormat,
		caller:      opts.IncludeLocation,
		name:        opts.Name,
		timeFormat:  TimeFormat,
		mutex:       mutex,
		writer:      newWriter(output),
		level:       level,
		lowestLevel: new(int32),
		sinks:       make(map[*Sink]struct{}),
	}

	if opts.TimeFormat != "" {
		l.timeFormat = opts.TimeFormat
	}

	atomic.StoreInt32(l.lowestLevel, int32(level))

	return l
}

type Sink struct {
	writer *writer
	level  Level
	json   bool
}

func NewSink(opts *SinkOptions) *Sink {
	return &Sink{
		level:  opts.Level,
		json:   opts.JSONFormat,
		writer: newWriter(opts.Output),
	}
}

func (l *sinkLogger) RegisterSink(sink *Sink) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, ok := l.sinks[sink]; ok {
		return
	}

	if sink.level < Level(atomic.LoadInt32(l.lowestLevel)) {
		atomic.StoreInt32(l.lowestLevel, int32(sink.level))
	}

	l.sinks[sink] = struct{}{}
}

func (l *sinkLogger) DeregisterSink(sink *Sink) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.sinks, sink)
}

// Log a message and a set of key/value pairs if the given level is at
// or more severe that the threshold configured in the Logger.
func (l *sinkLogger) Log(level Level, msg string, args ...interface{}) {
	if level < Level(atomic.LoadInt32(l.lowestLevel)) {
		return
	}

	t := time.Now()

	l.mutex.Lock()
	defer l.mutex.Unlock()

	var line []byte
	var lineJSON []byte
	for sink := range l.sinks {
		if level < Level(sink.level) {
			continue
		}

		if sink.json {
			if len(lineJSON) == 0 {
				lineJSON = l.logJSON(t, level, msg, args...)
			}
			sink.writer.Write(lineJSON)
		} else {
			if len(line) == 0 {
				line = l.log(t, level, msg, args...)
			}
			sink.writer.Write(line)
		}
		sink.writer.Flush(level)
	}

	if level < Level(l.level) {
		return
	}

	if l.json {
		if len(lineJSON) == 0 {
			lineJSON = l.logJSON(t, level, msg, args...)
		}
		l.writer.Write(line)
	} else {
		if len(line) == 0 {
			line = l.log(t, level, msg, args...)
		}
		l.writer.Write(line)
	}

	l.writer.Flush(level)
}

// Non-JSON logging format function
func (l *sinkLogger) log(t time.Time, level Level, msg string, args ...interface{}) []byte {
	ld := &lineDetails{
		t:       t,
		tfmt:    l.timeFormat,
		name:    l.name,
		caller:  l.caller,
		implied: l.implied,
	}
	return ld.build(level, msg, args...)
}

// JSON logging function
func (l *sinkLogger) logJSON(t time.Time, level Level, msg string, args ...interface{}) []byte {
	ld := &lineDetails{
		t:       t,
		tfmt:    l.timeFormat,
		name:    l.name,
		caller:  l.caller,
		implied: l.implied,
	}

	return ld.buildJSON(level, msg, args...)
}

// Emit the message and args at DEBUG level
func (l *sinkLogger) Debug(msg string, args ...interface{}) {
	l.Log(Debug, msg, args...)
}

// Emit the message and args at TRACE level
func (l *sinkLogger) Trace(msg string, args ...interface{}) {
	l.Log(Trace, msg, args...)
}

// Emit the message and args at INFO level
func (l *sinkLogger) Info(msg string, args ...interface{}) {
	l.Log(Info, msg, args...)
}

// Emit the message and args at WARN level
func (l *sinkLogger) Warn(msg string, args ...interface{}) {
	l.Log(Warn, msg, args...)
}

// Emit the message and args at ERROR level
func (l *sinkLogger) Error(msg string, args ...interface{}) {
	l.Log(Error, msg, args...)
}

// Indicate that the logger would emit TRACE level logs
func (l *sinkLogger) IsTrace() bool {
	return Level(atomic.LoadInt32(l.lowestLevel)) == Trace
}

// Indicate that the logger would emit DEBUG level logs
func (l *sinkLogger) IsDebug() bool {
	return Level(atomic.LoadInt32(l.lowestLevel)) <= Debug
}

// Indicate that the logger would emit INFO level logs
func (l *sinkLogger) IsInfo() bool {
	return Level(atomic.LoadInt32(l.lowestLevel)) <= Info
}

// Indicate that the logger would emit WARN level logs
func (l *sinkLogger) IsWarn() bool {
	return Level(atomic.LoadInt32(l.lowestLevel)) <= Warn
}

// Indicate that the logger would emit ERROR level logs
func (l *sinkLogger) IsError() bool {
	return Level(atomic.LoadInt32(l.lowestLevel)) <= Error
}

// Return a sub-Logger for which every emitted log message will contain
// the given key/value pairs. This is used to create a context specific
// Logger.
func (l *sinkLogger) With(args ...interface{}) Logger {
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
func (l *sinkLogger) NamedMultiSink(name string) MultiSinkLogger {
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
func (l *sinkLogger) ResetNamedMultiSink(name string) MultiSinkLogger {
	sl := *l

	sl.name = name

	return &sl
}

// Create a new sub-Logger that a name decending from the current name.
// This is used to create a subsystem specific Logger.
func (l *sinkLogger) Named(name string) Logger {
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
func (l *sinkLogger) ResetNamed(name string) Logger {
	sl := *l

	sl.name = name

	return &sl
}

// Update the logging level on-the-fly. This will affect all subloggers as
// well.
func (l *sinkLogger) SetLevel(level Level) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	currentLowest := atomic.LoadInt32(l.lowestLevel)
	if level < Level(currentLowest) {
		atomic.StoreInt32(l.lowestLevel, int32(level))
	}

	l.level = level
}

// Create a *log.Logger that will send it's data through this Logger. This
// allows packages that expect to be using the standard library log to actually
// use this logger.
func (l *sinkLogger) StandardLogger(opts *StandardLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &StandardLoggerOptions{}
	}

	return log.New(l.StandardWriter(opts), "", 0)
}

func (l *sinkLogger) StandardWriter(opts *StandardLoggerOptions) io.Writer {
	return &stdlogAdapter{
		log:         l,
		inferLevels: opts.InferLevels,
		forceLevel:  opts.ForceLevel,
	}
}
