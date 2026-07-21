package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

// mockNonTerminal replaces os.Stdin with a pipe so term.IsTerminal() returns false.
func mockNonTerminal(t *testing.T) func() {
	r, _, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	oldReader := reader
	os.Stdin = r
	reader = nil
	return func() {
		os.Stdin = oldStdin
		reader = oldReader
		r.Close()
	}
}

func TestCmdEditNoArgs(t *testing.T) {
	s := &core.Store{}
	got := captureStdout(func() {
		cmdEdit(s, nil)
	})
	if !contains(got, "Usage:") {
		t.Fatalf("expected usage, got: %s", got)
	}
}

func TestCmdEditNotFound(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	got := captureStdout(func() {
		cmdEdit(s, []string{"xyz"})
	})
	if !contains(got, "No command found") {
		t.Fatalf("expected not found, got: %s", got)
	}
}

func TestCmdEditAmbiguous(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
		{ID: "abc456", Context: "git", Title: "Reset", Command: "git reset"},
	}}
	got := captureStdout(func() {
		cmdEdit(s, []string{"abc"})
	})
	if !contains(got, "Ambiguous") {
		t.Fatalf("expected ambiguous, got: %s", got)
	}
}

func TestCmdEditTerminalRequired(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	restore := mockNonTerminal(t)
	defer restore()

	got := captureStdout(func() {
		cmdEdit(s, []string{"abc123"})
	})
	if !contains(got, "Terminal required") {
		t.Fatalf("expected terminal required, got: %s", got)
	}
}
