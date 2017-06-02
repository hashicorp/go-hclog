package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZap(t *testing.T) {
	t.Run("formats log entries", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing")

		str := buf.String()

		dataIdx := strings.IndexByte(str, '\t')

		// ts := str[:dataIdx]
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ]\ttest:\tthis is test\t{\"who\": \"programmer\", \"why\": \"testing\"}\n", rest)
	})

	t.Run("outputs stack traces", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(LoggerOptions{
			Name:   "test",
			Output: &buf,
		})

		logger.Stacktrace("who", "programmer", "why", "testing")

		lines := strings.Split(buf.String(), "\n")

		assert.Equal(t, "go.uber.org/zap.Stack", lines[1])
	})
}
