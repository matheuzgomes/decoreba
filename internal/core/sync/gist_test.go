package sync

import (
	"path/filepath"
	"testing"
)

func TestNewGistBackend(t *testing.T) {
	b := NewGistBackend("test-token")
	if b == nil {
		t.Fatal("expected non-nil backend")
	}
	if b.Name() != "gist" {
		t.Fatalf("name = %q, want gist", b.Name())
	}
}

func TestConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "sync.json" {
		t.Fatalf("expected sync.json, got %q", path)
	}
}

func TestSaveConfigEdgeCases(t *testing.T) {
	t.Run("creates directory and saves", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

		cfg := DefaultConfig()
		cfg.GistID = "test123"
		err := SaveConfig(cfg)
		if err != nil {
			t.Fatal(err)
		}

		loaded, err := LoadConfig()
		if err != nil {
			t.Fatal(err)
		}
		if loaded.GistID != "test123" {
			t.Fatalf("gist_id = %q, want test123", loaded.GistID)
		}
	})
}

func TestComputeStatusEdgeCases(t *testing.T) {
	t.Run("clean when both hashes changed but match last sync", func(t *testing.T) {
		cfg := &Config{LastLocalHash: "xyz", LastRemoteHash: "xyz"}
		if got := ComputeStatus(cfg, "xyz", "xyz"); got != StatusClean {
			t.Fatalf("got %q, want clean", got)
		}
	})
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
