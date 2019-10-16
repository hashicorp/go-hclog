package hclog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiLogger(t *testing.T) {
	t.Run("acts as normal logger by default", func(t *testing.T) {
		var buf bytes.Buffer

		logger := NewMulti(&LoggerOptions{
			Name:   "test",
			Output: buf,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing")

		str := buf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO]  test: this is test: who=programmer why=testing\n", rest)
	})

	t.Run("sends output to multiple loggers", func(t *testing.T) {
		var buf1 bytes.Buffer
		var buf2 bytes.Buffer

		logger := NewMulti(&LoggerOptions{
			Name:   "test",
			Output: buf1,
		})

		tee := logger.Tee(&LoggerOptions{
			Output: buf2,
		})

		logger.Info("this is test", "who", "programmer", "why", "testing")

		rest := bytes.SplitN(buf1.Bytes(), ' ', 2)[1]
		assert.Equal(t, "[INFO]  test: this is test: who=programmer why=testing\n", string(rest))

		rest := bytes.SplitN(buf2.Bytes(), ' ', 2)[1]
		assert.Equal(t, "[INFO]  test: this is test: who=programmer why=testing\n", string(rest))
	})
}
