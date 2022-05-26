package hclog

import (
	"testing"
)

var logger = NewNullLogger()

func TestNullLoggerIsEfficient(t *testing.T) {
	// Since statements like "IsWarn()", "IsError()", etc. are used to gate
	// actually writing warning and error statements, the null logger will
	// be faster and more efficient if it always returns false for these calls.
	assertFalse(t, logger.IsTrace())
	assertFalse(t, logger.IsDebug())
	assertFalse(t, logger.IsInfo())
	assertFalse(t, logger.IsWarn())
	assertFalse(t, logger.IsError())
}

func TestNullLoggerReturnsNullLoggers(t *testing.T) {

	// Sometimes the logger is asked to return subloggers.
	// These should also be a nullLogger.

	subLogger := logger.With()
	_, ok := subLogger.(*nullLogger)
	assertTrue(t, ok)

	subLogger = logger.Named("")
	_, ok = subLogger.(*nullLogger)
	assertTrue(t, ok)

	subLogger = logger.ResetNamed("")
	_, ok = subLogger.(*nullLogger)
	assertTrue(t, ok)
}

func TestStandardLoggerIsntNil(t *testing.T) {
	// Don't return a nil pointer for the standard logger,
	// lest it cause a panic.
	stdLogger := logger.StandardLogger(nil)
	if stdLogger == nil {
		t.Fatalf("expected non-nil")
	}
}
