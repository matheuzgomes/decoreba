//go:build !windows

package clipboard

import "os/exec"

func hideWindow(cmd *exec.Cmd) {}
