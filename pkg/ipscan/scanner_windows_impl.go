package ipscan

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

type WindowsScanner struct {
}

func NewWindowsScanner() *WindowsScanner {
	return &WindowsScanner{}
}

func (w *WindowsScanner) GetLocalSubnet() string {
	log.Printf("[WindowsScanner.GetLocalSubnet] Detecting local subnet...")
	subnet := w.getLocalSubnet()
	if subnet != "" {
		log.Printf("[WindowsScanner.GetLocalSubnet] Detected subnet: %s", subnet)
	} else {
		log.Printf("[WindowsScanner.GetLocalSubnet] Warning: Could not detect local subnet")
	}
	return subnet
}

func (w *WindowsScanner) Scan(subnet string) ([]Device, error) {
	log.Printf("[WindowsScanner.Scan] Starting scan of subnet: %s", subnet)
	startTime := time.Now()

	log.Printf("[WindowsScanner.Scan] Phase 1: Pinging subnet...")
	PingSubnet(w, subnet)

	log.Printf("[WindowsScanner.Scan] Phase 2: Reading ARP table...")
	devices := w.readARP()

	elapsed := time.Since(startTime)
	log.Printf("[WindowsScanner.Scan] Scan completed in %v, found %d devices in ARP table",
		elapsed.Round(time.Second), len(devices))

	return devices, nil
}

func (w *WindowsScanner) Ping(ip string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ping", "-n", "1", "-w", "300", ip)

	// Capture output for debugging (but don't print to avoid spam)
	err := cmd.Run()

	// Check if context was cancelled (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		// Timeout occurred - this is expected for unreachable hosts
		return fmt.Errorf("ping timeout for %s", ip)
	}

	return err
}

func (w *WindowsScanner) readARP() []Device {
	log.Printf("[WindowsScanner.readARP] Executing 'arp -a' command...")
	var devices []Device

	startTime := time.Now()
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		log.Printf("[WindowsScanner.readARP] Error executing arp command: %v", err)
		return devices
	}

	arpOutputTime := time.Since(startTime)
	log.Printf("[WindowsScanner.readARP] ARP command completed in %v, output length: %d bytes",
		arpOutputTime, len(out))

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	lineCount := 0
	parsedCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) >= 2 {
			ip := net.ParseIP(fields[0])
			if ip != nil {
				parsedCount++
				devices = append(devices, Device{
					IP:  fields[0],
					MAC: fields[1],
				})
				log.Printf("[WindowsScanner.readARP] Found device: IP=%s, MAC=%s", fields[0], fields[1])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[WindowsScanner.readARP] Error scanning ARP output: %v", err)
	}

	log.Printf("[WindowsScanner.readARP] Parsed %d devices from %d ARP table lines", parsedCount, lineCount)
	return devices
}

func (w *WindowsScanner) getLocalSubnet() string {
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

func windowsNetbiosName(ip string) string {
	return windowsNetbiosNameWithTimeout(ip, 5*time.Second)
}

func windowsNetbiosNameWithTimeout(ip string, timeout time.Duration) string {
	log.Printf("[windowsNetbiosName] Starting nbtstat for %s with timeout %v", ip, timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	startTime := time.Now()
	cmd := exec.CommandContext(ctx, "nbtstat", "-A", ip)

	// Use CombinedOutput to capture both stdout and stderr
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(startTime)

	if err != nil {
		// Check if context was cancelled (timeout)
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[windowsNetbiosName] nbtstat for %s timed out after %v", ip, elapsed)
			return ""
		}
		log.Printf("[windowsNetbiosName] nbtstat for %s failed: %v (took %v)", ip, err, elapsed)
		return ""
	}

	log.Printf("[windowsNetbiosName] nbtstat for %s completed in %v, output length: %d bytes",
		ip, elapsed, len(out))

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "<00>") && strings.Contains(line, "UNIQUE") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				name := fields[0]
				log.Printf("[windowsNetbiosName] Found NetBIOS name for %s: %s", ip, name)
				return name
			}
		}
	}

	log.Printf("[windowsNetbiosName] No NetBIOS name found for %s", ip)
	return ""
}
