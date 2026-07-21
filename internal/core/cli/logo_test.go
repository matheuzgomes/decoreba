package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRpt(t *testing.T) {
	t.Run("repeats string n times", func(t *testing.T) {
		if got := rpt("-", 3); got != "---" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("zero times", func(t *testing.T) {
		if got := rpt("x", 0); got != "" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("empty string", func(t *testing.T) {
		if got := rpt("", 5); got != "" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestShowVersionBox(t *testing.T) {
	got := captureStdout(func() {
		showVersionBox()
	})
	if !contains(got, "decoreba v") {
		t.Fatalf("missing version, got: %s", got)
	}
	if !contains(got, "╭") || !contains(got, "╰") {
		t.Fatalf("expected box drawing chars, got: %s", got)
	}
}

func TestIsFirstRun(t *testing.T) {
	t.Run("no config file returns true", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))
		if !isFirstRun() {
			t.Fatal("expected first run when no config exists")
		}
	})
	t.Run("existing config returns false", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))
		os.WriteFile(filepath.Join(tmp, "commands.json"), []byte("{}"), 0o644)
		if isFirstRun() {
			t.Fatalf("expected false when config exists at %s", tmp)
		}
	})
}
