package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/term"
)

const (
	stepFieldTitle   = 0
	stepFieldCommand = 1
	stepFieldCount   = 2
)

var stepFieldLabels = [stepFieldCount]string{"title", "command"}

// EditSteps opens an inline overlay for editing workflow steps.
// Uses the same redraw/close pattern as palette and addform.
// Raw mode must already be active.
func EditSteps(steps []core.WorkflowStep, width int, out io.Writer) ([]core.WorkflowStep, bool, error) {
	if out == nil {
		out = os.Stdout
	}
	e := &stepEditor{
		steps: append([]core.WorkflowStep(nil), steps...),
		focus: 0,
		width: width,
		out:   out,
	}
	e.width, e.height = readTermSize()
	if len(e.steps) == 0 {
		e.focus = -1
	}
	e.redraw()

	buf := make([]byte, 64)
	for {
		n, err := term.ReadInput(buf)
		if err != nil {
			e.close()
			return nil, false, err
		}

		done, cancelled := e.apply(parseKeys(buf[:n]))
		if done {
			e.close()
			if cancelled {
				return nil, true, nil
			}
			return e.steps, false, nil
		}
		e.redraw()
	}
}

type stepEditor struct {
	steps      []core.WorkflowStep
	focus      int
	width      int
	height     int
	lines      int
	parkedLine int
	out        io.Writer

	editing  bool
	editNew  bool
	editStep int
	editBuf  [stepFieldCount][]rune
	editCur  [stepFieldCount]int
	editFld  int
}

func (e *stepEditor) apply(events []keyEvent) (done, cancelled bool) {
	if e.editing {
		return e.applyEdit(events)
	}

	for _, ev := range events {
		switch ev.kind {
		case keyEsc, keyCancel:
			return true, true
		case keySave:
			return true, false
		case keyEnter:
			return true, false
		case keyEdit:
			if e.focus >= 0 && e.focus < len(e.steps) {
				e.startEdit(e.focus, false)
			}
		case keyStepAdd:
			e.startEdit(len(e.steps), true)
		case keyStepDelete:
			if e.focus >= 0 && e.focus < len(e.steps) {
				e.steps = append(e.steps[:e.focus], e.steps[e.focus+1:]...)
				if e.focus >= len(e.steps) {
					e.focus = len(e.steps) - 1
				}
			}
		case keyUp:
			if e.focus > 0 {
				e.focus--
			}
		case keyDown:
			if e.focus < len(e.steps)-1 {
				e.focus++
			}
		}
	}
	return false, false
}

func (e *stepEditor) applyEdit(events []keyEvent) (done, cancelled bool) {
	for _, ev := range events {
		switch ev.kind {
		case keyEsc, keyCancel:
			e.editing = false
			return false, false
		case keyEnter, keySave:
			title := string(e.editBuf[stepFieldTitle])
			cmd := string(e.editBuf[stepFieldCommand])
			if title == "" || cmd == "" {
				break
			}
			step := core.WorkflowStep{Title: title, Command: cmd}
			if e.editNew {
				e.steps = append(e.steps, step)
				e.focus = len(e.steps) - 1
			} else if e.editStep >= 0 && e.editStep < len(e.steps) {
				e.steps[e.editStep] = step
				e.focus = e.editStep
			}
			e.editing = false
			return false, false
		case keyTab:
			e.editFld = (e.editFld + 1) % stepFieldCount
			e.editCur[e.editFld] = len(e.editBuf[e.editFld])
		case keyShiftTab:
			e.editFld = (e.editFld - 1 + stepFieldCount) % stepFieldCount
			e.editCur[e.editFld] = len(e.editBuf[e.editFld])
		case keyLeft:
			if e.editCur[e.editFld] > 0 {
				e.editCur[e.editFld]--
			}
		case keyRight:
			if e.editCur[e.editFld] < len(e.editBuf[e.editFld]) {
				e.editCur[e.editFld]++
			}
		case keyBackspace:
			if e.editCur[e.editFld] > 0 {
				fld := e.editBuf[e.editFld]
				pos := e.editCur[e.editFld]
				e.editBuf[e.editFld] = append(fld[:pos-1], fld[pos:]...)
				e.editCur[e.editFld]--
			}
		case keyDelete:
			if e.editCur[e.editFld] < len(e.editBuf[e.editFld]) {
				fld := e.editBuf[e.editFld]
				pos := e.editCur[e.editFld]
				e.editBuf[e.editFld] = append(fld[:pos], fld[pos+1:]...)
			}
		case keyRune:
			fld := e.editBuf[e.editFld]
			pos := e.editCur[e.editFld]
			e.editBuf[e.editFld] = append(fld[:pos], append([]rune{ev.r}, fld[pos:]...)...)
			e.editCur[e.editFld]++
		}
	}
	return false, false
}

