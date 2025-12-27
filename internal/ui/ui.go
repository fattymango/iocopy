package ui

import (
	"context"
	"copy/internal/shared"
	"log"
)

type UI struct {
	app  Application
	ctx  context.Context
	port string
}

func NewUI(app Application, port string) *UI {
	return &UI{app: app, port: port}
}

func (u *UI) Startup(ctx context.Context) {
	u.ctx = ctx
	log.Println("[ui] UI started")
}

// Called by frontend to scan peers
func (u *UI) ScanPeers() []string {
	log.Println("[ui] Scanning for peers...")
	return u.app.FindReachableIPs(u.port)
}

// Called by frontend when user selects an IP
func (u *UI) Connect(ip string) error {
	log.Printf("[ui] Connecting to %s:%s", ip, u.port)
	return u.app.RunControl(ip, u.port)
}

// Optional: expose local IP
func (u *UI) LocalIP() string {
	return shared.GetLocalIP()
}
