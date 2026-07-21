package tui

import (
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestRunCommand(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		err := RunCommand(&core.Command{Command: "true"})
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
	t.Run("failing command", func(t *testing.T) {
		err := RunCommand(&core.Command{Command: "false"})
		if err == nil {
			t.Fatal("expected error for failing command")
		}
	})
}
