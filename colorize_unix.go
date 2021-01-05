// +build !windows

package hclog

import (
	"os"

	"github.com/mattn/go-isatty"
)

func withColor(w *writer) {
	switch w.color {
	case ColorOff, ForceColor:
		return
	case AutoColor:
		fd := os.Stdout.Fd()
		if isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd) {
			return
		}
		w.color = ColorOff
	}
}
