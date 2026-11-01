package main

import (
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildMenu збирає меню поточною мовою. Гарячі клавіші (Ctrl+O/F/B) обробляє
// фронтенд, тому в пунктах меню акселераторів немає — щоб не спрацьовували двічі.
func (a *App) buildMenu() *menu.Menu {
	m := menu.NewMenu()
	isMac := goruntime.GOOS == "darwin"
	about := func(_ *menu.CallbackData) { a.emit("show-about") }
	quit := func(_ *menu.CallbackData) { runtime.Quit(a.ctx) }

	if isMac {
		// На macOS «Про програму» і «Вийти» живуть у меню застосунку.
		app := m.AddSubmenu("mdview")
		app.AddText(a.tr("menu.about"), nil, about)
		app.AddSeparator()
		app.AddText(a.tr("menu.quit"), keys.CmdOrCtrl("q"), quit)
	}

	file := m.AddSubmenu(a.tr("menu.file"))
	file.AddText(a.tr("menu.open"), nil, func(_ *menu.CallbackData) { go a.openWithDialog() })
	file.AddText(a.tr("menu.saveAs"), nil, func(_ *menu.CallbackData) { go a.SaveAs() })
	if !isMac {
		file.AddSeparator()
		file.AddText(a.tr("menu.quit"), nil, quit)
	}

	if isMac {
		m.Append(menu.EditMenu()) // потрібне для Cmd+C / Cmd+A
	}

	view := m.AddSubmenu(a.tr("menu.view"))
	view.AddText(a.tr("menu.sidebar"), nil, func(_ *menu.CallbackData) { a.emit("menu-sidebar") })
	view.AddText(a.tr("menu.find"), nil, func(_ *menu.CallbackData) { a.emit("menu-find") })
	view.AddSeparator()

	theme := a.GetTheme()
	themes := view.AddSubmenu(a.tr("menu.theme"))
	for _, th := range []struct{ code, key string }{
		{"", "menu.themeAuto"}, {"light", "menu.themeLight"}, {"dark", "menu.themeDark"},
	} {
		code := th.code
		themes.AddRadio(a.tr(th.key), theme == code, nil, func(_ *menu.CallbackData) { go a.SetTheme(code) })
	}

	_, pref := a.langState()
	langs := view.AddSubmenu(a.tr("menu.language"))
	langs.AddRadio(a.tr("menu.auto"), pref == "", nil, func(_ *menu.CallbackData) { go a.SetLanguage("") })
	for _, l := range a.i18n.Languages() {
		code := l.Code
		langs.AddRadio(l.Name, pref == code, nil, func(_ *menu.CallbackData) { go a.SetLanguage(code) })
	}

	if !isMac {
		help := m.AddSubmenu(a.tr("menu.help"))
		help.AddText(a.tr("menu.about"), nil, about)
	}
	return m
}

func (a *App) openWithDialog() {
	if p, err := a.PickFile(); err == nil && p != "" {
		a.openExternal(p)
	}
}
