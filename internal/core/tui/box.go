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
var noColor bool

func SetNoColor(v bool) {
	noColor = v
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

func runeWidth(r rune) int {
	if r < 0x20 || r == 0x7f {
		return 0
	}
	switch {
	case r >= 0x0300 && r <= 0x036f, // Combining Diacritical Marks
		r >= 0x0483 && r <= 0x0489, // Cyrillic two dots, hundred thousands
		r >= 0x0591 && r <= 0x05bd, // Hebrew combining marks
		r >= 0x05bf && r <= 0x05bf,
		r >= 0x05c1 && r <= 0x05c2,
		r >= 0x05c4 && r <= 0x05c5,
		r >= 0x05c7 && r <= 0x05c7,
		r >= 0x0610 && r <= 0x061a, // Arabic combining marks
		r >= 0x064b && r <= 0x065f,
		r >= 0x0670 && r <= 0x0670,
		r >= 0x06d6 && r <= 0x06dc,
		r >= 0x06df && r <= 0x06e4,
		r >= 0x06e7 && r <= 0x06e8,
		r >= 0x06ea && r <= 0x06ed,
		r >= 0x0711 && r <= 0x0711,
		r >= 0x0730 && r <= 0x074a,
		r >= 0x07a6 && r <= 0x07b0,
		r >= 0x07eb && r <= 0x07f3,
		r >= 0x0816 && r <= 0x0819,
		r >= 0x081b && r <= 0x0823,
		r >= 0x0825 && r <= 0x0827,
		r >= 0x0829 && r <= 0x082d,
		r >= 0x0900 && r <= 0x0902,
		r >= 0x093a && r <= 0x093a,
		r >= 0x093c && r <= 0x093c,
		r >= 0x0941 && r <= 0x0948,
		r >= 0x094d && r <= 0x094d,
		r >= 0x0951 && r <= 0x0957,
		r >= 0x0962 && r <= 0x0963,
		r >= 0x0981 && r <= 0x0981,
		r >= 0x09bc && r <= 0x09bc,
		r >= 0x09c1 && r <= 0x09c4,
		r >= 0x09cd && r <= 0x09cd,
		r >= 0x09e2 && r <= 0x09e3,
		r >= 0x0a01 && r <= 0x0a02,
		r >= 0x0a3c && r <= 0x0a3c,
		r >= 0x0a41 && r <= 0x0a42,
		r >= 0x0a47 && r <= 0x0a48,
		r >= 0x0a4b && r <= 0x0a4d,
		r >= 0x0a70 && r <= 0x0a71,
		r >= 0x0a81 && r <= 0x0a82,
		r >= 0x0abc && r <= 0x0abc,
		r >= 0x0ac1 && r <= 0x0ac5,
		r >= 0x0ac7 && r <= 0x0ac8,
		r >= 0x0acd && r <= 0x0acd,
		r >= 0x0ae2 && r <= 0x0ae3,
		r >= 0x0b01 && r <= 0x0b01,
		r >= 0x0b3c && r <= 0x0b3c,
		r >= 0x0b3f && r <= 0x0b3f,
		r >= 0x0b41 && r <= 0x0b43,
		r >= 0x0b4d && r <= 0x0b4d,
		r >= 0x0b56 && r <= 0x0b56,
		r >= 0x0b82 && r <= 0x0b82,
		r >= 0x0bc0 && r <= 0x0bc0,
		r >= 0x0bcd && r <= 0x0bcd,
		r >= 0x0c3e && r <= 0x0c40,
		r >= 0x0c46 && r <= 0x0c48,
		r >= 0x0c4a && r <= 0x0c4d,
		r >= 0x0c55 && r <= 0x0c56,
		r >= 0x0cbc && r <= 0x0cbc,
		r >= 0x0cbf && r <= 0x0cbf,
		r >= 0x0cc6 && r <= 0x0cc6,
		r >= 0x0ccc && r <= 0x0ccd,
		r >= 0x0ce2 && r <= 0x0ce3,
		r >= 0x0d41 && r <= 0x0d43,
		r >= 0x0d4d && r <= 0x0d4d,
		r >= 0x0dca && r <= 0x0dca,
		r >= 0x0dd2 && r <= 0x0dd4,
		r >= 0x0dd6 && r <= 0x0dd6,
		r >= 0x0e31 && r <= 0x0e31,
		r >= 0x0e34 && r <= 0x0e3a,
		r >= 0x0e47 && r <= 0x0e4e,
		r >= 0x0eb1 && r <= 0x0eb1,
		r >= 0x0eb4 && r <= 0x0eb9,
		r >= 0x0ebb && r <= 0x0ebc,
		r >= 0x0ec8 && r <= 0x0ecd,
		r >= 0x0f18 && r <= 0x0f19,
		r >= 0x0f35 && r <= 0x0f35,
		r >= 0x0f37 && r <= 0x0f37,
		r >= 0x0f39 && r <= 0x0f39,
		r >= 0x0f71 && r <= 0x0f7e,
		r >= 0x0f80 && r <= 0x0f84,
		r >= 0x0f86 && r <= 0x0f87,
		r >= 0x0f90 && r <= 0x0f97,
		r >= 0x0f99 && r <= 0x0fbc,
		r >= 0x0fc6 && r <= 0x0fc6,
		r >= 0x102d && r <= 0x1030,
		r >= 0x1032 && r <= 0x1032,
		r >= 0x1036 && r <= 0x1037,
		r >= 0x1039 && r <= 0x1039,
		r >= 0x1058 && r <= 0x1059,
		r >= 0x1160 && r <= 0x11ff, // Hangul Jungseong/Jongseong (combining)
		r == 0x200b || r == 0x200c || r == 0x200d, // ZWNJ, ZWJ
		r >= 0x200e && r <= 0x200f, // LRM, RLM
		r >= 0x2028 && r <= 0x2029,
		r >= 0x202a && r <= 0x202e, // Bidi controls
		r >= 0x2060 && r <= 0x2064,
		r >= 0x2066 && r <= 0x206f,
		r >= 0x20d0 && r <= 0x20f0, // Combining Diacritical Symbols
		r >= 0x2cef && r <= 0x2cf1,
		r >= 0x2d7f && r <= 0x2d7f,
		r >= 0x2de0 && r <= 0x2dff,
		r >= 0xa66f && r <= 0xa672,
		r >= 0xa67c && r <= 0xa67d,
		r >= 0xa802 && r <= 0xa802,
		r >= 0xa806 && r <= 0xa806,
		r >= 0xa80b && r <= 0xa80b,
		r >= 0xa825 && r <= 0xa826,
		r >= 0xa8c4 && r <= 0xa8c4,
		r >= 0xa8e0 && r <= 0xa8f1,
		r >= 0xa926 && r <= 0xa92d,
		r >= 0xa947 && r <= 0xa951,
		r >= 0xa980 && r <= 0xa982,
		r >= 0xa9b3 && r <= 0xa9b3,
		r >= 0xa9b6 && r <= 0xa9b9,
		r >= 0xa9bc && r <= 0xa9bc,
		r >= 0xaa29 && r <= 0xaa2e,
		r >= 0xaa31 && r <= 0xaa32,
		r >= 0xaa35 && r <= 0xaa36,
		r >= 0xaa43 && r <= 0xaa43,
		r >= 0xaa4c && r <= 0xaa4c,
		r >= 0xaab0 && r <= 0xaab0,
		r >= 0xaab2 && r <= 0xaab4,
		r >= 0xaab7 && r <= 0xaab8,
		r >= 0xaabe && r <= 0xaabf,
		r >= 0xaac1 && r <= 0xaac1,
		r >= 0xabe5 && r <= 0xabe5,
		r >= 0xabe8 && r <= 0xabe8,
		r >= 0xabed && r <= 0xabed,
		r >= 0xfb1e && r <= 0xfb1e,
		r >= 0xfe00 && r <= 0xfe0f, // Variation Selectors
		r >= 0xfe20 && r <= 0xfe23: // Combining Half Marks
		return 0
	case r >= 0x1100 && r <= 0x115f,
		r >= 0x231a && r <= 0x231b,
		r >= 0x2329 && r <= 0x232a,
		r >= 0x23e9 && r <= 0x23f3,
		r >= 0x25fd && r <= 0x25fe,
		r >= 0x2614 && r <= 0x2615,
		r >= 0x2648 && r <= 0x2653,
		r >= 0x267f && r <= 0x267f,
		r >= 0x2693 && r <= 0x2693,
		r >= 0x26a1 && r <= 0x26a1,
		r >= 0x26aa && r <= 0x26ab,
		r >= 0x26bd && r <= 0x26be,
		r >= 0x26c4 && r <= 0x26c5,
		r >= 0x26ce && r <= 0x26ce,
		r >= 0x26d4 && r <= 0x26d4,
		r >= 0x26ea && r <= 0x26ea,
		r >= 0x26f2 && r <= 0x26f5,
		r >= 0x26fa && r <= 0x26fa,
		r >= 0x26fd && r <= 0x26fd,
		r >= 0x2702 && r <= 0x270d,
		r >= 0x270f && r <= 0x270f,
		r >= 0x2728 && r <= 0x2728,
		r >= 0x2744 && r <= 0x2744,
		r >= 0x274c && r <= 0x274e,
		r >= 0x2753 && r <= 0x2757,
		r >= 0x2763 && r <= 0x2764,
		r >= 0x2795 && r <= 0x2797,
		r >= 0x27a1 && r <= 0x27a1,
		r >= 0x27b0 && r <= 0x27bf,
		r >= 0x2b05 && r <= 0x2b07,
		r >= 0x2b1b && r <= 0x2b1c,
		r >= 0x2b50 && r <= 0x2b50,
		r >= 0x2b55 && r <= 0x2b55,
		r >= 0x2e80 && r <= 0x303e,
		r >= 0x3041 && r <= 0x3096,
		r >= 0x3099 && r <= 0x30ff,
		r >= 0x3105 && r <= 0x312f,
		r >= 0x3131 && r <= 0x318e,
		r >= 0x3190 && r <= 0x31e3,
		r >= 0x3200 && r <= 0x33ff,
		r >= 0x3400 && r <= 0x4dbf,
		r >= 0x4e00 && r <= 0x9fff,
		r >= 0xa000 && r <= 0xa4cf,
		r >= 0xac00 && r <= 0xd7af,
		r >= 0xf900 && r <= 0xfaff,
		r >= 0xfe10 && r <= 0xfe19,
		r >= 0xfe30 && r <= 0xfe6f,
		r >= 0xff01 && r <= 0xff60,
		r >= 0xffe0 && r <= 0xffe6,
		r >= 0x1b000 && r <= 0x1b0ff,
		r >= 0x1f000 && r <= 0x1f02f,
		r >= 0x1f030 && r <= 0x1f09f,
		r >= 0x1f0a0 && r <= 0x1f0ff,
		r >= 0x1f100 && r <= 0x1f64f,
		r >= 0x1f680 && r <= 0x1f6ff,
		r >= 0x1f900 && r <= 0x1f9ff,
		r >= 0x20000 && r <= 0x2fffd,
		r >= 0x30000 && r <= 0x3fffd:
		return 2
	}
	return 1
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
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		n += runeWidth(r)
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
		r, size := utf8.DecodeRuneInString(s[i:])
		w := runeWidth(r)
		if n+w > maxVisible {
			break
		}
		b.WriteString(s[i : i+size])
		i += size
		n += w
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

var (
	cachedWidth, cachedHeight int
	termSizeTick              int
)

// readTermSize fetches the terminal dimensions and applies safe defaults.
// Results are cached for 10 calls to avoid an ioctl syscall per keystroke.
func readTermSize() (width, height int) {
	termSizeTick++
	if termSizeTick%10 != 0 && cachedWidth > 0 {
		return cachedWidth, cachedHeight
	}
	width, height = term.GetSize()
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	cachedWidth, cachedHeight = width, height
	return
}

// --- context colors ---

func contextColor(name string) string {
	if noColor {
		return ""
	}
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

var contextColors = []int{39, 45, 51, 75, 81, 114, 150, 186, 215, 209, 203, 168, 141, 135, 79, 120}
var chipColors = []int{24, 30, 36, 60, 66, 95, 96, 102, 131, 138, 167, 173, 97, 133, 139, 174}
