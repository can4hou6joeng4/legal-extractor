package main

import (
	"embed"

	"legal-extractor/config"
	"legal-extractor/pkg/extractor"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Load Configuration
	cfg, err := config.LoadConfig("config/conf.yaml")
	if err != nil {
		println("Warning: Could not load config/conf.yaml, MCP OCR will be disabled.")
		println("Error:", err.Error())
	} else {
		// Configure Extractor
		extractor.SetMCPConfig(cfg.MCP.Bin, cfg.MCP.Args)
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "legal-extractor",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// 启用原生拖拽支持
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop:     true,
			DisableWebViewDrop: true,
			CSSDropProperty:    "--wails-drop-target",
			CSSDropValue:       "drop",
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
