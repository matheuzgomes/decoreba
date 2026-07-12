package core

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"runtime"
)

func isWSL() bool {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	return bytes.Contains(bytes.ToLower(data), []byte("microsoft")) ||
		bytes.Contains(bytes.ToLower(data), []byte("wsl"))
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		if isWSL() {
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
