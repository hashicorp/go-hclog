package log

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	_levelToBracket = map[Level]string{
		Debug: "[DEBUG]",
		Trace: "[TRACE]",
		Info:  "[INFO ]",
		Warn:  "[WARN ]",
		Error: "[ERROR]",
	}
)

func New(opts LoggerOptions) Logger {

	output := opts.Output
	if output == nil {
		output = os.Stdout
	}

	return &intLogger{
		m:      new(sync.Mutex),
		json:   opts.JSONFormat,
		caller: opts.IncludeLocation,
		name:   opts.Name,
		w:      bufio.NewWriter(output),
		level:  opts.Level,
	}
}

type intLogger struct {
	json   bool
	caller bool
	name   string

	// this is a pointer so that it's shared by any derived loggers, since
	// those derived loggers share the bufio.Writer as well.
	m     *sync.Mutex
	w     *bufio.Writer
	level Level

	implied []interface{}
}

var _ Logger = &intLogger{}

const TimeFormat = "2006-01-02T15:04:05.000Z0700"

// Log a message and a set of key/value pairs if the given level is at
// or more severe that the threshold configured in the Logger.
func (z *intLogger) Log(level Level, msg string, args ...interface{}) {
	if level < z.level {
		return
	}

	t := time.Now()

	z.m.Lock()
	defer z.m.Unlock()

	if z.json {
		z.logJson(t, level, msg, args...)
	} else {
		z.log(t, level, msg, args...)
	}

	z.w.Flush()
}

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
	//

	// Find the last separator.
	//
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

func (z *intLogger) log(t time.Time, level Level, msg string, args ...interface{}) {
	z.w.WriteString(t.Format(TimeFormat))
	z.w.WriteByte(' ')

	s, ok := _levelToBracket[level]
	if ok {
		z.w.WriteString(s)
	} else {
		z.w.WriteString("[UNKN ]")
	}

	if z.caller {
		if _, file, line, ok := runtime.Caller(3); ok {
			z.w.WriteByte(' ')
			z.w.WriteString(trimCallerPath(file))
			z.w.WriteByte(':')
			z.w.WriteString(strconv.Itoa(line))
			z.w.WriteByte(':')
		}
	}

	z.w.WriteByte(' ')

	if z.name != "" {
		z.w.WriteString(z.name)
		z.w.WriteString(": ")
	}

	z.w.WriteString(msg)

	args = append(z.implied, args...)

	if args != nil && len(args) > 0 {
		if len(args)%2 != 0 {
			args = append(args, "<unknown>")
		}

		z.w.WriteByte(':')

		for i := 0; i < len(args); i = i + 2 {
			val := fmt.Sprintf("%v", args[i+1])

			var quote string
			if strings.ContainsAny(val, " \t\n\r") {
				quote = `"`
			}

			fmt.Fprintf(z.w, " %s=%s%v%s", args[i], quote, val, quote)
		}
	}

	z.w.WriteString("\n")
}

func (z *intLogger) logJson(t time.Time, level Level, msg string, args ...interface{}) {
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

	if z.name != "" {
		vals["@module"] = z.name
	}

	if z.caller {
		if _, file, line, ok := runtime.Caller(3); ok {
			vals["@caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}

	if args != nil && len(args) > 0 {

		if len(args)%2 != 0 {
			args = append(args, "<unknown>")
		}

		for i := 0; i < len(args); i = i + 2 {
			if _, ok := args[i].(string); !ok {
				// As this is the logging function not much we can do here
				// without injecting into logs...
				continue
			}
			vals[args[i].(string)] = args[i+1]
		}
	}

	err := json.NewEncoder(z.w).Encode(vals)
	if err != nil {
		panic(err)
	}
}

func (z *intLogger) Debug(msg string, args ...interface{}) {
	z.Log(Debug, msg, args...)
}

func (z *intLogger) Trace(msg string, args ...interface{}) {
	z.Log(Trace, msg, args...)
}

func (z *intLogger) Info(msg string, args ...interface{}) {
	z.Log(Info, msg, args...)
}

func (z *intLogger) Warn(msg string, args ...interface{}) {
	z.Log(Warn, msg, args...)
}

func (z *intLogger) Error(msg string, args ...interface{}) {
	z.Log(Error, msg, args...)
}

func (z *intLogger) IsTrace() bool {
	return z.level >= Trace
}

func (z *intLogger) IsDebug() bool {
	return z.level >= Debug
}

func (z *intLogger) IsInfo() bool {
	return z.level >= Info
}

func (z *intLogger) IsWarn() bool {
	return z.level >= Warn
}

func (z *intLogger) IsError() bool {
	return z.level >= Error
}

func (z *intLogger) With(args ...interface{}) Logger {
	var nz intLogger = *z

	nz.implied = args

	return &nz
}

func (z *intLogger) Named(name string) Logger {
	var nz intLogger = *z

	if nz.name != "" {
		nz.name = nz.name + "." + name
	}

	return &nz
}

func (z *intLogger) Stacktrace(args ...interface{}) {
	t := time.Now()

	z.m.Lock()
	defer z.m.Unlock()

	if z.json {
		var buf bytes.Buffer

		sw := bufio.NewWriter(&buf)
		writeStacktrace(sw)
		sw.Flush()

		args = append(args, "stacktrace", buf.String())

		z.logJson(t, Trace, "", args...)
	} else {
		z.log(t, Trace, "", args...)
		writeStacktrace(z.w)
	}

	z.w.Flush()
}

func (z *intLogger) StandardLogger(inferLevels bool) *log.Logger {
	return log.New(&stdlogAdapter{z, inferLevels}, "", 0)
}
