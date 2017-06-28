package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func WrapperFunc(logger Logger, message string) {
	Helper()
	logger.Info(message)
}

func TestHelper(t *testing.T) {
	t.Run("formats log entries", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		// The test is going to assert based on the LINE NUMBER OF THIS CALL
		// so if you edit lines above you will need to change the assert below!
		logger.Info("this is test", "who", "programmer", "why", "testing")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] go-log/logger_line_test.go:28: test: this is test: who=programmer why=testing\n", rest)
	})

}
