package app

import (
	"copy/internal/control"
	"copy/internal/wire"
	"fmt"
	"log"
)

// runControl establishes control over the remote peer's keyboard and mouse
func (a *App) RunControl(targetIP, port string) error {
	log.Printf("[control] Attempting to connect to %s:%s...", targetIP, port)
	client, err := wire.NewClient(targetIP, port)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		client.Close()
		log.Printf("[control] Client connection closed")
	}()

	log.Printf("[control] Connected to %s:%s", targetIP, port)
	log.Printf("[control] Gaining control over remote device...")
	log.Printf("[control] Press Ctrl+Shift+B to stop control")

	// Create input controller with UI blackscreen (if available)
	var blackScreen control.BlackScreen
	if a.ui != nil {
		blackScreen = a.ui
	}
	controller := control.NewController(client, blackScreen)

	// Start controlling (this blocks until Ctrl+Shift+B or connection lost)
	log.Printf("[control] Starting controller...")
	err = controller.Start()
	if err != nil {
		log.Printf("[control] Controller error: %v", err)
		return fmt.Errorf("control session ended: %w", err)
	}

	log.Printf("[control] Control session ended normally")
	return nil
}
