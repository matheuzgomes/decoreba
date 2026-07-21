package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func TestCmdExportStdout(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset", Tags: []string{"fix"}},
		{ID: "2", Context: "docker", Title: "Prune", Command: "docker system prune", Notes: "cleanup"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdExport(s, nil)
	})
	if got == "" {
		t.Fatal("expected output")
	}
	if !contains(got, `"context": "git"`) {
		t.Fatalf("missing context in:\n%s", got)
	}
	if contains(got, `"usage_count"`) || contains(got, `"last_used_at"`) {
		t.Fatal("non-full export should not include meta fields")
	}
}

func TestCmdExportFull(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1, Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	got := captureStdout(func() {
		cmdExport(s, []string{"--full"})
	})
	if !contains(got, `"id": "1"`) {
		t.Fatalf("full export should include id in:\n%s", got)
	}
	if !contains(got, `"usage_count"`) {
		t.Fatal("full export should include usage_count")
	}
}

func TestCmdExportToFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "export.json")
	got := captureStdout(func() {
		cmdExport(s, []string{out})
	})
	if !contains(got, "Exported 1 commands to") {
		t.Fatalf("unexpected output: %s", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(data), `"command": "git reset"`) {
		t.Fatalf("unexpected file content: %s", string(data))
	}
}

func TestCmdExportFullToFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1, Commands: []core.Command{
		{ID: "1", Context: "git", Title: "Undo", Command: "git reset"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(tmp, "export-full.json")
	got := captureStdout(func() {
		cmdExport(s, []string{"--full", out})
	})
	if !contains(got, "Exported 1 commands to") {
		t.Fatalf("unexpected output: %s", got)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(data), `"id": "1"`) {
		t.Fatal("full export should include id")
	}
}
