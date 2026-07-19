package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/search"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
)

const (
	paletteHint = "↵ copy  ↑↓ nav  1-9 direct  ^e edit  ^x exec  ^s pin  esc cancel"
	maxVisible  = 9
)

// PaletteAction describes what the user chose to do with the selected command.
type PaletteAction int

const (
	ActionCopy    PaletteAction = iota
	ActionEdit
	ActionExecute
)

type palette struct {
	store        *core.Store
	chip         string
	query        []rune
	pool         []core.Command
	results      []search.Scored
	suggestions  []string
	sel          int
	scrollOffset int
	action       PaletteAction
	lines        int
	parkedLine   int
	width        int
	height       int
	out          io.Writer
}

// UseTTY forces the palette to write its UI to /dev/tty instead of stdout.
// Set this before RunPalette when stdout is captured (e.g. shell integration).
var UseTTY bool

func (p *palette) writer() io.Writer {
	if p.out != nil {
		return p.out
	}
	return os.Stdout
}

// RunPalette opens an interactive inline command palette. Returns the
// selected command, the action the user chose (copy or edit), or nil when
// the user cancels.
func RunPalette(store *core.Store, context, initialQuery string) (*core.Command, PaletteAction, error) {
	p := &palette{store: store, chip: context, action: ActionCopy, scrollOffset: 0}
	if UseTTY {
		f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			p.out = f
			defer f.Close()
		}
	}
	p.setPool()
	if initialQuery != "" {
		p.query = []rune(initialQuery)
	}
	p.refilter()

	restore, err := term.MakeRaw()
	if err != nil {
		return nil, ActionCopy, err
	}
	defer restore()

	p.width, p.height = readTermSize()
	p.redraw()

	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			p.close()
			return nil, ActionCopy, err
		}
		// A lone ESC byte may be the Esc key or the start of an arrow
		// sequence; if more bytes are already available, read them in.
		if n == 1 && buf[0] == 0x1b && term.InputAvailable(25) {
			if m, err := os.Stdin.Read(buf[1:]); err == nil {
				n += m
			}
		}
		done, chosen := p.apply(parseKeys(buf[:n]))
		if done {
			p.close()
			if chosen != nil && p.action == ActionCopy && hasVariables(chosen.Command) {
				resolved, cancelled, err := resolveCommand(chosen.Command)
				if err != nil {
					return nil, ActionCopy, err
				}
				if cancelled {
					return nil, ActionCopy, nil
				}
				chosen.Command = resolved
			}
			return chosen, p.action, nil
		}
		p.redraw()
	}
}

func (p *palette) apply(events []keyEvent) (done bool, chosen *core.Command) {
	for _, ev := range events {
		switch ev.kind {
		case keyEsc, keyCancel:
			return true, nil
		case keyEnter:
			if len(p.results) > 0 {
				p.action = ActionCopy
				return true, &p.results[p.sel].Cmd
			}
		case keyEdit:
			if len(p.results) > 0 {
				p.action = ActionEdit
				return true, &p.results[p.sel].Cmd
			}
		case keyExecute:
		if len(p.results) > 0 {
			p.action = ActionExecute
			return true, &p.results[p.sel].Cmd
		}
	case keySave:
		if len(p.results) > 0 {
			cmd := &p.results[p.sel].Cmd
			cmd.Pinned = !cmd.Pinned
			for i := range p.store.Commands {
				if p.store.Commands[i].ID == cmd.ID {
					p.store.Commands[i].Pinned = cmd.Pinned
					break
				}
			}
			_ = store.Save(p.store)
			p.refilter()
		}
	case keyUp:
			if p.sel > 0 {
				p.sel--
				if p.sel < p.scrollOffset {
					p.scrollOffset = p.sel
				}
			}
		case keyDown:
			if p.sel < len(p.results)-1 {
				p.sel++
				if p.sel >= p.scrollOffset+p.visibleCount() {
					p.scrollOffset = p.sel - p.visibleCount() + 1
				}
			}
		case keyBackspace:
			if len(p.query) > 0 {
				p.query = p.query[:len(p.query)-1]
				p.refilter()
			} else if p.chip != "" {
				// Backspace on the empty field drops the context chip and
				// expands the search to every context.
				p.chip = ""
				p.setPool()
				p.refilter()
			}
		case keyRune:
			if len(p.query) == 0 && ev.r >= '1' && ev.r <= '9' {
				if idx := int(ev.r - '1'); idx < p.visibleCount() {
					p.action = ActionCopy
					return true, &p.results[p.scrollOffset+idx].Cmd
				}
			}
			p.query = append(p.query, ev.r)
			p.refilter()
		}
	}
	return false, nil
}

func (p *palette) setPool() {
	p.pool = nil
	if p.chip == "" {
		p.pool = p.store.Commands
		return
	}
	for _, c := range p.store.Commands {
		if strings.EqualFold(c.Context, p.chip) {
			p.pool = append(p.pool, c)
		}
	}
}

func (p *palette) refilter() {
	p.results = search.Sort(p.pool, string(p.query))
	p.scrollOffset = 0
	if len(p.results) == 0 && len(p.query) > 1 {
		p.suggestions = search.Suggest(p.pool, string(p.query), 3)
	} else {
		p.suggestions = nil
	}
	if p.sel >= len(p.results) {
		p.sel = len(p.results) - 1
	}
	if p.sel < 0 {
		p.sel = 0
	}
}

func (p *palette) visibleCount() int {
	remaining := len(p.results) - p.scrollOffset
	if remaining <= 0 {
		return 0
	}
	n := remaining
	if n > maxVisible {
		n = maxVisible
	}
	if budget := p.height - 2; budget >= 0 && budget/2 < n {
		n = budget / 2
	}
	return n
}

func (p *palette) frameLines() int {
	itemLines := 1
	if len(p.results) > 0 {
		itemLines = p.visibleCount() * 2
	}
	return 1 + 1 + itemLines + 1 + 1
}

func (p *palette) inputCol() int {
	col := boxLeftPad + 2 + len(p.query)
	if p.chip != "" {
		col += len([]rune(p.chip)) + 1
	}
	return col
}

func (p *palette) searchLine() int { return 1 }

func (p *palette) redraw() {
	p.width, p.height = readTermSize()
	var b bytes.Buffer
	if p.lines > 0 && p.parkedLine > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", p.parkedLine)
	}
	b.WriteString("\r")
	b.Write(p.renderFrame())
	b.WriteString("\x1b[J")
	newLines := p.frameLines()
	up := newLines - 1 - p.searchLine()
	if up > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", up)
	}
	b.WriteString("\r")
	if col := p.inputCol(); col > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", col)
	}
	_, _ = p.writer().Write(b.Bytes())
	p.lines = newLines
	p.parkedLine = p.searchLine()
}

func (p *palette) close() {
	if p.lines > 0 && p.parkedLine > 0 {
		fmt.Fprintf(p.writer(), "\x1b[%dA", p.parkedLine)
	}
	_, _ = p.writer().Write([]byte("\r\x1b[J"))
	p.lines = 0
	p.parkedLine = 0
}
