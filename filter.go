package hclog

// FilterOut provides a simple way to build a list of log messages that
// can be queried and matched. This is meant to be used with the FilterOut
// option on Options to suppress log messages.
type FilterOut struct {
	messages map[string]struct{}
}

func (f *FilterOut) Add(msg string) {
	if f.messages == nil {
		f.messages = make(map[string]struct{})
	}

	f.messages[msg] = struct{}{}
}

func (f *FilterOut) FilterOut(level Level, msg string, args ...interface{}) bool {
	_, ok := f.messages[msg]
	return ok
}
