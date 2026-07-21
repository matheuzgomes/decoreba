package tui

import (
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/search"
)

func TestHighlight(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		got := highlight("docker ps", nil, 80)
		if !strings.HasPrefix(got, ansiSubtle) {
			t.Fatalf("should start with subtle: %q", got)
		}
		if !strings.Contains(got, "docker ps") {
			t.Fatalf("should contain command: %q", got)
		}
	})

	t.Run("with positions", func(t *testing.T) {
		got := highlight("docker container prune", []int{17, 18, 19, 20, 21}, 80)
		if !strings.Contains(got, ansiAccent+"prune"+ansiReset) {
			t.Fatalf("should highlight matched segment: %q", got)
		}
	})

	t.Run("scattered positions", func(t *testing.T) {
		got := highlight("git stash pop", []int{0, 4, 10}, 80)
		if !strings.Contains(got, ansiAccent+"g"+ansiReset) {
			t.Fatalf("should highlight first char: %q", got)
		}
	})

	t.Run("truncates to maxVisible", func(t *testing.T) {
		got := highlight("docker container prune", []int{0}, 5)
		w := visibleWidth(got)
		if w > 5 {
			t.Fatalf("width = %d, want ≤ 5: %q", w, got)
		}
		if w != 5 {
			t.Fatalf("width = %d, want 5: %q", w, got)
		}
	})

	t.Run("empty positions", func(t *testing.T) {
		got := highlight("test", []int{}, 80)
		if !strings.Contains(got, "test") {
			t.Fatalf("should contain text: %q", got)
		}
	})
}

func TestPaletteRenderFrameSuggestions(t *testing.T) {
	noColor = true
	defer func() { noColor = false }()

	p := &palette{
		store: &core.Store{Commands: []core.Command{
			{ID: "1", Context: "git", Title: "Undo", Command: "git reset"},
		}},
		chip:        "git",
		query:       []rune("zzz"),
		suggestions: []string{"Undo"},
	}
	p.width = 80
	p.height = 24

	frame := string(p.renderFrame())
	if !strings.Contains(frame, "Did you mean") {
		t.Fatalf("suggestions missing: %q", frame)
	}
	if !strings.Contains(frame, "Undo") {
		t.Fatalf("suggestion text missing: %q", frame)
	}
}

func TestPaletteRenderFrameScrolling(t *testing.T) {
	cmds := make([]core.Command, 15)
	for i := range cmds {
		cmds[i] = core.Command{ID: "x", Context: "git", Title: "Item", Command: "echo hi"}
	}
	results := make([]search.Scored, 15)
	for i := range results {
		results[i] = search.Scored{Cmd: cmds[i], Score: 10}
	}

	p := &palette{
		store:        &core.Store{},
		results:      results,
		scrollOffset: 0,
		sel:          0,
	}
	p.height = 24
	p.width = 80

	// visibleCount = min(15, 9, (24-2)/2=11) = 9 → remaining = 6
	frame := string(p.renderFrame())
	if !strings.Contains(frame, "6 more") {
		t.Fatalf("should show remaining count in: %q", frame)
	}
}

func TestPaletteRendererFrameNoResults(t *testing.T) {
	noColor = true
	defer func() { noColor = false }()

	p := &palette{
		store: &core.Store{},
	}
	p.width = 80
	p.height = 24

	frame := string(p.renderFrame())
	if !strings.Contains(frame, "no results") {
		t.Fatalf("should show no results: %q", frame)
	}
}
