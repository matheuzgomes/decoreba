package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeStatus(t *testing.T) {
	cfg := &Config{
		LastLocalHash:  "aaa",
		LastRemoteHash: "aaa",
	}

	t.Run("clean when hashes match", func(t *testing.T) {
		if got := ComputeStatus(cfg, "aaa", "aaa"); got != StatusClean {
			t.Fatalf("got %q, want clean", got)
		}
	})

	t.Run("no remote", func(t *testing.T) {
		if got := ComputeStatus(cfg, "aaa", ""); got != StatusNoRemote {
			t.Fatalf("got %q, want no-remote", got)
		}
	})

	t.Run("ahead", func(t *testing.T) {
		if got := ComputeStatus(cfg, "bbb", "aaa"); got != StatusAhead {
			t.Fatalf("got %q, want ahead", got)
		}
	})

	t.Run("behind", func(t *testing.T) {
		if got := ComputeStatus(cfg, "aaa", "bbb"); got != StatusBehind {
			t.Fatalf("got %q, want behind", got)
		}
	})

	t.Run("diverged", func(t *testing.T) {
		if got := ComputeStatus(cfg, "bbb", "ccc"); got != StatusDiverged {
			t.Fatalf("got %q, want diverged", got)
		}
	})

	t.Run("clean when last sync matches despite hash change", func(t *testing.T) {
		cfg2 := &Config{LastLocalHash: "aaa", LastRemoteHash: "aaa"}
		// Local has changed and remote has changed back
		if got := ComputeStatus(cfg2, "bbb", "aaa"); got != StatusAhead {
			t.Fatalf("got %q, want ahead", got)
		}
	})
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("hello"))
	h2 := HashBytes([]byte("hello"))
	h3 := HashBytes([]byte("world"))

	if h1 != h2 {
		t.Fatal("same input should produce same hash")
	}
	if h1 == h3 {
		t.Fatal("different input should produce different hash")
	}
	if len(h1) != 64 {
		t.Fatalf("hash length = %d, want 64", len(h1))
	}
}

func TestFileHash(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	os.WriteFile(path, []byte("hello"), 0o644)

	h, err := FileHash(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != 64 {
		t.Fatalf("hash length = %d", len(h))
	}

	if _, err := FileHash("/nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Backend != "gist" {
		t.Fatalf("backend = %q, want gist", cfg.Backend)
	}
	if cfg.AutoSync != "off" {
		t.Fatalf("auto_sync = %q, want off", cfg.AutoSync)
	}
	if cfg.Encrypt {
		t.Fatal("encrypt should default to false")
	}
}

func TestLoadConfigAndSaveConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	// First load returns default
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Backend != "gist" {
		t.Fatalf("backend = %q", cfg.Backend)
	}

	// Save custom config
	cfg.GistID = "abc123"
	cfg.Encrypt = true
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	// Load again
	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.GistID != "abc123" {
		t.Fatalf("gist_id = %q, want abc123", cfg2.GistID)
	}
	if !cfg2.Encrypt {
		t.Fatal("encrypt should be true")
	}
}

func TestLoadConfigCorrupted(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	path, _ := ConfigPath()
	os.MkdirAll(filepath.Dir(path), 0o700)
	os.WriteFile(path, []byte("{invalid"), 0o600)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Backend != "gist" {
		t.Fatal("corrupted config should fallback to default")
	}
}

func TestStorePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	path, err := StorePath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "commands.json" {
		t.Fatalf("got %q", path)
	}
}
