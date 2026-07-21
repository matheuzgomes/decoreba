package core

import (
	"strings"
	"testing"
	"time"
)

func TestIsWorkflow(t *testing.T) {
	if (&Command{Steps: nil}).IsWorkflow() {
		t.Fatal("nil steps should not be workflow")
	}
	if (&Command{Steps: []WorkflowStep{}}).IsWorkflow() {
		t.Fatal("empty steps should not be workflow")
	}
	if !(&Command{Steps: []WorkflowStep{{Title: "x", Command: "y"}}}).IsWorkflow() {
		t.Fatal("non-empty steps should be workflow")
	}
}

func TestGenID(t *testing.T) {
	m := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := GenID()
		if len(id) != 8 {
			t.Fatalf("GenID() = %q, want 8 hex chars", id)
		}
		if m[id] {
			t.Fatalf("duplicate id: %q", id)
		}
		m[id] = true
	}
}

func mkCmd(id, ctx, title, cmd string, tags ...string) Command {
	return Command{ID: id, Context: ctx, Title: title, Command: cmd, Tags: tags}
}

func TestFindByPrefix(t *testing.T) {
	s := &Store{Commands: []Command{
		mkCmd("abc123", "git", "Undo", "git reset"),
		mkCmd("abc456", "git", "Stash", "git stash"),
		mkCmd("def789", "docker", "Ps", "docker ps"),
	}}

	t.Run("exact", func(t *testing.T) {
		cmd, n := s.FindByPrefix("abc123")
		if n != 1 || cmd == nil || cmd.ID != "abc123" {
			t.Fatalf("got cmd=%v n=%d", cmd, n)
		}
	})
	t.Run("prefix", func(t *testing.T) {
		cmd, n := s.FindByPrefix("abc")
		if n != 2 {
			t.Fatalf("want 2 matches, got %d", n)
		}
		if cmd.ID != "abc456" {
			t.Fatalf("want last match abc456, got %s", cmd.ID)
		}
	})
	t.Run("no match", func(t *testing.T) {
		cmd, n := s.FindByPrefix("zzz")
		if n != 0 || cmd != nil {
			t.Fatalf("got cmd=%v n=%d", cmd, n)
		}
	})
	t.Run("empty store", func(t *testing.T) {
		cmd, n := (&Store{}).FindByPrefix("abc")
		if n != 0 || cmd != nil {
			t.Fatalf("got cmd=%v n=%d", cmd, n)
		}
	})
}

func TestRemoveByPrefix(t *testing.T) {
	s := &Store{Commands: []Command{
		mkCmd("abc123", "git", "Undo", "git reset"),
		mkCmd("def456", "docker", "Ps", "docker ps"),
	}}

	t.Run("success", func(t *testing.T) {
		s2 := &Store{Commands: append([]Command(nil), s.Commands...)}
		removed, n := s2.RemoveByPrefix("abc123")
		if n != 1 || removed == nil || removed.ID != "abc123" {
			t.Fatalf("got removed=%v n=%d", removed, n)
		}
		if len(s2.Commands) != 1 || s2.Commands[0].ID != "def456" {
			t.Fatal("remaining commands wrong")
		}
	})
	t.Run("not found", func(t *testing.T) {
		removed, n := s.RemoveByPrefix("zzz")
		if n != 0 || removed != nil {
			t.Fatalf("got removed=%v n=%d", removed, n)
		}
	})
	t.Run("ambiguous", func(t *testing.T) {
		s3 := &Store{Commands: []Command{
			mkCmd("abc123", "git", "A", "a"),
			mkCmd("abc456", "git", "B", "b"),
		}}
		removed, n := s3.RemoveByPrefix("abc")
		if n != 2 || removed != nil {
			t.Fatalf("got removed=%v n=%d", removed, n)
		}
		if len(s3.Commands) != 2 {
			t.Fatal("should not modify on ambiguous")
		}
	})
}

func TestReplace(t *testing.T) {
	s := &Store{Commands: []Command{
		mkCmd("abc", "git", "Undo", "git reset"),
		mkCmd("def", "docker", "Ps", "docker ps"),
	}}

	updated := mkCmd("abc", "git", "Redo", "git reflog")
	if !s.Replace(&updated) {
		t.Fatal("Replace should return true")
	}
	if s.Commands[0].Title != "Redo" || s.Commands[0].Command != "git reflog" {
		t.Fatalf("command not updated: %+v", s.Commands[0])
	}

	if s.Replace(&Command{ID: "zzz"}) {
		t.Fatal("Replace non-existent should return false")
	}
}

