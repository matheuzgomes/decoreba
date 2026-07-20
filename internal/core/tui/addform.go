package tui

import (
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

	addFormHint  = "tab next  ⇧tab previous  ^s save  ^w workflow  esc cancel"
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
	frame
	store         *core.Store
	fields        [fieldCount][]rune
	focus         int
	cursor        int
	errField      int
	errMsg        string
	errFlash      int
	contexts      []string
	editing       bool
	existing      *core.Command
	isWorkflow    bool
	workflowSteps []core.WorkflowStep
	confirmExit   bool
	width         int
	height        int
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
	f.frame = newFrame(nil)
	if UseTTY {
		ff, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			f.w = ff
			defer ff.Close()
		}
	}
	if existing != nil {
		f.fields[fieldContext] = []rune(existing.Context)
		f.fields[fieldTitle] = []rune(existing.Title)
		f.fields[fieldCommand] = []rune(existing.Command)
		f.fields[fieldTags] = []rune(strings.Join(existing.Tags, ", "))
		f.fields[fieldNotes] = []rune(existing.Notes)
		f.focus = fieldContext
		f.cursor = len(f.fields[fieldContext])
		if existing.IsWorkflow() {
			f.isWorkflow = true
			f.workflowSteps = existing.Steps
		}
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
		n, err := term.ReadInput(buf)
		if err != nil {
			f.close()
			return nil, err
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
			if f.isDirty() && !f.confirmExit {
				f.confirmExit = true
				f.errMsg = "Unsaved changes — Esc again to discard"
				f.errField = -1
				f.errFlash = 4
				continue
			}
			return true, nil
		case keySave:
			if c, ok := f.trySave(); ok {
				return true, c
			}
		case keyWorkflow:
			f.isWorkflow = !f.isWorkflow
			f.clearErrOnEdit()
		case keyEnter:
			if f.focus == fieldCommand && f.isWorkflow {
				f.close()
				steps, cancelled, err := EditSteps(f.workflowSteps, f.width, nil)
				f.lines = 0
				f.parkedLine = 0
				if err != nil {
					return true, nil
				}
				if !cancelled {
					f.workflowSteps = steps
				}
			} else if f.focus == fieldNotes {
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
	f.clearConfirm()
	if f.errField == f.focus {
		f.errField = -1
		f.errMsg = ""
		f.errFlash = 0
	}
}

// isDirty reports whether the form has unsaved changes.
func (f *addForm) isDirty() bool {
	if f.editing && f.existing != nil {
		if string(f.fields[fieldContext]) != f.existing.Context {
			return true
		}
		if string(f.fields[fieldTitle]) != f.existing.Title {
			return true
		}
		if string(f.fields[fieldCommand]) != f.existing.Command {
			return true
		}
		if string(f.fields[fieldTags]) != strings.Join(f.existing.Tags, ", ") {
			return true
		}
		if string(f.fields[fieldNotes]) != f.existing.Notes {
			return true
		}
		return false
	}
	for i := 0; i < fieldCount; i++ {
		if len(f.fields[i]) > 0 {
			return true
		}
	}
	return false
}

func (f *addForm) clearConfirm() {
	if f.confirmExit {
		f.confirmExit = false
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
	if cmdStr == "" && !f.isWorkflow {
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
		cmd.Pinned = f.existing.Pinned
	} else {
		cmd.ID = core.GenID()
	}

	if f.isWorkflow {
		cmd.Steps = f.workflowSteps
		cmd.Command = ""
	}
	return cmd, true
}

func (f *addForm) frameLines() int {
	n := 1 + 1 + fieldCount + 1 + 1
	if f.height > 0 && n > f.height {
		n = f.height
	}
	return n
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
	f.draw(f.renderFrame(), f.inputLine(), f.inputCol())
}

func (f *addForm) close() {
	f.dismiss()
}