func (e *stepEditor) startEdit(idx int, isNew bool) {
	e.editing = true
	e.editNew = isNew
	e.editStep = idx
	e.editFld = stepFieldTitle
	if isNew {
		e.editBuf[stepFieldTitle] = nil
		e.editBuf[stepFieldCommand] = nil
		e.editCur[stepFieldTitle] = 0
		e.editCur[stepFieldCommand] = 0
	} else if idx >= 0 && idx < len(e.steps) {
		e.editBuf[stepFieldTitle] = []rune(e.steps[idx].Title)
		e.editBuf[stepFieldCommand] = []rune(e.steps[idx].Command)
		e.editCur[stepFieldTitle] = len(e.editBuf[stepFieldTitle])
		e.editCur[stepFieldCommand] = len(e.editBuf[stepFieldCommand])
	}
}

func (e *stepEditor) frameLines() int {
	body := 0
	if len(e.steps) > 0 {
		body = len(e.steps) * 2
	} else if !e.editing {
		body = 1
	}
	editLines := 0
	if e.editing {
		editLines = 3
	}

	if maxBody := e.height - 4 - editLines; maxBody > 0 && body > maxBody {
		body = maxBody
		if body%2 == 1 {
			body--
		}
	}
	return 1 + 1 + body + editLines + 1 + 1
}

func (e *stepEditor) inputLine() int {
	if !e.editing {
		return 1
	}
	body := 0
	if len(e.steps) > 0 {
		body = len(e.steps) * 2
	}
	return 2 + body + 1 + e.editFld
}

func (e *stepEditor) inputCol() int {
	if !e.editing {
		return 0
	}
	return boxLeftPad + labelPad + len(e.editBuf[e.editFld])
}

func (e *stepEditor) redraw() {
	e.width, e.height = readTermSize()
	var b bytes.Buffer

	if e.lines > 0 && e.parkedLine > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", e.parkedLine)
	}
	b.WriteString("\r")
	b.Write(e.renderFrame())
	b.WriteString("\x1b[J")

	newLines := e.frameLines()
	up := newLines - 1 - e.inputLine()
	if up > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", up)
	}
	b.WriteString("\r")
	if col := e.inputCol(); col > 0 {
		fmt.Fprintf(&b, "\x1b[%dC", col)
	}

	_, _ = e.out.Write(b.Bytes())
	e.lines = newLines
	e.parkedLine = e.inputLine()
}

func (e *stepEditor) close() {
	if e.lines > 0 && e.parkedLine > 0 {
		fmt.Fprintf(e.out, "\x1b[%dA", e.parkedLine)
	}
	_, _ = e.out.Write([]byte("\r\x1b[J"))
	e.lines = 0
	e.parkedLine = 0
}

func (e *stepEditor) renderFrame() []byte {
	var b bytes.Buffer
	cw := boxContentWidth(e.width)

	header := ansiBold + fmt.Sprintf("%d steps", len(e.steps)) + ansiReset
	b.WriteString(renderBoxTop(e.width))
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(e.width, header, ""))

	if len(e.steps) == 0 && !e.editing {
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(e.width, ansiDim+truncate("no steps yet — ^n to add", cw)+ansiReset, ""))
	} else {
		for i, s := range e.steps {
			focused := i == e.focus && !e.editing
			fill := ""
			num := fmt.Sprintf("%d", i+1)
			numPadding := " " // space after number for single digit alignment
			if i+1 < 10 {
				numPadding = " "
			}

			titleBudget := cw - len(num) - 2
			title := truncate(s.Title, titleBudget)

			var line string
			if focused {
				line = ansiAccent + num + ansiReset + numPadding + ansiBold + title + ansiReset
				fill = ansiFocusBg
			} else {
				line = ansiDim + num + ansiReset + numPadding + title
			}
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(e.width, line, fill))

			cmdBudget := cw - 2
			cmdLine := "  " + ansiSubtle + truncate(s.Command, cmdBudget-2) + ansiReset
			b.WriteByte('\n')
			b.WriteString(renderBoxLine(e.width, cmdLine, fill))
		}
	}

	if e.editing {
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(e.width, ansiDim+strings.Repeat("─", cw)+ansiReset, ""))

		for fi := 0; fi < stepFieldCount; fi++ {
			focused := fi == e.editFld
			label := stepFieldLabels[fi]
			pad := labelPad - len([]rune(label))
			if pad < 1 {
				pad = 1
			}
			labelStr := label + strings.Repeat(" ", pad)

			var fieldLine string
			if focused {
				fieldLine = ansiBold + labelStr + ansiReset
				fieldLine += ansiFocusBg
			} else {
				fieldLine = ansiDim + labelStr + ansiReset
			}

			val := string(e.editBuf[fi])
			fieldBudget := cw - labelPad
			fieldLine += truncate(val, fieldBudget)

			b.WriteByte('\n')
			fill := ""
			if focused {
				fill = ansiFocusBg
			}
			b.WriteString(renderBoxLine(e.width, fieldLine, fill))
		}
	}

	hint := ""
	if e.editing {
		hint = "tab next   ↵ save   esc cancel"
	} else {
		hint = "^n add   ^d del   ^e edit   ↑↓ nav   enter done   esc cancel"
	}
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(e.width, ansiDim+truncate(hint, cw)+ansiReset, ""))
	b.WriteByte('\n')
	b.WriteString(renderBoxBottom(e.width))
	return b.Bytes()
}
