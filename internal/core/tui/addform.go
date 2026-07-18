package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/term"
)

const (
	fieldContext = 0
	fieldTitle   = 1
	fieldCommand = 2
	fieldTags    = 3
	fieldNotes   = 4
	fieldCount   = 5

	addFormHint  = "tab next   ⇧tab previous   ^s save   esc cancel"
	newCmdHeader = "+ new command"
	editCmdHeader = "edit command"
	labelPad     = 9
)

var fieldLabels = [fieldCount]string{
	"context",
	"title",
	"command",
	"tags",
	"notes",
}

type addForm struct {
	store      *core.Store
	fields     [fieldCount][]rune
	focus      int
	cursor     int
	errField   int
	errMsg     string
	errFlash   int
	contexts   []string
	editing    bool
	existing   *core.Command
	lines      int
	parkedLine int
	width      int
	height     int
	out        io.Writer
}

func (f *addForm) writer() io.Writer {
	if f.out != nil {
		return f.out
	}
	return os.Stdout
}

// RunAddForm opens an interactive inline command-creation form. Returns the
// new command or nil when the user cancels.
func RunAddForm(store *core.Store) (*core.Command, error) {
	return runForm(store, nil)
}

// RunEditForm opens the same interactive form pre-filled with an existing
// command. Returns the updated command (with the same ID) or nil on cancel.
func RunEditForm(store *core.Store, existing *core.Command) (*core.Command, error) {
	return runForm(store, existing)
}

func runForm(store *core.Store, existing *core.Command) (*core.Command, error) {
	f := &addForm{
		store:    store,
		errField: -1,
		contexts: collectContexts(store),
		editing:  existing != nil,
		existing: existing,
	}
	if existing != nil {
		f.fields[fieldContext] = []rune(existing.Context)
		f.fields[fieldTitle] = []rune(existing.Title)
		f.fields[fieldCommand] = []rune(existing.Command)
		f.fields[fieldTags] = []rune(strings.Join(existing.Tags, ", "))
		f.fields[fieldNotes] = []rune(existing.Notes)
		f.focus = fieldContext
		f.cursor = len(f.fields[fieldContext])
	}

	restore, err := term.MakeRaw()
	if err != nil {
		return nil, err
	}
	defer restore()

	f.width, f.height = readTermSize()
	f.redraw()

	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			f.close()
			return nil, err
		}
		if n == 1 && buf[0] == 0x1b && term.InputAvailable(25) {
			if m, err := os.Stdin.Read(buf[1:]); err == nil {
				n += m
			}
		}
		done, cmd := f.apply(parseKeys(buf[:n]))
		if done {
			f.close()
			return cmd, nil
		}
		f.redraw()
	}
}

func collectContexts(store *core.Store) []string {
	seen := make(map[string]string)
	var out []string
	for _, c := range store.Commands {
		k := strings.ToLower(c.Context)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = c.Context
		out = append(out, c.Context)
	}
	return out
}

func (f *addForm) apply(events []keyEvent) (done bool, cmd *core.Command) {
	for _, ev := range events {
		switch ev.kind {
		case keyEsc, keyCancel:
			return true, nil
		case keySave:
			if c, ok := f.trySave(); ok {
				return true, c
			}
		case keyEnter:
			if f.focus == fieldNotes {
				if c, ok := f.trySave(); ok {
					return true, c
				}
			} else {
				f.moveFocus(1)
			}
		case keyTab:
			if f.acceptSuggestion() {
				continue
			}
			f.moveFocus(1)
		case keyShiftTab:
			f.moveFocus(-1)
		case keyLeft:
			if f.cursor > 0 {
				f.cursor--
			}
		case keyRight:
			if f.acceptSuggestion() {
				continue
			}
			if f.cursor < len(f.fields[f.focus]) {
				f.cursor++
			}
		case keyBackspace:
			f.clearErrOnEdit()
			if f.cursor > 0 {
				fld := f.fields[f.focus]
				f.fields[f.focus] = append(fld[:f.cursor-1], fld[f.cursor:]...)
				f.cursor--
			}
		case keyDelete:
			f.clearErrOnEdit()
			if f.cursor < len(f.fields[f.focus]) {
				fld := f.fields[f.focus]
				f.fields[f.focus] = append(fld[:f.cursor], fld[f.cursor+1:]...)
			}
		case keyRune:
			f.clearErrOnEdit()
			fld := f.fields[f.focus]
			f.fields[f.focus] = append(fld[:f.cursor], append([]rune{ev.r}, fld[f.cursor:]...)...)
			f.cursor++
		}
	}
	return false, nil
}

