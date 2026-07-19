package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func frameLines(t *testing.T, frame []byte) []string {
	t.Helper()
	lines := strings.Split(string(frame), "\n")
	for _, l := range lines {
		if !strings.HasPrefix(l, "\x1b[2K\r") {
			t.Fatalf("line does not start with clear sequence: %q", l)
		}
	}
	return lines
}

func newTestPalette() *palette {
	return &palette{
		store: &core.Store{Commands: []core.Command{
			{ID: "1", Context: "docker", Title: "Remove stopped containers", Command: "docker container prune"},
			{ID: "2", Context: "docker", Title: "Follow container logs", Command: "docker logs -f nome_container"},
			{ID: "3", Context: "git", Title: "Undo last commit", Command: "git reset --soft HEAD~1"},
		}},
		chip:   "docker",
		width:  80,
		height: 24,
	}
}

func TestRenderFrameLayout(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()

	if got := len(p.results); got != 2 {
		t.Fatalf("pool filtered by chip: got %d results, want 2", got)
	}
	if got := p.frameLines(); got != 8 {
		t.Fatalf("frameLines = %d, want 8", got)
	}

	lines := frameLines(t, p.renderFrame())
	if len(lines) != 8 {
		t.Fatalf("rendered %d lines, want 8", len(lines))
	}
	if !strings.Contains(lines[0], boxTL) || !strings.Contains(lines[0], boxTR) {
		t.Fatalf("top border: %q", lines[0])
	}
	if !strings.Contains(lines[1], "docker") || !strings.Contains(lines[1], "›") {
		t.Fatalf("search line missing chip/separator: %q", lines[1])
	}
	color := contextColor("docker")
	if !strings.Contains(lines[1], color+"docker"+ansiReset) {
		t.Fatalf("chip should use context color: %q", lines[1])
	}
	if !strings.Contains(lines[2], "Remove stopped containers") {
		t.Fatalf("item 1 description: %q", lines[2])
	}
	if !strings.Contains(lines[3], "docker container prune") {
		t.Fatalf("item 1 command: %q", lines[3])
	}
	if !strings.Contains(lines[3], ansiSubtle) {
		t.Fatalf("command line should use subtle color: %q", lines[3])
	}
	if !strings.Contains(lines[6], "↵ copy") {
		t.Fatalf("hint line: %q", lines[6])
	}
	if !strings.Contains(lines[7], boxBL) || !strings.Contains(lines[7], boxBR) {
		t.Fatalf("bottom border: %q", lines[7])
	}
	if !strings.Contains(lines[2], ansiBold) {
		t.Fatalf("selected description should be bold: %q", lines[2])
	}
	if !strings.Contains(lines[2], ansiFocusBg) {
		t.Fatalf("selected row should have focus bg: %q", lines[2])
	}
	if strings.Contains(lines[4], ansiBold) {
		t.Fatalf("unselected description should not be bold: %q", lines[4])
	}
}

func TestRenderFrameEmptyResults(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.query = []rune("zzzz")
	p.refilter()

	if got := p.frameLines(); got != 5 {
		t.Fatalf("frameLines = %d, want 5", got)
	}
	lines := frameLines(t, p.renderFrame())
	if !strings.Contains(lines[2], "no results") {
		t.Fatalf("empty state line: %q", lines[2])
	}
}

func TestRenderFrameTruncatesToWidth(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()
	p.width = 40
	want := boxTotalWidth(40)
	for _, l := range frameLines(t, p.renderFrame()) {
		if n := visibleWidth(l); n != want {
			t.Fatalf("line has %d visible runes, want %d: %q", n, want, l)
		}
	}
}

func TestRenderFrameHighlightsMatches(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.query = []rune("prune")
	p.refilter()

	if len(p.results) != 1 {
		t.Fatalf("got %d results, want 1", len(p.results))
	}
	lines := frameLines(t, p.renderFrame())
	cmd := lines[3]
	if !strings.Contains(cmd, ansiAccent) {
		t.Fatalf("matched runes should be accent-colored: %q", cmd)
	}
	needle := ansiAccent + "prune" + ansiReset
	if !strings.Contains(cmd, needle) {
		t.Fatalf("expected %q inside command line: %q", needle, cmd)
	}
}

