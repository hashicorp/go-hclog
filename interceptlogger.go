package hclog

type Sink interface {
	Accept(name string, level Level, msg string, args []interface{})
}

type InterceptLogger struct {
	Logger

	Sinks []Sink
}

func (i *InterceptLogger) Debug(msg string, args ...interface{}) {
	i.Logger.Debug(msg, args...)

	var top []interface{}

	if il, ok := i.Logger.(interface {
		ImpliedArgs() []interface{}
	}); ok {
		top = il.ImpliedArgs()
	}

	cp := make([]interface{}, len(top)+len(args))

	copy(cp, top)
	copy(cp, args)

	for _, s := range i.Sinks {
		s.Accept(i.Name(), Debug, msg, cp)
	}
}

func (i *InterceptLogger) With(args ...interface{}) Logger {
	var sub InterceptLogger

	sub = *i

	sub.Logger = i.Logger.With(args...)

	return &sub
}

type SinkAdapter struct {
	*intLogger
}

func (s *SinkAdapter) Accept(level Level, msg string, args []interface{}) {
	s.intogger.Log(level, msg, args...)
}
