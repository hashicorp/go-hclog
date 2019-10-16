package hclog

import (
	"io"
	"log"
	"sync"
)

type MultiLogger interface {
	Tee(*LoggerOptions) Logger
	CloseTee(Logger)

	Logger
}

type multiLogger struct {
	// opts is the default logger options
	opts *LoggerOptions

	// root embedded logger that cannot be deregistered
	root Logger

	sinks map[Logger]struct{}
	mu    sync.RWMutex
}

// NewMulti returns a configured logger capable of being teed into multiple
// outputs.
func NewMulti(opts *LoggerOptions) MultiLogger {
	m := &multiLogger{
		opts:  opts,
		root:  New(opts),
		sinks: make(map[Logger]struct{}),
	}
	return m
}

// Tee a new Logger with distinct options from the root logger. Caller must
// call CloseTee when done.
func (m *multiLogger) Tee(opts *LoggerOptions) Logger {
	l := New(opts)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sinks[l] = struct{}{}

	return l
}

// CloseTee stops sending log lines to a teed logger.
func (m *multiLogger) CloseTee(l Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sinks, l)
}

func (m *multiLogger) Trace(msg string, args ...interface{}) {
	m.root.Log(Trace, msg, args...)
}

func (m *multiLogger) Debug(msg string, args ...interface{}) {
	m.root.Log(Debug, msg, args...)
}

func (m *multiLogger) Info(msg string, args ...interface{}) {
	m.root.Log(Info, msg, args...)
}

func (m *multiLogger) Warn(msg string, args ...interface{}) {
	m.root.Log(Warn, msg, args...)
}

func (m *multiLogger) Error(msg string, args ...interface{}) {
	m.root.Log(Error, msg, args...)
}

// IsTrace returns true if any loggers would emit TRACE level logs
func (m *multiLogger) IsTrace() bool {
	if m.root.IsTrace() {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for l := range m.sinks {
		if l.IsTrace() {
			return true
		}
	}

	return false
}

// IsDebug returns true if any loggers would emit DEBUG level logs
func (m *multiLogger) IsDebug() bool {
	if m.root.IsDebug() {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.sinks {
		if l.IsDebug() {
			return true
		}
	}

	return false
}

// IsInfo returns true if any loggers would emit INFO level logs
func (m *multiLogger) IsInfo() bool {
	if m.root.IsTrace() {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.sinks {
		if l.IsTrace() {
			return true
		}
	}

	return false
}

// IsWarn returns true if any loggers would emit WARN level logs
func (m *multiLogger) IsWarn() bool {
	if m.root.IsWarn() {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.sinks {
		if l.IsWarn() {
			return true
		}
	}

	return false
}

// IsError returns true if any loggers would emit WARN level logs
func (m *multiLogger) IsError() bool {
	if m.root.IsError() {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, l := range m.sinks {
		if l.IsError() {
			return true
		}
	}

	return false
}

// Return a sub-Logger of the root Logger.
func (m *multiLogger) With(args ...interface{}) Logger {
	return m.root.With(args...)
}

// Create a new sub-Logger of the root Logger with a name descending from the
// current name.  This is used to create a subsystem specific Logger.
func (m *multiLogger) Named(name string) Logger {
	return m.root.Named(name)
}

// Create a new sub-Logger of the root Logger with an explicit name. This
// ignores the current name. This is used to create a standalone logger that
// doesn't fall within the normal naming hierarchy.
func (m *multiLogger) ResetNamed(name string) Logger {
	return m.root.ResetNamed(name)
}

// Update the logging level of the root Logger on-the-fly. This will affect all
// subloggers but not tees.
func (m *multiLogger) SetLevel(level Level) {
	m.root.SetLevel(level)
}

// Create a *log.Logger that will send it's data through the this Logger. This
// allows packages that expect to be using the standard library log to actually
// use this logger.
func (m *multiLogger) StandardLogger(opts *StandardLoggerOptions) *log.Logger {
	if opts == nil {
		opts = &StandardLoggerOptions{}
	}

	return log.New(l.StandardWriter(opts), "", 0)
}

func (m *multiLogger) StandardWriter(opts *StandardLoggerOptions) io.Writer {
	return &stdlogAdapter{
		log:         n,
		inferLevels: opts.InferLevels,
		forceLevel:  opts.ForceLevel,
	}
}