func TestVisibleCountClamps(t *testing.T) {
	p := newTestPalette()
	p.store.Commands = nil
	for i := 0; i < 20; i++ {
		p.store.Commands = append(p.store.Commands, core.Command{
			Context: "docker", Title: "t", Command: "docker x",
		})
	}
	p.setPool()
	p.refilter()
	if got := p.visibleCount(); got != maxVisible {
		t.Fatalf("visibleCount = %d, want %d", got, maxVisible)
	}
	p.height = 8
	if got := p.visibleCount(); got != 3 {
		t.Fatalf("visibleCount with height 8 = %d, want 3", got)
	}
}

func TestInputCol(t *testing.T) {
	p := &palette{chip: "docker", query: []rune("pru")}
	if got := p.inputCol(); got != boxLeftPad+6+1+2+3 {
		t.Fatalf("inputCol = %d, want %d", got, boxLeftPad+6+1+2+3)
	}
	p.chip = ""
	if got := p.inputCol(); got != boxLeftPad+2+3 {
		t.Fatalf("inputCol without chip = %d, want %d", got, boxLeftPad+2+3)
	}
}

func TestApplyDigitRules(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()
	done, chosen := p.apply([]keyEvent{{kind: keyRune, r: '2'}})
	if !done || chosen == nil || chosen.ID != "2" {
		t.Fatalf("digit 2 should select second item: done=%v chosen=%v", done, chosen)
	}

	p = newTestPalette()
	p.setPool()
	p.refilter()
	done, _ = p.apply([]keyEvent{{kind: keyRune, r: '9'}})
	if done {
		t.Fatal("digit 9 with 2 results should not select")
	}
	if string(p.query) != "9" {
		t.Fatalf("out-of-range digit should be typed, query=%q", string(p.query))
	}

	p = newTestPalette()
	p.setPool()
	p.query = []rune("log")
	p.refilter()
	done, _ = p.apply([]keyEvent{{kind: keyRune, r: '1'}})
	if done {
		t.Fatal("digit with text in the field should not select")
	}
	if string(p.query) != "log1" {
		t.Fatalf("query = %q, want log1", string(p.query))
	}
}

func TestApplyBackspaceRemovesChip(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()

	done, _ := p.apply([]keyEvent{{kind: keyBackspace}})
	if done {
		t.Fatal("backspace should not close")
	}
	if p.chip != "" {
		t.Fatalf("chip = %q, want empty", p.chip)
	}
	if len(p.results) != 3 {
		t.Fatalf("pool should expand to all contexts, got %d results", len(p.results))
	}

	p.apply([]keyEvent{{kind: keyBackspace}})
	if len(p.results) != 3 {
		t.Fatal("second backspace should be a no-op")
	}
}

func TestApplyNavigationAndConfirm(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()

	p.apply([]keyEvent{{kind: keyDown}, {kind: keyDown}, {kind: keyDown}})
	if p.sel != 1 {
		t.Fatalf("sel = %d, want 1 (clamped)", p.sel)
	}
	p.apply([]keyEvent{{kind: keyUp}})
	if p.sel != 0 {
		t.Fatalf("sel = %d, want 0", p.sel)
	}

	done, chosen := p.apply([]keyEvent{{kind: keyEnter}})
	if !done || chosen == nil || chosen.ID != "1" {
		t.Fatalf("enter should confirm first item: done=%v chosen=%v", done, chosen)
	}
}

func TestApplyCancel(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()
	for _, ev := range []keyEvent{{kind: keyEsc}, {kind: keyCancel}} {
		done, chosen := p.apply([]keyEvent{ev})
		if !done || chosen != nil {
			t.Fatalf("cancel should close with nil: done=%v chosen=%v", done, chosen)
		}
	}
}

func TestApplyEnterWithNoResults(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.query = []rune("zzz")
	p.refilter()
	done, chosen := p.apply([]keyEvent{{kind: keyEnter}})
	if done || chosen != nil {
		t.Fatal("enter with no results should do nothing")
	}
}

func TestApplyRefiltersOnType(t *testing.T) {
	p := newTestPalette()
	p.setPool()
	p.refilter()
	p.apply([]keyEvent{{kind: keyRune, r: 'l'}, {kind: keyRune, r: 'o'}, {kind: keyRune, r: 'g'}})
	if len(p.results) != 1 || p.results[0].Cmd.ID != "2" {
		t.Fatalf("typing 'log' should leave only item 2: %+v", p.results)
	}
	p.sel = 5
	p.apply([]keyEvent{{kind: keyRune, r: 's'}})
	if p.sel >= len(p.results) && len(p.results) > 0 {
		t.Fatalf("sel = %d out of range after refilter", p.sel)
	}
}

