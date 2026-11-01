package main

import (
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetTheme() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Theme
}

// SetTheme: "" (системна), "light" або "dark".
func (a *App) SetTheme(theme string) {
	switch theme {
	case "", "light", "dark":
	default:
		return
	}
	a.mu.Lock()
	a.settings.Theme = theme
	a.mu.Unlock()

	if err := a.settings.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "settings save:", err)
	}
	if a.ctx != nil {
		a.applyNativeTheme()
		runtime.MenuSetApplicationMenu(a.ctx, a.buildMenu()) // оновити радіо-пункти
		runtime.EventsEmit(a.ctx, "theme-changed", theme)
	}
}

// applyNativeTheme перемикає тему рамки вікна (діє на Windows; на інших ОС — без ефекту).
func (a *App) applyNativeTheme() {
	switch a.GetTheme() {
	case "light":
		runtime.WindowSetLightTheme(a.ctx)
	case "dark":
		runtime.WindowSetDarkTheme(a.ctx)
	default:
		runtime.WindowSetSystemDefaultTheme(a.ctx)
	}
}

func (a *App) toast(msg string) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "toast", msg)
	}
}
