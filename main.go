package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

// Wails сам знаходить підкаталог з index.html усередині embed.FS.
//
//go:embed all:frontend
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "mdview",
		Width:     1000,
		Height:    760,
		MinWidth:  480,
		MinHeight: 320,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.localFileHandler(), // віддає картинки з диска
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true, // щоб webview не «переходив» на скинутий файл
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
		Menu:       app.buildMenu(),
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		Bind:       []interface{}{app},

		// Один екземпляр: повторний запуск (клік на іншому .md)
		// передає шлях у вже відкрите вікно.
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "7b0f5c1e-2a4d-4c55-9e0a-mdview-single-instance",
			OnSecondInstanceLaunch: app.onSecondInstance,
		},

		// macOS: файли з Finder приходять подією ОС, а не через os.Args.
		Mac: &mac.Options{
			OnFileOpen: app.onFileOpen,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
