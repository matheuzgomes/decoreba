package cli

import (
	"testing"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestContextCount(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{Context: "git"},
		{Context: "docker"},
		{Context: "git"},
	}}
	if got := contextCount(s); got != 2 {
		t.Fatalf("got %d, want 2", got)
	}
}

func TestContextCountEmpty(t *testing.T) {
	if got := contextCount(&core.Store{}); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestTopUsed(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{Title: "Stash", UsageCount: 5},
		{Title: "Reset", UsageCount: 10},
		{Title: "Log", UsageCount: 3},
	}}
	got := topUsed(s)
	if got.title != "Reset" || got.count != 10 {
		t.Fatalf("got %+v", got)
	}
}

func TestTopUsedNoCommands(t *testing.T) {
	got := topUsed(&core.Store{})
	if got.title != "—" || got.count != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestTopUsedAllZero(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{Title: "A", UsageCount: 0},
		{Title: "B", UsageCount: 0},
	}}
	got := topUsed(s)
	if got.title != "—" {
		t.Fatalf("got %q", got.title)
	}
}

func TestLastUsed(t *testing.T) {
	now := time.Now()
	s := &core.Store{Commands: []core.Command{
		{Title: "Old", LastUsedAt: now.Add(-2 * time.Hour)},
		{Title: "Recent", LastUsedAt: now.Add(-time.Hour)},
	}}
	got := lastUsed(s)
	if !contains(got, "Recent") {
		t.Fatalf("got %q", got)
	}
}

func TestLastUsedNone(t *testing.T) {
	got := lastUsed(&core.Store{Commands: []core.Command{
		{Title: "Never", LastUsedAt: time.Time{}},
	}})
	if got != "—" {
		t.Fatalf("got %q", got)
	}
}

func TestNewestCmd(t *testing.T) {
	now := time.Now()
	s := &core.Store{Commands: []core.Command{
		{Title: "Old", CreatedAt: now.Add(-2 * time.Hour)},
		{Title: "New", CreatedAt: now.Add(-time.Hour)},
	}}
	got := newestCmd(s)
	if !contains(got, "New") {
		t.Fatalf("got %q", got)
	}
}

func TestNewestCmdNoCommands(t *testing.T) {
	if got := newestCmd(&core.Store{}); got != "—" {
		t.Fatalf("got %q", got)
	}
}

func TestLastChanged(t *testing.T) {
	now := time.Now()
	s := &core.Store{Commands: []core.Command{
		{Title: "NoEdit", CreatedAt: now, UpdatedAt: now},
		{Title: "Edited", CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-time.Hour)},
	}}
	got := lastChanged(s)
	if !contains(got, "Edited") {
		t.Fatalf("got %q", got)
	}
}

func TestLastChangedNoEdits(t *testing.T) {
	now := time.Now()
	s := &core.Store{Commands: []core.Command{
		{Title: "A", CreatedAt: now, UpdatedAt: now},
	}}
	if got := lastChanged(s); got != "—" {
		t.Fatalf("got %q", got)
	}
}

func TestCountTagged(t *testing.T) {
	s := &core.Store{Commands: []core.Command{
		{ID: "1", Tags: []string{"git"}},
		{ID: "2"},
		{ID: "3", Tags: []string{"docker", "cleanup"}},
	}}
	if got := countTagged(s); got != 2 {
		t.Fatalf("got %d, want 2", got)
	}
}

func TestTimeAgo(t *testing.T) {
	t.Run("just now", func(t *testing.T) {
		if got := timeAgo(time.Now()); got != "just now" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("future clamped to just now", func(t *testing.T) {
		if got := timeAgo(time.Now().Add(time.Hour)); got != "just now" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("1 min", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-1 * time.Minute))
		if got != "1 min ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("5 min", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-5 * time.Minute))
		if got != "5 min ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("1 hour", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-1 * time.Hour))
		if got != "1 hour ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("3 hours", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-3 * time.Hour))
		if got != "3 hours ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("yesterday", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-26 * time.Hour))
		if got != "yesterday" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("today via sameDay", func(t *testing.T) {
		now := time.Now()
		// Same day, just a few hours ago → should be "X hours ago" (<24h)
		// or "today" if duration >= 24h but same day.
		got := timeAgo(now.Add(-2 * time.Hour))
		if got != "2 hours ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("3 days ago", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-3 * 24 * time.Hour))
		if got != "3 days ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("last week", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-10 * 24 * time.Hour))
		if got != "last week" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("2 weeks ago", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-18 * 24 * time.Hour))
		if got != "2 weeks ago" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("old date", func(t *testing.T) {
		got := timeAgo(time.Now().Add(-60 * 24 * time.Hour))
		if got == "" {
			t.Fatal("got empty")
		}
		// Should be a date like "May 20"
	})
}

func TestSameDay(t *testing.T) {
	now := time.Now()
	if !sameDay(now, now) {
		t.Fatal("same day should be true")
	}
	if sameDay(now, now.AddDate(0, 0, 1)) {
		t.Fatal("different days should be false")
	}
}

func TestCmdStats(t *testing.T) {
	now := time.Now()
	s := &core.Store{Commands: []core.Command{
		{
			ID: "1", Context: "git", Title: "Stash", Command: "git stash",
			UsageCount: 5, CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-time.Hour), Tags: []string{"save"},
		},
		{
			ID: "2", Context: "docker", Title: "Ps", Command: "docker ps",
			UsageCount: 10, CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-30 * time.Minute),
		},
	}}

	got := captureStdout(func() {
		cmdStats(s)
	})
	if !contains(got, "total") || !contains(got, "2 commands / 2 contexts") {
		t.Fatalf("missing total, got: %s", got)
	}
	if !contains(got, "top used") || !contains(got, "Ps") {
		t.Fatalf("missing top used, got: %s", got)
	}
	if !contains(got, "tagged") || !contains(got, "1 commands") {
		t.Fatalf("missing tagged, got: %s", got)
	}
}

func TestCmdStatsEmpty(t *testing.T) {
	got := captureStdout(func() {
		cmdStats(&core.Store{})
	})
	if !contains(got, "No commands yet") {
		t.Fatalf("got: %s", got)
	}
}
