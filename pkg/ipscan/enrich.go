package ipscan

import "runtime"

func EnrichDevices(devices []Device) []Device {
	for i := range devices {
		ip := devices[i].IP

		// Try DNS first
		if name := resolveHostname(ip); name != "" {
			devices[i].Hostname = name
			continue
		}

		// Windows NetBIOS fallback
		if runtime.GOOS == "windows" {
			if name := netbiosName(ip); name != "" {
				devices[i].Hostname = name
			}
		}
	}
	return devices
}
