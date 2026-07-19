package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/term"
)

// RunWorkflow opens an inline progress UI and executes the workflow steps
// interactively. Returns nil on success or when the user aborts.
func RunWorkflow(cmd *core.Command) error {
	if !cmd.IsWorkflow() {
		return nil
	}

	restore, err := term.MakeRaw()
	if err != nil {
		return err
	}
	defer restore()

	w := &workflowRunner{
		cmd:    cmd,
		step:   0,
		status: make([]rune, len(cmd.Steps)),
	}
	for i := range w.status {
		w.status[i] = ' '
	}
	w.status[0] = '→'
	w.draw()

	buf := make([]byte, 64)
	for w.step < len(cmd.Steps) {
		n, err := term.ReadInput(buf)
		if err != nil {
			w.clear()
			return err
		}

		handled := false
		for _, ev := range parseKeys(buf[:n]) {
			switch ev.kind {
			case keyEsc, keyCancel:
				w.clear()
				return nil
			case keyExecute:
				w.clear()
				for w.step < len(cmd.Steps) {
					w.runCurrent()
				}
				return nil
			case keyEnter:
				w.runCurrent()
				handled = true
			}
		}
		if handled {
			w.draw()
		}
	}
	w.clear()
	return nil
}

type workflowRunner struct {
	cmd    *core.Command
	step   int
	status []rune
	lines  int
}

func (w *workflowRunner) runCurrent() {
	step := w.cmd.Steps[w.step]
	w.status[w.step] = '…'
	w.clear()

	// Single running line — stays in scrollback as context.
	fmt.Printf("%s %s...\r\n", ansiAccent+"→"+ansiReset, step.Title)

	cmd := exec.Command("sh", "-c", step.Command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	// Blank line to separate command output from the next frame.
	fmt.Println()

	w.status[w.step] = '✓'
	w.step++
	if w.step < len(w.status) {
		w.status[w.step] = '→'
	}
}

func (w *workflowRunner) frameLines() int {
	return 2 + len(w.cmd.Steps)*2 + 2
}

func (w *workflowRunner) draw() {
	w.lines = w.frameLines()
	width, _ := readTermSize()

	var b bytes.Buffer
	cw := boxContentWidth(width)

	header := fmt.Sprintf("%s (%d/%d)", w.cmd.Title, w.step+1, len(w.cmd.Steps))
	b.WriteString(renderBoxTop(width))
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(width, ansiBold+truncate(header, cw)+ansiReset, ""))

	for i, step := range w.cmd.Steps {
		indicator := " "
		switch w.status[i] {
		case '→':
			indicator = ansiAccent + "→" + ansiReset
		case '✓':
			indicator = ansiDim + "✓" + ansiReset
		case '✗':
			indicator = ansiWarn + "✗" + ansiReset
		}

		text := step.Title
		fill := ""
		switch w.status[i] {
		case '→':
			text = ansiBold + text + ansiReset
			fill = ansiFocusBg
		case '✓', ' ':
			text = ansiDim + text + ansiReset
		case '✗':
			text = ansiWarn + text + ansiReset
		}

		line := indicator + " " + text
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(width, line, fill))

		cmdLine := "  " + ansiSubtle + truncate(step.Command, cw-2) + ansiReset
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(width, cmdLine, fill))
	}

	hint := "enter next   ^x run all   esc abort"
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(width, ansiDim+truncate(hint, cw)+ansiReset, ""))
	b.WriteByte('\n')
	b.WriteString(renderBoxBottom(width))

	up := w.lines - 1
	if up > 0 {
		fmt.Fprintf(&b, "\x1b[%dA", up)
	}
	b.WriteString("\r")

	_, _ = os.Stdout.Write(b.Bytes())
}

func (w *workflowRunner) clear() {
	if w.lines > 0 {
		fmt.Fprintf(os.Stdout, "\x1b[%dB", w.lines-1)
		for i := 0; i < w.lines; i++ {
			fmt.Fprint(os.Stdout, "\r\x1b[2K")
			if i < w.lines-1 {
				fmt.Fprint(os.Stdout, "\x1b[1A")
			}
		}
		w.lines = 0
	}
}
