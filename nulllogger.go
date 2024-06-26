// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MIT

package hclog

import (
	"io"
	"log"
)

// NewNullLogger instantiates a Logger for which all calls
// will succeed without doing anything.
// Useful for testing purposes.
func NewNullLogger() LogImpl {
	return &nullLogger{}
}

type nullLogger struct{}

func (l *nullLogger) Log(_ Level, _ string, _ ...interface{}) {}

func (l *nullLogger) LogRecord(_ Record) {}

func (l *nullLogger) Trace(_ string, _ ...interface{}) {}

func (l *nullLogger) Debug(_ string, _ ...interface{}) {}

func (l *nullLogger) Info(_ string, _ ...interface{}) {}

func (l *nullLogger) Warn(_ string, _ ...interface{}) {}

func (l *nullLogger) Error(_ string, _ ...interface{}) {}

func (l *nullLogger) IsTrace() bool { return false }

func (l *nullLogger) IsDebug() bool { return false }

func (l *nullLogger) IsInfo() bool { return false }

func (l *nullLogger) IsWarn() bool { return false }

func (l *nullLogger) IsError() bool { return false }

func (l *nullLogger) ImpliedArgs() []interface{} { return []interface{}{} }

func (l *nullLogger) With(_ ...interface{}) LogImpl { return l }

func (l *nullLogger) Name() string { return "" }

func (l *nullLogger) Named(_ string) LogImpl { return l }

func (l *nullLogger) ResetNamed(_ string) LogImpl { return l }

func (l *nullLogger) SetLevel(_ Level) {}

func (l *nullLogger) Level() Level { return NoLevel }

func (l *nullLogger) StandardLogger(opts *StandardLoggerOptions) *log.Logger {
	return log.New(l.StandardWriter(opts), "", log.LstdFlags)
}

func (l *nullLogger) StandardWriter(_ *StandardLoggerOptions) io.Writer {
	return io.Discard
}
