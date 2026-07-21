package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func TestCmdImportFullFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	cmds := []core.Command{
		{ID: "x1", Context: "git", Title: "Stash", Command: "git stash"},
	}
	data, _ := json.Marshal(cmds)
	src := filepath.Join(tmp, "import.json")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{src})
	})
	if !contains(got, "Imported 1 commands") {
		t.Fatalf("unexpected output: %s", got)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(loaded.Commands))
	}
}

func TestCmdImportSimplifiedFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	simplified := []exportCmd{
		{Context: "docker", Title: "Ps", Command: "docker ps", Tags: []string{"list"}},
	}
	data, _ := json.Marshal(simplified)
	src := filepath.Join(tmp, "import.json")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{src})
	})
	if !contains(got, "Imported 1 commands") {
		t.Fatalf("unexpected output: %s", got)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 1 {
		t.Fatalf("got %d commands, want 1", len(loaded.Commands))
	}
	if loaded.Commands[0].Context != "docker" {
		t.Fatalf("got context %q", loaded.Commands[0].Context)
	}
}

func TestCmdImportSkipDuplicates(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1, Commands: []core.Command{
		{ID: "x1", Context: "git", Title: "Stash", Command: "git stash"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	cmds := []core.Command{
		{ID: "x2", Context: "git", Title: "Stash", Command: "git stash"},
	}
	data, _ := json.Marshal(cmds)
	src := filepath.Join(tmp, "import.json")
	if err := os.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{src})
	})
	if !contains(got, "skipped 1") {
		t.Fatalf("expected skip, got: %s", got)
	}
}

func TestCmdImportEmpty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(tmp, "import.json")
	if err := os.WriteFile(src, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{src})
	})
	if !contains(got, "Imported 0") {
		t.Fatalf("unexpected: %s", got)
	}
}
