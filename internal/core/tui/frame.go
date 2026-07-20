package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// frame manages terminal cursor positioning for inline overlays. Every
// overlay (palette, addform, step editor, list browser, workflow runner)
// embeds frame and calls draw/dismiss instead of duplicating the redraw/
// close pattern.
//
// Usage:
//
//	f := newFrame(os.Stdout)
//	for ... {
//	    body := renderFrame()
//	    f.draw(body, cursorLine, cursorCol)
//	    // read input, apply events...
//	}
//	f.dismiss()
type frame struct {
	lines      int
	parkedLine int
	w          io.Writer
}

func newFrame(w io.Writer) frame {
	if w == nil {
		w = os.Stdout
	}
	return frame{w: w}
}

// draw renders the overlay body at the current terminal position. On the
// first call, it draws from the current line. On subsequent calls, it
// moves the cursor up by the previously parked distance so the frame
// appears in-place — no scrollback pollution.
//
// cursorLine is the frame-relative line (0 = top) where the cursor should
// land. cursorCol, when provided, positions the cursor horizontally on
// that line. Use it for input fields; omit it for static overlays.
func (f *frame) draw(body []byte, cursorLine int, cursorCol ...int) {
	var b bytes.Buffer
	if f.lines > 0 && f.parkedLine > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", f.parkedLine)
	}
	b.WriteString("\r")
	b.Write(body)
	b.WriteString("\x1b[J")

	newLines := bytes.Count(body, []byte{'\n'}) + 1
	up := newLines - 1 - cursorLine
	if up > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", up)
	}
	b.WriteString("\r")
	if len(cursorCol) > 0 && cursorCol[0] > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", cursorCol[0])
	}

	_, _ = f.w.Write(b.Bytes())
	f.lines = newLines
	f.parkedLine = cursorLine
}

// dismiss clears the overlay from the terminal, returning the cursor to
// the position it was at when the overlay first appeared.
func (f *frame) dismiss() {
	if f.lines > 0 && f.parkedLine > 0 {
		fmt.Fprintf(f.w, "\x1b[%dA", f.parkedLine)
	}
	_, _ = f.w.Write([]byte("\r\x1b[J"))
	f.lines = 0
	f.parkedLine = 0
}
