package hclog

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromStandardSLogger(t *testing.T) {
	newSimple := func(level slog.Leveler, buf *bytes.Buffer) Logger {
		var lvar *slog.LevelVar = nil
		if v, ok := level.(*slog.LevelVar); ok {
			lvar = v
		}
		return FromStandardSLogger(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{}
				}
				return a
			},
		})), lvar)
	}
	t.Run("trace", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(SlogLevelTrace, &buf)
		logger.Trace("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, "level=DEBUG-4 msg=\"this is test\" who=programmer why=testing\n", buf.String())

		assert.Equal(t, Trace, logger.GetLevel())
		assert.Equal(t, true, logger.IsTrace())
		assert.Equal(t, true, logger.IsDebug())
		assert.Equal(t, true, logger.IsInfo())
		assert.Equal(t, true, logger.IsWarn())
		assert.Equal(t, true, logger.IsError())
	})
	t.Run("debug", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelDebug, &buf)
		logger.Debug("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, "level=DEBUG msg=\"this is test\" who=programmer why=testing\n", buf.String())

		assert.Equal(t, Debug, logger.GetLevel())
		assert.Equal(t, false, logger.IsTrace())
		assert.Equal(t, true, logger.IsDebug())
		assert.Equal(t, true, logger.IsInfo())
		assert.Equal(t, true, logger.IsWarn())
		assert.Equal(t, true, logger.IsError())
	})
	t.Run("info", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelInfo, &buf)
		logger.Info("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, "level=INFO msg=\"this is test\" who=programmer why=testing\n", buf.String())

		assert.Equal(t, Info, logger.GetLevel())
		assert.Equal(t, false, logger.IsTrace())
		assert.Equal(t, false, logger.IsDebug())
		assert.Equal(t, true, logger.IsInfo())
		assert.Equal(t, true, logger.IsWarn())
		assert.Equal(t, true, logger.IsError())
	})
	t.Run("warn", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelWarn, &buf)
		logger.Warn("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, "level=WARN msg=\"this is test\" who=programmer why=testing\n", buf.String())

		assert.Equal(t, Warn, logger.GetLevel())
		assert.Equal(t, false, logger.IsTrace())
		assert.Equal(t, false, logger.IsDebug())
		assert.Equal(t, false, logger.IsInfo())
		assert.Equal(t, true, logger.IsWarn())
		assert.Equal(t, true, logger.IsError())
	})
	t.Run("error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelError, &buf)
		logger.Error("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, "level=ERROR msg=\"this is test\" who=programmer why=testing\n", buf.String())

		assert.Equal(t, Error, logger.GetLevel())
		assert.Equal(t, false, logger.IsTrace())
		assert.Equal(t, false, logger.IsDebug())
		assert.Equal(t, false, logger.IsInfo())
		assert.Equal(t, false, logger.IsWarn())
		assert.Equal(t, true, logger.IsError())
	})
	t.Run("off", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(SlogLevelOff, &buf)
		logger.Error("this is test", "who", "programmer", "why", "testing")
		assert.Equal(t, 0, buf.Len())

		assert.Equal(t, Off, logger.GetLevel())
		assert.Equal(t, false, logger.IsTrace())
		assert.Equal(t, false, logger.IsDebug())
		assert.Equal(t, false, logger.IsInfo())
		assert.Equal(t, false, logger.IsWarn())
		assert.Equal(t, false, logger.IsError())
	})
	t.Run("var level", func(t *testing.T) {
		var buf bytes.Buffer
		var lvar slog.LevelVar
		logger := newSimple(&lvar, &buf)

		// Default LevelVar is Info, so it won't log trace msg
		logger.Trace("trace msg")
		assert.Equal(t, 0, buf.Len())
		assert.Equal(t, Info, logger.GetLevel())

		// Info message is logged
		logger.Info("info msg")
		assert.Equal(t, "level=INFO msg=\"info msg\"\n", buf.String())
		assert.Equal(t, Info, logger.GetLevel())

		// Set it to trace level to log trace msg
		buf.Reset()
		lvar.Set(SlogLevelTrace)
		logger.Trace("trace msg")
		assert.Equal(t, "level=DEBUG-4 msg=\"trace msg\"\n", buf.String())
		assert.Equal(t, Trace, logger.GetLevel())
	})
	t.Run("name related", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelInfo, &buf)
		n1Logger := logger.Named("n1")
		n2Logger := n1Logger.Named("n2")

		assert.Equal(t, "", logger.Name())
		assert.Equal(t, "n1", n1Logger.Name())
		assert.Equal(t, "n1.n2", n2Logger.Name())

		logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg age=1\n", buf.String())

		buf.Reset()
		n1Logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg n1.age=1\n", buf.String())

		buf.Reset()
		n2Logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg n1.n2.age=1\n", buf.String())

		n2Logger = n2Logger.ResetNamed("")
		buf.Reset()
		n2Logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg age=1\n", buf.String())
		assert.Equal(t, "", n2Logger.Name())

		n1Logger = n1Logger.ResetNamed("n11")
		buf.Reset()
		n1Logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg n11.age=1\n", buf.String())
		assert.Equal(t, "n11", n1Logger.Name())

		logger = logger.ResetNamed("n0")
		buf.Reset()
		logger.Info("msg", "age", 1)
		assert.Equal(t, "level=INFO msg=msg n0.age=1\n", buf.String())
		assert.Equal(t, "n0", logger.Name())
	})
	t.Run("with args", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(slog.LevelInfo, &buf)
		logger = logger.With("foo", "bar")
		assert.Equal(t, []interface{}{"foo", "bar"}, logger.ImpliedArgs())
		logger.Info("msg")
		assert.Equal(t, "level=INFO msg=msg foo=bar\n", buf.String())
	})
	t.Run("standard logger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := newSimple(SlogLevelTrace, &buf)

		stdlogger := logger.StandardLogger(&StandardLoggerOptions{InferLevels: true})
		stdwriter := logger.StandardWriter(&StandardLoggerOptions{InferLevels: true})

		cases := []struct {
			input  string
			expect string
		}{
			{
				input:  "[TRACE] msg",
				expect: "level=DEBUG-4 msg=msg\n",
			},
			{
				input:  "[DEBUG] msg",
				expect: "level=DEBUG msg=msg\n",
			},
			{
				input:  "[INFO] msg",
				expect: "level=INFO msg=msg\n",
			},
			{
				input:  "[WARN] msg",
				expect: "level=WARN msg=msg\n",
			},
			{
				input:  "[ERROR] msg",
				expect: "level=ERROR msg=msg\n",
			},
		}

		for _, cc := range cases {
			buf.Reset()
			stdlogger.Println(cc.input)
			assert.Equal(t, cc.expect, buf.String())

			buf.Reset()
			stdwriter.Write([]byte(cc.input))
			assert.Equal(t, cc.expect, buf.String())
		}
	})
}
