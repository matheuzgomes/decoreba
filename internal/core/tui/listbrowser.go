package tui

import (
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
	frame
	store   *core.Store
	entries []contextEntry
	sel     int
	scroll  int
	width   int
	height  int
	onPin   func(*core.Command)
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
	b.frame = newFrame(nil)
	if len(onPin) > 0 {
		b.onPin = onPin[0]
	}

	if UseTTY {
		f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			b.w = f
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
	b.draw(b.renderFrame(), 1)
}

func (b *listBrowser) close() {
	b.dismiss()
}
