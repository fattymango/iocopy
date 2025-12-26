package app

import (
	"copy/pkg/ipscan"
	"log"
	"runtime"
)

func scanAndFilterDevices() []ipscan.Device {
	var scanner ipscan.Scanner
	if runtime.GOOS == "linux" {
		scanner = ipscan.NewLinuxScanner()
	} else if runtime.GOOS == "windows" {
		scanner = ipscan.NewWindowsScanner()
	} else {
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}

	subnet := scanner.GetLocalSubnet()
	if subnet == "" {
		log.Fatalf("Could not determine local subnet")
	}
	log.Printf("[scan] Scanning subnet: %s", subnet)

	devices, err := scanner.Scan(subnet)
	if err != nil {
		log.Fatalf("Error scanning subnet: %v", err)
	}

	log.Printf("[scan] Found %d devices, filtering...", len(devices))
	devices = ipscan.FilterDevices(devices)
	devices = ipscan.DetectDevicesType(devices)

	return devices
}

func findReachableIPs(port string) []string {
	localIP := getLocalIP()
	log.Printf("[scan] Scanning for reachable peers...")

	devices := scanAndFilterDevices()
	var reachableIPs []string

	for _, device := range devices {
		if device.IP == localIP {
			continue
		}

		if isReachable(device.IP, port) {
			reachableIPs = append(reachableIPs, device.IP)
			log.Printf("[scan] Found reachable peer: %s (%s)", device.IP, device.Hostname)
		}
	}

	return reachableIPs
}