func TestBumpUsage(t *testing.T) {
	now := time.Now()
	s := &Store{Commands: []Command{
		{ID: "abc", UsageCount: 5},
		{ID: "def", UsageCount: 0},
	}}

	s.BumpUsage("abc")
	if s.Commands[0].UsageCount != 6 {
		t.Fatalf("usage = %d, want 6", s.Commands[0].UsageCount)
	}
	if s.Commands[0].LastUsedAt.Before(now) {
		t.Fatal("LastUsedAt should be updated")
	}

	s.BumpUsage("zzz")
	// no-op, should not panic
}

func TestTogglePin(t *testing.T) {
	s := &Store{Commands: []Command{
		{ID: "abc", Pinned: false},
		{ID: "def", Pinned: true},
	}}

	if got := s.TogglePin("abc"); !got {
		t.Fatal("expected pinned=true")
	}
	if !s.Commands[0].Pinned {
		t.Fatal("command should be pinned")
	}

	if got := s.TogglePin("def"); got {
		t.Fatal("expected pinned=false")
	}
	if s.Commands[1].Pinned {
		t.Fatal("command should be unpinned")
	}

	if got := s.TogglePin("zzz"); got {
		t.Fatal("non-existent should return false")
	}
}

func TestFilterByContext(t *testing.T) {
	s := &Store{Commands: []Command{
		mkCmd("1", "git", "A", "a"),
		mkCmd("2", "Git", "B", "b"),
		mkCmd("3", "docker", "C", "c"),
		mkCmd("4", "DOCKER", "D", "d"),
	}}

	t.Run("empty returns all", func(t *testing.T) {
		if got := s.FilterByContext(""); len(got) != 4 {
			t.Fatalf("got %d, want 4", len(got))
		}
	})
	t.Run("case insensitive", func(t *testing.T) {
		got := s.FilterByContext("git")
		if len(got) != 2 {
			t.Fatalf("got %d, want 2", len(got))
		}
	})
	t.Run("no match", func(t *testing.T) {
		if got := s.FilterByContext("nonexistent"); len(got) != 0 {
			t.Fatalf("got %d, want 0", len(got))
		}
	})
}

func TestMerge(t *testing.T) {
	existing := []Command{
		mkCmd("1", "git", "Undo", "git reset"),
		mkCmd("2", "docker", "ps", "docker ps"),
	}

	t.Run("import new", func(t *testing.T) {
		s := &Store{Commands: append([]Command(nil), existing...)}
		incoming := []Command{
			mkCmd("3", "git", "Stash", "git stash"),     // new
			mkCmd("4", "git", "Undo", "git reset"),       // dup by context+command
			mkCmd("5", "Docker", "ps", "docker ps"),      // dup by lowercase match
		}
		imported, skipped := s.Merge(incoming)
		if imported != 1 || skipped != 2 {
			t.Fatalf("imported=%d skipped=%d, want 1/2", imported, skipped)
		}
		if len(s.Commands) != 3 {
			t.Fatalf("total = %d, want 3", len(s.Commands))
		}
	})

	t.Run("all duplicate", func(t *testing.T) {
		s := &Store{Commands: append([]Command(nil), existing...)}
		imported, skipped := s.Merge(existing)
		if imported != 0 || skipped != 2 {
			t.Fatalf("imported=%d skipped=%d", imported, skipped)
		}
	})

	t.Run("empty incoming", func(t *testing.T) {
		s := &Store{Commands: append([]Command(nil), existing...)}
		imported, skipped := s.Merge(nil)
		if imported != 0 || skipped != 0 {
			t.Fatalf("imported=%d skipped=%d", imported, skipped)
		}
	})
}

func TestStoreKey(t *testing.T) {
	k1 := storeKey(Command{Context: "Git", Command: "git reset"})
	k2 := storeKey(Command{Context: "git", Command: "GIT reset"})
	if k1 != k2 {
		t.Fatalf("storeKey(%q) != storeKey(%q)", k1, k2)
	}
	if !strings.Contains(k1, "\x00") {
		t.Fatal("storeKey should use null separator")
	}
}
