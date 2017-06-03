package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	t.Run("formats log entries", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(LoggerOptions{
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

		logger := New(LoggerOptions{
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

		logger := New(LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Stacktrace("who", "programmer", "why", "testing")

		lines := strings.Split(buf.String(), "\n")

		assert.Equal(t, "github.com/hashicorp/log.(*intLogger).Stacktrace", lines[1])
	})

	t.Run("includes the caller location", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(LoggerOptions{
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
		assert.Equal(t, "[INFO ] log/logger_test.go:76: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})
}
