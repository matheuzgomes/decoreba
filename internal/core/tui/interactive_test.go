package tui

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func mockTerminal() func() {
	origMakeRaw := makeRaw
	origReadInput := readInput
	origReadSize := readSize

	makeRaw = func() (func(), error) {
		return func() {}, nil
	}
	readSize = func() (int, int) {
		return 80, 24
	}

	return func() {
		makeRaw = origMakeRaw
		readInput = origReadInput
		readSize = origReadSize
	}
}

func setReadInput(seqs ...[]byte) func() {
	orig := readInput
	call := 0
	readInput = func(buf []byte) (int, error) {
		if call >= len(seqs) {
			buf[0] = 0x1b
			return 1, nil
		}
		s := seqs[call]
		call++
		copy(buf, s)
		return len(s), nil
	}
	return func() { readInput = orig }
}

// --- RunPalette ---

func TestRunPaletteCancel(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x1b})()

	s := &core.Store{}
	cmd, action, err := RunPalette(s, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil command on ESC")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d, want ActionCopy", action)
	}
}

func TestRunPaletteMakeRawError(t *testing.T) {
	defer mockTerminal()()
	makeRaw = func() (func(), error) {
		return nil, errors.New("term error")
	}

	s := &core.Store{}
	cmd, action, err := RunPalette(s, "", "")
	if err == nil || !strings.Contains(err.Error(), "term error") {
		t.Fatalf("expected term error, got %v", err)
	}
	if cmd != nil {
		t.Fatal("expected nil command")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d", action)
	}
}

func TestRunPaletteTerminalTooSmall(t *testing.T) {
	defer mockTerminal()()
	readSize = func() (int, int) { return 80, 2 }

	s := &core.Store{}
	cmd, action, err := RunPalette(s, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d", action)
	}
}

func TestRunPaletteSelectFirst(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x0d})()

	s := &core.Store{}
	s.Commands = []core.Command{
		{ID: "1", Context: "test", Title: "first", Command: "echo 1"},
	}
	cmd, action, err := RunPalette(s, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command on Enter")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d, want ActionCopy", action)
	}
	if cmd.ID != "1" {
		t.Fatalf("got cmd %s, want 1", cmd.ID)
	}
}

func TestRunPaletteEditAction(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x05})()

	s := &core.Store{}
	s.Commands = []core.Command{
		{ID: "1", Context: "test", Title: "first", Command: "echo 1"},
	}
	cmd, action, err := RunPalette(s, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command on ^E")
	}
	if action != ActionEdit {
		t.Fatalf("action=%d, want ActionEdit", action)
	}
}

func TestRunPaletteWithContext(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x0d})()

	s := &core.Store{}
	s.Commands = []core.Command{
		{ID: "1", Context: "work", Title: "task", Command: "make"},
		{ID: "2", Context: "home", Title: "chore", Command: "clean"},
	}
	cmd, _, err := RunPalette(s, "home", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.ID != "2" {
		t.Fatalf("got %s, want 2", cmd.ID)
	}
}

func TestRunPaletteApplyExecuteAndConfirm(t *testing.T) {
	defer mockTerminal()()
	// First 0x18 enters confirm-exec mode, second 0x18 confirms
	defer setReadInput([]byte{0x18}, []byte{0x18})()

	s := &core.Store{}
	s.Commands = []core.Command{
		{ID: "1", Context: "test", Title: "cmd", Command: "echo"},
	}
	cmd, action, err := RunPalette(s, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command")
	}
	if action != ActionExecute {
		t.Fatalf("action=%d, want ActionExecute", action)
	}
}

// --- RunAddForm / RunEditForm ---

func TestRunAddFormCancel(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x1b})()

	s := &core.Store{}
	cmd, err := RunAddForm(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil on cancel")
	}
}

func TestRunAddFormMakeRawError(t *testing.T) {
	defer mockTerminal()()
	makeRaw = func() (func(), error) {
		return nil, errors.New("raw fail")
	}

	s := &core.Store{}
	cmd, err := RunAddForm(s)
	if err == nil || !strings.Contains(err.Error(), "raw fail") {
		t.Fatalf("expected raw error, got %v", err)
	}
	if cmd != nil {
		t.Fatal("expected nil command")
	}
}

