package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func WrapperFunc(logger Logger, message string) {
	Helper()
	logger.Info(message)
}

func TestHelper(t *testing.T) {
	t.Run("uses Helper()to change caller location", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		// The test is going to assert based on the LINE NUMBER OF THIS CALL
		// so if you edit lines above you will need to change the assert below!
		WrapperFunc(logger, "sup")

		str := buf.String()

		dataIdx := strings.IndexByte(str, ' ')

		rest := str[dataIdx+1:]

		assert.Equal(t, "[INFO ] go-log/logger_line_test.go:29: test: sup\n", rest)
	})

	t.Run("uses Helper() to change caller location with JSON format", func(t *testing.T) {
		var buf bytes.Buffer

		logger := New(&LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
			JSONFormat:      true,
		})

		// The test is going to assert based on the LINE NUMBER OF THIS CALL
		// so if you edit lines above you will need to change the assert below!
		WrapperFunc(logger, "sup JSON")

		str := buf.String()

		type LogData struct {
			Caller string `json:"@caller"`
		}

		logData := &LogData{}

		if err := json.Unmarshal([]byte(str), logData); err != nil {
			t.Fatalf("Failed to convert JSON log data into struct: %s", err)
		}

		search := "github.com/hashicorp/go-log"
		dataIdx := strings.Index(logData.Caller, search)
		location := logData.Caller[dataIdx+len(search)+1:]

		assert.Equal(t, "logger_line_test.go:52", location)
	})

}
