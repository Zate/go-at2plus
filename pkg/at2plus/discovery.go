package at2plus

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// DiscoveryResult represents a discovered AirTouch device
type DiscoveryResult struct {
	IP string
}

// Discover searches for AirTouch 2+ devices on the network.
// It scans the local subnet on port 9200.
// The context controls the overall discovery timeout.
// If the context has no deadline, a 3-second timeout is applied.
func Discover(ctx context.Context) ([]DiscoveryResult, error) {
	var results []DiscoveryResult

	// Apply default timeout if context has no deadline
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
	}

	// Find local IP and scan /24
	ips, err := getLocalIPs()
	if err != nil {
		return nil, fmt.Errorf("get local IPs: %w", err)
	}

	type scanResult struct {
		ip string
		ok bool
	}

	// Count total IPs to scan
	count := 0
	for range ips {
		count += 254 // 1-254 for each /24 subnet
	}

	// Use buffered channel to prevent goroutine leaks
	resultsCh := make(chan scanResult, count)
	var wg sync.WaitGroup

	for _, ip := range ips {
		// Assume /24 subnet
		// e.g. 192.168.1.x
		baseIP := ip.Mask(net.CIDRMask(24, 32))
		baseIP[3] = 0

		for i := 1; i < 255; i++ {
			targetIP := net.IP{baseIP[0], baseIP[1], baseIP[2], byte(i)}
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				var d net.Dialer
				dialCtx, dialCancel := context.WithTimeout(ctx, 200*time.Millisecond)
				defer dialCancel()
				conn, err := d.DialContext(dialCtx, "tcp", net.JoinHostPort(ip, "9200"))
				if err == nil {
					conn.Close()
					resultsCh <- scanResult{ip: ip, ok: true}
				} else {
					resultsCh <- scanResult{ip: ip, ok: false}
				}
			}(targetIP.String())
		}
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results until channel is closed or context is done
	for res := range resultsCh {
		if res.ok {
			results = append(results, DiscoveryResult{IP: res.ip})
		}
		// Check context between results
		select {
		case <-ctx.Done():
			return results, nil
		default:
		}
	}

	return results, nil
}

func getLocalIPs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}
