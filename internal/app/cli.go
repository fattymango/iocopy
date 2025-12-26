package app

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func promptUserSelection(reachableIPs []string, scanner *bufio.Scanner) string {
	fmt.Println("\n=== Reachable Peers ===")
	for i, ip := range reachableIPs {
		fmt.Printf("%d. %s\n", i+1, ip)
	}
	fmt.Print("\nSelect peer number (or 'r' to search again, 'q' to quit): ")

	if !scanner.Scan() {
		return ""
	}

	input := strings.TrimSpace(scanner.Text())

	if input == "q" || input == "Q" {
		os.Exit(0)
	}

	if input == "r" || input == "R" {
		return ""
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(reachableIPs) {
		fmt.Printf("Invalid selection. Please choose 1-%d, 'r' to search again, or 'q' to quit\n", len(reachableIPs))
		return ""
	}

	selectedIP := reachableIPs[choice-1]
	fmt.Printf("Selected: %s\n\n", selectedIP)
	return selectedIP
}
