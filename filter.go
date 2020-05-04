package hclog

import "strings"

// MessageFilter provides a simple way to build a list of log messages that
// can be queried and matched. This is meant to be used with the Filter
// option on Options to suppress log messages. This does not hold any mutexs
// within itself, so normal usage would be to Add entries at setup and none after
// FilterOut is going to be called. FilterOut is called with a mutex held within
// the Logger, so that doesn't need to use a mutex. Example usage:
//
//	f := new(MessageFilter)
//	f.Add("Noisy log message text")
//	appLogger.Filter = f.FilterOut
type MessageFilter struct {
	messages map[string]struct{}
}

// Add a message to be filtered. Do not call this after FilterOut is to be called
// due to concurrency issues.
func (f *MessageFilter) Add(msg string) {
	if f.messages == nil {
		f.messages = make(map[string]struct{})
	}

	f.messages[msg] = struct{}{}
}

// Return true if the given message is known.
func (f *MessageFilter) FilterOut(level Level, msg string, args ...interface{}) bool {
	_, ok := f.messages[msg]
	return ok
}

// PrefixFilter is a simple type to match a message string that has a common prefix.
type PrefixFilter string

// Matches a message that starts with the prefix.
func (p PrefixFilter) FilterOut(level Level, msg string, args ...interface{}) bool {
	return strings.HasPrefix(msg, string(p))
}

// FilterFuncs is a slice of functions that will called to see if a log entry
// should be filtered or not. It stops calling functions once at least one returns
// true.
type FilterFuncs []func(level Level, msg string, args ...interface{}) bool

// Calls each function until one of them returns true
func (ff FilterFuncs) FilterOut(level Level, msg string, args ...interface{}) bool {
	for _, f := range ff {
		if f(level, msg, args...) {
			return true
		}
	}

	return false
}
