# AirTouch 2+ Protocol Documentation

This document describes the AirTouch 2+ communication protocol as implemented in this package, based on the `AirTouch2PlusCommunicationProtocolV1.1.pdf` specification.

## Overview

The AirTouch 2+ system uses a TCP/IP protocol on port **9200**.
The protocol is packet-based, with a fixed header, address fields, message ID, message type, data length, variable data, and a CRC16 Modbus checksum.

## Packet Structure

| Field | Length | Description |
|---|---|---|
| Header | 2 bytes | Fixed `0x55 0x55` |
| Address | 2 bytes | `0x80 0xB0` (Standard) or `0x90 0xB0` (Extended) |
| Message ID | 1 byte | Arbitrary ID, echoed in response |
| Message Type | 1 byte | `0xC0` (Control/Status) or `0x1F` (Extended) |
| Data Length | 2 bytes | Length of the Data field (Big Endian) |
| Data | N bytes | Variable length payload |
| CRC | 2 bytes | CRC16 Modbus of Address+ID+Type+Len+Data |

## Message Types

### Control & Status (0xC0)

Used for controlling Groups and ACs, and querying their status.

#### Group Control (Subtype 0x20)
- **Control**: Turn groups On/Off, set percentage, set Turbo.
- **Structure**: List of 4-byte group control structures.

#### Group Status (Subtype 0x21)
- **Query**: Send empty payload with subtype 0x21.
- **Response**: List of 8-byte group status structures.
- **Fields**: Power state, Open percentage, Turbo support, Spill status.

#### AC Control (Subtype 0x22)
- **Control**: Turn ACs On/Off, set Mode (Heat/Cool/etc), Fan Speed, Setpoint.
- **Structure**: List of 4-byte AC control structures.

#### AC Status (Subtype 0x23)
- **Query**: Send empty payload with subtype 0x23.
- **Response**: List of 10-byte AC status structures.
- **Fields**: Power, Mode, Fan Speed, Setpoint, Temperature, Error Code.

### Extended Messages (0x1F)

Used for querying capabilities and names.

#### AC Ability (Subtype 0x11)
- **Query**: Send `0xFF 0x11 [ACNum]`.
- **Response**: Detailed capabilities including supported modes, fan speeds, and setpoint ranges.

#### Group Name (Subtype 0x12)
- **Query**: Send `0xFF 0x12 [GroupNum]`.
- **Response**: 8-byte name string for each group.

#### AC Error (Subtype 0x10)
- **Query**: Send `0xFF 0x10 [ACNum]`.
- **Response**: Error string if any.

## Discovery

The official spec does not define discovery. This package implements:
1. **TCP Port Scan**: Scans the local subnet for devices listening on port 9200.
2. **UDP Broadcast**: (Placeholder) Listens on ports 49004/49005 (used by newer AirTouch models).

## References

- `AirTouch2PlusCommunicationProtocolV1.1.pdf`
