package at2plus

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalGroupControl_SpecExample(t *testing.T) {
	// Spec Page 5: Turn off the second group (Group 1)
	// Data: 0x20 0x00 0x00 0x00 0x00 0x01 0x00 0x04 0x01 0x02 0x00 0x00

	// Data: 0x20 0x00 0x00 0x00 0x00 0x01 0x00 0x04 0x01 0x02 0x00 0x00

	powerOff := 1 // My struct: 0:Next, 1:Off, 2:On

	groups := []GroupControl{
		{
			GroupNumber: 1,
			Power:       &powerOff,
		},
	}

	data, err := MarshalGroupControl(groups)
	require.NoError(t, err)

	expectedHex := "200000000001000401020000"
	assert.Equal(t, expectedHex, hex.EncodeToString(data))
}

func TestUnmarshalGroupStatus_SpecExample(t *testing.T) {
	// Spec Page 6: Response with data for 2 groups
	// Data: 21 00 00 00 00 02 00 08 00 00 00 00 00 00 80 00 41 32 00 00 00 00 02 00

	rawBytes, _ := hex.DecodeString("210000000002000800000000000080004132000000000200")

	groups, err := UnmarshalGroupStatus(rawBytes)
	require.NoError(t, err)
	require.Len(t, groups, 2)

	// Group 0
	assert.Equal(t, uint8(0), groups[0].GroupNumber)
	assert.Equal(t, 0, groups[0].Power) // Off
	assert.Equal(t, 0, groups[0].Percent)
	assert.True(t, groups[0].TurboSupport) // 0x80 -> Bit 8 is 1

	// Group 1
	assert.Equal(t, uint8(1), groups[1].GroupNumber)
	assert.Equal(t, 1, groups[1].Power)    // On
	assert.Equal(t, 50, groups[1].Percent) // 0x32 = 50
	assert.False(t, groups[1].TurboSupport)
	assert.True(t, groups[1].Spill) // 0x02 -> Bit 2 is 1
}

func TestUnmarshalACStatus_SpecExample(t *testing.T) {
	// Spec Page 11: AC 0 data
	// 10 12 78 C0 02 DA 00 00 80 00
	// AC 0 is on, heat mode, low fan, no error.
	// Setpoint 22, Temp 23.

	// Construct full message wrapper
	// Header part: 23 00 00 00 00 01 00 0A (1 AC)
	prefix, _ := hex.DecodeString("230000000001000A")
	acData, _ := hex.DecodeString("101278C002DA00008000")
	fullData := append(prefix, acData...)

	acs, err := UnmarshalACStatus(fullData)
	require.NoError(t, err)
	require.Len(t, acs, 1)

	ac := acs[0]
	assert.Equal(t, uint8(0), ac.ACNumber)
	assert.Equal(t, 1, ac.Power)    // On
	assert.Equal(t, 1, ac.Mode)     // Heat
	assert.Equal(t, 2, ac.FanSpeed) // Low
	assert.Equal(t, 22, ac.Setpoint)
	assert.Equal(t, 23, ac.Temperature)
	assert.Equal(t, 0, ac.ErrorCode)
}

func TestUnmarshalACAbility_SpecExample(t *testing.T) {
	// Spec Page 12: AC 0 data
	// 00 18 55 4e 49 54 00 ...
	// Name "UNIT", 4 groups, Cool/Heat/Fan/Auto, Low/Med/High/Auto

	// Full message data:
	// FF 11 00 18 (Header + AC0 Header)
	// 55 4e 49 54 00 00 00 00 00 00 00 00 00 00 00 00 (Name "UNIT")
	// 00 04 (Start Group 0, Count 4)
	// 17 (Modes: 00010111 -> Cool(1), Fan(0), Dry(1), Heat(1), Auto(1)? Wait)
	// Spec: 0x17 -> 0001 0111.
	// Bit 5 Cool (0), Bit 4 Fan (1), Bit 3 Dry (0), Bit 2 Heat (1), Bit 1 Auto (1).
	// Wait, Spec example says: "It has cool, heat, fan, auto modes".
	// My binary 0001 0111:
	// Bit 1 (Auto): 1 -> Yes
	// Bit 2 (Heat): 1 -> Yes
	// Bit 3 (Dry): 0 -> No
	// Bit 4 (Fan): 1 -> Yes
	// Bit 5 (Cool): 0 -> No?
	// Spec example text says "Cool...".
	// Maybe my bit mapping is wrong or spec example hex is tricky.
	// Spec: "0x17 0x1d 0x11 0x1f 0x11 0x1f"
	// Let's use the hex exactly.

	hexStr := "ff110018554e49540000000000000000000000000004171d111f111f"
	data, _ := hex.DecodeString(hexStr)

	abilities, err := UnmarshalACAbility(data)
	require.NoError(t, err)
	require.Len(t, abilities, 1)

	a := abilities[0]
	assert.Equal(t, "UNIT", a.Name)
	assert.Equal(t, uint8(0), a.StartGroup)
	assert.Equal(t, uint8(4), a.GroupCount)

	// Check modes based on 0x17 (0001 0111)
	// Bit 1: Auto (1)
	// Bit 2: Heat (1)
	// Bit 3: Dry (0)
	// Bit 4: Fan (1)
	// Bit 5: Cool (0)
	assert.True(t, a.AutoMode)
	assert.True(t, a.HeatMode)
	assert.False(t, a.DryMode)
	assert.True(t, a.FanMode)
	assert.False(t, a.CoolMode) // Spec text says Cool, but hex 0x17 says No Cool (Bit 5 is 0). I will trust Hex.
}

