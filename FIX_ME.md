# Code Review Findings

This document summarizes issues identified during code review of the go-at2plus repository.

**Status: All issues have been fixed.**

## Fixed Issues

### 1. Ignored Errors in CLI Argument Parsing

**Status:** Fixed in `afd96ef`

**Location:** `cmd/at2plus/commands.go`

**Issue:** `strconv.Atoi` errors were silently discarded using `_`.

**Fix:** Added error handling that exits with a meaningful message when invalid input is provided.

---

### 2. Discovery Timeout Break Does Not Exit Loop

**Status:** Fixed in `cf1e8f7`

**Location:** `pkg/at2plus/discovery.go`

**Issue:** The `break` statement inside a `select` only exited the `select`, not the enclosing `for` loop.

**Fix:** Used a labeled break (`break collectResults`) to properly exit the loop on timeout.

---

### 3. Incomplete TCP Read

**Status:** Fixed in `be21188`

**Location:** `pkg/at2plus/client.go`

**Issue:** `net.Conn.Read` does not guarantee reading the requested number of bytes.

**Fix:** Replaced `conn.Read` with `io.ReadFull` to ensure all requested bytes are read before processing.

---

### 4. No Bounds Check on Data Length

**Status:** Fixed in `e77cb48`

**Location:** `pkg/at2plus/client.go` and `pkg/at2plus/protocol.go`

**Issue:** The `dataLen` value was read directly from network data and used to allocate memory without validation.

**Fix:** Added `MaxDataLen` constant (1024 bytes) and validation before memory allocation. Packets exceeding this limit are skipped.

---

### 5. Debug Output to Stdout

**Status:** Fixed in `ca7d9b7`

**Location:** `pkg/at2plus/client.go`

**Issue:** Library code wrote directly to stdout with `fmt.Printf`.

**Fix:** Removed the debug print statement. The error is already handled by skipping the packet.

---

### 6. No Input Validation in CLI for Group/AC Numbers

**Status:** Fixed in `f59ffc1`

**Location:** `cmd/at2plus/commands.go`

**Issue:** Group numbers (0-15) and AC numbers (0-7) were not validated before sending to the device.

**Fix:** Added range validation that exits with an error message if values are out of bounds.

---

## Informational Notes

### No TLS/Encryption

**Status:** Not a bug

**Context:** Research confirms the AirTouch 2+ protocol uses plaintext TCP on port 9200. The device does not support TLS. This is typical for embedded IoT devices designed for local network use.

**Source:** Verified via [HendX/AirTouch2Plus](https://github.com/HendX/AirTouch2Plus) reference implementation which uses `NWParameters.tcp` without TLS configuration.

**Recommendation:** Document this limitation for users. If security is a concern, users should isolate AirTouch devices on a separate VLAN.

---

### Missing Test Coverage

The following components lack unit tests:

| Component | File |
|-----------|------|
| Client | `pkg/at2plus/client.go` |
| Discovery | `pkg/at2plus/discovery.go` |
| MarshalACControl | `pkg/at2plus/messages.go` |
| CLI commands | `cmd/at2plus/commands.go` |

**Recommendation:** Add tests, particularly for:
- Client connection handling and error paths
- Marshal/Unmarshal round-trip validation
- CLI input validation

---

## Summary Table

| # | Issue | Status | Commit |
|---|-------|--------|--------|
| 1 | Ignored `strconv.Atoi` errors | Fixed | `afd96ef` |
| 2 | `break` doesn't exit `for` loop | Fixed | `cf1e8f7` |
| 3 | Incomplete TCP read | Fixed | `be21188` |
| 4 | No `dataLen` bounds check | Fixed | `e77cb48` |
| 5 | Debug output to stdout | Fixed | `ca7d9b7` |
| 6 | No CLI input range validation | Fixed | `f59ffc1` |
