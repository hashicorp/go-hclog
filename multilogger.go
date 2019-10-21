package hclog

import (
	"io"
	"log"
)

type wrapperLogger struct {
	Logger Logger
}

// Args are alternating key, val pairs
// keys must be strings
// vals can be any type, but display is implementation specific
// Emit a message and key/value pairs at the TRACE level
func (w *wrapperLogger) Trace(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the DEBUG level
func (w *wrapperLogger) Debug(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the INFO level
func (w *wrapperLogger) Info(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the WARN level
func (w *wrapperLogger) Warn(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the ERROR level
func (w *wrapperLogger) Error(msg string, args ...interface{}) {
	panic("not implemented")
}

// Indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (w *wrapperLogger) IsTrace() bool {
	panic("not implemented")
}

// Indicate if DEBUG logs would be emitted. This and the other Is* guards
func (w *wrapperLogger) IsDebug() bool {
	panic("not implemented")
}

// Indicate if INFO logs would be emitted. This and the other Is* guards
func (w *wrapperLogger) IsInfo() bool {
	panic("not implemented")
}

// Indicate if WARN logs would be emitted. This and the other Is* guards
func (w *wrapperLogger) IsWarn() bool {
	panic("not implemented")
}

// Indicate if ERROR logs would be emitted. This and the other Is* guards
func (w *wrapperLogger) IsError() bool {
	panic("not implemented")
}

// Creates a sublogger that will always have the given key/value pairs
func (w *wrapperLogger) With(args ...interface{}) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (w *wrapperLogger) Named(name string) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (w *wrapperLogger) ResetNamed(name string) hclog.Logger {
	panic("not implemented")
}

// Updates the level. This should affect all sub-loggers as well. If an
// implementation cannot update the level on the fly, it should no-op.
func (w *wrapperLogger) SetLevel(level hclog.Level) {
	panic("not implemented")
}

// Return a value that conforms to the stdlib log.Logger interface
func (w *wrapperLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	panic("not implemented")
}

// Return a value that conforms to io.Writer, which can be passed into log.SetOutput()
func (w *wrapperLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("not implemented")
}

func InstallSub(root, sub Logger) {
	if ml, ok := root.(*MultiLogger); ok {
		ml.AddSub(sub)
	}
}

type MultiLogger struct {
	wrapperLogger

	Sub map[Logger]struct{}
}

func (m *MultiLogger) AddSub(sub Logger) {
	m.Sub[sub] = struct{}{}
}

// Args are alternating key, val pairs
// keys must be strings
// vals can be any type, but display is implementation specific
// Emit a message and key/value pairs at the TRACE level
func (m *MultiLogger) Trace(msg string, args ...interface{}) {
	m.Root.Trace(msg, args...)
}

// Emit a message and key/value pairs at the DEBUG level
func (m *MultiLogger) Debug(msg string, args ...interface{}) {
	m.Root.Trace(msg, args...)
}

// Emit a message and key/value pairs at the INFO level
func (m *MultiLogger) Info(msg string, args ...interface{}) {
	m.Root.Trace(msg, args...)
}

// Emit a message and key/value pairs at the WARN level
func (m *MultiLogger) Warn(msg string, args ...interface{}) {
	m.Root.Trace(msg, args...)
}

// Emit a message and key/value pairs at the ERROR level
func (m *MultiLogger) Error(msg string, args ...interface{}) {
	m.Root.Trace(msg, args...)

	for sub, _ := range m.Sub {
		sub.Error(msg, args...)
	}
}

// Indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (m *MultiLogger) IsTrace() bool {
	return m.lowest(Trace)
}

// Indicate if DEBUG logs would be emitted. This and the other Is* guards
func (m *MultiLogger) IsDebug() bool {
	panic("not implemented")
}

// Indicate if INFO logs would be emitted. This and the other Is* guards
func (m *MultiLogger) IsInfo() bool {
	panic("not implemented")
}

// Indicate if WARN logs would be emitted. This and the other Is* guards
func (m *MultiLogger) IsWarn() bool {
	panic("not implemented")
}

// Indicate if ERROR logs would be emitted. This and the other Is* guards
func (m *MultiLogger) IsError() bool {
	panic("not implemented")
}

// Creates a sublogger that will always have the given key/value pairs
func (m *MultiLogger) With(args ...interface{}) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (m *MultiLogger) Named(name string) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (m *MultiLogger) ResetNamed(name string) hclog.Logger {
	panic("not implemented")
}

// Updates the level. This should affect all sub-loggers as well. If an
// implementation cannot update the level on the fly, it should no-op.
func (m *MultiLogger) SetLevel(level hclog.Level) {
	panic("not implemented")
}

// Return a value that conforms to the stdlib log.Logger interface
func (m *MultiLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	panic("not implemented")
}

// Return a value that conforms to io.Writer, which can be passed into log.SetOutput()
func (m *MultiLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("not implemented")
}

type WithLogger struct {
	wrapperLogger
	args []interface{}
}

func (w *WithLogger) addArgs(args []interface{}) []interface{} {
	cp := make([]interface{}, len(w.args)+len(args))

	copy(cp, w.args)
	copy(cp, args)

	return cp
}

// Args are alternating key, val pairs
// keys must be strings
// vals can be any type, but display is implementation specific
// Emit a message and key/value pairs at the TRACE level
func (w *WithLogger) Trace(msg string, args ...interface{}) {
	w.upper.Trace(msg, w.addArgs(args)...)
}

// Emit a message and key/value pairs at the DEBUG level
func (w *WithLogger) Debug(msg string, args ...interface{}) {
	w.upper.Debug(msg, w.addArgs(args)...)
	panic("not implemented")
}

// Emit a message and key/value pairs at the INFO level
func (w *WithLogger) Info(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the WARN level
func (w *WithLogger) Warn(msg string, args ...interface{}) {
	panic("not implemented")
}

// Emit a message and key/value pairs at the ERROR level
func (w *WithLogger) Error(msg string, args ...interface{}) {
	panic("not implemented")
}

// Indicate if TRACE logs would be emitted. This and the other Is* guards
// are used to elide expensive logging code based on the current level.
func (w *WithLogger) IsTrace() bool {
	panic("not implemented")
}

// Indicate if DEBUG logs would be emitted. This and the other Is* guards
func (w *WithLogger) IsDebug() bool {
	panic("not implemented")
}

// Indicate if INFO logs would be emitted. This and the other Is* guards
func (w *WithLogger) IsInfo() bool {
	panic("not implemented")
}

// Indicate if WARN logs would be emitted. This and the other Is* guards
func (w *WithLogger) IsWarn() bool {
	panic("not implemented")
}

// Indicate if ERROR logs would be emitted. This and the other Is* guards
func (w *WithLogger) IsError() bool {
	panic("not implemented")
}

// Creates a sublogger that will always have the given key/value pairs
func (w *WithLogger) With(args ...interface{}) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all it's own logs
// without losing context.
func (w *WithLogger) Named(name string) hclog.Logger {
	panic("not implemented")
}

// Create a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (w *WithLogger) ResetNamed(name string) hclog.Logger {
	panic("not implemented")
}

// Updates the level. This should affect all sub-loggers as well. If an
// implementation cannot update the level on the fly, it should no-op.
func (w *WithLogger) SetLevel(level hclog.Level) {
	panic("not implemented")
}

// Return a value that conforms to the stdlib log.Logger interface
func (w *WithLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	panic("not implemented")
}

// Return a value that conforms to io.Writer, which can be passed into log.SetOutput()
func (w *WithLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	panic("not implemented")
}