func TestApplyScrollPagination(t *testing.T) {
	p := newTestPalette()
	// Add 20 items so we have more than maxVisible (9)
	for i := 0; i < 18; i++ {
		p.store.Commands = append(p.store.Commands, core.Command{
			ID:      "x",
			Context: "docker",
			Title:   "extra",
			Command: "docker ps",
		})
	}
	p.chip = ""
	p.setPool()
	p.refilter()

	if got := len(p.results); got != 21 {
		t.Fatalf("results = %d, want 21", got)
	}
	if got := p.visibleCount(); got != maxVisible {
		t.Fatalf("visibleCount = %d, want %d", got, maxVisible)
	}
	if p.scrollOffset != 0 {
		t.Fatalf("scrollOffset = %d, want 0", p.scrollOffset)
	}

	// Scroll down to the last visible item on the first page.
	for i := 0; i < maxVisible-1; i++ {
		p.apply([]keyEvent{{kind: keyDown}})
	}
	if p.sel != 8 {
		t.Fatalf("sel = %d, want 8 (last of page 0)", p.sel)
	}
	if p.scrollOffset != 0 {
		t.Fatalf("scrollOffset = %d, want 0 (still page 0)", p.scrollOffset)
	}

	// One more down should scroll to page 1.
	p.apply([]keyEvent{{kind: keyDown}})
	if p.sel != 9 {
		t.Fatalf("sel = %d, want 9", p.sel)
	}
	if p.scrollOffset != 1 {
		t.Fatalf("scrollOffset = %d, want 1", p.scrollOffset)
	}

	// Scroll back up all the way to return to page 0.
	for i := 0; i < 9; i++ {
		p.apply([]keyEvent{{kind: keyUp}})
	}
	if p.sel != 0 {
		t.Fatalf("sel = %d, want 0", p.sel)
	}
	if p.scrollOffset != 0 {
		t.Fatalf("scrollOffset = %d, want 0", p.scrollOffset)
	}

	// Scroll all the way down.
	for i := 0; i < 20; i++ {
		p.apply([]keyEvent{{kind: keyDown}})
	}
	if p.sel != 20 {
		t.Fatalf("sel = %d, want 20 (last item)", p.sel)
	}
	// scrollOffset = sel - visibleCount + 1 = 20 - 9 + 1 = 12
	if p.scrollOffset != 12 {
		t.Fatalf("scrollOffset = %d, want 12", p.scrollOffset)
	}

	// Refilter resets scrollOffset.
	p.query = []rune("zzz")
	p.refilter()
	if p.scrollOffset != 0 {
		t.Fatalf("scrollOffset after refilter = %d, want 0", p.scrollOffset)
	}
}

func TestRedrawAndCloseSequences(t *testing.T) {
	var out bytes.Buffer
	p := newTestPalette()
	p.setPool()
	p.refilter()
	p.out = &out
	p.width, p.height = readTermSize()

	p.redraw()
	frame1 := out.String()
	if !strings.HasPrefix(frame1, "\r\x1b[2K\r") {
		t.Fatalf("redraw must start at the overlay top (no move-up): %q", frame1)
	}
	if !strings.Contains(frame1, "\x1b[6A\r") {
		t.Fatalf("redraw must park the cursor back at the search line: %q", frame1)
	}
	if p.lines != 8 || p.parkedLine != 1 {
		t.Fatalf("p.lines=%d parked=%d, want 8/1", p.lines, p.parkedLine)
	}

	out.Reset()
	p.query = []rune("zzz")
	p.refilter()
	p.redraw()
	frame2 := out.String()
	if !strings.HasPrefix(frame2, "\x1b[1A\r") {
		t.Fatalf("redraw after shrink must climb from search line: %q", frame2)
	}
	if !strings.Contains(frame2, "\x1b[J") {
		t.Fatalf("redraw must clear leftover lines: %q", frame2)
	}
	if p.lines != 5 {
		t.Fatalf("p.lines = %d, want 5", p.lines)
	}

	out.Reset()
	p.close()
	got := out.String()
	if !strings.HasPrefix(got, "\x1b[1A") || !strings.HasSuffix(got, "\r\x1b[J") {
		t.Fatalf("close sequence = %q", got)
	}
	if p.lines != 0 {
		t.Fatalf("p.lines = %d after close, want 0", p.lines)
	}
}
