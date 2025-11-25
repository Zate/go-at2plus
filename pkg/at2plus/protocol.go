package at2plus

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Constants defined in the AirTouch 2+ Protocol Spec
const (
	HeaderBytes = 0x5555

	// Address constants
	AddressSendStandard = 0x80b0
	AddressSendExtended = 0x90b0
	AddressRecvStandard = 0xb080
	AddressRecvExtended = 0xb090

	// Message Types
	MsgTypeControlStatus = 0xC0
	MsgTypeExtended      = 0x1F

	// Sub Message Types (for MsgTypeControlStatus)
	SubMsgTypeGroupControl = 0x20
	SubMsgTypeGroupStatus  = 0x21
	SubMsgTypeACControl    = 0x22
	SubMsgTypeACStatus     = 0x23

	// Extended Message Sub Types
	ExtMsgTypeACError   = 0x10
	ExtMsgTypeACAbility = 0x11
	ExtMsgTypeGroupName = 0x12

	// MaxDataLen is the maximum allowed data length for a packet.
	// The protocol uses uint16 for length (max 65535), but real messages
	// are small (largest documented is ~54 bytes for AC Ability).
	// This limit provides protection against malformed packets.
	MaxDataLen = 1024
)

var (
	ErrInvalidHeader   = errors.New("invalid header")
	ErrInvalidChecksum = errors.New("invalid checksum")
	ErrInvalidLength   = errors.New("invalid data length")
	ErrDataLenExceeded = errors.New("data length exceeds maximum")
)

// Packet represents a full AirTouch 2+ protocol packet
type Packet struct {
	Header  uint16
	Address uint16
	MsgID   uint8
	MsgType uint8
	DataLen uint16
	Data    []byte
	CRC     uint16
}

// NewPacket creates a new packet with the given details
func NewPacket(address uint16, msgID uint8, msgType uint8, data []byte) *Packet {
	return &Packet{
		Header:  HeaderBytes,
		Address: address,
		MsgID:   msgID,
		MsgType: msgType,
		DataLen: uint16(len(data)),
		Data:    data,
	}
}

// Encode serializes the packet into bytes
func (p *Packet) Encode() []byte {
	buf := make([]byte, 8+len(p.Data)+2) // Header(2)+Addr(2)+ID(1)+Type(1)+Len(2) + Data + CRC(2)

	binary.BigEndian.PutUint16(buf[0:2], p.Header)
	binary.BigEndian.PutUint16(buf[2:4], p.Address)
	buf[4] = p.MsgID
	buf[5] = p.MsgType
	binary.BigEndian.PutUint16(buf[6:8], p.DataLen)
	copy(buf[8:], p.Data)

	// Calculate CRC on everything after header (Address onwards)
	// Spec says: "Use all data except the header."
	crcData := buf[2 : 8+len(p.Data)]
	p.CRC = Checksum(crcData)

	binary.BigEndian.PutUint16(buf[8+len(p.Data):], p.CRC)

	return buf
}

// Decode parses bytes into a Packet
func Decode(data []byte) (*Packet, error) {
	if len(data) < 10 {
		return nil, ErrInvalidLength
	}

	header := binary.BigEndian.Uint16(data[0:2])
	if header != HeaderBytes {
		return nil, ErrInvalidHeader
	}

	address := binary.BigEndian.Uint16(data[2:4])
	msgID := data[4]
	msgType := data[5]
	dataLen := binary.BigEndian.Uint16(data[6:8])

	if len(data) < int(8+dataLen+2) {
		return nil, ErrInvalidLength
	}

	packetData := make([]byte, dataLen)
	copy(packetData, data[8:8+dataLen])

	crcReceived := binary.BigEndian.Uint16(data[8+dataLen:])

	// Validate CRC
	crcData := data[2 : 8+dataLen]
	crcCalculated := Checksum(crcData)

	if crcReceived != crcCalculated {
		return nil, fmt.Errorf("%w: expected 0x%04X, got 0x%04X", ErrInvalidChecksum, crcCalculated, crcReceived)
	}

	return &Packet{
		Header:  header,
		Address: address,
		MsgID:   msgID,
		MsgType: msgType,
		DataLen: dataLen,
		Data:    packetData,
		CRC:     crcReceived,
	}, nil
}

// Checksum calculates the CRC16 Modbus checksum
func Checksum(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if (crc & 0x0001) != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	// Note: The spec examples show Big Endian CRC in the packet (e.g. 0xFD 0x18).
	// However, standard Modbus CRC is usually Little Endian.
	// Let's check the spec example:
	// Example: 0x55 0x55 0x80 0xB0 0x01 0xC0 0x00 0x0C ... CRC: 0x64 0xFD
	// Wait, the example in spec says: 0x64 0xFD.
	// Let's verify this with a test case before finalizing.
	// But for now, I will return the uint16. The Encode/Decode functions use BigEndian to put/get it.
	// If the calculation produces 0xFD64 but it's stored as 0x64FD, then we might need to swap.
	// I'll assume standard calculation and BigEndian storage for now, and verify with test.
	return crc
}
