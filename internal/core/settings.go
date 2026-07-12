package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	FontScale   float64 `json:"font_scale"`
	AlwaysOnTop bool    `json:"always_on_top"`
}

func DefaultSettings() Settings {
	return Settings{
		Width:       560,
		Height:      440,
		FontScale:   1.0,
		AlwaysOnTop: true,
	}
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(dir, "decoreba")
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(appDir, "settings.json"), nil
}

func LoadSettings() (Settings, error) {
	path, err := settingsPath()
	if err != nil {
		return DefaultSettings(), err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return DefaultSettings(), err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings(), nil
	}
	if s.Width < 400 {
		s.Width = 400
	}
	if s.Height < 280 {
		s.Height = 280
	}
	return s, nil
}

func SaveSettings(s Settings) error {
	path, err := settingsPath()
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
