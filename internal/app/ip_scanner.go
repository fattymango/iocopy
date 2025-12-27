package app

import (
	"copy/internal/shared"
	"copy/pkg/ipscan"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

func (a *App) FindReachableIPs(port string) []string {
	localIP := shared.GetLocalIP()
	log.Printf("[scan] Scanning for reachable peers...")

	devices := scanAndFilterDevices()
	var reachableIPs []string

	log.Printf("[scan] Found %d devices, checking reachability...", len(devices))

	wg := sync.WaitGroup{}
	for _, device := range devices {
		if device.IP == localIP {
			continue
		}
		wg.Add(1)
		go func(device ipscan.Device) {
			defer wg.Done()
			if isReachable(device.IP, port) {
				reachableIPs = append(reachableIPs, device.IP)
				log.Printf("[scan] Found reachable peer: %s (%s)", device.IP, device.Hostname)
			}
		}(device)
	}

	wg.Wait()
	log.Printf("[scan] All reachability checks completed")

	return reachableIPs
}

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
	log.Printf("[scan] Filtered %d devices, detecting types...", len(devices))
	devices = ipscan.DetectDevicesType(devices)

	return devices
}

func isReachable(ip, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
