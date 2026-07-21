package tui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/matheuzgomes/decoreba/internal/core"
)

// RunWorkflow opens an inline progress UI and executes the workflow steps
// interactively. Returns nil on success or when the user aborts.
func RunWorkflow(cmd *core.Command) error {
	if !cmd.IsWorkflow() {
		return nil
	}

	restore, err := makeRaw()
	if err != nil {
		return err
	}
	defer restore()

	w := &workflowRunner{
		cmd:    cmd,
		step:   0,
		status: make([]rune, len(cmd.Steps)),
	}
	w.overlay.init(nil)
	if UseTTY {
		if f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
			w.w = f
			defer f.Close()
		}
	}
	for i := range w.status {
		w.status[i] = ' '
	}
	w.status[0] = '→'
	w.redraw()

	buf := make([]byte, 64)
	for w.step < len(cmd.Steps) {
		n, err := readInput(buf)
		if err != nil {
			w.overlay.close()
			return err
		}

		handled := false
		for _, ev := range parseKeys(buf[:n]) {
			switch ev.kind {
			case keyEsc, keyCancel:
				w.overlay.close()
				return nil
			case keyExecute:
				if w.confirmExec {
					w.overlay.close()
					for w.step < len(cmd.Steps) {
						w.runCurrent()
					}
					return nil
				}
				w.confirmExec = true
				w.redraw()
				handled = true
			case keyEnter:
				w.confirmExec = false
				w.runCurrent()
				handled = true
			case keyRune:
				if w.confirmExec {
					if ev.r == 'y' || ev.r == 'Y' {
						w.overlay.close()
						for w.step < len(cmd.Steps) {
							w.runCurrent()
						}
						return nil
					}
					w.confirmExec = false
					handled = true
				}
			}
		}
		if handled {
			w.redraw()
		}
	}
	w.overlay.close()
	return nil
}

type workflowRunner struct {
	overlay
	cmd         *core.Command
	step        int
	status      []rune
	confirmExec bool
}

func (w *workflowRunner) runCurrent() {
	step := w.cmd.Steps[w.step]
	w.status[w.step] = '…'
	w.overlay.close()

	// Single running line — stays in scrollback as context.
	fmt.Printf("%s %s...\r\n", ansiAccent+"→"+ansiReset, step.Title)

	cmd := exec.Command("sh", "-c", step.Command)
	if UseTTY {
		tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
		if err == nil {
			cmd.Stdin = tty
			cmd.Stdout = tty
			cmd.Stderr = tty
			defer tty.Close()
		}
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()

	fmt.Println()

	if err != nil {
		w.status[w.step] = '✗'
	} else {
		w.status[w.step] = '✓'
	}
	w.step++
	if w.step < len(w.status) {
		w.status[w.step] = '→'
	}
}

func (w *workflowRunner) frameLines() int {
	return 2 + len(w.cmd.Steps)*2 + 2
}

func (w *workflowRunner) renderContent(wid, ht int) ([]byte, int, int) {
	cw := boxContentWidth(wid)

	var b bytes.Buffer
	header := fmt.Sprintf("%s (%d/%d)", w.cmd.Title, w.step+1, len(w.cmd.Steps))
	b.WriteString(renderBoxTop(wid))
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(wid, ansiBold+truncate(header, cw)+ansiReset, ""))

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
		b.WriteString(renderBoxLine(wid, line, fill))

		cmdLine := "  " + ansiSubtle + truncate(step.Command, cw-2) + ansiReset
		b.WriteByte('\n')
		b.WriteString(renderBoxLine(wid, cmdLine, fill))
	}

	hint := "enter next  ^x run all  esc cancel"
	if w.confirmExec {
		hint = "Run all remaining steps?  y/yes  n/no"
	}
	b.WriteByte('\n')
	b.WriteString(renderBoxLine(wid, ansiDim+truncate(hint, cw)+ansiReset, ""))
	b.WriteByte('\n')
	b.WriteString(renderBoxBottom(wid))

	return b.Bytes(), 1, 0
}

func (w *workflowRunner) redraw() {
	w.refresh(w.renderContent)
}