func (f *addForm) clearErrOnEdit() {
	if f.errField == f.focus {
		f.errField = -1
		f.errMsg = ""
		f.errFlash = 0
	}
}

func (f *addForm) moveFocus(delta int) {
	f.focus = (f.focus + delta + fieldCount) % fieldCount
	f.cursor = len(f.fields[f.focus])
}

func (f *addForm) contextSuggestion() string {
	if f.focus != fieldContext {
		return ""
	}
	typed := f.fields[fieldContext]
	if len(typed) == 0 || f.cursor != len(typed) {
		return ""
	}
	lower := strings.ToLower(string(typed))
	for _, ctx := range f.contexts {
		cr := []rune(ctx)
		if len(cr) <= len(typed) {
			continue
		}
		if strings.HasPrefix(strings.ToLower(ctx), lower) {
			return string(cr[len(typed):])
		}
	}
	return ""
}

func (f *addForm) acceptSuggestion() bool {
	sug := f.contextSuggestion()
	if sug == "" {
		return false
	}
	f.fields[fieldContext] = append(f.fields[fieldContext], []rune(sug)...)
	f.cursor = len(f.fields[fieldContext])
	f.clearErrOnEdit()
	return true
}

func (f *addForm) trySave() (*core.Command, bool) {
	ctx := strings.TrimSpace(string(f.fields[fieldContext]))
	cmdStr := strings.TrimSpace(string(f.fields[fieldCommand]))
	if ctx == "" {
		f.focus = fieldContext
		f.cursor = len(f.fields[fieldContext])
		f.errField = fieldContext
		f.errMsg = "context is required"
		f.errFlash = 3
		return nil, false
	}
	if cmdStr == "" {
		f.focus = fieldCommand
		f.cursor = len(f.fields[fieldCommand])
		f.errField = fieldCommand
		f.errMsg = "command is required"
		f.errFlash = 3
		return nil, false
	}

	var tags []string
	for _, t := range strings.Split(string(f.fields[fieldTags]), ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	now := time.Now()
	cmd := &core.Command{
		Context:   strings.ToLower(ctx),
		Title:     strings.TrimSpace(string(f.fields[fieldTitle])),
		Command:   cmdStr,
		Tags:      tags,
		Notes:     strings.TrimSpace(string(f.fields[fieldNotes])),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if f.editing && f.existing != nil {
		cmd.ID = f.existing.ID
		cmd.CreatedAt = f.existing.CreatedAt
		cmd.UsageCount = f.existing.UsageCount
		cmd.LastUsedAt = f.existing.LastUsedAt
	} else {
		cmd.ID = core.GenID()
	}
	return cmd, true
}

func (f *addForm) frameLines() int {
	return 1 + 1 + fieldCount + 1 + 1
}

func (f *addForm) inputCol() int {
	prefix := f.fieldPrefixDisplay(f.focus, f.cursor)
	return boxLeftPad + labelPad + visibleWidth(prefix)
}

func (f *addForm) inputLine() int {
	return 2 + f.focus
}

func (f *addForm) redraw() {
	f.width, f.height = readTermSize()
	var b bytes.Buffer
	if f.lines > 0 && f.parkedLine > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", f.parkedLine)
	}
	b.WriteString("\r")
	b.Write(f.renderFrame())
	b.WriteString("\x1b[J")
	newLines := f.frameLines()
	up := newLines - 1 - f.inputLine()
	if up > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", up)
	}
	b.WriteString("\r")
	if col := f.inputCol(); col > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", col)
	}
	_, _ = f.writer().Write(b.Bytes())
	f.lines = newLines
	f.parkedLine = f.inputLine()
}

func (f *addForm) close() {
	if f.lines > 0 && f.parkedLine > 0 {
		fmt.Fprintf(f.writer(), "\x1b[%dA", f.parkedLine)
	}
	_, _ = f.writer().Write([]byte("\r\x1b[J"))
	f.lines = 0
	f.parkedLine = 0
}
