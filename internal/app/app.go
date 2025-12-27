package app

import (
	"bufio"
	"copy/internal/wire"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
	log.Printf("[app] Local IP: %s", getLocalIP())
	err := a.startServer()
	if err != nil {
		return err
	}
	// Scan and let user choose IP or search again
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Scan and find reachable IPs
		reachableIPs := findReachableIPs(a.port)

		if len(reachableIPs) == 0 {
			fmt.Println("\n[app] No reachable peers found.")
			fmt.Print("Press Enter to search again, or 'q' to quit: ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "q" || input == "Q" {
				break
			}
			continue
		}

		// Prompt user to select an IP
		selectedIP := promptUserSelection(reachableIPs, scanner)
		if selectedIP == "" {
			// User chose to refresh or invalid input, loop again to rescan
			continue
		}

		// Connect to selected IP and gain control (blocks until Ctrl+Shift+B or connection fails)
		log.Printf("[app] Connecting to %s:%s...", selectedIP, a.port)
		if err := runControl(selectedIP, a.port); err != nil {
			log.Printf("[app] Control session ended: %v", err)
			fmt.Println("\nControl session ended. Returning to peer selection...")
			// Loop again to rescan and prompt
		}
	}

	return nil
}
