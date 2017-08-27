package hclog

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Run("formats log entries", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] test: this is test: who=programmer why=testing\n", rest)
	})

	t.Run("quotes values with spaces", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})

	t.Run("outputs stack traces", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("who", "programmer", "why", "testing", Stacktrace())

		lines := strings.Split(buf.String(), "\n")

		require.True(t, len(lines) > 1)

		assert.Equal(t, "github.com/hashicorp/go-hclog.Stacktrace", lines[1])
	})

	t.Run("outputs stack traces with it's given a name", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("who", "programmer", "why", "testing", "foo", Stacktrace())

		lines := strings.Split(buf.String(), "\n")

		require.True(t, len(lines) > 1)

		assert.Equal(t, "github.com/hashicorp/go-hclog.Stacktrace", lines[1])
	})

	t.Run("includes the caller location", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		// This test will break if you move this around, it's line dependent, just fyi
		assert.Equal(t, "[INFO ] go-hclog/logger_test.go:98: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})
}

func TestLogger_JSON(t *testing.T) {
	t.Run("json formatting", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing is fun")

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, raw["@message"], "this is test")
		assert.Equal(t, raw["who"], "programmer")
		assert.Equal(t, raw["why"], "testing is fun")
	})
	t.Run("json formatting error type", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:       "test",
			Output:     &buf,
			JSONFormat: true,
		})

		errMsg := errors.New("this is an error")
		logger.Info("this is test", "who", "programmer", "err", errMsg)

		b := buf.Bytes()

		var raw map[string]interface{}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, raw["@message"], "this is test")
		assert.Equal(t, raw["who"], "programmer")
		assert.Equal(t, raw["err"], errMsg.Error())
	})
}

func BenchmarkLogger(b *testing.B) {
	b.Run("info with 10 pairs", func(b *testing.B) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		for i := 0; i < b.N; i++ {
			logger.Info("this is some message",
				"name", "foo",
				"what", "benchmarking yourself",
				"why", "to see what's slow",
				"k4", "value",
				"k5", "value",
				"k6", "value",
				"k7", "value",
				"k8", "value",
				"k9", "value",
				"k10", "value",
			)
		}
	})
}
