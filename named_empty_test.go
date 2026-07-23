package hclog

import (
	"bytes"
	"strings"
	"testing"
)

func TestNamedEmptyIsNoop(t *testing.T) {
	var buf bytes.Buffer
	l := New(&LoggerOptions{Name: "app", Level: Info, Output: &buf, DisableTime: true})
	l.Named("").Info("hi")
	out := buf.String()
	if !strings.Contains(out, "app:") && !strings.Contains(out, "app") {
		t.Fatalf("expected app name in log: %q", out)
	}
	if strings.Contains(out, "app.") {
		t.Fatalf("unexpected empty name segment: %q", out)
	}
	buf.Reset()
	l.Named("http").Named("").Info("x")
	out = buf.String()
	if !strings.Contains(out, "app.http") {
		t.Fatalf("expected app.http: %q", out)
	}
}
