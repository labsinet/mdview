package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxHistory = 50

type HistoryEntry struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	OpenedAt string `json:"openedAt"` // RFC3339; рядок, бо генератор біндингів Wails не знає time.Time
	Exists   bool      `json:"exists"` // заповнюється при читанні, у файл не пишемо суттєвого
}

type History struct {
	mu      sync.Mutex
	file    string
	entries []HistoryEntry
}

// NewHistory кладе history.json у каталог конфігурації користувача:
//
//	Windows: %APPDATA%\mdview\history.json
//	macOS:   ~/Library/Application Support/mdview/history.json
//	Linux:   ~/.config/mdview/history.json
func NewHistory() (*History, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir = filepath.Join(dir, "mdview")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	h := &History{file: filepath.Join(dir, "history.json")}
	h.load()
	return h, nil
}

func (h *History) load() {
	data, err := os.ReadFile(h.file)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &h.entries) // зіпсований файл просто ігноруємо
}

// save пише атомарно: тимчасовий файл + rename.
func (h *History) save() error {
	data, err := json.MarshalIndent(h.entries, "", "  ")
	if err != nil {
		return err
	}
	tmp := h.file + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, h.file)
}

// Add ставить файл на початок списку, прибираючи дублікат.
func (h *History) Add(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := []HistoryEntry{{
		Path:     path,
		Name:     filepath.Base(path),
		OpenedAt: time.Now().Format(time.RFC3339),
	}}
	for _, e := range h.entries {
		if e.Path != path {
			out = append(out, e)
		}
	}
	if len(out) > maxHistory {
		out = out[:maxHistory]
	}
	h.entries = out
	return h.save()
}

func (h *History) List() []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	res := make([]HistoryEntry, len(h.entries))
	for i, e := range h.entries {
		_, err := os.Stat(e.Path)
		e.Exists = err == nil
		res[i] = e
	}
	return res
}

func (h *History) Remove(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := h.entries[:0:0]
	for _, e := range h.entries {
		if e.Path != path {
			out = append(out, e)
		}
	}
	h.entries = out
	return h.save()
}

func (h *History) Clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = nil
	return h.save()
}
