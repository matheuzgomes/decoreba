package tui

import (
	"bytes"
	"strconv"
	"strings"
)

const listHint = "↵ select  ↑↓ nav  1-9 direct  esc cancel"

func (b *listBrowser) renderFrame() []byte {
	var buf bytes.Buffer
	cw := boxContentWidth(b.width)

	buf.WriteString(renderBoxTop(b.width))
	buf.WriteByte('\n')
	buf.WriteString(renderBoxLine(b.width, "  list contexts", ""))
	buf.WriteByte('\n')

	visible := b.visibleCount()
	for i := 0; i < visible; i++ {
		entry := b.entries[b.scroll+i]
		num := strconv.Itoa(b.scroll + i + 1)

		var row strings.Builder
		if b.scroll+i == b.sel {
			row.WriteString(ansiAccent + num + ansiReset + " ")
			row.WriteString(contextColor(entry.Name) + entry.Name + ansiReset)
			row.WriteString("  " + ansiDim + strconv.Itoa(entry.Count) + ansiReset)
		} else {
			row.WriteString(ansiDim + num + ansiReset + " ")
			row.WriteString(contextColor(entry.Name) + entry.Name + ansiReset)
			row.WriteString("  " + ansiDim + strconv.Itoa(entry.Count) + ansiReset)
		}

		fill := ""
		if b.scroll+i == b.sel {
			fill = ansiFocusBg
		}

		buf.WriteByte('\n')
		buf.WriteString(renderBoxLine(b.width, row.String(), fill))
	}

	buf.WriteByte('\n')
	buf.WriteString(renderBoxLine(b.width, ansiDim+truncate(listHint, cw)+ansiReset, ""))
	buf.WriteByte('\n')
	buf.WriteString(renderBoxBottom(b.width))
	return buf.Bytes()
}
