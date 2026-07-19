package tui

import (
	"os"
	"os/exec"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func RunCommand(cmd *core.Command) error {
	c := exec.Command("sh", "-c", cmd.Command)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
