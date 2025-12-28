package main

import (
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

type App struct {
	port   string
	stream *Stream
}

func NewApp(port string) (*App, error) {
	return &App{
		port:   port,
		stream: NewStream(),
	}, nil
}

func (a *App) Run() error {
	// Start video stream server
	a.stream.Start(":9000")

	return wails.Run(&options.App{
		Title:  "Remote Viewer",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			a.stream, // 👈 expose to frontend
		},
	})
}
