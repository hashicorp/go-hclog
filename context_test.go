package hclog

import (
	"bytes"
	"context"
	"testing"
)

func TestContext_simpleLogger(t *testing.T) {
	l := L()
	ctx := WithContext(context.Background(), l)
	if l != FromContext(ctx) {
		t.Fatalf("expected equal")
	}
}

func TestContext_empty(t *testing.T) {
	if L() != FromContext(context.Background()) {
		t.Fatalf("expected equal")
	}
}

func TestContext_fields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&LoggerOptions{
		Level:  Debug,
		Output: &buf,
	})

	// Insert the logger with fields
	ctx := WithContext(context.Background(), l, "hello", "world")
	l = FromContext(ctx)
	requireNotNil(t, l)

	// Log something so we can test the output that the field is there
	l.Debug("test")
	requireContains(t, buf.String(), "hello")
}
