package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func findLevel(level Level) zapcore.Level {
	switch level {
	case Debug:
		return zapcore.DebugLevel
	case Trace:
		return zapcore.DebugLevel
	case Info:
		return zapcore.InfoLevel
	case Warn:
		return zapcore.WarnLevel
	case Error:
		return zapcore.ErrorLevel
	default:
		// Should we log here? How ironic is that.
		return zapcore.DebugLevel
	}
}

var (
	_levelToBracket = map[zapcore.Level]string{
		zapcore.DebugLevel: "[DEBUG]",
		zapcore.InfoLevel:  "[INFO ]",
		zapcore.WarnLevel:  "[WARN ]",
		zapcore.ErrorLevel: "[ERROR]",
	}
)

func encodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToBracket[l]
	if ok {
		enc.AppendString(s)
	} else {
		enc.AppendString(l.String())
	}
}

func New(opts LoggerOptions) Logger {
	zlevel := findLevel(opts.Level)
	level := zap.NewAtomicLevelAt(zlevel)

	enccfg := zap.NewProductionEncoderConfig()
	enccfg.EncodeLevel = encodeLevel
	enccfg.EncodeTime = zapcore.ISO8601TimeEncoder

	enc := zapcore.NewConsoleEncoder(enccfg)

	var (
		out zapcore.WriteSyncer
		err error
	)

	if opts.Output != nil {
		out = zapcore.AddSync(opts.Output)
	} else {
		out, _, err = zap.Open("stdout")
		if err != nil {
			panic(err)
		}
	}

	errSink, _, err := zap.Open("stderr")
	if err != nil {
		panic(err)
	}

	bo := []zap.Option{
		zap.ErrorOutput(errSink),
	}

	if opts.IncludeLocation {
		bo = append(bo, zap.AddCaller())
	}

	zl := zap.New(zapcore.NewCore(enc, out, level), bo...)

	if opts.Name != "" {
		zl = zl.Named(opts.Name + ":")
	}

	return &zapLogger{
		zl:     zl,
		zs:     zl.Sugar(),
		zlevel: zlevel,
		level:  opts.Level,
	}
}

type zapLogger struct {
	zl     *zap.Logger
	zs     *zap.SugaredLogger
	zlevel zapcore.Level
	level  Level
}

var _ Logger = &zapLogger{}

func (z *zapLogger) Debug(msg string, args ...interface{}) {
	z.zs.Debugw(msg, args...)
}

func (z *zapLogger) Trace(msg string, args ...interface{}) {
	z.zs.Debugw(msg, args...)
}

func (z *zapLogger) Info(msg string, args ...interface{}) {
	z.zs.Infow(msg, args...)
}

func (z *zapLogger) Warn(msg string, args ...interface{}) {
	z.zs.Warnw(msg, args...)
}

func (z *zapLogger) Error(msg string, args ...interface{}) {
	z.zs.Errorw(msg, args...)
}

func (z *zapLogger) IsTrace() bool {
	return z.level >= Trace
}

func (z *zapLogger) IsDebug() bool {
	return z.level >= Debug
}

func (z *zapLogger) IsInfo() bool {
	return z.level >= Info
}

func (z *zapLogger) IsWarn() bool {
	return z.level >= Warn
}

func (z *zapLogger) IsError() bool {
	return z.level >= Error
}

func (z *zapLogger) With(args ...interface{}) Logger {
	sz := z.zs.With(args...)

	return &zapLogger{
		zl:     sz.Desugar(),
		zs:     sz,
		zlevel: z.zlevel,
		level:  z.level,
	}
}

func (z *zapLogger) Named(name string) Logger {
	sz := z.zs.Named(name)

	return &zapLogger{
		zl:     sz.Desugar(),
		zs:     sz,
		zlevel: z.zlevel,
		level:  z.level,
	}
}

func (z *zapLogger) Stacktrace(args ...interface{}) {
	zl := z.zl.WithOptions(zap.AddStacktrace(z.zlevel))
	zl.Sugar().Debugw("", args...)
}

func (z *zapLogger) StandardLogger(inferLevels bool) *log.Logger {
	return log.New(&stdlogAdapter{z, inferLevels}, "", 0)
}
