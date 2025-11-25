package at2plus

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test cases from the Spec

func TestChecksum_SpecExample_TurnOffGroup2(t *testing.T) {
	// Example from Spec Page 5: Turn off the second group
	// Packet: 0x55 0x55 0x80 0xB0 0x01 0xC0 0x00 0x0C 0x20 0x00 0x00 0x00 0x00 0x01 0x00 0x04 0x01 0x02 0x00 0x00
	// CRC: 0x64 0xFD

	// Data to CRC (Address -> Data end):
	// 80 B0 01 C0 00 0C 20 00 00 00 00 01 00 04 01 02 00 00
	input, _ := hex.DecodeString("80B001C0000C200000000001000401020000")
	expectedCRC := uint16(0x64FD)

	crc := Checksum(input)
	// Note: If Checksum returns 0xFD64, then we need to handle endianness.
	// Let's see what happens.
	assert.Equal(t, expectedCRC, crc, "CRC should match spec example")
}

func TestEncode_SpecExample_TurnOffGroup2(t *testing.T) {
	// Example from Spec Page 5
	// 0x55 0x55 0x80 0xB0 0x01 0xC0 0x00 0x0C 0x20 0x00 0x00 0x00 0x00 0x01 0x00 0x04 0x01 0x02 0x00 0x00 0x64 0xFD

	data, _ := hex.DecodeString("200000000001000401020000")
	p := NewPacket(AddressSendStandard, 0x01, MsgTypeControlStatus, data)

	encoded := p.Encode()
	expectedHex := "555580b001c0000c20000000000100040102000064fd"

	assert.Equal(t, expectedHex, hex.EncodeToString(encoded))
}

func TestDecode_SpecExample_GroupStatusResponse(t *testing.T) {
	// Example from Spec Page 6: AirTouch 2+ response with data for 2 groups
	// 55 55 B0 80 01 C0 00 18 21 00 00 00 00 02 00 08 00 00 00 00 00 00 80 00 41 32 00 00 00 00 02 00 83 2F

	rawBytes, _ := hex.DecodeString("5555B08001C00018210000000002000800000000000080004132000000000200832F")

	p, err := Decode(rawBytes)
	assert.NoError(t, err)
	assert.Equal(t, uint16(0xB080), p.Address)
	assert.Equal(t, uint8(0x01), p.MsgID)
	assert.Equal(t, uint8(0xC0), p.MsgType)
	assert.Equal(t, uint16(0x18), p.DataLen)
	assert.Equal(t, uint16(0x832F), p.CRC)
}
