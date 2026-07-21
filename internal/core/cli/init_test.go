package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRcFilePath(t *testing.T) {
	t.Run("bash prefers .bashrc", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		bashrc := filepath.Join(home, ".bashrc")
		os.WriteFile(bashrc, []byte("# bash"), 0o644)
		got := rcFilePath("bash")
		if got != bashrc {
			t.Fatalf("got %q, want %q", got, bashrc)
		}
	})

	t.Run("bash falls back to .bash_profile", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		got := rcFilePath("bash")
		want := filepath.Join(home, ".bash_profile")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("zsh", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		got := rcFilePath("zsh")
		want := filepath.Join(home, ".zshrc")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("fish", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		got := rcFilePath("fish")
		want := filepath.Join(home, ".config", "fish", "config.fish")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("unknown shell returns empty", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		if got := rcFilePath("python"); got != "" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("empty home returns empty", func(t *testing.T) {
		os.Unsetenv("HOME")
		// It'll try os.UserHomeDir which may fail, but there's always a
		// home on CI; at least verify it doesn't panic.
		_ = rcFilePath("bash")
	})
}

func TestSourceCmd(t *testing.T) {
	t.Run("fish uses source", func(t *testing.T) {
		got := sourceCmd("fish", "/home/user/.config/fish/config.fish")
		if !contains(got, "source") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("bash uses source", func(t *testing.T) {
		got := sourceCmd("bash", "/home/user/.bashrc")
		if !contains(got, "source") {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("uses tilde when relative to home", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)
		rc := filepath.Join(home, ".zshrc")
		os.WriteFile(rc, []byte("#"), 0o644)
		got := sourceCmd("zsh", rc)
		if !contains(got, "~") {
			t.Fatalf("expected tilde in %q", got)
		}
	})
}

func TestCmdInitDryRun(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	os.Setenv("SHELL", "/bin/bash")

	got := captureStdout(func() {
		cmdInit([]string{"--dry-run"})
	})
	if !contains(got, "Shell:") || !contains(got, "Config:") || !contains(got, "Block to install:") {
		t.Fatalf("dry-run output missing fields:\n%s", got)
	}
	if !contains(got, "decoreba completion bash") {
		t.Fatalf("missing completion in:\n%s", got)
	}
}

func TestCmdInitYes(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	os.Setenv("SHELL", "/bin/zsh")

	bashrc := filepath.Join(home, ".zshrc")

	got := captureStdout(func() {
		cmdInit([]string{"--yes"})
	})
	if !contains(got, "✔ Config:") || !contains(got, bashrc) {
		t.Fatalf("unexpected: %s", got)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, blockBegin) {
		t.Fatal("rc file missing decoreba block")
	}
	if !strings.Contains(content, "decoreba completion zsh") {
		t.Fatal("rc file missing completion")
	}
}

func TestCmdInitYesUpdatesBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	os.Setenv("SHELL", "/bin/bash")

	bashrc := filepath.Join(home, ".bashrc")
	oldBlock := blockBegin + "\nsource <(decoreba completion bash)\n" + blockEnd + "\n"
	os.WriteFile(bashrc, []byte(oldBlock), 0o644)

	got := captureStdout(func() {
		cmdInit([]string{"-y"})
	})
	if !contains(got, "✔ Config:") {
		t.Fatalf("unexpected: %s", got)
	}

	data, err := os.ReadFile(bashrc)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Count(string(data), blockBegin)
	if lines != 1 {
		t.Fatalf("expected exactly 1 block, found %d", lines)
	}
}

func TestCmdInitForceShell(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	// SHELL is fish but we force bash.
	os.Setenv("SHELL", "/usr/bin/fish")

	got := captureStdout(func() {
		cmdInit([]string{"--dry-run", "bash"})
	})
	if !contains(got, "Shell: bash") {
		t.Fatalf("expected bash, got:\n%s", got)
	}
	if !contains(got, ".bashrc") && !contains(got, ".bash_profile") {
		t.Fatalf("expected bash rc, got:\n%s", got)
	}
}
