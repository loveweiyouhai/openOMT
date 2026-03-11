package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	app := NewApp()

	// 嵌入的前端根目录为 frontend/dist（Vite 构建输出）
	rootFS, err := fs.Sub(frontendAssets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	err = wails.Run(&options.App{
		Title:     "",
		Width:     1000,
		Height:    700,
		MinWidth:  800,
		MinHeight: 500,
		AssetServer: &assetserver.Options{
			Assets: rootFS,
		},
		OnStartup:  app.Startup,
		OnShutdown: app.Shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			DisableWindowIcon: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
