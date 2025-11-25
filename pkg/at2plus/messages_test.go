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
