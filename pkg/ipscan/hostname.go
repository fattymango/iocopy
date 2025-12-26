package ipscan

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

func resolveHostname(ip string) string {
	return resolveHostnameWithTimeout(ip, 5*time.Second)
}

func resolveHostnameWithTimeout(ip string, timeout time.Duration) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Use a channel to get the result
	type result struct {
		names []string
		err   error
	}
	resultChan := make(chan result, 1)
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Catch any panics from DNS lookup
				resultChan <- result{names: nil, err: fmt.Errorf("panic: %v", r)}
			}
		}()
		names, err := net.LookupAddr(ip)
		resultChan <- result{names: names, err: err}
	}()
	
	select {
	case res := <-resultChan:
		if res.err != nil || len(res.names) == 0 {
			return ""
		}
		return strings.TrimSuffix(res.names[0], ".")
	case <-ctx.Done():
		// Timeout occurred
		return ""
	}
}
