package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version можна задати під час збірки:
//
//	wails build -ldflags "-X main.Version=1.0.0"
var Version = "0.3.0"

// about.json вбудовується в бінарник: тут автор, компанія, сайт.
//
//go:embed about.json
var aboutJSON []byte

type aboutMeta struct {
	Author  string `json:"author"`
	Company string `json:"company"`
	Website string `json:"website"`
}

type AboutInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Author    string `json:"author"`
	Company   string `json:"company"`
	Website   string `json:"website"`
	ConfigDir string `json:"configDir"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func (a *App) GetAboutInfo() AboutInfo {
	dir, _ := configDir()
	var meta aboutMeta
	_ = json.Unmarshal(aboutJSON, &meta)
	return AboutInfo{
		Name:      "mdview",
		Version:   Version,
		Author:    meta.Author,
		Company:   meta.Company,
		Website:   meta.Website,
		ConfigDir: dir,
		OS:        goruntime.GOOS,
		Arch:      goruntime.GOARCH,
	}
}

type I18nInfo struct {
	Lang      string            `json:"lang"`      // активна мова
	Pref      string            `json:"pref"`      // вибір користувача ("" = системна)
	Strings   map[string]string `json:"strings"`   // переклади
	Languages []LangInfo        `json:"languages"` // доступні мови
}

func (a *App) langState() (lang, pref string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.lang, a.settings.Language
}

func (a *App) GetI18n() I18nInfo {
	lang, pref := a.langState()
	return I18nInfo{Lang: lang, Pref: pref, Strings: a.i18n.Strings(lang), Languages: a.i18n.Languages()}
}

// SetLanguage: "" = мова системи. Зберігає вибір, перебудовує меню
// і повідомляє фронтенд.
func (a *App) SetLanguage(code string) {
	if code != "" && !a.i18n.Has(code) {
		return
	}
	a.mu.Lock()
	a.settings.Language = code
	a.lang = a.i18n.Resolve(code)
	a.mu.Unlock()

	if err := a.settings.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "settings save:", err)
	}
	if a.ctx != nil {
		runtime.MenuSetApplicationMenu(a.ctx, a.buildMenu())
		runtime.EventsEmit(a.ctx, "lang-changed", a.GetI18n())
	}
}

// tr — переклад для рядків, що показує Go (меню, діалоги).
func (a *App) tr(key string, kv ...string) string {
	lang, _ := a.langState()
	return a.i18n.T(lang, key, kv...)
}

func (a *App) emit(event string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event)
	}
}
