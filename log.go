package hclog

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// TimeFormat to use for logging. This is a version of RFC3339 that contains
// contains millisecond precision
const TimeFormat = "2006-01-02T15:04:05.000Z0700"

// errJsonUnsupportedTypeMsg is included in log json entries, if an arg cannot be serialized to json
const errJsonUnsupportedTypeMsg = "logging contained values that don't serialize to json"

var (
	_levelToBracket = map[Level]string{
		Debug: "[DEBUG]",
		Trace: "[TRACE]",
		Info:  "[INFO] ",
		Warn:  "[WARN] ",
		Error: "[ERROR]",
	}
)

type lineDetails struct {
	t       time.Time
	tfmt    string
	name    string
	caller  bool
	implied []interface{}
}

func (l *lineDetails) build(level Level, msg string, args ...interface{}) []byte {
	var line bytes.Buffer

	line.WriteString(l.t.Format(l.tfmt))
	line.WriteByte(' ')

	s, ok := _levelToBracket[level]
	if ok {
		line.WriteString(s)
	} else {
		line.WriteString("[?????]")
	}

	if l.caller {
		if _, file, l, ok := runtime.Caller(4); ok {
			line.WriteByte(' ')
			line.WriteString(trimCallerPath(file))
			line.WriteByte(':')
			line.WriteString(strconv.Itoa(l))
			line.WriteByte(':')
		}
	}

	line.WriteByte(' ')

	if l.name != "" {
		line.WriteString(l.name)
		line.WriteString(": ")
	}

	line.WriteString(msg)

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

		line.WriteByte(':')
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
					val = renderSlice(v)
					raw = true
				} else {
					val = fmt.Sprintf("%v", st)
				}
			}

			line.WriteByte(' ')
			line.WriteString(args[i].(string))
			line.WriteByte('=')

			if !raw && strings.ContainsAny(val, " \t\n\r") {
				line.WriteByte('"')
				line.WriteString(val)
				line.WriteByte('"')
			} else {
				line.WriteString(val)
			}
		}
	}
	line.WriteString("\n")

	if stacktrace != "" {
		line.WriteString(string(stacktrace))
	}

	return line.Bytes()

}
func (l *lineDetails) buildJSON(level Level, msg string, args ...interface{}) []byte {
	var line bytes.Buffer
	vals := l.jsonMapEntry(l.t, level, msg)
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
	err := json.NewEncoder(&line).Encode(vals)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			plainVal := l.jsonMapEntry(l.t, level, msg)
			plainVal["@warn"] = errJsonUnsupportedTypeMsg

			json.NewEncoder(&line).Encode(plainVal)
		}
	}

	return line.Bytes()
}

func (l *lineDetails) jsonMapEntry(t time.Time, level Level, msg string) map[string]interface{} {
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
		if _, file, line, ok := runtime.Caller(5); ok {
			vals["@caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return vals
}

// Cleanup a path by returning the last 2 segments of the path only.
func trimCallerPath(path string) string {
	// lovely borrowed from zap
	// nb. To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.

	// Find the last separator.
	idx := strings.LastIndexByte(path, '/')
	if idx == -1 {
		return path
	}

	// Find the penultimate separator.
	idx = strings.LastIndexByte(path[:idx], '/')
	if idx == -1 {
		return path
	}

	return path[idx+1:]
}

func renderSlice(v reflect.Value) string {
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
