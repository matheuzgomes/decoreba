package tui

import (
	"bytes"
	"strings"
	"testing"

	"decoreba/internal/core"
)

func newTestAddForm() *addForm {
	return &addForm{
		store: &core.Store{Commands: []core.Command{
			{ID: "1", Context: "docker", Title: "prune", Command: "docker container prune"},
			{ID: "2", Context: "git", Title: "undo", Command: "git reset --soft HEAD~1"},
		}},
		errField: -1,
		contexts: []string{"docker", "git"},
		width:    80,
		height:   24,
	}
}

func TestAddFormRenderLayout(t *testing.T) {
	f := newTestAddForm()
	if got := f.frameLines(); got != 9 {
		t.Fatalf("frameLines = %d, want 9", got)
	}
	lines := frameLines(t, f.renderFrame())
	if len(lines) != 9 {
		t.Fatalf("rendered %d lines, want 9", len(lines))
	}
	if !strings.Contains(lines[0], boxTL) || !strings.Contains(lines[0], boxTR) {
		t.Fatalf("top border: %q", lines[0])
	}
	if !strings.Contains(lines[1], "●") || !strings.Contains(lines[1], addFormHeader) {
		t.Fatalf("header: %q", lines[1])
	}
	if !strings.Contains(lines[1], boxV) {
		t.Fatalf("header must sit inside the box: %q", lines[1])
	}
	if !strings.Contains(lines[2], "context") {
		t.Fatalf("context field: %q", lines[2])
	}
	if !strings.Contains(lines[3], "title") {
		t.Fatalf("title field: %q", lines[3])
	}
	if !strings.Contains(lines[4], "command") {
		t.Fatalf("command field: %q", lines[4])
	}
	if !strings.Contains(lines[5], "tags") {
		t.Fatalf("tags field: %q", lines[5])
	}
	if !strings.Contains(lines[6], "notes") {
		t.Fatalf("notes field: %q", lines[6])
	}
	if !strings.Contains(lines[7], addFormHint) {
		t.Fatalf("hint: %q", lines[7])
	}
	if !strings.Contains(lines[8], boxBL) || !strings.Contains(lines[8], boxBR) {
		t.Fatalf("bottom border: %q", lines[8])
	}
	if !strings.Contains(lines[2], ansiFocusBg) {
		t.Fatalf("focused field should have bg: %q", lines[2])
	}
}

func TestAddFormContextGhostAndAccept(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("doc")
	f.cursor = 3
	f.focus = fieldContext

	if got := f.contextSuggestion(); got != "ker" {
		t.Fatalf("suggestion = %q, want ker", got)
	}
	frame := string(f.renderFrame())
	if !strings.Contains(frame, "doc") || !strings.Contains(frame, ansiDim+"ker"+ansiReset) {
		t.Fatalf("ghost text missing: %q", frame)
	}

	done, cmd := f.apply([]keyEvent{{kind: keyTab}})
	if done || cmd != nil {
		t.Fatal("tab accept should not save")
	}
	if string(f.fields[fieldContext]) != "docker" {
		t.Fatalf("context = %q, want docker", string(f.fields[fieldContext]))
	}
	if f.focus != fieldContext {
		t.Fatalf("focus should stay on context after accept, got %d", f.focus)
	}
}

func TestAddFormRightAcceptsSuggestion(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("gi")
	f.cursor = 2
	f.apply([]keyEvent{{kind: keyRight}})
	if string(f.fields[fieldContext]) != "git" {
		t.Fatalf("context = %q, want git", string(f.fields[fieldContext]))
	}
}

