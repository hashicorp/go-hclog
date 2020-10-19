// +build windows

package hclog

import (
	"os"

	colorable "github.com/mattn/go-colorable"
)

func withColor(w *writer) {
	switch w.color {
	case ColorOff:
		return
	case ForceColor, AutoColor:
		if fi, ok := w.w.(*os.File); ok {
			w.w = colorable.NewColorable(fi)
			return
		}
		w.color = ColorOff
	}
}
