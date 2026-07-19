package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matheuzgomes/decoreba/internal/core"
)

// ConfigDir returns the config directory path without creating it.
// Respects the DECOREBA_CONFIG environment variable.
func ConfigDir() (string, error) {
	if env := os.Getenv("DECOREBA_CONFIG"); env != "" {
		return filepath.Dir(env), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "decoreba"), nil
}

// ConfigPath returns the full path to commands.json, creating the directory
// if needed.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "commands.json"), nil
}

func StatPath(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func Load() (*core.Store, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		s := &core.Store{Version: 1, Commands: seedCommands()}
		if saveErr := Save(s); saveErr != nil {
			return nil, saveErr
		}
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var s core.Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupted commands file (%s): %w", path, err)
	}
	return &s, nil
}

// Save writes atomically: data lands in a .tmp file and is renamed over the
// target, so a kill mid-write cannot corrupt the store.
func Save(s *core.Store) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
