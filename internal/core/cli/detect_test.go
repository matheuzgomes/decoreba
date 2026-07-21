package cli

import (
	"os"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestDetectContext(t *testing.T) {
	t.Run("git signal with matching store", func(t *testing.T) {
		d := t.TempDir()
		os.MkdirAll(d+"/.git", 0o755)
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		s := &core.Store{Commands: []core.Command{
			{Context: "git", Title: "x", Command: "y"},
		}}
		if got := detectContext(s); got != "git" {
			t.Fatalf("got %q, want git", got)
		}
	})

	t.Run("git signal without store context returns lowercase", func(t *testing.T) {
		d := t.TempDir()
		os.MkdirAll(d+"/.git", 0o755)
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		s := &core.Store{Commands: []core.Command{
			{Context: "docker", Title: "x", Command: "y"},
		}}
		if got := detectContext(s); got != "git" {
			t.Fatalf("got %q, want git", got)
		}
	})

	t.Run("docker signal via Dockerfile", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(d+"/Dockerfile", []byte("FROM alpine"), 0o644)
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		s := &core.Store{}
		if got := detectContext(s); got != "docker" {
			t.Fatalf("got %q, want docker", got)
		}
	})

	t.Run("no signals returns empty", func(t *testing.T) {
		d := t.TempDir()
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		s := &core.Store{}
		if got := detectContext(s); got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("go.mod detected", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(d+"/go.mod", []byte("module test"), 0o644)
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		if got := detectContext(&core.Store{}); got != "go" {
			t.Fatalf("got %q, want go", got)
		}
	})

	t.Run("package.json detected", func(t *testing.T) {
		d := t.TempDir()
		os.WriteFile(d+"/package.json", []byte("{}"), 0o644)
		orig, _ := os.Getwd()
		os.Chdir(d)
		defer os.Chdir(orig)

		if got := detectContext(&core.Store{}); got != "npm" {
			t.Fatalf("got %q, want npm", got)
		}
	})
}
