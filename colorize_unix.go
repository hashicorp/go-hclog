package hclog

import "io"

// colorize, on Unix systems, is a no-op, since unix streams accept
// ASCII color codes by default.
func colorize(output io.Writer) io.Writer {
	return output
}
