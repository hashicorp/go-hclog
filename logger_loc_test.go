package hclog

import (
	"bytes"
	"strings"
	"testing"
)

// This file contains tests that are sensitive to their location in the file,
// because they contain line numbers. They're basically "quarantined" from the
// other tests because they break all the time when new tests are added.

func TestLoggerLoc(t *testing.T) {
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
		rest := str[dataIdx+1:]

		// This test will break if you move this around, it's line dependent, just fyi
		assertEqual(t, "[INFO]  go-hclog/logger_loc_test.go:23: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})

	t.Run("includes the caller location excluding helper functions", func(t *testing.T) {
		var buf bytes.Buffer

		logMe := func(l Logger) {
			l.Info("this is test", "who", "programmer", "why", "testing is fun")
		}

		logger := New(&LoggerOptions{
			Name:                     "test",
			Output:                   &buf,
			IncludeLocation:          true,
			AdditionalLocationOffset: 1,
		})

		logMe(logger)

		str := buf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]

		// This test will break if you move this around, it's line dependent, just fyi
		assertEqual(t, "[INFO]  go-hclog/logger_loc_test.go:47: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})

}
