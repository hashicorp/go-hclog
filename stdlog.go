package log

import (
	"bytes"
	"strings"
)

type stdlogAdapter struct {
	hl          Logger
	inferLevels bool
}

func (s *stdlogAdapter) Write(data []byte) (int, error) {
	str := string(bytes.TrimRight(data, " \t\n"))

	if s.inferLevels {
		level, str := s.pickLevel(str)
		switch level {
		case Trace:
			s.hl.Trace(str)
		case Debug:
			s.hl.Debug(str)
		case Info:
			s.hl.Info(str)
		case Warn:
			s.hl.Warn(str)
		case Error:
			s.hl.Error(str)
		default:
			s.hl.Info(str)
		}
	} else {
		s.hl.Info(str)
	}

	return len(data), nil
}

func (s *stdlogAdapter) pickLevel(str string) (Level, string) {
	switch {
	case strings.HasPrefix(str, "[DEBUG]"):
		return Debug, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[TRACE]"):
		return Trace, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[INFO]"):
		return Info, strings.TrimSpace(str[6:])
	case strings.HasPrefix(str, "[WARN]"):
		return Warn, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERROR]"):
		return Error, strings.TrimSpace(str[7:])
	case strings.HasPrefix(str, "[ERR]"):
		return Error, strings.TrimSpace(str[5:])
	default:
		return Info, str
	}
}
