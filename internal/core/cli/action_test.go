package cli

import (
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func TestHandleActionResultShellOutputExecute(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc", UsageCount: 0, Title: "test", Command: "echo hi"},
	}}

	got := captureStdout(func() {
		handleActionResult(s, &s.Commands[0], tui.ActionExecute, true)
	})
	if !contains(got, "EXEC:echo hi") {
		t.Fatalf("expected EXEC: prefix, got: %s", got)
	}
	if s.Commands[0].UsageCount != 1 {
		t.Fatalf("expected usage bump, got %d", s.Commands[0].UsageCount)
	}
}

func TestHandleActionResultShellOutputCopy(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "abc", UsageCount: 0, Title: "test", Command: "echo hi"},
	}}

	got := captureStdout(func() {
		handleActionResult(s, &s.Commands[0], tui.ActionCopy, true)
	})
	if !contains(got, "✓ echo hi") {
		t.Fatalf("expected ✓ prefix, got: %s", got)
	}
	if s.Commands[0].UsageCount != 1 {
		t.Fatalf("expected usage bump, got %d", s.Commands[0].UsageCount)
	}
}

func TestHandleActionResultShellOutputWorkflow(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{
			ID: "abc", UsageCount: 0, Title: "deploy", Command: "",
			Steps: []core.WorkflowStep{{Title: "build", Command: "go build"}},
		},
	}}

	_ = captureStdout(func() {
		handleActionResult(s, &s.Commands[0], tui.ActionCopy, true)
	})
	if s.Commands[0].UsageCount != 1 {
		t.Fatalf("expected usage bump for workflow, got %d", s.Commands[0].UsageCount)
	}
}

func TestBumpUsage(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Commands: []core.Command{
		{ID: "abc", UsageCount: 2, Title: "test", Command: "echo hi"},
	}}
	if err := store.Save(s); err != nil {
		t.Fatal(err)
	}

	bumpUsage(s, &s.Commands[0])
	if s.Commands[0].UsageCount != 3 {
		t.Fatalf("usage = %d, want 3", s.Commands[0].UsageCount)
	}
	if s.Commands[0].LastUsedAt.IsZero() {
		t.Fatal("LastUsedAt should be set")
	}

	// Verify it was saved
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Commands[0].UsageCount != 3 {
		t.Fatalf("persisted usage = %d, want 3", loaded.Commands[0].UsageCount)
	}
}

func TestConfirmCopy(t *testing.T) {
	s := &core.Store{}
	cmd := &core.Command{Command: "echo hello"}
	// confirmCopy prints output and tries clipboard — just verify it doesn't panic.
	confirmCopy(s, cmd)
}

func TestHumanCount(t *testing.T) {
	tests := []struct {
		data []byte
		want string
	}{
		{[]byte(`{"version":1,"commands":[{"id":"a","context":"g","title":"t","command":"c"}]}`), "1 command"},
		{[]byte(`{"version":1,"commands":[{},{},{}]}`), "3 commands"},
		{[]byte(`{"version":1,"commands":[]}`), "0 commands"},
		{[]byte(`not json`), "8 bytes"},
	}
	for _, tt := range tests {
		got := humanCount(tt.data)
		if got != tt.want {
			t.Errorf("humanCount(%q) = %q, want %q", string(tt.data), got, tt.want)
		}
	}
}
