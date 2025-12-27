package ui

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

type AssetServer struct {
	Options *assetserver.Options
}

func NewAssetServer() (*AssetServer, error) {
	// Create asset server with embedded HTML
	htmlFS, err := fs.Sub(frontendAssets, "frontend")
	if err != nil {
		log.Printf("[frontend] Failed to create sub FS: %v", err)
		return nil, fmt.Errorf("failed to create sub FS: %v", err)
	}

	assetServer := &assetserver.Options{
		Assets: htmlFS,
	}
	return &AssetServer{Options: assetServer}, nil
}

func (a *AssetServer) GetServer() *assetserver.Options {
	return a.Options
}
