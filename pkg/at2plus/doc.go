// Package at2plus provides a client for communicating with
// AirTouch 2+ air conditioning controllers over TCP/IP.
//
// # Basic Usage
//
//	ctx := context.Background()
//	client, err := at2plus.NewClient(ctx, "192.168.1.50")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	status, err := client.GetGroupStatus(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Configuration
//
// The client can be configured using functional options:
//
//	client, err := at2plus.NewClient(ctx, "192.168.1.50",
//	    at2plus.WithConnectTimeout(10*time.Second),
//	    at2plus.WithRequestTimeout(5*time.Second),
//	    at2plus.WithLogger(slog.Default()),
//	)
//
// # Protocol
//
// This package implements the AirTouch 2+ Communication Protocol v1.1.
// The protocol uses TCP port 9200 by default and does not support TLS.
// For security, isolate AirTouch devices on a dedicated VLAN.
package at2plus