func TestRunAddFormSaveAndQuit(t *testing.T) {
	defer mockTerminal()()
	s := &core.Store{}

	inputs := [][]byte{
		[]byte("myctx"),       // type context
		{0x09},                // tab to title
		[]byte("my title"),    // type title
		{0x09},                // tab to command
		[]byte("echo hi"),     // type command
		{0x13},                // ^s to save
	}
	defer setReadInput(inputs...)()

	cmd, err := RunAddForm(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command after save")
	}
	if cmd.Context != "myctx" {
		t.Fatalf("context=%q, want myctx", cmd.Context)
	}
	if cmd.Title != "my title" {
		t.Fatalf("title=%q", cmd.Title)
	}
	if cmd.Command != "echo hi" {
		t.Fatalf("command=%q", cmd.Command)
	}
}

func TestRunAddFormSaveOnNotesEnter(t *testing.T) {
	defer mockTerminal()()
	s := &core.Store{}

	inputs := [][]byte{
		[]byte("ctx"),
		{0x09},             // tab to title
		[]byte("title"),
		{0x09},             // tab to command
		[]byte("cmd"),
		{0x09},             // tab to tags
		{0x09},             // tab to notes
		{0x0d},             // Enter on notes => save
	}
	defer setReadInput(inputs...)()

	cmd, err := RunAddForm(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.Context != "ctx" {
		t.Fatalf("context=%q", cmd.Context)
	}
}

func TestRunAddFormCancelDirtyThenEsc(t *testing.T) {
	defer mockTerminal()()
	s := &core.Store{}

	inputs := [][]byte{
		[]byte("x"),
		{0x1b},  // first ESC => confirm exit
		{0x1b},  // second ESC => discard
	}
	defer setReadInput(inputs...)()

	cmd, err := RunAddForm(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil after discard")
	}
}

func TestRunEditFormCancel(t *testing.T) {
	defer mockTerminal()()
	s := &core.Store{}
	existing := &core.Command{ID: "e1", Context: "ctx", Title: "old", Command: "echo old"}

	defer setReadInput([]byte{0x1b})()
	cmd, err := RunEditForm(s, existing)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil")
	}
}

// --- EditSteps ---

func TestEditStepsCancel(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x1b})()

	var buf bytes.Buffer
	steps, cancelled, err := EditSteps(nil, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if steps != nil {
		t.Fatal("expected nil steps on cancel")
	}
	if !cancelled {
		t.Fatal("expected cancelled=true")
	}
}

func TestEditStepsSaveEmpty(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x0d})()

	var buf bytes.Buffer
	steps, cancelled, err := EditSteps(nil, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("expected cancelled=false")
	}
	if len(steps) != 0 {
		t.Fatalf("got %d steps, want 0", len(steps))
	}
}

func TestEditStepsReadInputError(t *testing.T) {
	defer mockTerminal()()
	readInput = func(buf []byte) (int, error) {
		return 0, errors.New("read fail")
	}

	var buf bytes.Buffer
	steps, cancelled, err := EditSteps(nil, &buf)
	if err == nil || !strings.Contains(err.Error(), "read fail") {
		t.Fatalf("expected read error, got %v", err)
	}
	if steps != nil {
		t.Fatal("expected nil steps")
	}
	if cancelled {
		t.Fatal("expected cancelled=false")
	}
}

// --- ResolveCommandInteractive ---

func TestResolveInteractiveNoVars(t *testing.T) {
	cmd, cancelled, err := ResolveCommandInteractive("echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("cancelled should be false")
	}
	if cmd != "echo hello" {
		t.Fatalf("got %q", cmd)
	}
}

func TestResolveInteractiveHasVarsMakeRawError(t *testing.T) {
	defer mockTerminal()()
	makeRaw = func() (func(), error) {
		return nil, errors.New("raw err")
	}

	_, _, err := ResolveCommandInteractive("echo {{name}}")
	if err == nil || !strings.Contains(err.Error(), "raw err") {
		t.Fatalf("expected raw err, got %v", err)
	}
}

func TestResolveInteractiveVarsCancel(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x1b})()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd, cancelled, err := ResolveCommandInteractive("echo {{name}}")
	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatal(err)
	}
	if !cancelled {
		t.Fatal("expected cancelled")
	}
	if cmd != "" {
		t.Fatalf("cmd=%q, want empty", cmd)
	}
}

