package app

import (
	"context"
	"log"
)

type UI struct {
	app *App
	ctx context.Context
}

func NewUI(app *App) *UI {
	return &UI{app: app}
}

func (u *UI) Startup(ctx context.Context) {
	u.ctx = ctx
	log.Println("[ui] UI started")
}

// Called by frontend to scan peers
func (u *UI) ScanPeers() []string {
	log.Println("[ui] Scanning for peers...")
	return findReachableIPs(u.app.port)
}

// Called by frontend when user selects an IP
func (u *UI) Connect(ip string) error {
	log.Printf("[ui] Connecting to %s:%s", ip, u.app.port)
	return runControl(ip, u.app.port)
}

// Optional: expose local IP
func (u *UI) LocalIP() string {
	return getLocalIP()
}
