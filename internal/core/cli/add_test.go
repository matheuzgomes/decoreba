package cli

import (
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func TestCmdAddFallbackFull(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	input := "tmux\nNew Session\ntmux new-session -s work\ntmux, session\nCreates a new session\n"
	restore := mockStdin(t, input)
	defer restore()

	got := captureStdout(func() {
		cmdAddFallback(s)
	})
	if !contains(got, "tmux") || !contains(got, "Command saved") {
		t.Fatalf("unexpected output: %s", got)
	}
	if len(s.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(s.Commands))
	}
	cmd := s.Commands[0]
	if cmd.Context != "tmux" {
		t.Fatalf("context=%q", cmd.Context)
	}
	if cmd.Title != "New Session" {
		t.Fatalf("title=%q", cmd.Title)
	}
	if cmd.Command != "tmux new-session -s work" {
		t.Fatalf("command=%q", cmd.Command)
	}
	if len(cmd.Tags) != 2 || cmd.Tags[0] != "tmux" || cmd.Tags[1] != "session" {
		t.Fatalf("tags=%v", cmd.Tags)
	}
	if cmd.Notes != "Creates a new session" {
		t.Fatalf("notes=%q", cmd.Notes)
	}
}

func TestCmdAddFallbackMissingRequired(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	// Empty context and command.
	restore := mockStdin(t, "\n\n\n\n\n")
	defer restore()

	got := captureStdout(func() {
		cmdAddFallback(s)
	})
	if !contains(got, "Context and command are required") {
		t.Fatalf("unexpected: %s", got)
	}
	if len(s.Commands) != 0 {
		t.Fatal("should not save when required fields are missing")
	}
}

func TestCmdAddFallbackEmptyTagsNotes(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	restore := mockStdin(t, "git\nCommit\ngit commit -m \"msg\"\n\n\n")
	defer restore()

	got := captureStdout(func() {
		cmdAddFallback(s)
	})
	if !contains(got, "Command saved") {
		t.Fatalf("unexpected: %s", got)
	}
	if len(s.Commands) != 1 {
		t.Fatalf("got %d", len(s.Commands))
	}
	if s.Commands[0].Tags != nil && len(s.Commands[0].Tags) != 0 {
		t.Fatalf("tags=%v", s.Commands[0].Tags)
	}
}

func TestCmdAddFallbackLowercaseContext(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	restore := mockStdin(t, "DOCKER\nPs\ndocker ps\n\n\n")
	defer restore()

	cmdAddFallback(s)
	if s.Commands[0].Context != "docker" {
		t.Fatalf("context should be lowercased, got %q", s.Commands[0].Context)
	}
}
