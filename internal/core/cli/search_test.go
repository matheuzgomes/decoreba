package cli

import (
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestRunSearchFallbackNoResults(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}}

	got := captureStdout(func() {
		runSearchFallback(s, "docker", "")
	})
	if !contains(got, "No commands in context") {
		t.Fatalf("expected no-commands message, got: %s", got)
	}
	if !contains(got, "Available contexts") {
		t.Fatalf("expected context list, got: %s", got)
	}
}

func TestRunSearchFallbackQueryFromArg(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash pop", Command: "git stash pop"},
		{ID: "2", Context: "git", Title: "Apply stash", Command: "git stash apply"},
	}}

	restore := mockStdin(t, "1\n")
	defer restore()

	got := captureStdout(func() {
		runSearchFallback(s, "git", "stash")
	})
	if !contains(got, "[1] (git) Stash pop") {
		t.Fatalf("expected numbered list, got: %s", got)
	}
}

func TestRunSearchFallbackQueryFromPrompt(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Stash pop", Command: "git stash pop"},
	}}

	restore := mockStdin(t, "stash\n1\n")
	defer restore()

	got := captureStdout(func() {
		runSearchFallback(s, "", "")
	})
	if !contains(got, "[1]") {
		t.Fatalf("expected numbered list, got: %s", got)
	}
}

func TestFallbackSelectorValidChoice(t *testing.T) {
	cmds := []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
		{ID: "2", Context: "git", Title: "Reset", Command: "git reset"},
	}

	restore := mockStdin(t, "2\n")
	defer restore()

	got, err := fallbackSelector(cmds)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ID != "2" {
		t.Fatalf("got id=%v", got)
	}
}

func TestFallbackSelectorEmptyChoice(t *testing.T) {
	cmds := []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}

	restore := mockStdin(t, "\n")
	defer restore()

	got, err := fallbackSelector(cmds)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil on empty choice")
	}
}

func TestFallbackSelectorInvalidChoice(t *testing.T) {
	cmds := []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}

	restore := mockStdin(t, "abc\n")
	defer restore()

	got, err := fallbackSelector(cmds)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil on invalid choice")
	}
}

func TestFallbackSelectorOutOfRange(t *testing.T) {
	cmds := []core.Command{
		{ID: "1", Context: "git", Title: "Stash", Command: "git stash"},
	}

	restore := mockStdin(t, "99\n")
	defer restore()

	got, err := fallbackSelector(cmds)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil on out-of-range choice")
	}
}