func TestAddFormTabShiftTabNavigate(t *testing.T) {
	f := newTestAddForm()
	f.apply([]keyEvent{{kind: keyTab}})
	if f.focus != fieldTitle {
		t.Fatalf("focus = %d, want title", f.focus)
	}
	f.apply([]keyEvent{{kind: keyTab}, {kind: keyTab}})
	if f.focus != fieldTags {
		t.Fatalf("focus = %d, want tags", f.focus)
	}
	f.apply([]keyEvent{{kind: keyShiftTab}})
	if f.focus != fieldCommand {
		t.Fatalf("focus = %d, want command", f.focus)
	}
	f.focus = fieldContext
	f.apply([]keyEvent{{kind: keyShiftTab}})
	if f.focus != fieldNotes {
		t.Fatalf("shift-tab from first should wrap to notes, got %d", f.focus)
	}
}

func TestAddFormEnterAdvancesThenSaves(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("tmux")
	f.fields[fieldCommand] = []rune("tmux ls")
	f.focus = fieldContext
	f.cursor = len(f.fields[fieldContext])

	done, cmd := f.apply([]keyEvent{{kind: keyEnter}})
	if done || cmd != nil {
		t.Fatal("enter on non-last field should advance")
	}
	if f.focus != fieldTitle {
		t.Fatalf("focus = %d, want title", f.focus)
	}

	f.focus = fieldNotes
	done, cmd = f.apply([]keyEvent{{kind: keyEnter}})
	if !done || cmd == nil {
		t.Fatal("enter on notes should save")
	}
	if cmd.Context != "tmux" || cmd.Command != "tmux ls" {
		t.Fatalf("saved cmd: %+v", cmd)
	}
}

func TestAddFormCtrlSSavesFromAnyField(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("git")
	f.fields[fieldCommand] = []rune("git status")
	f.focus = fieldTitle
	done, cmd := f.apply([]keyEvent{{kind: keySave}})
	if !done || cmd == nil {
		t.Fatal("ctrl+s should save")
	}
	if cmd.Context != "git" || cmd.Command != "git status" {
		t.Fatalf("saved: %+v", cmd)
	}
}

func TestAddFormValidationKeepsData(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldTitle] = []rune("my title")
	f.fields[fieldNotes] = []rune("note")
	f.focus = fieldNotes

	done, cmd := f.apply([]keyEvent{{kind: keySave}})
	if done || cmd != nil {
		t.Fatal("save with empty context must not close")
	}
	if f.errField != fieldContext || f.errMsg == "" {
		t.Fatalf("expected context error, errField=%d msg=%q", f.errField, f.errMsg)
	}
	if f.focus != fieldContext {
		t.Fatalf("focus should jump to context, got %d", f.focus)
	}
	if f.errFlash <= 0 {
		t.Fatal("errFlash should be set so the label blinks warn")
	}
	if string(f.fields[fieldTitle]) != "my title" || string(f.fields[fieldNotes]) != "note" {
		t.Fatal("typed data must be preserved")
	}

	f.fields[fieldContext] = []rune("docker")
	f.cursor = len(f.fields[fieldContext])
	f.focus = fieldNotes
	done, cmd = f.apply([]keyEvent{{kind: keySave}})
	if done || cmd != nil {
		t.Fatal("save with empty command must not close")
	}
	if f.errField != fieldCommand {
		t.Fatalf("errField = %d, want command", f.errField)
	}
	if f.focus != fieldCommand {
		t.Fatalf("focus = %d, want command", f.focus)
	}
	frame := string(f.renderFrame())
	if !strings.Contains(frame, "command is required") {
		t.Fatalf("hint should show error: %q", frame)
	}
	if !strings.Contains(frame, ansiWarn+"command") {
		t.Fatalf("errored label should flash warn: %q", frame)
	}
}

func TestAddFormCancel(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("x")
	done, cmd := f.apply([]keyEvent{{kind: keyEsc}})
	if !done || cmd != nil {
		t.Fatalf("esc should cancel: done=%v cmd=%v", done, cmd)
	}
}

func TestAddFormTagsChips(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldTags] = []rune("docker, prod")
	f.focus = fieldTags
	f.cursor = len(f.fields[fieldTags])
	frame := string(f.renderFrame())
	if !strings.Contains(frame, "\x1b[48;5;") {
		t.Fatalf("tags should render with bg chips: %q", frame)
	}
	if !strings.Contains(frame, " docker ") || !strings.Contains(frame, "prod") {
		t.Fatalf("chip contents missing: %q", frame)
	}
}

