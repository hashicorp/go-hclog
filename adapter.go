package hclogslog

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"golang.org/x/exp/slog"
)

func Adapt(l hclog.Logger) slog.Handler {
	return &Handler{l: l}
}

type Handler struct {
	l      hclog.Logger
	prefix string
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
func (h *Handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	switch {
	case lvl < slog.LevelDebug:
		return h.l.IsTrace()
	case lvl < slog.LevelInfo:
		return h.l.IsDebug()
	case lvl < slog.LevelWarn:
		return h.l.IsInfo()
	case lvl < slog.LevelError:
		return h.l.IsWarn()
	default:
		return h.l.IsError()
	}
}

var basicTranslate = map[slog.Level]hclog.Level{
	slog.LevelDebug - 4: hclog.Trace,
	slog.LevelDebug:     hclog.Debug,
	slog.LevelInfo:      hclog.Info,
	slog.LevelWarn:      hclog.Warn,
	slog.LevelError:     hclog.Error,
}

func (h *Handler) translateLevel(lvl slog.Level) hclog.Level {
	if tl, ok := basicTranslate[lvl]; ok {
		return tl
	}

	switch {
	case lvl < slog.LevelDebug:
		return hclog.Trace
	case lvl < slog.LevelInfo:
		return hclog.Debug
	case lvl < slog.LevelWarn:
		return hclog.Info
	case lvl < slog.LevelError:
		return hclog.Warn
	default:
		return hclog.Error
	}
}

// Handle handles the Record.
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
func (h *Handler) Handle(_ context.Context, rec slog.Record) error {
	attrs := make([]any, 0, rec.NumAttrs()*2)

	rec.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindGroup {
			for _, subA := range a.Value.Group() {
				attrs = append(attrs, h.prefix+a.Key+"."+subA.Key, subA.Value.Any())
			}
		} else {
			attrs = append(attrs, h.prefix+a.Key, a.Value.Any())
		}
		return true
	})

	h.l.Log(h.translateLevel(rec.Level), rec.Message, attrs...)
	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	args := make([]any, 0, len(attrs))

	for _, a := range attrs {
		if a.Value.Kind() == slog.KindGroup {
			for _, subA := range a.Value.Group() {
				args = append(args, h.prefix+a.Key+"."+subA.Key, subA.Value.Any())
			}
		} else {
			args = append(args, h.prefix+a.Key, a.Value.Any())
		}
	}

	return &Handler{
		l:      h.l.With(args...),
		prefix: h.prefix,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		l:      h.l,
		prefix: name + ".",
	}
}

var _ slog.Handler = &Handler{}
