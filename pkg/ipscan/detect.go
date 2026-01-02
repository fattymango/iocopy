package ipscan

import (
	"log"
	"strings"
	"sync"
)

// DetectDevicesType detects the type of each device using the OUI vendor list and hostname heuristics
func DetectDevicesType(d []Device) []Device {
	ouiDB, err := loadOUITable()
	if err != nil {
		log.Printf("[DetectDevicesType] Warning: OUI table unavailable (%v), using hostname heuristics only", err)
		// Continue without OUI lookup - will use hostname heuristics instead
	} else {
		log.Printf("[DetectDevicesType] OUI table loaded successfully")
	}

	var wg sync.WaitGroup
	log.Printf("[DetectDevicesType] Starting %d device type detection goroutines", len(d))
	for i := range d {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()
			if err != nil {
				detectDeviceType(&d[i], nil)
			} else {
				detectDeviceType(&d[i], ouiDB)
			}
		}(i)
	}

	wg.Wait()
	log.Printf("[DetectDevicesType] All device type detection goroutines completed")
	return d
}

func detectDeviceType(d *Device, ouiDB map[string]string) {
	// Step 0: ignore invalid MAC
	if d.MAC == "" || d.MAC == "00:00:00:00:00:00" || d.MAC == "00-00-00-00-00-00" {
		d.Type = "Unknown"
		return
	}

	// Step 1: MAC OUI lookup
	if ouiDB != nil {
		macClean := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(d.MAC, ":", ""), "-", ""))
		if len(macClean) >= 6 {
			oui := macClean[:6]
			if vendor, ok := ouiDB[oui]; ok {
				vendor = strings.ToLower(vendor)
				switch {
				case strings.Contains(vendor, "apple"), strings.Contains(vendor, "samsung"), strings.Contains(vendor, "xiaomi"):
					d.Type = "Phone/Tablet"
					return
				case strings.Contains(vendor, "tp-link"), strings.Contains(vendor, "netgear"), strings.Contains(vendor, "cisco"):
					d.Type = "Router"
					return
				case strings.Contains(vendor, "hp"), strings.Contains(vendor, "canon"), strings.Contains(vendor, "epson"):
					d.Type = "Printer"
					return
				default:
					// Vendor unknown, fallback to hostname
				}
			}
		}
	}

	// Step 2: Hostname heuristic fallback
	h := strings.ToLower(d.Hostname)
	h = strings.ReplaceAll(h, "-", "")
	h = strings.ReplaceAll(h, "_", "")

	switch {
	case strings.Contains(h, "desktop"), strings.Contains(h, "laptop"):
		d.Type = "PC"
	case strings.Contains(h, "iphone"), strings.Contains(h, "ipad"), strings.Contains(h, "galaxy"):
		d.Type = "Phone/Tablet"
	case strings.Contains(h, "router"), strings.Contains(h, "gateway"):
		d.Type = "Router"
	case strings.Contains(h, "printer"):
		d.Type = "Printer"
	default:
		d.Type = "Unknown"
	}
}
