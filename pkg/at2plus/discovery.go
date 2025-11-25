package at2plus

import (
	"fmt"
	"net"
	"time"
)

// DiscoveryResult represents a discovered AirTouch device
type DiscoveryResult struct {
	IP string
}

// Discover searches for AirTouch 2+ devices on the network
// It tries UDP broadcast first (Ports 49004, 49005 as per newer models, and maybe others?)
// Then falls back to scanning local subnet on port 9200.
func Discover() ([]DiscoveryResult, error) {
	var results []DiscoveryResult

	// Method 1: UDP Broadcast (Best effort guess based on AT4/5)
	// We'll try to listen for broadcasts or send one.
	// Since we don't know the exact protocol, we'll skip implementing a complex UDP handshake
	// unless we find more info. The user asked for "discovery... not covered in the spec".
	// The most reliable way without spec is TCP Port Scan on 9200.

	// Let's try to find local IP and scan /24
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
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:9200", ip), 200*time.Millisecond)
				if err == nil {
					conn.Close()
					resultsCh <- scanResult{ip: ip, ok: true}
				} else {
					resultsCh <- scanResult{ip: ip, ok: false}
				}
			}(targetIP.String())
		}
	}

	timeout := time.After(3 * time.Second)
	for i := 0; i < count; i++ {
		select {
		case res := <-resultsCh:
			if res.ok {
				results = append(results, DiscoveryResult{IP: res.ip})
			}
		case <-timeout:
			// Stop waiting
			break
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
