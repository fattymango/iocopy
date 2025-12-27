package app

import (
	"copy/internal/wire"
	"embed"
	"io/fs"
	"log"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

const defaultPort = "8080"

type App struct {
	port   string
	server *wire.Server
}

func NewApp(port string) (*App, error) {
	// Don't create server immediately - create it lazily when needed
	// This prevents issues when Wails tries to generate bindings
	return &App{
		port: port,
	}, nil
}

//go:embed all:frontend
var frontendAssets embed.FS

func (a *App) Run() error {
	log.Printf("[app] Starting on %s", runtime.GOOS)
	log.Printf("[app] Local IP: %s", getLocalIP())

	if err := a.startServer(); err != nil {
		return err
	}

	ui := NewUI(a)

	// Create asset server with embedded HTML
	htmlFS, err := fs.Sub(frontendAssets, "frontend")
	if err != nil {
		log.Printf("[frontend] Failed to create sub FS: %v", err)
		htmlFS = frontendAssets
	}

	assetServer := &assetserver.Options{
		Assets: htmlFS,
	}

	return wails.Run(&options.App{
		Title:  "IOCopy",
		Width:  420,
		Height: 500,

		AssetServer: assetServer,

		OnStartup: ui.Startup,
		Bind: []interface{}{
			ui,
		},
	})
}
