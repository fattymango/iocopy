package ipscan

import "strings"

// DetectDevicesType detects the type of each device using the OUI vendor list and hostname heuristics
func DetectDevicesType(d []Device) []Device {
	ouiDB, err := loadOUITable()
	if err != nil {
		// Could not load OUI table, fallback to hostname only
		for i := range d {
			detectDeviceType(&d[i], nil)
		}
		return d
	}

	for i := range d {
		detectDeviceType(&d[i], ouiDB)
	}
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
