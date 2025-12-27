package app

import (
	"copy/internal/shared"
	"copy/internal/ui"
	"copy/internal/wire"
	"fmt"
	"log"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
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

func (a *App) Run() error {
	log.Printf("[app] Starting on %s", runtime.GOOS)
	log.Printf("[app] Local IP: %s", shared.GetLocalIP())

	if err := a.startServer(); err != nil {
		return err
	}

	assetServer, err := ui.NewAssetServer()
	if err != nil {
		return fmt.Errorf("failed to create asset server: %v", err)
	}

	ui := ui.NewUI(a, a.port)

	return wails.Run(&options.App{
		Title:  "IOCopy",
		Width:  420,
		Height: 500,

		AssetServer: assetServer.GetServer(),

		OnStartup: ui.Startup,
		Bind: []interface{}{
			ui,
		},
	})
}
