package ipscan

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

const ieeeOuiURL = "http://standards-oui.ieee.org/oui.txt"

// Global OUI map (once loaded)
var ouiOnce sync.Once

// Download and parse the IEEE OUI file
func loadOUITable() (map[string]string, error) {
	ouiMap := make(map[string]string)

	resp, err := http.Get(ieeeOuiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse lines that contain an OUI prefix
		// Example: "00-04-F2   (hex)  Some Vendor Inc."
		if strings.Contains(line, "(hex)") {
			parts := strings.Split(line, "(hex)")
			if len(parts) >= 2 {
				// Left of "(hex)" part: "00-04-F2"
				oui := strings.TrimSpace(parts[0])
				// Right part contains vendor name
				vendor := strings.TrimSpace(parts[1])

				// Normalize prefix (remove hyphens)
				key := strings.ToUpper(strings.ReplaceAll(oui, "-", ""))

				ouiMap[key] = vendor
			}
		}
	}
	if scanner.Err() != nil {
		return nil, fmt.Errorf("error scanning OUI file: %v", scanner.Err())
	}

	return ouiMap, nil
}

// Lookup vendor name for a MAC
func lookupVendor(mac string) string {
	// Ensure table loaded once
	var err error
	var ouiMap map[string]string
	ouiOnce.Do(func() {
		ouiMap, err = loadOUITable()
	})
	if err != nil {
		return "Unknown"
	}

	// Clean MAC (remove colons/hyphens)
	clean := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, ":", ""), "-", ""))
	if len(clean) < 6 {
		return "Unknown"
	}

	// First 6 hex characters = OUI
	prefix := clean[:6]

	if vendor, ok := ouiMap[prefix]; ok {
		return vendor
	}
	return "Unknown"
}
