package hclog

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
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
	level       *int32
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
		level:       new(int32),
		lowestLevel: new(int32),
		sinks:       make(map[*Sink]struct{}),
	}

	if opts.TimeFormat != "" {
		l.timeFormat = opts.TimeFormat
	}

	atomic.StoreInt32(l.lowestLevel, int32(level))
	atomic.StoreInt32(l.level, int32(level))

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

	if sink.level < Level(atomic.LoadInt32(l.lowestLevel)) {
		atomic.StoreInt32(l.lowestLevel, int32(sink.level))
	}

	if _, ok := l.sinks[sink]; ok {
		return
	}

	l.sinks[sink] = struct{}{}
}

func (l *sinkLogger) DeregisterSink(sink *Sink) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	delete(l.sinks, sink)
}

// Level returns the current level
func (l *sinkLogger) Level() Level {
	return Level(atomic.LoadInt32(l.level))
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

	for sink := range l.sinks {
		if level < Level(sink.level) {
			continue
		}

		lwriter := l.writer
		l.writer = sink.writer

		if sink.json {
			l.logJSON(t, level, msg, args...)
		} else {
			l.log(t, level, msg, args...)
		}
		l.writer.Flush(level)

		l.writer = lwriter

	}

	if level < Level(atomic.LoadInt32(l.level)) {
		return
	}

	if l.json {
		l.logJSON(t, level, msg, args...)
	} else {
		l.log(t, level, msg, args...)
	}

	l.writer.Flush(level)
}

// Non-JSON logging format function
func (l *sinkLogger) log(t time.Time, level Level, msg string, args ...interface{}) {
	l.writer.WriteString(t.Format(l.timeFormat))
	l.writer.WriteByte(' ')

	s, ok := _levelToBracket[level]
	if ok {
		l.writer.WriteString(s)
	} else {
		l.writer.WriteString("[?????]")
	}

	if l.caller {
		if _, file, line, ok := runtime.Caller(3); ok {
			l.writer.WriteByte(' ')
			l.writer.WriteString(trimCallerPath(file))
			l.writer.WriteByte(':')
			l.writer.WriteString(strconv.Itoa(line))
			l.writer.WriteByte(':')
		}
	}

	l.writer.WriteByte(' ')

	if l.name != "" {
		l.writer.WriteString(l.name)
		l.writer.WriteString(": ")
	}

	l.writer.WriteString(msg)

	args = append(l.implied, args...)

	var stacktrace CapturedStacktrace

	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			cs, ok := args[len(args)-1].(CapturedStacktrace)
			if ok {
				args = args[:len(args)-1]
				stacktrace = cs
			} else {
				args = append(args, "<unknown>")
			}
		}

		l.writer.WriteByte(':')

	FOR:
		for i := 0; i < len(args); i = i + 2 {
			var (
				val string
				raw bool
			)

			switch st := args[i+1].(type) {
			case string:
				val = st
			case int:
				val = strconv.FormatInt(int64(st), 10)
			case int64:
				val = strconv.FormatInt(int64(st), 10)
			case int32:
				val = strconv.FormatInt(int64(st), 10)
			case int16:
				val = strconv.FormatInt(int64(st), 10)
			case int8:
				val = strconv.FormatInt(int64(st), 10)
			case uint:
				val = strconv.FormatUint(uint64(st), 10)
			case uint64:
				val = strconv.FormatUint(uint64(st), 10)
			case uint32:
				val = strconv.FormatUint(uint64(st), 10)
			case uint16:
				val = strconv.FormatUint(uint64(st), 10)
			case uint8:
				val = strconv.FormatUint(uint64(st), 10)
			case CapturedStacktrace:
				stacktrace = st
				continue FOR
			case Format:
				val = fmt.Sprintf(st[0].(string), st[1:]...)
			default:
				v := reflect.ValueOf(st)
				if v.Kind() == reflect.Slice {
					val = l.renderSlice(v)
					raw = true
				} else {
					val = fmt.Sprintf("%v", st)
				}
			}

			l.writer.WriteByte(' ')
			l.writer.WriteString(args[i].(string))
			l.writer.WriteByte('=')

			if !raw && strings.ContainsAny(val, " \t\n\r") {
				l.writer.WriteByte('"')
				l.writer.WriteString(val)
				l.writer.WriteByte('"')
			} else {
				l.writer.WriteString(val)
			}
		}
	}

	l.writer.WriteString("\n")

	if stacktrace != "" {
		l.writer.WriteString(string(stacktrace))
	}
}

