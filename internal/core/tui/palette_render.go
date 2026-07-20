package tui

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func (p *palette) renderFrame() []byte {
	var b bytes.Buffer
	cw := boxContentWidth(p.width)

	b.WriteString(renderBoxTop(p.width))

	var search strings.Builder
	prefix := 2
	if p.chip != "" {
		chipRunes := len([]rune(p.chip))
		search.WriteString(contextColor(p.chip) + p.chip + ansiReset + " ")
		prefix += chipRunes + 1
	}
	search.WriteString(ansiDim + "›" + ansiReset + " ")
	search.WriteString(truncate(string(p.query), cw-prefix))
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(p.width, search.String(), ""))

	if len(p.results) == 0 {
		b.WriteByte('\n')
		if len(p.suggestions) > 0 {
			hint := "Did you mean: " + strings.Join(p.suggestions, ", ")
			b.WriteString(renderBoxLine(p.width, ansiDim+truncate(hint, cw)+ansiReset, ""))
		} else {
			b.WriteString(renderBoxLine(p.width, ansiDim+truncate("no results", cw)+ansiReset, ""))
		}
	} else {
		vis := p.visibleCount()
		for i := 0; i < vis; i++ {
			r := p.results[p.scrollOffset+i]
			num := strconv.Itoa(p.scrollOffset + i + 1)
			star := ""
			if r.Cmd.Pinned {
				star = "★ "
			}
			if r.Cmd.IsWorkflow() {
				star = "▶ " + star
			}

			titleRaw := r.Cmd.Title
			if r.Cmd.IsWorkflow() {
				titleRaw += fmt.Sprintf(" (%d steps)", len(r.Cmd.Steps))
			}
			titleBudget := cw - len(num) - 1 - len([]rune(star))
			title := truncate(titleRaw, titleBudget)
			var row strings.Builder
			if r.Cmd.Pinned || r.Cmd.IsWorkflow() {
				if r.Cmd.IsWorkflow() {
					row.WriteString(ansiAccent + "▶" + ansiReset + " ")
				}
				if r.Cmd.Pinned {
					row.WriteString(ansiWarn + "★" + ansiReset + " ")
				}
			}
			if p.scrollOffset+i == p.sel {
				row.WriteString(ansiAccent + num + ansiReset + " " + ansiBold + title + ansiReset)
			} else {
				row.WriteString(ansiDim + num + ansiReset + " " + title)
			}
			fill := ""
			if p.scrollOffset+i == p.sel {
				fill = ansiFocusBg
			}
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(p.width, row.String(), fill))

			var cmdRow string
			if r.Cmd.IsWorkflow() {
				preview := make([]string, len(r.Cmd.Steps))
				for s, step := range r.Cmd.Steps {
					preview[s] = step.Title
				}
				cmdRow = "  " + ansiDim + truncate(strings.Join(preview, " → "), cw-2) + ansiReset
			} else {
				cmdRow = "  " + highlight(r.Cmd.Command, r.Pos, cw-2)
			}
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(p.width, cmdRow, fill))
		}

		remaining := len(p.results) - p.scrollOffset - vis
		if remaining > 0 {
			more := fmt.Sprintf("  … %d more", remaining)
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(p.width, ansiDim+more+ansiReset, ""))
		}
	}
	hint := paletteHint
	if p.confirmExec {
		hint = paletteExecHint
	}
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(p.width, ansiDim+truncate(hint, cw)+ansiReset, ""))
	b.WriteByte('\n')
	b.WriteString(renderBoxBottom(p.width))
	return b.Bytes()
}

func highlight(cmd string, pos []int, maxVisible int) string {
	matched := make(map[int]bool, len(pos))
	for _, i := range pos {
		matched[i] = true
	}
	var b strings.Builder
	b.WriteString(ansiSubtle)
	accent := false
	visible := 0
	for i, r := range []rune(cmd) {
		if visible >= maxVisible {
			break
		}
		if matched[i] && !accent {
			b.WriteString(ansiReset + ansiAccent)
			accent = true
		} else if !matched[i] && accent {
			b.WriteString(ansiReset + ansiSubtle)
			accent = false
		}
		b.WriteRune(r)
		visible += runeWidth(r)
	}
	b.WriteString(ansiReset)
	return b.String()
}
