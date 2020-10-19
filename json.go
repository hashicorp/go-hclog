package hclog

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

type jsonFormat struct {
	timeFormat   string
	callerOffset int
}

func (j jsonFormat) Format(name string, level Level, msg string, args ...interface{}) []byte {
	t := time.Now()
	vals := j.jsonMapEntry(t, name, level, msg)

	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			cs, ok := args[len(args)-1].(CapturedStacktrace)
			if ok {
				args = args[:len(args)-1]
				vals["stacktrace"] = cs
			} else {
				extra := args[len(args)-1]
				args = append(args[:len(args)-1], MissingKey, extra)
			}
		}

		for i := 0; i < len(args); i = i + 2 {
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

			var key string

			switch st := args[i].(type) {
			case string:
				key = st
			default:
				key = fmt.Sprintf("%s", st)
			}
			vals[key] = val
		}
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(vals)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			plainVal := j.jsonMapEntry(t, name, level, msg)
			plainVal["@warn"] = errJsonUnsupportedTypeMsg

			json.NewEncoder(buf).Encode(plainVal)
		}
	}
	return buf.Bytes()
}

func (j jsonFormat) jsonMapEntry(t time.Time, name string, level Level, msg string) map[string]interface{} {
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

	if name != "" {
		vals["@module"] = name
	}

	if j.callerOffset > 0 {
		if _, file, line, ok := runtime.Caller(j.callerOffset + 1); ok {
			vals["@caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}
	return vals
}