func TestResolveInteractiveVarsEnterValue(t *testing.T) {
	defer mockTerminal()()
	// readVar reads one byte at a time, so send each byte individually + Enter
	defer setReadInput(
		[]byte{'w'}, []byte{'o'}, []byte{'r'}, []byte{'l'}, []byte{'d'}, []byte{0x0d},
	)()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd, cancelled, err := ResolveCommandInteractive("echo {{name}}")
	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("unexpected cancel")
	}
	if cmd != "echo world" {
		t.Fatalf("cmd=%q, want echo world", cmd)
	}
}

func TestResolveInteractiveVarWithDefault(t *testing.T) {
	defer mockTerminal()()
	// Enter with default
	defer setReadInput([]byte{0x0d})()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd, cancelled, err := ResolveCommandInteractive("echo {{name:default}}")
	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err != nil {
		t.Fatal(err)
	}
	if cancelled {
		t.Fatal("unexpected cancel")
	}
	if cmd != "echo default" {
		t.Fatalf("cmd=%q, want echo default", cmd)
	}
}

func TestResolveInteractiveReadInputError(t *testing.T) {
	defer mockTerminal()()
	readInput = func(buf []byte) (int, error) {
		return 0, errors.New("read err")
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_, _, err := ResolveCommandInteractive("echo {{name}}")
	w.Close()
	os.Stdout = oldStdout
	_ = r

	if err == nil || !strings.Contains(err.Error(), "read err") {
		t.Fatalf("expected read err, got %v", err)
	}
}

// --- RunWorkflow ---

func TestRunWorkflowNonWorkflow(t *testing.T) {
	cmd := &core.Command{ID: "1", Title: "simple", Command: "echo"}
	err := RunWorkflow(cmd)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunWorkflowCancel(t *testing.T) {
	defer mockTerminal()()
	defer setReadInput([]byte{0x1b})()

	cmd := &core.Command{
		ID:    "wf1",
		Title: "steps",
		Steps: []core.WorkflowStep{
			{Title: "step1", Command: "echo 1"},
		},
	}
	err := RunWorkflow(cmd)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunWorkflowReadInputError(t *testing.T) {
	defer mockTerminal()()
	readInput = func(buf []byte) (int, error) {
		return 0, errors.New("read err")
	}

	cmd := &core.Command{
		ID:    "wf2",
		Title: "steps",
		Steps: []core.WorkflowStep{
			{Title: "s1", Command: "echo 1"},
		},
	}
	err := RunWorkflow(cmd)
	if err == nil || !strings.Contains(err.Error(), "read err") {
		t.Fatalf("expected read err, got %v", err)
	}
}

func TestRunWorkflowMakeRawError(t *testing.T) {
	defer mockTerminal()()
	makeRaw = func() (func(), error) {
		return nil, errors.New("raw err")
	}

	cmd := &core.Command{
		ID:    "wf3",
		Title: "steps",
		Steps: []core.WorkflowStep{
			{Title: "s1", Command: "echo 1"},
		},
	}
	err := RunWorkflow(cmd)
	if err == nil || !strings.Contains(err.Error(), "raw err") {
		t.Fatalf("expected raw err, got %v", err)
	}
}

// --- RunListBrowser ---

func TestRunListBrowserEmpty(t *testing.T) {
	s := &core.Store{}
	cmd, action, err := RunListBrowser(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d", action)
	}
}

func TestRunListBrowserSelectThenPaletteCancel(t *testing.T) {
	defer mockTerminal()()
	// first read: Enter (select first context)
	// RunListBrowser then calls RunPalette which needs its own readInput
	inputs := [][]byte{
		{0x0d}, // Enter in browser
		{0x1b}, // ESC in palette
	}
	defer setReadInput(inputs...)()

	s := &core.Store{}
	s.Commands = []core.Command{
		{ID: "1", Context: "work", Title: "task", Command: "make"},
	}
	cmd, action, err := RunListBrowser(s)
	if err != nil {
		t.Fatal(err)
	}
	if cmd != nil {
		t.Fatal("expected nil after palette cancel")
	}
	if action != ActionCopy {
		t.Fatalf("action=%d", action)
	}
}
