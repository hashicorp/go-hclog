package hclogslog

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
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
}
