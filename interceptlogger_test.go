package hclog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterceptLogger(t *testing.T) {
	t.Run("sends output to registered sinks", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})

		intercept.RegisterSink(sink)
		defer intercept.DeregisterSink(sink)

		intercept.Debug("test log", "who", "programmer")

		str := sbuf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]

		assert.Equal(t, "[DEBUG] test log: who=programmer\n", rest)

	})

	t.Run("sink includes with arguments", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Name:   "with_test",
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})
		intercept.RegisterSink(sink)
		defer intercept.DeregisterSink(sink)

		derived := intercept.With("a", 1, "b", 2)
		derived = derived.With("c", 3)

		derived.Info("test1")
		output := buf.String()
		dataIdx := strings.IndexByte(output, ' ')
		rest := output[dataIdx+1:]

		assert.Equal(t, "[INFO]  with_test: test1: a=1 b=2 c=3\n", rest)

		// Ensure intercept works
		output = sbuf.String()
		dataIdx = strings.IndexByte(output, ' ')
		rest = output[dataIdx+1:]

		assert.Equal(t, "[INFO]  with_test: test1: a=1 b=2 c=3\n", rest)
	})

	t.Run("sink includes name", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Name:   "with_test",
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})
		intercept.RegisterSink(sink)
		defer intercept.DeregisterSink(sink)

		httpLogger := intercept.Named("http")

		httpLogger.Info("test1")
		output := buf.String()
		dataIdx := strings.IndexByte(output, ' ')
		rest := output[dataIdx+1:]

		assert.Equal(t, "[INFO]  with_test.http: test1\n", rest)

		// Ensure intercept works
		output = sbuf.String()
		dataIdx = strings.IndexByte(output, ' ')
		rest = output[dataIdx+1:]

		assert.Equal(t, "[INFO]  with_test.http: test1\n", rest)
	})

	t.Run("intercepting logger can create logger with reset name", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Name:   "with_test",
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})
		intercept.RegisterSink(sink)
		defer intercept.DeregisterSink(sink)

		httpLogger := intercept.ResetNamed("http")

		httpLogger.Info("test1")
		output := buf.String()
		dataIdx := strings.IndexByte(output, ' ')
		rest := output[dataIdx+1:]

		assert.Equal(t, "[INFO]  http: test1\n", rest)

		// Ensure intercept works
		output = sbuf.String()
		dataIdx = strings.IndexByte(output, ' ')
		rest = output[dataIdx+1:]

		assert.Equal(t, "[INFO]  http: test1\n", rest)
	})

	t.Run("Intercepting logger sink can deregister itself", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Name:   "with_test",
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})
		intercept.RegisterSink(sink)
		intercept.DeregisterSink(sink)

		intercept.Info("test1")

		assert.Equal(t, "", sbuf.String())
	})

	t.Run("Sinks accept different log formats", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		intercept := NewInterceptLogger(&LoggerOptions{
			Level:  Info,
			Output: &buf,
		})

		sink := NewSinkAdapter(&LoggerOptions{
			Level:      Debug,
			Output:     &sbuf,
			JSONFormat: true,
		})

		intercept.RegisterSink(sink)
		defer intercept.DeregisterSink(sink)

		intercept.Info("this is a test", "who", "caller")
		output := buf.String()
		dataIdx := strings.IndexByte(output, ' ')
		rest := output[dataIdx+1:]

		assert.Equal(t, "[INFO]  this is a test: who=caller\n", rest)

		b := sbuf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "this is a test", raw["@message"])
		assert.Equal(t, "caller", raw["who"])
	})
}
