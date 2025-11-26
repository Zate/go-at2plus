package at2plus

import (
	"context"
	"net"
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
		return nil, err
	}

	type scanResult struct {
		ip string
		ok bool
	}

	resultsCh := make(chan scanResult)
	count := 0

	for _, ip := range ips {
		// Assume /24 subnet
		// e.g. 192.168.1.x
		baseIP := ip.Mask(net.CIDRMask(24, 32))
		baseIP[3] = 0

		for i := 1; i < 255; i++ {
			targetIP := net.IP{baseIP[0], baseIP[1], baseIP[2], byte(i)}
			count++
			go func(ip string) {
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

collectResults:
	for i := 0; i < count; i++ {
		select {
		case res := <-resultsCh:
			if res.ok {
				results = append(results, DiscoveryResult{IP: res.ip})
			}
		case <-ctx.Done():
			// Context canceled or timed out
			break collectResults
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
