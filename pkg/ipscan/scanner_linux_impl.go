package ipscan

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

type LinuxScanner struct {
}

func NewLinuxScanner() *LinuxScanner {
	return &LinuxScanner{}
}

func (l *LinuxScanner) GetLocalSubnet() string {
	return l.getLocalSubnet()
}

func (l *LinuxScanner) Scan(subnet string) ([]Device, error) {
	PingSubnet(l, subnet)
	return l.readARP(), nil
}

func (l *LinuxScanner) Ping(ip string) error {
	return exec.Command("ping", "-c", "1", "-W", "1", ip).Run()
}

func (l *LinuxScanner) getLocalSubnet() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipnet := addr.(*net.IPNet)
			ip := ipnet.IP.To4()
			if ip == nil {
				continue
			}
			network := ip.Mask(ipnet.Mask)
			ones, _ := ipnet.Mask.Size()
			return fmt.Sprintf("%s/%d", network, ones)
		}
	}
	return ""
}

func (l *LinuxScanner) readARP() []Device {
	var devices []Device

	file, _ := os.Open("/proc/net/arp")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 4 {
			devices = append(devices, Device{
				IP:  fields[0],
				MAC: fields[3],
			})
		}
	}
	return devices
}

func linuxNetbiosName(ip string) string {
	// NetBIOS does not exist on Linux
	return ""
}
