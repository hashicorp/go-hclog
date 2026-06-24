package hclogslog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
)

func TestAdapter(t *testing.T) {
	t.Run("translates a specific level", func(t *testing.T) {
		var output bytes.Buffer

		log := hclog.New(&hclog.LoggerOptions{
			Name:   "test1",
			Level:  hclog.Debug,
			Output: &output,
		})

		logger := slog.New(Adapt(log))

		logger.Log(context.Background(), slog.LevelDebug-1, "message 0")

		logger.Debug("message 1", "key", "value")

		logger.Info("message 2", "a", 1, "long", "has spaces")

		ts := time.Now()

		logger.Warn("message 3", "b", ts)

		logger.Error("message 4", "c", 1*time.Minute)

		o := output.String()

		// slog reformulates the time, so we need to that too before
		// comparing it's output.

		cleanTime := time.Unix(0, ts.UnixNano()).In(ts.Location())

		require.NotContains(t, o, "message 0")
		require.Contains(t, o, "message 1: key=value\n")
		require.Contains(t, o, "message 2: a=1 long=\"has spaces\"\n")
		require.Contains(t, o, "message 3: b=\""+cleanTime.String()+"\"")
		require.Contains(t, o, "message 4: c=1m0s")
	})

	t.Run("translates groups", func(t *testing.T) {
		var output bytes.Buffer

		log := hclog.New(&hclog.LoggerOptions{
			Name:   "test1",
			Level:  hclog.Debug,
			Output: &output,
		})

		logger := slog.New(Adapt(log))

		ctx := context.Background()

		logger.LogAttrs(ctx, slog.LevelInfo, "message 1", slog.Group("x", slog.String("a", "thing")))

		logger.WithGroup("y").Warn("message 2", "z", 1)

		o := output.String()

		require.Contains(t, o, "message 1: x.a=thing\n")
		require.Contains(t, o, "message 2: y.z=1\n")
	})

	t.Run("translates nested groups", func(t *testing.T) {
		var output bytes.Buffer

		log := hclog.New(&hclog.LoggerOptions{
			Name:   "test1",
			Level:  hclog.Debug,
			Output: &output,
		})

		logger := slog.New(Adapt(log))

		ctx := context.Background()

		logger.LogAttrs(ctx, slog.LevelInfo, "message 1",
			slog.Group("x", slog.Group("n", slog.String("a", "thing"))))

		logger.WithGroup("y").Warn("message 2", "z", 1)

		o := output.String()

		require.Contains(t, o, "message 1: x.n.a=thing\n")
		require.Contains(t, o, "message 2: y.z=1\n")
	})

	t.Run("handles groups with no name", func(t *testing.T) {
		var output bytes.Buffer

		log := hclog.New(&hclog.LoggerOptions{
			Name:   "test1",
			Level:  hclog.Debug,
			Output: &output,
		})

		logger := slog.New(Adapt(log))

		ctx := context.Background()

		logger.LogAttrs(ctx, slog.LevelInfo, "message 1",
			slog.Group("", slog.String("a", "thing")))

		logger.WithGroup("").Warn("message 2", "z", 1)

		o := output.String()

		require.Contains(t, o, "message 1: a=thing\n")
		require.Contains(t, o, "message 2: z=1\n")
	})

	t.Run("handles attrs in a group with no name", func(t *testing.T) {
		var output bytes.Buffer

		log := hclog.New(&hclog.LoggerOptions{
			Name:   "test1",
			Level:  hclog.Debug,
			Output: &output,
		})

		logger := slog.New(Adapt(log))

		ctx := context.Background()

		logger.LogAttrs(ctx, slog.LevelInfo, "message 1",
			slog.Group("x", slog.String("", "thing")))

		logger.WithGroup("y").Warn("message 2", "", 1)

		o := output.String()

		require.Contains(t, o, "message 1: x.0=thing\n")
		require.Contains(t, o, "message 2: y.0=1\n")
	})
}

func TestSlogtest(t *testing.T) {
	var output bytes.Buffer

	log := hclog.New(&hclog.LoggerOptions{
		Name:       "test1",
		Level:      hclog.Debug,
		Output:     &output,
		JSONFormat: true,
	})

	err := slogtest.TestHandler(Adapt(log), func() []map[string]any {
		var entries []map[string]any

		dec := json.NewDecoder(bytes.NewReader(output.Bytes()))

		for {
			var e map[string]any

			err := dec.Decode(&e)
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Error(err)
			}

			e["time"] = e["@timestamp"]
			e["level"] = e["@level"]
			e["msg"] = e["@message"]

			for k, v := range e {
				if strings.Contains(k, ".") {
					parts := strings.Split(k, ".")

					top := e

					for _, p := range parts[:len(parts)-1] {
						sub, ok := top[p].(map[string]any)
						if !ok {
							sub = make(map[string]any)
							top[p] = sub
						}

						top = sub
					}

					top[parts[len(parts)-1]] = v
				}
			}

			entries = append(entries, e)
		}

		return entries
	})
	if err != nil {
		// Strip out the zero time issue, we can't handle it yet.

		if je, ok := err.(interface {
			Unwrap() []error
		}); ok {
			var filtered []string
			for _, sub := range je.Unwrap() {
				if !strings.Contains(sub.Error(), "should ignore a zero Record.Time") {
					filtered = append(filtered, sub.Error())
				}
			}

			if len(filtered) > 0 {
				err = errors.New(strings.Join(filtered, "\n"))
			} else {
				err = nil
			}
		}
	}

	if err != nil {
		t.Error(err)
	}
}
