package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func TestConfigDir(t *testing.T) {
	t.Run("env var", func(t *testing.T) {
		t.Setenv("DECOREBA_CONFIG", "/tmp/decoreba/test/commands.json")
		dir, err := ConfigDir()
		if err != nil {
			t.Fatal(err)
		}
		if dir != "/tmp/decoreba/test" {
			t.Fatalf("ConfigDir = %q, want /tmp/decoreba/test", dir)
		}
	})

	t.Run("env var dir only", func(t *testing.T) {
		t.Setenv("DECOREBA_CONFIG", "/custom/dir/")
		dir, err := ConfigDir()
		if err != nil {
			t.Fatal(err)
		}
		if dir != "/custom/dir" {
			t.Fatalf("ConfigDir = %q, want /custom/dir", dir)
		}
	})
}

func TestConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "commands.json" {
		t.Fatalf("got %q, want commands.json", filepath.Base(path))
	}
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		t.Fatal("ConfigPath should create the directory")
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	original := &core.Store{
		Version: 1,
		Commands: []core.Command{
			{ID: "abc", Context: "git", Title: "Undo", Command: "git reset"},
		},
	}
	if err := Save(original); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != 1 || len(loaded.Commands) != 1 || loaded.Commands[0].ID != "abc" {
		t.Fatalf("loaded = %+v", loaded)
	}
}

func TestSaveAtomicity(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s := &core.Store{Version: 1}
	if err := Save(s); err != nil {
		t.Fatal(err)
	}

	path, _ := ConfigPath()
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Fatal(".tmp file should not exist after Save")
	}
}

func TestLoadFirstRunSeeds(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	s, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != 1 {
		t.Fatalf("version = %d, want 1", s.Version)
	}
	if len(s.Commands) == 0 {
		t.Fatal("seed should have commands")
	}
	// Verify seed persisted
	path, _ := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fromDisk core.Store
	if err := json.Unmarshal(data, &fromDisk); err != nil {
		t.Fatal(err)
	}
	if len(fromDisk.Commands) != len(s.Commands) {
		t.Fatalf("disk has %d commands, want %d", len(fromDisk.Commands), len(s.Commands))
	}
}

func TestLoadCorruptedFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("DECOREBA_CONFIG", filepath.Join(tmp, "commands.json"))

	cfgPath, _ := ConfigPath()
	os.WriteFile(cfgPath, []byte("{not json"), 0o600)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for corrupted file")
	}
}

func TestConfigDirDefault(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if dir == "" {
		t.Fatal("ConfigDir returned empty")
	}
	// Should end with /decoreba
	if filepath.Base(dir) != "decoreba" {
		t.Fatalf("ConfigDir = %q, want .../decoreba", dir)
	}
}

func TestStatPath(t *testing.T) {
	t.Run("path exists", func(t *testing.T) {
		tmp := t.TempDir()
		f := filepath.Join(tmp, "test.txt")
		os.WriteFile(f, []byte("hello"), 0o644)
		info, err := StatPath(f)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 5 {
			t.Fatalf("size = %d", info.Size())
		}
	})

	t.Run("path not exists", func(t *testing.T) {
		_, err := StatPath("/nonexistent/path/xyz")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestLoadNonExistentDir(t *testing.T) {
	t.Setenv("DECOREBA_CONFIG", "/nonexistent/path/commands.json")
	_, err := Load()
	if err != nil {
		t.Logf("expected error for impossible path: %v", err)
	}
}