func (l *sinkLogger) renderSlice(v reflect.Value) string {
	var buf bytes.Buffer

	buf.WriteRune('[')

	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}

		sv := v.Index(i)

		var val string

		switch sv.Kind() {
		case reflect.String:
			val = sv.String()
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			val = strconv.FormatInt(sv.Int(), 10)
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val = strconv.FormatUint(sv.Uint(), 10)
		default:
			val = fmt.Sprintf("%v", sv.Interface())
		}

		if strings.ContainsAny(val, " \t\n\r") {
			buf.WriteByte('"')
			buf.WriteString(val)
			buf.WriteByte('"')
		} else {
			buf.WriteString(val)
		}
	}

	buf.WriteRune(']')

	return buf.String()
}

// JSON logging function
func (l *sinkLogger) logJSON(t time.Time, level Level, msg string, args ...interface{}) {
	vals := l.jsonMapEntry(t, level, msg)
	args = append(l.implied, args...)

	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			cs, ok := args[len(args)-1].(CapturedStacktrace)
			if ok {
				args = args[:len(args)-1]
				vals["stacktrace"] = cs
			} else {
				args = append(args, "<unknown>")
			}
		}

		for i := 0; i < len(args); i = i + 2 {
			if _, ok := args[i].(string); !ok {
				// As this is the logging function not much we can do here
				// without injecting into logs...
				continue
			}
			val := args[i+1]
			switch sv := val.(type) {
			case error:
				// Check if val is of type error. If error type doesn't
				// implement json.Marshaler or encoding.TextMarshaler
				// then set val to err.Error() so that it gets marshaled
				switch sv.(type) {
				case json.Marshaler, encoding.TextMarshaler:
				default:
					val = sv.Error()
				}
			case Format:
				val = fmt.Sprintf(sv[0].(string), sv[1:]...)
			}

			vals[args[i].(string)] = val
		}
	}

	err := json.NewEncoder(l.writer).Encode(vals)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			plainVal := l.jsonMapEntry(t, level, msg)
			plainVal["@warn"] = errJsonUnsupportedTypeMsg

			json.NewEncoder(l.writer).Encode(plainVal)
		}
	}
}

func (l sinkLogger) jsonMapEntry(t time.Time, level Level, msg string) map[string]interface{} {
	vals := map[string]interface{}{
		"@message":   msg,
		"@timestamp": t.Format("2006-01-02T15:04:05.000000Z07:00"),
	}

	var levelStr string
	switch level {
	case Error:
		levelStr = "error"
	case Warn:
		levelStr = "warn"
	case Info:
		levelStr = "info"
	case Debug:
		levelStr = "debug"
	case Trace:
		levelStr = "trace"
	default:
		levelStr = "all"
	}

	vals["@level"] = levelStr

	if l.name != "" {
		vals["@module"] = l.name
	}

	if l.caller {
		if _, file, line, ok := runtime.Caller(4); ok {
			vals["@caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return vals
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
	return Level(atomic.LoadInt32(l.level)) == Trace
}

// Indicate that the logger would emit DEBUG level logs
func (l *sinkLogger) IsDebug() bool {
	return Level(atomic.LoadInt32(l.level)) <= Debug
}

// Indicate that the logger would emit INFO level logs
func (l *sinkLogger) IsInfo() bool {
	return Level(atomic.LoadInt32(l.level)) <= Info
}

// Indicate that the logger would emit WARN level logs
func (l *sinkLogger) IsWarn() bool {
	return Level(atomic.LoadInt32(l.level)) <= Warn
}

// Indicate that the logger would emit ERROR level logs
func (l *sinkLogger) IsError() bool {
	return Level(atomic.LoadInt32(l.level)) <= Error
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
	atomic.StoreInt32(l.level, int32(level))
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
