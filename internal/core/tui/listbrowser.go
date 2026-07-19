package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/term"
)

type contextEntry struct {
	Name  string
	Count int
}

const maxVisibleContexts = 20

type listBrowser struct {
	store      *core.Store
	entries    []contextEntry
	sel        int
	scroll     int
	lines      int
	parkedLine int
	width      int
	height     int
	out        io.Writer
	onPin      func(*core.Command)
}

func (b *listBrowser) writer() io.Writer {
	if b.out != nil {
		return b.out
	}
	return os.Stdout
}

func RunListBrowser(s *core.Store, onPin ...func(*core.Command)) (*core.Command, PaletteAction, error) {
	counts := map[string]int{}
	for _, c := range s.Commands {
		counts[c.Context]++
	}
	if len(counts) == 0 {
		return nil, ActionCopy, nil
	}

	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]contextEntry, len(names))
	for i, name := range names {
		entries[i] = contextEntry{Name: name, Count: counts[name]}
	}

	b := &listBrowser{
		store:   s,
		entries: entries,
	}
	if len(onPin) > 0 {
		b.onPin = onPin[0]
	}

	if UseTTY {
		f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			b.out = f
			defer f.Close()
		}
	}

	restore, err := term.MakeRaw()
	if err != nil {
		return nil, ActionCopy, err
	}
	defer restore()

	b.width, b.height = readTermSize()
	b.redraw()

	buf := make([]byte, 64)
	for {
		n, err := term.ReadInput(buf)
		if err != nil {
			b.close()
			return nil, ActionCopy, err
		}
		done := b.apply(parseKeys(buf[:n]))
		if done {
			b.close()
			ctx := b.entries[b.sel].Name
			if b.onPin != nil {
				return RunPalette(s, ctx, "", b.onPin)
			}
			return RunPalette(s, ctx, "")
		}
		b.redraw()
	}
}

func (b *listBrowser) apply(events []keyEvent) (done bool) {
	for _, ev := range events {
		switch ev.kind {
		case keyEsc, keyCancel:
			return true
		case keyEnter:
			return true
		case keyUp:
			if b.sel > 0 {
				b.sel--
				if b.sel < b.scroll {
					b.scroll = b.sel
				}
			}
		case keyDown:
			if b.sel < len(b.entries)-1 {
				b.sel++
				if b.sel >= b.scroll+b.visibleCount() {
					b.scroll = b.sel - b.visibleCount() + 1
				}
			}
		case keyRune:
			if ev.r >= '1' && ev.r <= '9' {
				if idx := int(ev.r - '1'); idx < b.visibleCount() {
					b.sel = b.scroll + idx
					return true
				}
			}
		}
	}
	return false
}

func (b *listBrowser) visibleCount() int {
	remaining := len(b.entries) - b.scroll
	if remaining <= 0 {
		return 0
	}
	n := remaining
	if n > maxVisibleContexts {
		n = maxVisibleContexts
	}
	if budget := b.height - 4; budget >= 0 && budget < n {
		n = budget
	}
	return n
}

func (b *listBrowser) frameLines() int {
	return 4 + b.visibleCount()
}

func (b *listBrowser) redraw() {
	b.width, b.height = readTermSize()
	var buf bytes.Buffer
	if b.lines > 0 && b.parkedLine > 0 {
		fmt.Fprintf(&buf, "\x1b[%dA", b.parkedLine)
	}
	buf.WriteString("\r")
	buf.Write(b.renderFrame())
	buf.WriteString("\x1b[J")
	newLines := b.frameLines()
	up := newLines - 1
	if up > 0 {
		fmt.Fprintf(&buf, "\x1b[%dA", up)
	}
	buf.WriteString("\r")
	_, _ = b.writer().Write(buf.Bytes())
	b.lines = newLines
	b.parkedLine = 1
}

func (b *listBrowser) close() {
	if b.lines > 0 && b.parkedLine > 0 {
		fmt.Fprintf(b.writer(), "\x1b[%dA", b.parkedLine)
	}
	_, _ = b.writer().Write([]byte("\r\x1b[J"))
	b.lines = 0
	b.parkedLine = 0
}
