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

func TestCmdImportPet(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(tmp, "snippets.toml")
	if err := os.WriteFile(src, []byte(`
[[snippets]]
description = "List files"
command = "ls -la"
tag = ["file"]
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{"pet", "--path", src})
	})
	if !contains(got, "Imported: 1") {
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

func TestCmdImportPetDryRun(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1, Commands: []core.Command{
		{ID: "x1", Context: "pet-imported", Title: "Existing", Command: "echo existing"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	src := filepath.Join(tmp, "snippets.toml")
	if err := os.WriteFile(src, []byte(`
[[snippets]]
description = "New"
command = "echo new"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{"pet", "--path", src, "--dry-run"})
	})
	if !contains(got, "Would import: 1") {
		t.Fatalf("expected dry-run report, got: %s", got)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 1 {
		t.Fatalf("expected store unchanged (1 command), got %d", len(loaded.Commands))
	}
}

func TestCmdImportNavi(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	cheatDir := filepath.Join(tmp, "cheats")
	if err := os.MkdirAll(cheatDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cheatDir, "git.cheat"), []byte(`% git

# Stash
git stash

# Log
git log
`), 0o644); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdImport(s, []string{"navi", "--path", cheatDir})
	})
	if !contains(got, "Imported: 2") {
		t.Fatalf("unexpected output: %s", got)
	}
}

func TestResolveContexts(t *testing.T) {
	s := &core.Store{Version: 1, Commands: []core.Command{
		{ID: "x1", Context: "Docker", Title: "Ps", Command: "docker ps"},
	}}

	cmds := []core.Command{
		{ID: "x2", Context: "docker", Title: "Stop", Command: "docker stop"},
		{ID: "x3", Context: "unique", Title: "Build", Command: "go build"},
	}

	resolveContexts(s, cmds)

	if cmds[0].Context != "Docker" {
		t.Errorf("resolved context = %q, want 'Docker' (preserve casing)", cmds[0].Context)
	}
	if cmds[1].Context != "unique" {
		t.Errorf("non-matching context changed to %q", cmds[1].Context)
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
