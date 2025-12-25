package main

import (
	"copy/pkg/ipscan"
	"fmt"
	"os"
	"runtime"
)

func main() {
	var scanner ipscan.Scanner
	if runtime.GOOS == "linux" {
		scanner = ipscan.NewLinuxScanner()
	} else if runtime.GOOS == "windows" {
		scanner = ipscan.NewWindowsScanner()
	} else {
		fmt.Println("Unsupported OS")
		os.Exit(1)
	}
	subnet := scanner.GetLocalSubnet()
	fmt.Println("Scanning:", subnet)

	// pingSubnet(scanner, subnet)

	devices, err := scanner.Scan(subnet)
	if err != nil {
		fmt.Println("Error scanning subnet:", err)
		return
	}
	devices = ipscan.EnrichDevices(devices)
	devices = ipscan.FilterDevices(devices)
	devices = ipscan.DetectDevicesType(devices)
	fmt.Printf("\n%-15s %-25s %-20s %-15s\n", "IP", "HOSTNAME", "MAC", "TYPE")
	for _, d := range devices {
		fmt.Printf("%-15s %-25s %-20s %-15s\n", d.IP, d.Hostname, d.MAC, d.Type)
	}

}
