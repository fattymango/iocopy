package ipscan

import (
	"net"
	"sort"
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
		if d.MAC == "" || d.MAC == "00:00:00:00:00:00" || d.MAC == "00-00-00-00-00-00" {
			continue // ignore invalid MAC
		}
		deviceMap[d.MAC] = append(deviceMap[d.MAC], d)
	}

	// Step 2: pick best IP per MAC
	var filtered []Device
	for _, devList := range deviceMap {
		// Sort by: hostname first, then reachable, then lowest IP
		sort.Slice(devList, func(i, j int) bool {
			if devList[i].Hostname != "" && devList[j].Hostname == "" {
				return true
			}
			if devList[i].Hostname == "" && devList[j].Hostname != "" {
				return false
			}

			// check reachability
			ri, rj := isReachable(devList[i].IP), isReachable(devList[j].IP)
			if ri && !rj {
				return true
			}
			if !ri && rj {
				return false
			}

			// fallback: lowest IP
			return devList[i].IP < devList[j].IP
		})

		// first in the sorted list is “best”
		filtered = append(filtered, devList[0])
	}

	return filtered
}
