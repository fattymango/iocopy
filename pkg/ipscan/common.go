package ipscan

import (
	"net"
	"sync"
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
	ip, ipnet, _ := net.ParseCIDR(subnet)
	var wg sync.WaitGroup

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); IncIP(ip) {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			scanner.Ping(ip)
		}(ip.String())
	}
	wg.Wait()
}
