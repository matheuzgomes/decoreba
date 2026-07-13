package core

import (
	"bytes"
	"errors"
	"os/exec"
	"runtime"
)

func CopyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
		hideWindow(cmd)
	default:
		if IsWSL() {
			if path, err := exec.LookPath("clip.exe"); err == nil {
				cmd = exec.Command(path)
			}
		}
		if cmd == nil {
			if path, err := exec.LookPath("wl-copy"); err == nil {
				cmd = exec.Command(path)
			} else if path, err := exec.LookPath("xclip"); err == nil {
				cmd = exec.Command(path, "-selection", "clipboard")
			} else if path, err := exec.LookPath("xsel"); err == nil {
				cmd = exec.Command(path, "--clipboard", "--input")
			} else if path, err := exec.LookPath("clip"); err == nil {
				cmd = exec.Command(path)
			} else {
				return errors.New("no clipboard tool found (install xclip, xsel or wl-clipboard)")
			}
		}
	}
	cmd.Stdin = bytes.NewBufferString(text)
	return cmd.Run()
}
