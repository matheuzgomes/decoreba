package cli

import (
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
)

// detectContext scans the current directory for well-known signals and
// returns the first context name the user already has in their vault.
// Returns "" when no match is found.
func detectContext(s *core.Store) string {
	signals := []struct {
		check   func() bool
		context string
	}{
		{func() bool { return dirExists(".git") }, "git"},
		{func() bool { return fileExists("Dockerfile") || fileExists("docker-compose.yml") || fileExists("docker-compose.yaml") }, "docker"},
		{func() bool { return fileExists("go.mod") }, "go"},
		{func() bool { return fileExists("Makefile") }, "make"},
		{func() bool { return fileExists("package.json") }, "npm"},
		{func() bool { return fileExists("Cargo.toml") }, "rust"},
		{func() bool { return fileExists("pyproject.toml") || fileExists("requirements.txt") }, "python"},
		{func() bool { return fileExists("Gemfile") }, "ruby"},
		{func() bool { return fileExists("CMakeLists.txt") }, "cmake"},
		{func() bool { return fileExists("deno.json") || fileExists("deno.jsonc") }, "deno"},
		{func() bool { return fileExists("kubectl") || dirExists(".kube") }, "kubernetes"},
		{func() bool { return fileExists(".tmux.conf") || fileExists("tmux.conf") }, "tmux"},
	}

	for _, sig := range signals {
		if sig.check() {
			for _, c := range s.Commands {
				if strings.EqualFold(c.Context, sig.context) {
					return c.Context
				}
			}
			// Signal found but no matching context yet — try the lowercase name.
			return sig.context
		}
	}
	return ""
}

func dirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}
