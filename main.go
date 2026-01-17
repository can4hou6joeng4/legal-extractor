package main

import (
	"embed"

	"log/slog"
	"os"

	"legal-extractor/internal/app"
	"legal-extractor/internal/config"
	"legal-extractor/internal/extractor"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load Configuration
	configPath := os.Getenv("LEGAL_EXTRACTOR_CONFIG")
	if configPath == "" {
		configPath = "config/conf.yaml"
	}

	if err := config.Init(configPath); err != nil {
		slog.Warn("Could not initialize config", "path", configPath, "error", err)
	}

	// Get MCP configuration
	mcpCfg := config.GetMCP()

	// Initialize Extractor
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ext := extractor.NewExtractor(mcpCfg.Bin, mcpCfg.Args, logger)

	// Create an instance of the app structure
	application := app.NewApp(ext)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "legal-extractor",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        application.Startup,
		Bind: []interface{}{
			application,
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
