package testsub

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

func TestSub(t *testing.T) {
	t.Run("includes the caller location", func(t *testing.T) {
		var buf bytes.Buffer

		logger := hclog.New(&hclog.LoggerOptions{
			Name:            "test",
			Output:          &buf,
			IncludeLocation: true,
		})

		DoWork(logger)

		str := buf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]

		// This test will break if you move this around, it's line dependent, just fyi
		assert.Equal(t, "[INFO]  testsub/work.go:6: test: this is test: who=programmer why=\"testing is fun\"\n", rest)
	})
}
