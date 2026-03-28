package ui

import (
	"fmt"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// NewStreamAssetServer creates an asset server for the stream client
func NewStreamAssetServer() (*AssetServer, error) {
	// Create asset server with embedded HTML
	htmlFS, err := fs.Sub(frontendAssets, "frontend")
	if err != nil {
		log.Printf("[stream-frontend] Failed to create sub FS: %v", err)
		return nil, fmt.Errorf("failed to create sub FS: %v", err)
	}

	assetServer := &assetserver.Options{
		Assets: htmlFS,
	}
	return &AssetServer{Options: assetServer}, nil
}




