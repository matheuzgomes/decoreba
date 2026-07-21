package tui

import (
	"strings"
	"testing"
)

func TestBoxTotalWidth(t *testing.T) {
	tests := []struct {
		term int
		want int
	}{
		{0, 80},
		{-1, 80},
		{39, 39},
		{40, 40},
		{80, 80},
		{96, 96},
		{97, 96},
		{200, 96},
	}
	for _, tt := range tests {
		got := boxTotalWidth(tt.term)
		if got != tt.want {
			t.Errorf("boxTotalWidth(%d) = %d, want %d", tt.term, got, tt.want)
		}
	}
}

func TestBoxContentWidth(t *testing.T) {
	if got := boxContentWidth(80); got != 76 {
		t.Fatalf("boxContentWidth(80) = %d, want 76", got)
	}
}

func TestBoxTopBottom(t *testing.T) {
	top := renderBoxTop(40)
	if !strings.HasPrefix(top, "\x1b[2K\r") {
		t.Fatalf("top should start with clear: %q", top)
	}
	if !strings.Contains(top, boxTL) || !strings.Contains(top, boxTR) {
		t.Fatalf("top missing corners: %q", top)
	}

	bot := renderBoxBottom(40)
	if !strings.Contains(bot, boxBL) || !strings.Contains(bot, boxBR) {
		t.Fatalf("bottom missing corners: %q", bot)
	}
}

func TestRenderBoxLine(t *testing.T) {
	ansiFocusBg = "\x1b[48;5;236m"
	ansiReset = "\x1b[0m"
	ansiBorder = "\x1b[38;5;240m"

	line := renderBoxLine(40, "hello", "")
	if !strings.Contains(line, "hello") {
		t.Fatalf("content missing: %q", line)
	}
	if !strings.Contains(line, boxV) {
		t.Fatalf("vertical bar missing: %q", line)
	}

	lineWithBg := renderBoxLine(40, "hi", ansiFocusBg)
	if !strings.Contains(lineWithBg, ansiFocusBg) {
		t.Fatalf("fill bg missing: %q", lineWithBg)
	}
}

func TestRenderBoxLineTruncation(t *testing.T) {
	long := strings.Repeat("x", 100)
	line := renderBoxLine(40, long, "")
	inner := visibleWidth(line)
	if inner > 40 {
		t.Fatalf("line width = %d, should be ≤ 40: %q", inner, line)
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		r    rune
		want int
	}{
		{'a', 1},
		{'ç', 1},
		{'\x01', 0},
		{'\x7f', 0},
		{'\u0300', 0}, // combining accent
		{'\u200b', 0}, // zero-width space
		{'\u4e00', 2}, // CJK
		{'\uFF01', 2}, // fullwidth
	}
	for _, tt := range tests {
		got := runeWidth(tt.r)
		if got != tt.want {
			t.Errorf("runeWidth(%U %q) = %d, want %d", tt.r, tt.r, got, tt.want)
		}
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"hello", 5},
		{"", 0},
		{"héllo", 5},
		{"\x1b[31mred\x1b[0m", 3},
		{"a\x00b", 2}, // null byte skipped, a+b counted
		{"日本語", 6}, // 3 CJK chars × 2
		{"a\u0300", 1}, // a + combining = 1
	}
	for _, tt := range tests {
		got := visibleWidth(tt.s)
		if got != tt.want {
			t.Errorf("visibleWidth(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestTruncateVisible(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 5, "hello"},
		{"hello", 3, "hel" + ansiReset},
		{"hello", 0, ""},
		{"hé", 1, "h" + ansiReset},
		{"\x1b[31mred\x1b[0m", 2, "\x1b[31mre\x1b[0m"},
		{"日本語", 4, "日本" + ansiReset},
	}
	for _, tt := range tests {
		got := truncateVisible(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncateVisible(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"hello", 5, "hello"},
		{"hello", 3, "hel"},
		{"hello", 0, ""},
		{"日本語", 2, "日本"},
	}
	for _, tt := range tests {
		got := truncate(tt.s, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
		}
	}
}

func TestContextColor(t *testing.T) {
	noColor = false
	ansiDim = "\x1b[2m"
	c := contextColor("")
	if c != ansiDim {
		t.Fatalf("empty context color = %q, want dim", c)
	}

	c = contextColor("git")
	if c == "" || c == ansiDim {
		t.Fatalf("git context color = %q", c)
	}

	noColor = true
	c = contextColor("git")
	if c != "" {
		t.Fatalf("with noColor, contextColor should be empty: %q", c)
	}
	noColor = false
}

func TestSetNoColor(t *testing.T) {
	// Save original values
	origReset := ansiReset
	origDim := ansiDim
	origBold := ansiBold

	SetNoColor(true)
	if ansiReset != "" {
		t.Fatal("ansiReset should be empty when noColor")
	}
	if ansiDim != "" {
		t.Fatal("ansiDim should be empty when noColor")
	}

	// Restore for other tests
	ansiReset = origReset
	ansiDim = origDim
	ansiBold = origBold
	noColor = false
}
