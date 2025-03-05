package hclog

import (
	"runtime"
)

type Logger struct {
	impl LogImpl
}

func (v *Logger) log(level Level, msg string, args ...interface{}) {
	if level < v.impl.GetLevel() {
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

// Emit the message and args at the given level
func (v *Logger) Log(level Level, msg string, args ...interface{}) {
	v.log(level, msg, args...)
}

// Emit the message and args at TRACE level
func (v *Logger) Trace(msg string, args ...interface{}) {
	v.log(Trace, msg, args...)
}

// Emit the message and args at DEBUG level
func (v *Logger) Debug(msg string, args ...interface{}) {
	v.log(Debug, msg, args...)
}

// Emit the message and args at INFO level
func (v *Logger) Info(msg string, args ...interface{}) {
	v.log(Info, msg, args...)
}

// Emit the message and args at WARN level
func (v *Logger) Warn(msg string, args ...interface{}) {
	v.log(Warn, msg, args...)
}

// Emit the message and args at ERROR level
func (v Logger) Error(msg string, args ...interface{}) {
	v.log(Error, msg, args...)
}

func (v Logger) IsTrace() bool {
	return Debug >= v.impl.GetLevel()
}

func (v Logger) IsDebug() bool {
	return Debug >= v.impl.GetLevel()
}

func (v Logger) IsInfo() bool {
	return Debug >= v.impl.GetLevel()
}

func (v Logger) IsWarn() bool {
	return Debug >= v.impl.GetLevel()
}

func (v Logger) IsError() bool {
	return Debug >= v.impl.GetLevel()
}

func (v Logger) Named(name string) Logger {
	return Logger{
		v.impl.Named(name),
	}
}

func (v Logger) ResetNamed(name string) Logger {
	return Logger{
		v.impl.ResetNamed(name),
	}
}

func (v Logger) With(args ...interface{}) Logger {
	return Logger{
		v.impl.With(args...),
	}
}

func (v *Logger) SetLevel(level Level) {
	v.impl.SetLevel(level)
}

func (v *Logger) GetLevel() Level {
	return v.impl.GetLevel()
}

func AsValue(log LogImpl) Logger {
	return Logger{
		impl: log,
	}
}
