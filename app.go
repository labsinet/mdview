package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// RenderResult — те, що отримує фронтенд.
type RenderResult struct {
	Path string `json:"path"`
	Name string `json:"name"`
	HTML string `json:"html"`
}

type App struct {
	ctx      context.Context
	history  *History
	settings *Settings
	i18n     *I18n
	lang     string // активна мова (під a.mu)

	mu         sync.Mutex
	initial    string // файл, з яким запустили програму
	current    string // файл, що відкритий зараз
	watchedDir string
	watcher    *fsnotify.Watcher
	debounce   *time.Timer
}

func NewApp() *App {
	a := &App{
		settings: LoadSettings(),
		i18n:     LoadI18n(),
	}
	a.lang = a.i18n.Resolve(a.settings.Language)
	// Windows/Linux: шлях приходить аргументом командного рядка.
	a.initial = firstFileArg(os.Args[1:])
	return a
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.applyNativeTheme()

	h, err := NewHistory()
	if err != nil {
		fmt.Fprintln(os.Stderr, "history disabled:", err)
		h = &History{file: os.DevNull}
	}
	a.history = h

	// Перетягування файлу у вікно.
	runtime.OnFileDrop(ctx, func(_, _ int, paths []string) {
		for _, p := range paths {
			if isMarkdownFile(p) {
				a.openExternal(p)
				return
			}
		}
		if len(paths) > 0 {
			runtime.EventsEmit(ctx, "open-error", a.tr("ui.unsupported"))
		}
	})

	w, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintln(os.Stderr, "file watching disabled:", err)
		return
	}
	a.watcher = w
	go a.watchLoop()
}

func (a *App) shutdown(_ context.Context) {
	if a.watcher != nil {
		_ = a.watcher.Close()
	}
}

// ---------- Методи, доступні з JS (window.go.main.App.*) ----------

// GetInitialFile повертає файл, з яким запущено програму (або "").
func (a *App) GetInitialFile() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	p := a.initial
	a.initial = ""
	return p
}

// OpenFile читає, рендерить, додає в історію і ставить на спостереження.
func (a *App) OpenFile(path string) (*RenderResult, error) {
	res, err := a.render(path)
	if err != nil {
		return nil, err
	}
	if err := a.history.Add(res.Path); err != nil {
		fmt.Fprintln(os.Stderr, "history save:", err)
	}
	a.watch(res.Path)
	return res, nil
}

// PickFile показує системний діалог вибору файлу.
func (a *App) PickFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: a.tr("dialog.open"),
		Filters: []runtime.FileFilter{
			{DisplayName: a.tr("dialog.filterMd"), Pattern: "*.md;*.markdown"},
			{DisplayName: a.tr("dialog.filterAll"), Pattern: "*.*"},
		},
	})
}

// GetHighlightCSS віддає CSS підсвічування коду (світла + темна теми).
func (a *App) GetHighlightCSS() string { return HighlightCSS() }

func (a *App) GetHistory() []HistoryEntry { return a.history.List() }
func (a *App) RemoveFromHistory(path string) error { return a.history.Remove(path) }
func (a *App) ClearHistory() error { return a.history.Clear() }

// ---------- Внутрішня логіка ----------

func (a *App) render(path string) (*RenderResult, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	src, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	html, err := RenderMarkdown(src, filepath.Dir(abs))
	if err != nil {
		return nil, err
	}
	return &RenderResult{Path: abs, Name: filepath.Base(abs), HTML: html}, nil
}

// watch стежить за каталогом файлу (а не за самим файлом): так працюють
// редактори, що зберігають через тимчасовий файл + rename.
func (a *App) watch(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.current = path
	if a.watcher == nil {
		return
	}
	dir := filepath.Dir(path)
	if dir == a.watchedDir {
		return
	}
	if a.watchedDir != "" {
		_ = a.watcher.Remove(a.watchedDir)
	}
	if err := a.watcher.Add(dir); err != nil {
		fmt.Fprintln(os.Stderr, "watch:", err)
		a.watchedDir = ""
		return
	}
	a.watchedDir = dir
}

func (a *App) watchLoop() {
	for {
		select {
		case ev, ok := <-a.watcher.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			a.mu.Lock()
			match := filepath.Clean(ev.Name) == a.current
			if match {
				if a.debounce != nil {
					a.debounce.Stop()
				}
				a.debounce = time.AfterFunc(120*time.Millisecond, a.reloadCurrent)
			}
			a.mu.Unlock()
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintln(os.Stderr, "watcher:", err)
		}
	}
}

func (a *App) reloadCurrent() {
	a.mu.Lock()
	path := a.current
	a.mu.Unlock()
	if path == "" {
		return
	}
	res, err := a.render(path)
	if err != nil {
		return // файл може бути на мить недоступний під час збереження
	}
	runtime.EventsEmit(a.ctx, "file-changed", res)
}

// openExternal викликається, коли файл прийшов ззовні (2-й запуск / Finder).
func (a *App) openExternal(path string) {
	if a.ctx == nil { // подія прийшла ще до startup
		a.mu.Lock()
		a.initial = path
		a.mu.Unlock()
		return
	}
	res, err := a.OpenFile(path)
	if err != nil {
		runtime.EventsEmit(a.ctx, "open-error", err.Error())
		return
	}
	runtime.WindowUnminimise(a.ctx)
	runtime.Show(a.ctx)
	runtime.EventsEmit(a.ctx, "file-opened", res)
}

func (a *App) onSecondInstance(data options.SecondInstanceData) {
	if p := firstFileArg(data.Args); p != "" {
		a.openExternal(p)
	} else if a.ctx != nil {
		runtime.WindowUnminimise(a.ctx)
		runtime.Show(a.ctx)
	}
}

func (a *App) onFileOpen(path string) { a.openExternal(path) }

// localFileHandler віддає локальні картинки за адресою /_local?p=...
// Лише безпечні типи, щоб сторінка не могла читати довільні файли.
func (a *App) localFileHandler() http.Handler {
	allowed := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".webp": true, ".svg": true, ".bmp": true, ".ico": true, ".avif": true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_local" {
			http.NotFound(w, r)
			return
		}
		p := r.URL.Query().Get("p")
		if p == "" || !allowed[strings.ToLower(filepath.Ext(p))] {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, p)
	})
}

// firstFileArg вибирає перший аргумент, що є існуючим файлом
// (відсіює прапорці, які може додати `wails dev`).
func firstFileArg(args []string) string {
	for _, s := range args {
		if strings.HasPrefix(s, "-") {
			continue
		}
		if st, err := os.Stat(s); err == nil && !st.IsDir() {
			if abs, err := filepath.Abs(s); err == nil {
				return abs
			}
			return s
		}
	}
	return ""
}

func isMarkdownFile(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".md", ".markdown", ".mdown", ".mkd", ".txt":
		return true
	}
	return false
}
