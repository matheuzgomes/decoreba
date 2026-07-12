package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ConfigPath() (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	return path, nil
}

func StatPath(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "decoreba")
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "commands.json"), nil
}

func Load() (*Store, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		s := &Store{Version: 1, Commands: seedCommands()}
		if saveErr := Save(s); saveErr != nil {
			return nil, saveErr
		}
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("corrupted commands file (%s): %w", path, err)
	}
	return &s, nil
}

func Save(s *Store) error {
	path, err := configPath()
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
