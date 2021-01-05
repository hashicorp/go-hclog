package hclog

import (
	"bytes"
	"fmt"
	"io"
)

type writer struct {
	b     bytes.Buffer
	w     io.Writer
	color ColorOption
}

func newWriter(out io.Writer, color ColorOption) *writer {
	w := &writer{w: out, color: color}
	withColor(w)
	return w
}

func (w *writer) Flush(level Level) (err error) {
	msg := w.b.Bytes()

	if w.color != ColorOff {
		msg = _levelToColor[level](msg)
	}

	if lw, ok := w.w.(LevelWriter); ok {
		_, err = lw.LevelWrite(level, msg)
	} else {
		_, err = w.w.Write(msg)
	}
	w.b.Reset()
	return err
}

var _levelToColor = map[Level]func([]byte) []byte{
	Debug: newColor(252), // light grey
	Trace: newColor(82),  // bright green
	Info:  newColor(27),  // blue
	Warn:  newColor(226), // yellow
	Error: newColor(196), // red
}

func newColor(code uint) func([]byte) []byte {
	return func(src []byte) []byte {
		return []byte(fmt.Sprintf("\x1b[38;5;%dm%s\x1B[0m", code, src))
	}
}

func (w *writer) Write(p []byte) (int, error) {
	return w.b.Write(p)
}

func (w *writer) WriteByte(c byte) error {
	return w.b.WriteByte(c)
}

func (w *writer) WriteString(s string) (int, error) {
	return w.b.WriteString(s)
}

// LevelWriter is the interface that wraps the LevelWrite method.
type LevelWriter interface {
	LevelWrite(level Level, p []byte) (n int, err error)
}

// LeveledWriter writes all log messages to the standard writer,
// except for log levels that are defined in the overrides map.
type LeveledWriter struct {
	standard  io.Writer
	overrides map[Level]io.Writer
}

// NewLeveledWriter returns an initialized LeveledWriter.
//
// standard will be used as the default writer for all log levels,
// except for log levels that are defined in the overrides map.
func NewLeveledWriter(standard io.Writer, overrides map[Level]io.Writer) *LeveledWriter {
	return &LeveledWriter{
		standard:  standard,
		overrides: overrides,
	}
}

// Write implements io.Writer.
func (lw *LeveledWriter) Write(p []byte) (int, error) {
	return lw.standard.Write(p)
}

// LevelWrite implements LevelWriter.
func (lw *LeveledWriter) LevelWrite(level Level, p []byte) (int, error) {
	w, ok := lw.overrides[level]
	if !ok {
		w = lw.standard
	}
	return w.Write(p)
}
