package ipscan

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type WindowsScanner struct {
}

func NewWindowsScanner() *WindowsScanner {
	return &WindowsScanner{}
}

func (s *WindowsScanner) GetLocalSubnet() string {
	return getLocalSubnet()
}

func (s *WindowsScanner) Scan(subnet string) ([]Device, error) {
	PingSubnet(s, subnet)
	return readARP(), nil
}

func (s *WindowsScanner) Ping(ip string) error {
	return exec.Command("ping", "-n", "1", "-w", "300", ip).Run()
}

func readARP() []Device {
	var devices []Device

	out, _ := exec.Command("arp", "-a").Output()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && net.ParseIP(fields[0]) != nil {
			devices = append(devices, Device{
				IP:  fields[0],
				MAC: fields[1],
			})
		}
	}
	return devices
}

func getLocalSubnet() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, iface := range ifaces {
		// Must be up and not loopback
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip virtual / tunnel interfaces (common on Windows)
		if strings.Contains(strings.ToLower(iface.Name), "virtual") ||
			strings.Contains(strings.ToLower(iface.Name), "vmware") ||
			strings.Contains(strings.ToLower(iface.Name), "loopback") ||
			strings.Contains(strings.ToLower(iface.Name), "tunnel") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4()
			if ip == nil {
				continue
			}

			network := ip.Mask(ipNet.Mask)
			ones, _ := ipNet.Mask.Size()

			return fmt.Sprintf("%s/%d", network.String(), ones)
		}
	}
	return ""
}

func netbiosName(ip string) string {
	out, err := exec.Command("nbtstat", "-A", ip).Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "<00>") && strings.Contains(line, "UNIQUE") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				return fields[0]
			}
		}
	}
	return ""
}
