package cli

import (
	"testing"
)

func TestCmdShellBash(t *testing.T) {
	got := captureStdout(func() {
		cmdShell(nil)
	})
	if !contains(got, "__decoreba_widget") {
		t.Fatalf("bash widget missing, got: %s", got)
	}
	if !contains(got, "bind") {
		t.Fatalf("bash widget missing bind, got: %s", got)
	}
}

func TestCmdShellBashExplicit(t *testing.T) {
	got := captureStdout(func() {
		cmdShell([]string{"bash"})
	})
	if !contains(got, "__decoreba_widget") {
		t.Fatalf("bash widget missing, got: %s", got)
	}
}

func TestCmdShellZsh(t *testing.T) {
	got := captureStdout(func() {
		cmdShell([]string{"zsh"})
	})
	if !contains(got, "__decoreba_widget") {
		t.Fatalf("zsh widget missing, got: %s", got)
	}
	if !contains(got, "zle") {
		t.Fatalf("zsh widget missing zle, got: %s", got)
	}
}
