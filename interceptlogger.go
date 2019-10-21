package hclog

import (
	"sync"
	"sync/atomic"
)

type InterceptLogger struct {
	Logger

	sync.Mutex
	sinkCount *int32
	Sinks     map[SinkAdapter]struct{}
}

func NewInterceptLogger(root Logger) *InterceptLogger {
	intercept := &InterceptLogger{
		Logger:    root,
		sinkCount: new(int32),
		Sinks:     make(map[SinkAdapter]struct{}),
	}

	atomic.StoreInt32(intercept.sinkCount, 0)

	return intercept
}

func (i *InterceptLogger) Debug(msg string, args ...interface{}) {
	i.Logger.Debug(msg, args...)
	if atomic.LoadInt32(i.sinkCount) == 0 {
		return
	}

	i.Lock()
	defer i.Unlock()
	for s := range i.Sinks {
		s.Accept(i.Name(), Debug, msg, i.retrieveImplied(args...)...)
	}
}

func (i *InterceptLogger) Trace(msg string, args ...interface{}) {
	i.Logger.Trace(msg, args...)
	if atomic.LoadInt32(i.sinkCount) == 0 {
		return
	}

	i.Lock()
	defer i.Unlock()
	for s := range i.Sinks {
		s.Accept(i.Name(), Trace, msg, i.retrieveImplied(args...)...)
	}
}

func (i *InterceptLogger) Info(msg string, args ...interface{}) {
	i.Logger.Info(msg, args...)
	if atomic.LoadInt32(i.sinkCount) == 0 {
		return
	}

	i.Lock()
	defer i.Unlock()
	for s := range i.Sinks {
		s.Accept(i.Name(), Info, msg, i.retrieveImplied(args...)...)
	}
}

func (i *InterceptLogger) Warn(msg string, args ...interface{}) {
	i.Logger.Warn(msg, args...)
	if atomic.LoadInt32(i.sinkCount) == 0 {
		return
	}

	i.Lock()
	defer i.Unlock()
	for s := range i.Sinks {
		s.Accept(i.Name(), Warn, msg, i.retrieveImplied(args...)...)
	}
}

func (i *InterceptLogger) Error(msg string, args ...interface{}) {
	i.Logger.Error(msg, args...)
	if atomic.LoadInt32(i.sinkCount) == 0 {
		return
	}

	i.Lock()
	defer i.Unlock()
	for s := range i.Sinks {
		s.Accept(i.Name(), Error, msg, i.retrieveImplied(args...)...)
	}
}

func (i *InterceptLogger) retrieveImplied(args ...interface{}) []interface{} {
	top := i.Logger.ImpliedArgs()

	cp := make([]interface{}, len(top)+len(args))

	copy(cp, top)
	copy(cp, args)

	return cp
}

func (i *InterceptLogger) With(args ...interface{}) Logger {
	var sub InterceptLogger

	sub = *i

	sub.Logger = i.Logger.With(args...)

	return &sub
}

func (i *InterceptLogger) SubscribeWith(opts *LoggerOptions) func() {
	logger := New(opts)
	sink := &sinkAdapter{logger.(*intLogger)}

	i.Sinks[sink] = struct{}{}

	atomic.AddInt32(i.sinkCount, 1)

	return func() {
		delete(i.Sinks, sink)
		atomic.AddInt32(i.sinkCount, -1)
	}
}

type sinkAdapter struct {
	*intLogger
}

func (s *sinkAdapter) Accept(name string, level Level, msg string, args ...interface{}) {
	// TODO handle name
	s.intLogger.Log(level, msg, args...)
}
