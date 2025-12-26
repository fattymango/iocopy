package ipscan

import (
	"log"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func EnrichDevices(devices []Device) []Device {
	log.Printf("[EnrichDevices] Starting enrichment for %d devices", len(devices))
	startTime := time.Now()

	// Use a worker pool to limit concurrent operations (max 10 at a time)
	const maxWorkers = 10
	workChan := make(chan int, len(devices))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var completed int64

	// Queue all device indices
	for i := range devices {
		workChan <- i
	}
	close(workChan)

	log.Printf("[EnrichDevices] Starting %d worker goroutines", maxWorkers)

	// Start workers
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("[EnrichDevices] Worker %d started", workerID)

			for idx := range workChan {
				ip := devices[idx].IP
				deviceNum := atomic.AddInt64(&completed, 1)

				log.Printf("[EnrichDevices] Worker %d: Processing device %d/%d: %s",
					workerID, deviceNum, len(devices), ip)

				// Skip enrichment for multicast, broadcast, and invalid IPs
				parsedIP := net.ParseIP(ip)
				if parsedIP == nil {
					log.Printf("[EnrichDevices] Worker %d: Skipping invalid IP: %s", workerID, ip)
					continue
				}

				if parsedIP.IsMulticast() || parsedIP.IsUnspecified() {
					log.Printf("[EnrichDevices] Worker %d: Skipping multicast/unspecified IP: %s", workerID, ip)
					continue
				}

				// Try DNS first with timeout
				log.Printf("[EnrichDevices] Worker %d: Attempting DNS lookup for %s", workerID, ip)
				dnsStart := time.Now()
				if name := resolveHostnameWithTimeout(ip, 1*time.Second); name != "" {
					dnsElapsed := time.Since(dnsStart)
					mu.Lock()
					devices[idx].Hostname = name
					mu.Unlock()
					log.Printf("[EnrichDevices] Worker %d: Found DNS hostname for %s: %s (took %v)",
						workerID, ip, name, dnsElapsed)
					continue
				}
				dnsElapsed := time.Since(dnsStart)
				log.Printf("[EnrichDevices] Worker %d: DNS lookup for %s failed or timed out (took %v)",
					workerID, ip, dnsElapsed)

				// Windows NetBIOS fallback with timeout (skip for certain IPs)
				if runtime.GOOS == "windows" && shouldTryNetBIOS(parsedIP) {
					log.Printf("[EnrichDevices] Worker %d: Attempting NetBIOS lookup for %s", workerID, ip)
					nbStart := time.Now()
					if name := windowsNetbiosNameWithTimeout(ip, 1*time.Second); name != "" {
						nbElapsed := time.Since(nbStart)
						mu.Lock()
						devices[idx].Hostname = name
						mu.Unlock()
						log.Printf("[EnrichDevices] Worker %d: Found NetBIOS name for %s: %s (took %v)",
							workerID, ip, name, nbElapsed)
					} else {
						nbElapsed := time.Since(nbStart)
						log.Printf("[EnrichDevices] Worker %d: NetBIOS lookup for %s failed or timed out (took %v)",
							workerID, ip, nbElapsed)
					}
				} else if runtime.GOOS == "windows" {
					log.Printf("[EnrichDevices] Worker %d: Skipping NetBIOS for %s (not suitable)", workerID, ip)
				}

				log.Printf("[EnrichDevices] Worker %d: Completed enrichment for %s", workerID, ip)
			}

			log.Printf("[EnrichDevices] Worker %d finished", workerID)
		}(w)
	}

	log.Printf("[EnrichDevices] Waiting for all enrichment operations to complete...")
	wg.Wait()
	elapsed := time.Since(startTime)
	log.Printf("[EnrichDevices] Completed enrichment in %v", elapsed.Round(time.Second))

	return devices
}

// shouldTryNetBIOS returns true if the IP is suitable for NetBIOS lookup
func shouldTryNetBIOS(ip net.IP) bool {
	// Skip private multicast ranges, link-local, etc.
	if ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
		return false
	}

	// Only try NetBIOS for private IP ranges (192.168.x.x, 10.x.x.x, 172.16-31.x.x)
	if ip4 := ip.To4(); ip4 != nil {
		// 192.168.0.0/16
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 10.0.0.0/8
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
	}

	return false
}
