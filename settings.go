package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// configDir: %APPDATA%\mdview, ~/Library/Application Support/mdview, ~/.config/mdview
func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "mdview")
	return dir, os.MkdirAll(dir, 0o755)
}

type Settings struct {
	Language string `json:"language"` // "" = мова системи
	Theme    string `json:"theme"`    // "" = системна, "light", "dark"

	file string
}

func LoadSettings() *Settings {
	s := &Settings{}
	if dir, err := configDir(); err == nil {
		s.file = filepath.Join(dir, "settings.json")
		if data, err := os.ReadFile(s.file); err == nil {
			_ = json.Unmarshal(data, s)
		}
	}
	return s
}

func (s *Settings) Save() error {
	if s.file == "" {
		return nil
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.file)
}
