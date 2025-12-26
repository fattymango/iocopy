package ipscan

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type Device struct {
	IP       string
	MAC      string
	Hostname string
	Type     string // New field for device type
}

func IncIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

func PingSubnet(scanner Scanner, subnet string) {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		log.Printf("[PingSubnet] Error parsing subnet %s: %v", subnet, err)
		return
	}

	log.Printf("[PingSubnet] Starting to ping subnet: %s", subnet)
	
	var wg sync.WaitGroup
	var totalIPs int64
	var completedIPs int64

	// Count total IPs first
	for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); IncIP(currentIP) {
		totalIPs++
	}

	log.Printf("[PingSubnet] Total IPs to ping: %d", totalIPs)

	// Reset for actual pinging
	ip, ipnet, _ = net.ParseCIDR(subnet)
	startTime := time.Now()
	
	// Start a progress logger in a separate goroutine
	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()
	
	go func() {
		for range progressTicker.C {
			completed := atomic.LoadInt64(&completedIPs)
			if completed > 0 && completed < totalIPs {
				progress := float64(completed) / float64(totalIPs) * 100
				elapsed := time.Since(startTime)
				log.Printf("[PingSubnet] Progress: %d/%d (%.1f%%) - Elapsed: %v", 
					completed, totalIPs, progress, elapsed.Round(time.Second))
			}
		}
	}()

	for currentIP := ip.Mask(ipnet.Mask); ipnet.Contains(currentIP); IncIP(currentIP) {
		wg.Add(1)
		ipStr := currentIP.String()
		go func(ip string) {
			defer wg.Done()
			err := scanner.Ping(ip)
			completed := atomic.AddInt64(&completedIPs, 1)
			
			// Log every 25 completions for more frequent updates
			if completed%25 == 0 {
				progress := float64(completed) / float64(totalIPs) * 100
				elapsed := time.Since(startTime)
				log.Printf("[PingSubnet] Progress: %d/%d (%.1f%%) - Elapsed: %v", 
					completed, totalIPs, progress, elapsed.Round(time.Second))
			}
			
			if err != nil && completed%50 == 0 {
				// Only log errors occasionally to avoid spam
				log.Printf("[PingSubnet] Ping error for %s: %v (showing every 50th error)", ip, err)
			}
		}(ipStr)
	}
	
	log.Printf("[PingSubnet] Waiting for all ping operations to complete...")
	wg.Wait()
	elapsed := time.Since(startTime)
	log.Printf("[PingSubnet] Completed pinging all %d IPs in %v", totalIPs, elapsed.Round(time.Second))
}
