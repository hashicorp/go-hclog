package hclog

import (
	"bytes"
	"testing"
)

func BenchmarkLoggerMemory(b *testing.B) {
	var buf bytes.Buffer

	logger := New(&LoggerOptions{
		Name:   "test",
		Output: &buf,
		Level:  Info,
	})

	for i := 0; i < b.N; i++ {
		logger.Trace("this is some message",
			"name", "foo",
			"what", "benchmarking yourself",
		)

		if buf.Len() != 0 {
			panic("oops")
		}
	}
}
