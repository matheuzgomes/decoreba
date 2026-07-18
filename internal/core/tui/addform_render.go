package tui

import (
	"bytes"
	"fmt"
	"strings"
)

func chipStyle(tag string) string {
	h := 0
	for _, r := range tag {
		h = h*31 + int(r)
	}
	if h < 0 {
		h = -h
	}
	bg := chipColors[h%len(chipColors)]
	return fmt.Sprintf("\x1b[48;5;%dm\x1b[38;5;255m", bg)
}

func renderChip(tag string) string {
	return chipStyle(tag) + " " + tag + " " + ansiReset
}

func renderTagsDisplay(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, ",")
	var b strings.Builder
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString(renderChip(trimmed))
	}
	return b.String()
}

func (f *addForm) fieldValueDisplay(idx int) string {
	val := string(f.fields[idx])
	if idx == fieldTags {
		return renderTagsDisplay(val)
	}
	return val
}

func (f *addForm) fieldPrefixDisplay(idx int, runeEnd int) string {
	runes := f.fields[idx]
	if runeEnd > len(runes) {
		runeEnd = len(runes)
	}
	prefix := string(runes[:runeEnd])
	if idx == fieldTags {
		return renderTagsDisplay(prefix)
	}
	return prefix
}

func (f *addForm) renderFieldContent(idx int) string {
	label := fieldLabels[idx]
	pad := labelPad - len([]rune(label))
	if pad < 1 {
		pad = 1
	}
	labelStr := label + strings.Repeat(" ", pad)

	focused := idx == f.focus
	errored := idx == f.errField && f.errFlash > 0

	var b strings.Builder
	switch {
	case errored:
		b.WriteString(ansiWarn + labelStr + ansiReset)
	case focused:
		b.WriteString(ansiBold + labelStr + ansiReset)
	default:
		b.WriteString(ansiDim + labelStr + ansiReset)
	}
	if focused {
		b.WriteString(ansiFocusBg)
	}

	budget := boxContentWidth(f.width) - labelPad
	if budget < 1 {
		budget = 1
	}

	if idx == fieldContext && focused {
		if sug := f.contextSuggestion(); sug != "" {
			shown := truncate(string(f.fields[idx]), budget)
			ghostBudget := budget - len([]rune(shown))
			ghost := truncate(sug, ghostBudget)
			b.WriteString(shown)
			if ghost != "" {
				b.WriteString(ansiDim + ghost + ansiReset)
				if focused {
					b.WriteString(ansiFocusBg)
				}
			}
			return b.String()
		}
	}

	val := f.fieldValueDisplay(idx)
	if idx == fieldTags {
		b.WriteString(truncateVisible(val, budget))
		if focused {
			b.WriteString(ansiFocusBg)
		}
	} else {
		b.WriteString(truncate(val, budget))
	}
	return b.String()
}

func (f *addForm) renderFrame() []byte {
	var b bytes.Buffer
	ctx := strings.TrimSpace(string(f.fields[fieldContext]))
	dotColor := contextColor(ctx)

	b.WriteString(renderBoxTop(f.width))

	headerText := newCmdHeader
	if f.editing {
		headerText = editCmdHeader
	}
	header := dotColor + "●" + ansiReset + " " + ansiDim + headerText + ansiReset
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(f.width, header, ""))

	for i := 0; i < fieldCount; i++ {
		fill := ""
		if i == f.focus {
			fill = ansiFocusBg
		}
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(f.width, f.renderFieldContent(i), fill))
	}

	hint := addFormHint
	hintColor := ansiDim
	hintFill := ""
	if f.errMsg != "" {
		hint = f.errMsg
		hintColor = ansiWarn
	}
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(f.width, hintColor+truncate(hint, boxContentWidth(f.width))+ansiReset, hintFill))

	b.WriteByte('\n')
	b.WriteString(renderBoxBottom(f.width))

	if f.errFlash > 0 {
		f.errFlash--
	}
	return b.Bytes()
}
