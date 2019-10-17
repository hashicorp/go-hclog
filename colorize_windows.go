// +build windows

package hclog

import (
	"io"
	"os"

	colorable "github.com/mattn/go-colorable"
)

// colorize, on Windows systems, requires a wrapper to support
// Windows file streams
func colorize(output io.Writer) io.Writer {
	if fi, ok := output.(*os.File); ok {
		output = colorable.NewColorable(fi)
	} else {
		panic("Cannot enable coloring of non-file Writers")
	}
	return output
}