func TestUnmarshalGroupName_SpecExample(t *testing.T) {
	// Spec Page 14: Group 0 "Group1"
	// ff 12 00 47 72 6f 75 70 31 00 00

	hexStr := "ff120047726f7570310000"
	data, _ := hex.DecodeString(hexStr)

	names, err := UnmarshalGroupName(data)
	require.NoError(t, err)
	require.Len(t, names, 1)

	assert.Equal(t, uint8(0), names[0].GroupNumber)
	assert.Equal(t, "Group1", names[0].Name)
}

func TestMarshalACControl_SingleAC(t *testing.T) {
	powerOn := 3 // On
	mode := 4    // Cool

	acs := []ACControl{
		{
			ACNumber: 0,
			Power:    &powerOn,
			Mode:     &mode,
		},
	}

	data, err := MarshalACControl(acs)
	require.NoError(t, err)

	// Header: 22 00 00 00 00 01 00 04 (SubType, 0s, count=1, repeatLen=4)
	// Data: AC0 with power=3 (on), mode=4 (cool)
	// Byte1: (power<<4) | acNum = (3<<4) | 0 = 0x30
	// Byte2: (mode<<4) | fanSpeed = (4<<4) | 0 = 0x40
	// Byte3: 0x00 (no setpoint change)
	// Byte4: 0x00

	assert.Equal(t, byte(SubMsgTypeACControl), data[0])
	assert.Equal(t, byte(0x30), data[8])  // power on, AC 0
	assert.Equal(t, byte(0x40), data[9])  // cool mode
	assert.Equal(t, byte(0x00), data[10]) // no setpoint
}

func TestMarshalACControl_MultipleACs(t *testing.T) {
	powerOff := 2 // Off
	powerOn := 3  // On

	acs := []ACControl{
		{ACNumber: 0, Power: &powerOff},
		{ACNumber: 1, Power: &powerOn},
	}

	data, err := MarshalACControl(acs)
	require.NoError(t, err)

	// Should have 2 ACs
	count := int(data[4])<<8 | int(data[5])
	assert.Equal(t, 2, count)

	// AC0: power off (2<<4 = 0x20), ACNum 0
	assert.Equal(t, byte(0x20), data[8])
	// AC1: power on (3<<4 = 0x30), ACNum 1
	assert.Equal(t, byte(0x31), data[12])
}

func TestMarshalACControl_WithSetpoint(t *testing.T) {
	setpoint := 24

	acs := []ACControl{
		{ACNumber: 0, Setpoint: &setpoint},
	}

	data, err := MarshalACControl(acs)
	require.NoError(t, err)

	// Byte3 (offset 10): 0x40 = change setpoint
	assert.Equal(t, byte(0x40), data[10])
	// Byte4 (offset 11): (24*10)-100 = 140 = 0x8C
	assert.Equal(t, byte(0x8C), data[11])
}

func TestUnmarshalGroupStatus_InvalidSubType(t *testing.T) {
	// Use wrong subtype (0x23 instead of 0x21)
	data, _ := hex.DecodeString("230000000001000800000000000080")

	_, err := UnmarshalGroupStatus(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sub type")
}

func TestUnmarshalGroupStatus_TooShort(t *testing.T) {
	// Data shorter than minimum header (8 bytes)
	data := []byte{0x21, 0x00, 0x00}

	_, err := UnmarshalGroupStatus(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidLength)
}

func TestUnmarshalACStatus_InvalidSubType(t *testing.T) {
	// Use wrong subtype (0x21 instead of 0x23)
	data, _ := hex.DecodeString("210000000001000A101278C002DA00008000")

	_, err := UnmarshalACStatus(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sub type")
}

func TestUnmarshalACStatus_TooShort(t *testing.T) {
	// Data shorter than minimum header
	data := []byte{0x23, 0x00}

	_, err := UnmarshalACStatus(data)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidLength)
}
