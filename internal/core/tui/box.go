package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/matheuzgomes/decoreba/internal/core/term"
)

// --- ANSI and box drawing constants ---
var (
	ansiReset   = "\x1b[0m"
	ansiDim     = "\x1b[2m"
	ansiBold    = "\x1b[1m"
	ansiAccent  = "\x1b[36;1m"
	ansiSubtle  = "\x1b[90m"
	ansiWarn    = "\x1b[33;1m"
	ansiFocusBg = "\x1b[48;5;236m"
	ansiBorder  = "\x1b[38;5;240m"
)

// SetNoColor suppresses all ANSI escape codes when v is true.
// Must be called before any rendering.
func SetNoColor(v bool) {
	if v {
		ansiReset = ""
		ansiDim = ""
		ansiBold = ""
		ansiAccent = ""
		ansiSubtle = ""
		ansiWarn = ""
		ansiFocusBg = ""
		ansiBorder = ""
	}
}

const (
	boxTL      = "╭"
	boxTR      = "╮"
	boxBL      = "╰"
	boxBR      = "╯"
	boxH       = "─"
	boxV       = "│"
	boxLeftPad = 2
)

func boxTotalWidth(termWidth int) int {
	if termWidth <= 0 {
		termWidth = 80
	}
	if termWidth < 40 {
		return termWidth
	}
	if termWidth > 96 {
		return 96
	}
	return termWidth
}

func boxContentWidth(termWidth int) int {
	return boxTotalWidth(termWidth) - 4
}

func renderBoxTop(termWidth int) string {
	w := boxTotalWidth(termWidth)
	return "\x1b[2K\r" + ansiBorder + boxTL + strings.Repeat(boxH, w-2) + boxTR + ansiReset
}

func renderBoxBottom(termWidth int) string {
	w := boxTotalWidth(termWidth)
	return "\x1b[2K\r" + ansiBorder + boxBL + strings.Repeat(boxH, w-2) + boxBR + ansiReset
}

func renderBoxLine(termWidth int, content string, fillBg string) string {
	w := boxTotalWidth(termWidth)
	inner := w - 2
	budget := inner - 2
	if budget < 1 {
		budget = 1
	}
	shown := truncateVisible(content, budget)
	pad := budget - visibleWidth(shown)
	if pad < 0 {
		pad = 0
	}

	var b strings.Builder
	b.WriteString("\x1b[2K\r")
	b.WriteString(ansiBorder)
	b.WriteString(boxV)
	b.WriteString(ansiReset)
	if fillBg != "" {
		b.WriteString(fillBg)
	}
	b.WriteByte(' ')
	b.WriteString(shown)
	if pad > 0 {
		b.WriteString(strings.Repeat(" ", pad))
	}
	b.WriteByte(' ')
	if fillBg != "" {
		b.WriteString(ansiReset)
	}
	b.WriteString(ansiBorder)
	b.WriteString(boxV)
	b.WriteString(ansiReset)
	return b.String()
}

func visibleWidth(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		if s[i] < 0x20 {
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n++
	}
	return n
}

func truncateVisible(s string, maxVisible int) string {
	if maxVisible <= 0 {
		return ""
	}
	if visibleWidth(s) <= maxVisible {
		return s
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(s); {
		if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '[' {
			start := i
			i += 2
			for i < len(s) && (s[i] < 0x40 || s[i] > 0x7e) {
				i++
			}
			if i < len(s) {
				i++
			}
			b.WriteString(s[start:i])
			continue
		}
		if s[i] < 0x20 {
			i++
			continue
		}
		if n >= maxVisible {
			break
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		b.WriteString(s[i : i+size])
		i += size
		n++
	}
	b.WriteString(ansiReset)
	return b.String()
}

func truncate(s string, maxVisible int) string {
	if maxVisible <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxVisible {
		return s
	}
	return string(r[:maxVisible])
}

// readTermSize fetches the terminal dimensions and applies safe defaults.
func readTermSize() (width, height int) {
	width, height = term.GetSize()
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return
}

// --- context colors ---

func contextColor(name string) string {
	if name == "" {
		return ansiDim
	}
	h := 0
	for _, r := range name {
		h = h*31 + int(r)
	}
	if h < 0 {
		h = -h
	}
	return fmt.Sprintf("\x1b[38;5;%dm", contextColors[h%len(contextColors)])
}

var contextColors = []int{39, 45, 51, 75, 81, 114, 150, 186, 215, 209, 203, 168, 141, 135, 105, 69}
var chipColors = []int{24, 30, 36, 60, 66, 95, 96, 102, 131, 138, 167, 173, 97, 133, 139, 173}
