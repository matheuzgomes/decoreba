package sync

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/matheuzgomes/decoreba/internal/core/store"
)

const TokenEnvVar = "DECOREBA_GIST_TOKEN"

type Config struct {
	Backend  string `json:"backend"`
	GistID   string `json:"gist_id,omitempty"`
	AutoSync string `json:"auto_sync"`
	Encrypt  bool   `json:"encrypt"`

	LastLocalHash  string `json:"last_local_hash"`
	LastRemoteHash string `json:"last_remote_hash"`
}

const (
	StatusClean    = "clean"
	StatusAhead    = "ahead"
	StatusBehind   = "behind"
	StatusDiverged = "diverged"
	StatusNoRemote = "no-remote"
)

func DefaultConfig() *Config {
	return &Config{
		Backend:  "gist",
		AutoSync: "off",
	}
}

func ConfigPath() (string, error) {
	dir, err := store.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sync.json"), nil
}

func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return DefaultConfig(), err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), nil
	}
	if cfg.AutoSync == "" {
		cfg.AutoSync = "off"
	}
	return &cfg, nil
}

func SaveConfig(c *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func StorePath() (string, error) {
	return store.ConfigPath()
}

func FileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h), nil
}

func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func ComputeStatus(cfg *Config, localHash, remoteHash string) string {
	if remoteHash == "" {
		return StatusNoRemote
	}
	if localHash == remoteHash {
		return StatusClean
	}
	if localHash == cfg.LastLocalHash && remoteHash == cfg.LastRemoteHash {
		return StatusClean
	}

	if localHash == cfg.LastLocalHash && remoteHash != cfg.LastRemoteHash {
		return StatusBehind
	}
	if remoteHash == cfg.LastRemoteHash && localHash != cfg.LastLocalHash {
		return StatusAhead
	}
	return StatusDiverged
}
