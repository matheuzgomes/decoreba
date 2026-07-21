package tui

import (
	"bytes"
	"strings"
	"testing"
)

func TestOverlayInit(t *testing.T) {
	var o overlay
	o.init(nil)
	if o.w == nil {
		t.Fatal("writer should be set to os.Stdout when nil")
	}

	var buf bytes.Buffer
	o.init(&buf)
	if o.w != &buf {
		t.Fatal("writer should be the passed writer")
	}
}

func TestOverlayClose(t *testing.T) {
	t.Run("empty writer writes clear", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf, lines: 0, parkedLine: 1}
		o.close()
		got := buf.String()
		if got != "\r\x1b[J" {
			t.Fatalf("close with lines=0 should just clear: %q", got)
		}
	})

	t.Run("dismisses lines", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf, lines: 3, parkedLine: 2}
		o.close()
		out := buf.String()
		if !strings.HasPrefix(out, "\x1b[2A") {
			t.Fatalf("should move up by parkedLine: %q", out)
		}
		if !strings.HasSuffix(out, "\r\x1b[J") {
			t.Fatalf("should clear: %q", out)
		}
		if o.lines != 0 || o.parkedLine != 0 {
			t.Fatal("state should be reset")
		}
	})

	t.Run("clear without move", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf, lines: 5, parkedLine: 0}
		o.close()
		got := buf.String()
		if got != "\r\x1b[J" {
			t.Fatalf("close with parked=0 should just clear: %q", got)
		}
	})
}

func TestOverlayUnsafeDraw(t *testing.T) {
	t.Run("first draw", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf}
		o.unsafeDraw([]byte("line1\nline2"), 1, 3)
		out := buf.String()
		if !strings.HasPrefix(out, "\r") {
			t.Fatalf("should start with CR: %q", out)
		}
		if !strings.Contains(out, "\x1b[J") {
			t.Fatalf("should clear after: %q", out)
		}
		// cursor positioning: newLines=2, cursorLine=1 → up=0
		// cursorCol=3 → \x1b[3C
		if !strings.Contains(out, "\x1b[3C") {
			t.Fatalf("should position cursor at col 3: %q", out)
		}
	})

	t.Run("redraw move up", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf, lines: 3, parkedLine: 2}
		o.unsafeDraw([]byte("shorter"), 0, 0)
		out := buf.String()
		if !strings.HasPrefix(out, "\x1b[2A\r") {
			t.Fatalf("should move up first: %q", out)
		}
	})

	t.Run("zero cursor col", func(t *testing.T) {
		var buf bytes.Buffer
		o := overlay{w: &buf}
		o.unsafeDraw([]byte("hello"), 0, 0)
		out := buf.String()
		if strings.Contains(out, "\x1b[C") {
			t.Fatal("should not position cursor when col=0")
		}
	})
}
