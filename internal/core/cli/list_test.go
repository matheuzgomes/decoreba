package cli

import (
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func TestPrintContexts(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
		{ID: "2", Context: "docker", Title: "Ps", Command: "docker ps"},
		{ID: "3", Context: "git", Title: "Reset", Command: "git reset"},
	}}

	got := captureStdout(func() {
		printContexts(s)
	})
	if !contains(got, "git (2)") {
		t.Fatalf("missing git count, got: %s", got)
	}
	if !contains(got, "docker (1)") {
		t.Fatalf("missing docker, got: %s", got)
	}
}

func TestPrintContextsEmpty(t *testing.T) {
	got := captureStdout(func() {
		printContexts(&core.Store{})
	})
	if !contains(got, "No commands saved yet") {
		t.Fatalf("got: %s", got)
	}
}

func TestPrintContextCommands(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc", Context: "git", Title: "Stash", Command: "git stash", UsageCount: 1},
		{ID: "def", Context: "git", Title: "Reset", Command: "git reset", UsageCount: 10},
	}}

	got := captureStdout(func() {
		printContextCommands(s, "git")
	})
	if !contains(got, "[abc]") {
		t.Fatalf("missing id in: %s", got)
	}
	if !contains(got, "git stash") {
		t.Fatalf("missing command in: %s", got)
	}
}

func TestPrintContextCommandsNoMatch(t *testing.T) {
	got := captureStdout(func() {
		printContextCommands(&core.Store{Commands: []core.Command{
			{Context: "git", Title: "Stash", Command: "git stash"},
		}}, "docker")
	})
	if !contains(got, "No commands in context") {
		t.Fatalf("got: %s", got)
	}
}

func TestCmdListFallbackNoArgs(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdList(s, nil, false)
	})
	if !contains(got, "Available contexts:") {
		t.Fatalf("expected context list, got: %s", got)
	}
}

func TestCmdListFallbackWithContext(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdList(s, []string{"git"}, false)
	})
	if !contains(got, "[1]") {
		t.Fatalf("expected command list, got: %s", got)
	}
}
