package cli

import (
	"strings"
	"testing"
)

func TestCompletionIncludesAllCommands(t *testing.T) {
	expected := []string{
		"add", "list", "ls", "rm", "remove",
		"edit", "stats", "init", "shell",
		"sync", "mcp", "export", "import",
		"version", "help", "completion",
	}

	for _, cmd := range expected {
		if !strings.Contains(completionBash, cmd) {
			t.Errorf("bash completion missing %q", cmd)
		}
		if !strings.Contains(completionZsh, cmd) {
			t.Errorf("zsh completion missing %q", cmd)
		}
		if !strings.Contains(completionFish, cmd) {
			t.Errorf("fish completion missing %q", cmd)
		}
	}
}

func TestCompletionBashSyntax(t *testing.T) {
	if !strings.Contains(completionBash, "complete -F _decoreba_completion") {
		t.Fatal("bash completion should register the function")
	}
}

func TestCompletionZshSyntax(t *testing.T) {
	if !strings.Contains(completionZsh, "#compdef decoreba") {
		t.Fatal("zsh completion should start with #compdef")
	}
}

func TestCompletionFishSyntax(t *testing.T) {
	if !strings.Contains(completionFish, "complete -c decoreba") {
		t.Fatal("fish completion should have complete commands")
	}
}

func TestCmdCompletion(t *testing.T) {
	t.Run("default bash", func(t *testing.T) {
		got := captureStdout(func() {
			cmdCompletion(nil)
		})
		if !contains(got, "_decoreba_completion") {
			t.Fatalf("bash completion missing, got: %s", got)
		}
	})
	t.Run("bash", func(t *testing.T) {
		got := captureStdout(func() {
			cmdCompletion([]string{"bash"})
		})
		if !contains(got, "complete -F _decoreba_completion") {
			t.Fatalf("bash completion missing, got: %s", got)
		}
	})
	t.Run("zsh", func(t *testing.T) {
		got := captureStdout(func() {
			cmdCompletion([]string{"zsh"})
		})
		if !contains(got, "#compdef decoreba") {
			t.Fatalf("zsh completion missing, got: %s", got)
		}
	})
	t.Run("fish", func(t *testing.T) {
		got := captureStdout(func() {
			cmdCompletion([]string{"fish"})
		})
		if !contains(got, "complete -c decoreba") {
			t.Fatalf("fish completion missing, got: %s", got)
		}
	})
}
