package ipscan

import (
	"net"
	"sort"
	"sync"
	"time"
)

// Ping or TCP check
func isReachable(ip string) bool {
	conn, err := net.DialTimeout("tcp", ip+":80", 500*time.Millisecond)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}

func FilterDevices(devices []Device) []Device {
	// Step 1: group by MAC
	deviceMap := make(map[string][]Device)
	for _, d := range devices {
		if d.MAC == "" ||
			d.MAC == "00:00:00:00:00:00" ||
			d.MAC == "00-00-00-00-00-00" {
			continue
		}
		deviceMap[d.MAC] = append(deviceMap[d.MAC], d)
	}

	// Step 2: collect unique IPs
	ipSet := make(map[string]struct{})
	for _, devs := range deviceMap {
		for _, d := range devs {
			ipSet[d.IP] = struct{}{}
		}
	}

	// Step 3: check reachability concurrently
	reachable := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for ip := range ipSet {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			r := isReachable(ip)
			mu.Lock()
			reachable[ip] = r
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	// Step 4: pick best device per MAC
	var filtered []Device

	for _, devList := range deviceMap {
		sort.Slice(devList, func(i, j int) bool {
			// Prefer hostname
			if devList[i].Hostname != "" && devList[j].Hostname == "" {
				return true
			}
			if devList[i].Hostname == "" && devList[j].Hostname != "" {
				return false
			}

			// Prefer reachable
			ri := reachable[devList[i].IP]
			rj := reachable[devList[j].IP]
			if ri && !rj {
				return true
			}
			if !ri && rj {
				return false
			}

			// Fallback: lowest IP
			return devList[i].IP < devList[j].IP
		})

		filtered = append(filtered, devList[0])
	}

	return filtered
}
