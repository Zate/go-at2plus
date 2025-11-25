# go-at2plus

A professional, industry-standard Go package for controlling **AirTouch 2+** air conditioner controllers.

[![Go Report Card](https://goreportcard.com/badge/github.com/zberg/go-at2plus)](https://goreportcard.com/report/github.com/zberg/go-at2plus)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Features

- **Full Protocol Support**: Implements the complete AirTouch 2+ V1.1 specification.
- **Discovery**: Automatically finds AirTouch 2+ units on your local network.
- **Control & Status**:
  - Turn Groups On/Off, set percentages.
  - Turn ACs On/Off, set Mode (Cool, Heat, etc), Fan Speed, and Temperature.
  - Query real-time status of all units.
- **Extended Info**: Fetch Group names and AC capabilities.
- **CLI Tool**: Includes a powerful command-line interface for testing and automation.

## Installation

```bash
go get github.com/zberg/go-at2plus
```

To install the CLI tool:

```bash
go install github.com/zberg/go-at2plus/cmd/at2plus@latest
```

## Usage

### Library

```go
package main

import (
    "fmt"
    "github.com/zberg/go-at2plus/pkg/at2plus"
)

func main() {
    // Connect to the unit
    client, err := at2plus.NewClient("192.168.1.50")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    // Get Group Status
    groups, err := client.GetGroupStatus()
    if err != nil {
        panic(err)
    }

    for _, g := range groups {
        fmt.Printf("Group %d: Power=%d, Open=%d%%\n", g.GroupNumber, g.Power, g.Percent)
    }

    // Turn off Group 1
    off := 1 // 1=Off
    err = client.SetGroupControl([]at2plus.GroupControl{
        {
            GroupNumber: 1,
            Power:       &off,
        },
    })
}
```

### CLI

```bash
# Discover devices
at2plus discover

# Check status
at2plus status --ip 192.168.1.50

# Turn on Group 0 and set to 80%
at2plus control-group 0 --power on --percent 80 --ip 192.168.1.50

# Set AC 0 to Cool mode, 24 degrees
at2plus control-ac 0 --mode cool --temp 24 --ip 192.168.1.50
```

## Documentation

See [PROTOCOL.md](PROTOCOL.md) for details on the communication protocol.

## License

MIT
