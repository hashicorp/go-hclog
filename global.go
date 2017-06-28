package log

import (
	"runtime"
	"sync"
)

var (
	protect    sync.Once
	def        Logger
	helperLock sync.Mutex
	helpers    map[string]struct{}

	// The options used to create the Default logger. These are
	// read only when the Default logger is created, so set them
	// as soon as the process starts.
	DefaultOptions = &LoggerOptions{
		Level:  DefaultLevel,
		Output: DefaultOutput,
	}
)

// Return a logger that is held globally. This can be a good starting
// place, and then you can use .With() and .Name() to create sub-loggers
// to be used in more specific contexts.
func Default() Logger {
	protect.Do(func() {
		def = New(DefaultOptions)
	})

	return def
}

// A short alias for Default()
func L() Logger {
	return Default()
}

// Helper marks the calling function as a log helper function when used with
// IncludeLocation. Instead of logging the line number in the helper func it
// will log the line number in the function calling the helper.
func Helper() {
	helperLock.Lock()
	defer helperLock.Unlock()
	if helpers == nil {
		helpers = make(map[string]struct{})
	}
	helpers[callerName(1)] = struct{}{}
}

// callerName gives the function name (qualified with a package path)
// for the caller after skip frames (where 0 means the current function).
//
// Copied from go stdlib testing package
func callerName(skip int) string {
	// Make room for the skip PC.
	var pc [2]uintptr
	n := runtime.Callers(skip+2, pc[:]) // skip + runtime.Callers + callerName
	if n == 0 {
		panic("testing: zero callers found")
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}
