package hclog

import (
	"runtime"
)

type Logger struct {
	impl LogImpl
}

func (v *Logger) log(level Level, msg string, args ...interface{}) {
	// Return early if the level is set to 'off' or is a lower level than the logger is configured for.
	if level == Off || level < v.impl.Level() {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, this function's caller]
	runtime.Callers(3, pcs[:])
	pc = pcs[0]

	r := Record{
		Level:    level,
		Msg:      msg,
		CallerPc: pc,
	}

	r.SetArgs(args)

	v.impl.LogRecord(r)
}

// Log emits a message and key/value pairs at a provided log level.
// Args are alternating key, val pairs.
// Keys must be strings.
// Vals can be any type, but display is implementation specific.
func (v *Logger) Log(level Level, msg string, args ...interface{}) {
	v.log(level, msg, args...)
}

// Trace emits a message and key/value pairs at the TRACE level.
func (v *Logger) Trace(msg string, args ...interface{}) {
	v.log(Trace, msg, args...)
}

// Debug emits a message and key/value pairs at the DEBUG level.
func (v *Logger) Debug(msg string, args ...interface{}) {
	v.log(Debug, msg, args...)
}

// Info emits a message and key/value pairs at the INFO level.
func (v *Logger) Info(msg string, args ...interface{}) {
	v.log(Info, msg, args...)
}

// Warn emits a message and key/value pairs at the WARN level.
func (v *Logger) Warn(msg string, args ...interface{}) {
	v.log(Warn, msg, args...)
}

// Error emits a message and key/value pairs at the ERROR level.
func (v *Logger) Error(msg string, args ...interface{}) {
	v.log(Error, msg, args...)
}

// IsTrace indicates if TRACE logs would be emitted.
// This and the other Is* guards are used to elide expensive logging code based on the current level.
func (v *Logger) IsTrace() bool {
	return v.Level() <= Trace
}

// IsDebug indicates if DEBUG logs would be emitted. This and the other Is* guards.
// This and the other Is* guards are used to elide expensive logging code based on the current level.
func (v *Logger) IsDebug() bool {
	return v.Level() <= Debug
}

// IsInfo indicates if INFO logs would be emitted. This and the other Is* guards.
// This and the other Is* guards are used to elide expensive logging code based on the current level.
func (v *Logger) IsInfo() bool {
	return v.Level() <= Info
}

// IsWarn indicates if WARN logs would be emitted. This and the other Is* guards.
// This and the other Is* guards are used to elide expensive logging code based on the current level.
func (v *Logger) IsWarn() bool {
	return v.Level() <= Warn
}

// IsError indicates if ERROR logs would be emitted. This and the other Is* guards.
// This and the other Is* guards are used to elide expensive logging code based on the current level.
func (v *Logger) IsError() bool {
	return v.Level() <= Error
}

// Named creates a logger that will prepend the name string on the front of all messages.
// If the logger already has a name, the new value will be appended to the current
// name. That way, a major subsystem can use this to decorate all its own logs
// without losing context.
func (v *Logger) Named(name string) Logger {
	return Logger{
		v.impl.Named(name),
	}
}

// ResetNamed creates a logger that will prepend the name string on the front of all messages.
// This sets the name of the logger to the value directly, unlike Named which honor
// the current name as well.
func (v *Logger) ResetNamed(name string) Logger {
	return Logger{
		v.impl.ResetNamed(name),
	}
}

// With creates a sublogger that will always have the given key/value pairs.
func (v *Logger) With(args ...interface{}) Logger {
	return Logger{
		v.impl.With(args...),
	}
}

// SetLevel updates the level. This should affect all related loggers as well,
// unless they were created with IndependentLevels. If an
// implementation cannot update the level on the fly, it should no-op.
func (v *Logger) SetLevel(level Level) {
	v.impl.SetLevel(level)
}

// Level returns the current level.
func (v *Logger) Level() Level {
	return v.impl.Level()
}
