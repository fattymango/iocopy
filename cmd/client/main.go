package main

import (
	"copy/internal/ui"
	"flag"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

func main() {
	wsAddr := flag.String("ws", "localhost:8080", "WebSocket server address to connect to")
	flag.Parse()

	log.Println("=== Screen Streaming Client ===")
	log.Printf("Connecting to WebSocket: ws://%s/ws", *wsAddr)

	// Create asset server for stream client
	assetServer, err := ui.NewStreamAssetServer()
	if err != nil {
		log.Fatalf("failed to create asset server: %v", err)
	}

	// Create UI with WebSocket address
	uiInstance := ui.NewStreamUI(*wsAddr)

	// Run Wails app
	err = wails.Run(&options.App{
		Title:  "Screen Stream Viewer",
		Width:  1280,
		Height: 720,
		MinWidth:  800,
		MinHeight: 600,

		AssetServer: assetServer.GetServer(),

		OnStartup: uiInstance.Startup,
		Bind: []interface{}{
			uiInstance,
		},
	})

	if err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}

