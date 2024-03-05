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

func TestLoggerMemory(t *testing.T) {
	var buf bytes.Buffer

	logger := New(&LoggerOptions{
		Name:   "test",
		Output: &buf,
		Level:  Info,
	})

	avg := testing.AllocsPerRun(100, func() {
		logger.Trace("this is some message",
			"name", "foo",
			"what", "benchmarking yourself",
		)

		if buf.Len() != 0 {
			panic("oops")
		}
	})

	if avg != 0 {
		t.Fatalf("ignored logs allocated")
	}
}
