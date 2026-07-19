package cli

import (
	"fmt"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func cmdStats(s *core.Store) {
	if len(s.Commands) == 0 {
		fmt.Println("No commands yet. Start with 'decoreba add'.")
		return
	}

	ctxCount := contextCount(s)
	top := topUsed(s)
	last := lastUsed(s)
	newest := newestCmd(s)
	changed := lastChanged(s)
	tagged := countTagged(s)

	fmt.Printf("      total  %d commands / %d contexts\n", len(s.Commands), ctxCount)
	fmt.Printf("  top used  %s (%d×)\n", top.title, top.count)
	fmt.Printf(" last used  %s\n", last)
	fmt.Printf("       new  %s\n", newest)
	fmt.Printf("   changed  %s\n", changed)
	fmt.Printf("    tagged  %d commands have tags\n", tagged)
}

func contextCount(s *core.Store) int {
	seen := make(map[string]bool, len(s.Commands)/2)
	for _, c := range s.Commands {
		seen[c.Context] = true
	}
	return len(seen)
}

type topCmd struct {
	title string
	count int
}

func topUsed(s *core.Store) topCmd {
	var best topCmd
	for _, c := range s.Commands {
		if c.UsageCount > best.count {
			best.title = c.Title
			best.count = c.UsageCount
		}
	}
	if best.title == "" {
		best.title = "—"
	}
	return best
}

func lastUsed(s *core.Store) string {
	var latest time.Time
	var title string
	for _, c := range s.Commands {
		if c.LastUsedAt.After(latest) {
			latest = c.LastUsedAt
			title = c.Title
		}
	}
	if title == "" || latest.IsZero() {
		return "—"
	}
	return fmt.Sprintf("%s (%s)", title, timeAgo(latest))
}

func newestCmd(s *core.Store) string {
	var latest time.Time
	var title string
	for _, c := range s.Commands {
		if c.CreatedAt.After(latest) {
			latest = c.CreatedAt
			title = c.Title
		}
	}
	if title == "" {
		return "—"
	}
	return fmt.Sprintf("%s (added %s)", title, timeAgo(latest))
}

func lastChanged(s *core.Store) string {
	var latest time.Time
	var title string
	for _, c := range s.Commands {
		if c.UpdatedAt.After(c.CreatedAt) && c.UpdatedAt.After(latest) {
			latest = c.UpdatedAt
			title = c.Title
		}
	}
	if title == "" {
		return "—"
	}
	return fmt.Sprintf("%s (edited %s)", title, timeAgo(latest))
}

func countTagged(s *core.Store) int {
	n := 0
	for _, c := range s.Commands {
		if len(c.Tags) > 0 {
			n++
		}
	}
	return n
}

// timeAgo returns a human-readable relative time string.
func timeAgo(t time.Time) string {
	d := time.Since(t)
	if d < 0 {
		d = 0
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	switch {
	case sameDay(t, now):
		return "today"
	case sameDay(t, yesterday):
		return "yesterday"
	}

	days := int(d.Hours() / 24)
	switch {
	case days < 7:
		return fmt.Sprintf("%d days ago", days)
	case days < 14:
		return "last week"
	case days < 30:
		return fmt.Sprintf("%d weeks ago", days/7)
	default:
		return t.Format("Jan 2")
	}
}

func sameDay(a, b time.Time) bool {
	ya, ma, da := a.Date()
	yb, mb, db := b.Date()
	return ya == yb && ma == mb && da == db
}
