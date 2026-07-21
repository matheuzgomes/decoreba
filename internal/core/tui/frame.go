package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// overlay manages terminal cursor positioning for inline overlays.
// Every overlay component (palette, addform, list browser, step editor,
// workflow runner) embeds overlay and delegates to refresh()/close().
//
// Usage:
//
//	o.init(nil)
//	for {
//	    o.refresh(func(w, h int) ([]byte, int, int) {
//	        return renderContent(), cursorLine, cursorCol
//	    })
//	    // read input, apply events...
//	}
//	o.close()
type overlay struct {
	w          io.Writer
	lines      int
	parkedLine int
	width      int
	height     int
}

func (o *overlay) init(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	o.w = w
}

// refresh redraws the overlay at the current terminal position.
// render returns (body, cursorLine, cursorCol).
// cursorCol of 0 means no horizontal cursor positioning.
func (o *overlay) refresh(render func(int, int) ([]byte, int, int)) {
	w, h := readSize()
	o.width, o.height = w, h
	body, cursorLine, cursorCol := render(w, h)
	o.unsafeDraw(body, cursorLine, cursorCol)
}

// unsafeDraw writes body and repositions the cursor. It is the lowest-level
// cursor math; prefer refresh().
func (o *overlay) unsafeDraw(body []byte, cursorLine, cursorCol int) {
	var b bytes.Buffer
	if o.lines > 0 && o.parkedLine > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", o.parkedLine)
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
	if cursorCol > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", cursorCol)
	}

	_, _ = o.w.Write(b.Bytes())
	o.lines = newLines
	o.parkedLine = cursorLine
}

// close dismisses the overlay from the terminal.
func (o *overlay) close() {
	if o.lines > 0 && o.parkedLine > 0 {
		fmt.Fprintf(o.w, "\x1b[%dA", o.parkedLine)
	}
	_, _ = o.w.Write([]byte("\r\x1b[J"))
	o.lines = 0
	o.parkedLine = 0
}