func TestAddFormTagsSplitOnSave(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("docker")
	f.fields[fieldCommand] = []rune("docker ps")
	f.fields[fieldTags] = []rune(" a , b, ,c ")
	done, cmd := f.apply([]keyEvent{{kind: keySave}})
	if !done || cmd == nil {
		t.Fatal("save should succeed")
	}
	want := []string{"a", "b", "c"}
	if len(cmd.Tags) != 3 || cmd.Tags[0] != want[0] || cmd.Tags[1] != want[1] || cmd.Tags[2] != want[2] {
		t.Fatalf("tags = %v, want %v", cmd.Tags, want)
	}
}

func TestAddFormCursorEdit(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("doker")
	f.cursor = 2
	f.apply([]keyEvent{{kind: keyRune, r: 'c'}})
	if string(f.fields[fieldContext]) != "docker" {
		t.Fatalf("insert: %q", string(f.fields[fieldContext]))
	}
	f.cursor = 6
	f.apply([]keyEvent{{kind: keyLeft}, {kind: keyBackspace}})
	if string(f.fields[fieldContext]) != "dockr" {
		t.Fatalf("backspace: %q", string(f.fields[fieldContext]))
	}
	f.fields[fieldContext] = []rune("docker")
	f.cursor = 3
	f.apply([]keyEvent{{kind: keyDelete}})
	if string(f.fields[fieldContext]) != "docer" {
		t.Fatalf("delete: %q", string(f.fields[fieldContext]))
	}
}

func TestAddFormRedrawAndClose(t *testing.T) {
	var out bytes.Buffer
	f := newTestAddForm()
	f.out = &out
	f.redraw()
	if f.lines != 9 || f.parkedLine != 2 {
		t.Fatalf("lines=%d parked=%d", f.lines, f.parkedLine)
	}
	frame1 := out.String()
	if !strings.HasPrefix(frame1, "\r\x1b[2K\r") {
		t.Fatalf("first redraw: %q", frame1)
	}
	if !strings.Contains(frame1, "\x1b[6A\r") {
		t.Fatalf("should park at context line from bottom: %q", frame1)
	}
	if !strings.Contains(frame1, "\x1b[11C") {
		t.Fatalf("caret should sit after the label pad: %q", frame1)
	}

	out.Reset()
	f.focus = fieldCommand
	f.cursor = 0
	f.redraw()
	frame2 := out.String()
	if !strings.HasPrefix(frame2, "\x1b[2A\r") {
		t.Fatalf("second redraw must climb to top first: %q", frame2)
	}
	if f.parkedLine != 4 {
		t.Fatalf("parkedLine = %d, want 4", f.parkedLine)
	}

	out.Reset()
	f.close()
	got := out.String()
	if !strings.HasPrefix(got, "\x1b[4A") || !strings.HasSuffix(got, "\r\x1b[J") {
		t.Fatalf("close sequence = %q", got)
	}
	if f.lines != 0 {
		t.Fatalf("lines after close = %d", f.lines)
	}
}

func TestAddFormDotTakesContextColor(t *testing.T) {
	f := newTestAddForm()
	f.fields[fieldContext] = []rune("docker")
	frame := string(f.renderFrame())
	color := contextColor("docker")
	if !strings.Contains(frame, color+"●"+ansiReset) {
		t.Fatalf("dot should use context color %q in %q", color, frame)
	}
}

func TestAddFormBoxMatchesWidth(t *testing.T) {
	f := newTestAddForm()
	f.width = 80
	wantInner := boxTotalWidth(80)
	for _, l := range frameLines(t, f.renderFrame()) {
		if n := visibleWidth(l); n != wantInner {
			t.Fatalf("line visible width = %d, want %d: %q", n, wantInner, l)
		}
	}
}
