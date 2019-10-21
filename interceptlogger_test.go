package hclog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInterceptLogger(t *testing.T) {
	t.Run("sends output to registered sinks", func(t *testing.T) {
		var buf bytes.Buffer
		var sbuf bytes.Buffer

		root := New(&LoggerOptions{
			Level:  Info,
			Output: &buf,
		})

		intercept := NewInterceptLogger(root)

		unsubscribe := intercept.SubscribeWith(&LoggerOptions{
			Level:  Debug,
			Output: &sbuf,
		})
		defer unsubscribe()

		intercept.Debug("test log", "who", "programmer")

		str := sbuf.String()
		dataIdx := strings.IndexByte(str, ' ')
		rest := str[dataIdx+1:]

		assert.Equal(t, "[DEBUG] test log: who=programmer\n", rest)

	})
}
