package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/matheuzgomes/decoreba/internal/core/store"
)

const maxBackups = 5

func BackupStore() error {
	path, err := store.ConfigPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	for i := maxBackups - 1; i >= 1; i-- {
		old := filepath.Join(dir, fmt.Sprintf("commands.json.mcp.%d", i))
		next := filepath.Join(dir, fmt.Sprintf("commands.json.mcp.%d", i+1))
		if _, err := os.Stat(old); err == nil {
			os.Rename(old, next)
		}
	}
	dst := filepath.Join(dir, "commands.json.mcp.1")
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		return err
	}
	return nil
}
