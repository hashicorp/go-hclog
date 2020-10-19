package hclog

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func newFormatter(opts *LoggerOptions) formatter {
	if opts.JSONFormat {
		return jsonFormat{timeFormat: opts.timeFormat(), callerOffset: opts.callerOffset()}
	}
	return textFormat{timeFormat: opts.timeFormat(), callerOffset: opts.callerOffset()}
}

type formatter interface {
	Format(name string, level Level, msg string, args ...interface{}) []byte
}

type textFormat struct {
	timeFormat   string
	callerOffset int
}

func (l textFormat) Format(name string, level Level, msg string, args ...interface{}) []byte {
	t := time.Now()
	buf := new(bytes.Buffer)

	if len(l.timeFormat) > 0 {
		buf.WriteString(t.Format(l.timeFormat))
		buf.WriteByte(' ')
	}

	s, ok := _levelToBracket[level]
	if ok {
		buf.WriteString(s)
	} else {
		buf.WriteString("[?????]")
	}

	if l.callerOffset > 0 {
		if _, file, line, ok := runtime.Caller(l.callerOffset); ok {
			buf.WriteByte(' ')
			buf.WriteString(trimCallerPath(file))
			buf.WriteByte(':')
			buf.WriteString(strconv.Itoa(line))
			buf.WriteByte(':')
		}
	}

	buf.WriteByte(' ')

	if name != "" {
		buf.WriteString(name)
		buf.WriteString(": ")
	}

	buf.WriteString(msg)

	var stacktrace CapturedStacktrace

	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			cs, ok := args[len(args)-1].(CapturedStacktrace)
			if ok {
				args = args[:len(args)-1]
				stacktrace = cs
			} else {
				extra := args[len(args)-1]
				args = append(args[:len(args)-1], MissingKey, extra)
			}
		}

		buf.WriteByte(':')

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
			case Hex:
				val = "0x" + strconv.FormatUint(uint64(st), 16)
			case Octal:
				val = "0" + strconv.FormatUint(uint64(st), 8)
			case Binary:
				val = "0b" + strconv.FormatUint(uint64(st), 2)
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

			buf.WriteByte(' ')
			switch st := args[i].(type) {
			case string:
				buf.WriteString(st)
			default:
				buf.WriteString(fmt.Sprintf("%s", st))
			}
			buf.WriteByte('=')

			if !raw && strings.ContainsAny(val, " \t\n\r") {
				buf.WriteByte('"')
				buf.WriteString(val)
				buf.WriteByte('"')
			} else {
				buf.WriteString(val)
			}
		}
	}

	buf.WriteString("\n")

	if stacktrace != "" {
		buf.WriteString(string(stacktrace))
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

var _levelToBracket = map[Level]string{
	Debug: "[DEBUG]",
	Trace: "[TRACE]",
	Info:  "[INFO] ",
	Warn:  "[WARN] ",
	Error: "[ERROR]",
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
