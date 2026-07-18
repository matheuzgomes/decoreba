package tui

import (
	"bytes"
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
		b.WriteString(renderBoxLine(p.width, ansiDim+truncate("no results", cw)+ansiReset, ""))
	} else {
		for i := 0; i < p.visibleCount(); i++ {
			r := p.results[i]
			num := strconv.Itoa(i + 1)
			titleBudget := cw - len(num) - 1
			title := truncate(r.Cmd.Title, titleBudget)
			var row strings.Builder
			if i == p.sel {
				row.WriteString(ansiAccent + num + ansiReset + " " + ansiBold + title + ansiReset)
			} else {
				row.WriteString(ansiDim + num + ansiReset + " " + title)
			}
			fill := ""
			if i == p.sel {
				fill = ansiFocusBg
			}
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(p.width, row.String(), fill))

			cmdRow := "  " + highlight(r.Cmd.Command, r.Pos, cw-2)
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(p.width, cmdRow, fill))
		}
	}

	b.WriteByte('\n')
	b.WriteString(renderBoxLine(p.width, ansiDim+truncate(paletteHint, cw)+ansiReset, ""))
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
		visible++
	}
	b.WriteString(ansiReset)
	return b.String()
}
