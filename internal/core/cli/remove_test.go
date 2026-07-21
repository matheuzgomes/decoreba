package cli

import (
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func TestCmdRemove(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
		{ID: "def456", Context: "docker", Title: "Ps", Command: "docker ps"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdRemove(s, []string{"abc123"})
	})
	if !contains(got, "✓ Removed") {
		t.Fatalf("expected success, got: %s", got)
	}
	if len(s.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(s.Commands))
	}
	if s.Commands[0].ID != "def456" {
		t.Fatal("wrong command removed")
	}
}

func TestCmdRemoveNoArgs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdRemove(s, nil)
	})
	if !contains(got, "Usage:") {
		t.Fatalf("expected usage, got: %s", got)
	}
}

func TestCmdRemoveNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdRemove(s, []string{"xyz"})
	})
	if !contains(got, "No command found") {
		t.Fatalf("expected not found, got: %s", got)
	}
}

func TestCmdRemoveAmbiguous(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
		{ID: "abc456", Context: "git", Title: "Reset", Command: "git reset"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdRemove(s, []string{"abc"})
	})
	if !contains(got, "Ambiguous") {
		t.Fatalf("expected ambiguous, got: %s", got)
	}
}

func TestCmdRemovePersists(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc123", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	cmdRemove(s, []string{"abc123"})
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 0 {
		t.Fatal("remove should persist to store")
	}
}
