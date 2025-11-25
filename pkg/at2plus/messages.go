package at2plus

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// GroupControl represents a command to control a group
type GroupControl struct {
	GroupNumber uint8 // 0-15
	Power       *int  // 0: Next, 1: Off, 2: On, 3: Turbo (Mapped from spec: 001->Next, 010->Off, 011->On, 101->Turbo)
	Value       *int  // 0: Dec, 1: Inc, 2: Set (Mapped from spec: 010->Dec, 011->Inc, 100->Set)
	Percent     *int  // 0-100
}

// GroupStatus represents the status of a group
type GroupStatus struct {
	GroupNumber  uint8
	Power        int // 0: Off, 1: On, 3: Turbo
	Percent      int
	TurboSupport bool
	Spill        bool
}

// ACControl represents a command to control an AC
type ACControl struct {
	ACNumber uint8 // 0-7
	Power    *int  // 1: Toggle, 2: Off, 3: On, 4: Away, 5: Sleep
	Mode     *int  // 0: Auto, 1: Heat, 2: Dry, 3: Fan, 4: Cool
	FanSpeed *int  // 0: Auto, 1: Quiet, 2: Low, 3: Med, 4: High, 5: Powerful, 6: Turbo
	Setpoint *int  // 10-35
}

// ACStatus represents the status of an AC
type ACStatus struct {
	ACNumber    uint8
	Power       int // 0: Off, 1: On, ...
	Mode        int
	FanSpeed    int
	Setpoint    int
	Temperature int
	Turbo       bool
	Bypass      bool
	Spill       bool
	Timer       bool
	ErrorCode   int
}

// MarshalGroupControl creates the byte payload for a Group Control message
// Spec: 0x20 (Sub Type) + NormalLen(0) + RepeatCount(1) + RepeatLen(4) + Data
func MarshalGroupControl(groups []GroupControl) ([]byte, error) {
	// Header part: 8 bytes
	// Byte1: SubType (0x20)
	// Byte2: 0
	// Byte3: Normal Len (0)
	// Byte4: 0
	// Byte5: Repeat Count
	// Byte6: 0
	// Byte7: Repeat Len (4)
	// Byte8: 0

	count := len(groups)
	buf := make([]byte, 8+count*4)

	buf[0] = SubMsgTypeGroupControl
	binary.BigEndian.PutUint16(buf[4:6], uint16(count))
	binary.BigEndian.PutUint16(buf[6:8], 0x04)

	for i, g := range groups {
		offset := 8 + i*4
		// Byte 1: Bit6-1 Group Number
		if g.GroupNumber > 15 {
			return nil, errors.New("invalid group number")
		}
		buf[offset] = g.GroupNumber

		// Byte 2: Bit8-6 Group Setting Value, Bit3-1 Power
		var b2 uint8
		if g.Value != nil {
			// Map: 0->010(2), 1->011(3), 2->100(4)
			val := 0
			switch *g.Value {
			case 0:
				val = 2 // Dec
			case 1:
				val = 3 // Inc
			case 2:
				val = 4 // Set
			}
			b2 |= uint8(val << 5)
		}
		if g.Power != nil {
			// Map: 0->001(1), 1->010(2), 2->011(3), 3->101(5)
			val := 0
			switch *g.Power {
			case 0:
				val = 1 // Next
			case 1:
				val = 2 // Off
			case 2:
				val = 3 // On
			case 3:
				val = 5 // Turbo
			}
			b2 |= uint8(val)
		}
		buf[offset+1] = b2

		// Byte 3: Percentage
		if g.Percent != nil {
			buf[offset+2] = uint8(*g.Percent)
		}

		// Byte 4: 0
	}
	return buf, nil
}

// UnmarshalGroupStatus parses the byte payload of a Group Status message
func UnmarshalGroupStatus(data []byte) ([]GroupStatus, error) {
	if len(data) < 8 {
		return nil, ErrInvalidLength
	}

	subType := data[0]
	if subType != SubMsgTypeGroupStatus {
		return nil, fmt.Errorf("invalid sub type for group status: %x", subType)
	}

	count := int(binary.BigEndian.Uint16(data[4:6]))
	repeatLen := int(binary.BigEndian.Uint16(data[6:8]))

	expectedLen := 8 + count*repeatLen
	if len(data) < expectedLen {
		return nil, ErrInvalidLength
	}

	groups := make([]GroupStatus, 0, count)
	for i := 0; i < count; i++ {
		offset := 8 + i*repeatLen
		chunk := data[offset : offset+repeatLen]

		// Byte 1: Bit8-7 Power, Bit6-1 Group Num
		power := (chunk[0] >> 6) & 0x03
		groupNum := chunk[0] & 0x3F

		// Byte 2: Bit7-1 Open Percentage
		percent := chunk[1] & 0x7F

		// Byte 7: Bit8 Turbo Support, Bit2 Spill
		turboSupport := (chunk[6] & 0x80) != 0
		spill := (chunk[6] & 0x02) != 0

		groups = append(groups, GroupStatus{
			GroupNumber:  groupNum,
			Power:        int(power),
			Percent:      int(percent),
			TurboSupport: turboSupport,
			Spill:        spill,
		})
	}
	return groups, nil
}

// MarshalACControl creates the byte payload for an AC Control message
func MarshalACControl(acs []ACControl) ([]byte, error) {
	count := len(acs)
	buf := make([]byte, 8+count*4)

	buf[0] = SubMsgTypeACControl
	binary.BigEndian.PutUint16(buf[4:6], uint16(count))
	binary.BigEndian.PutUint16(buf[6:8], 0x04)

	for i, ac := range acs {
		offset := 8 + i*4

		// Byte 1: Bit8-5 Power, Bit4-1 AC Number
		var b1 uint8
		if ac.Power != nil {
			// Map: 1->0001, 2->0010, 3->0011, 4->0100, 5->0101
			// Spec: 1:Change, 2:Off, 3:On, 4:Away, 5:Sleep
			// Wait, my struct comments mapped 1:Toggle, 2:Off, 3:On...
			// Spec: 0001: Change on/off, 0010: Off, 0011: On, 0100: Away, 0101: Sleep
			val := 0
			switch *ac.Power {
			case 1:
				val = 1
			case 2:
				val = 2
			case 3:
				val = 3
			case 4:
				val = 4
			case 5:
				val = 5
			}
			b1 |= uint8(val << 4)
		}
		b1 |= ac.ACNumber & 0x0F
		buf[offset] = b1

		// Byte 2: Bit8-5 Mode, Bit4-1 Fan Speed
		var b2 uint8
		if ac.Mode != nil {
			b2 |= uint8(*ac.Mode << 4)
		}
		if ac.FanSpeed != nil {
			b2 |= uint8(*ac.FanSpeed)
		}
		buf[offset+1] = b2

		// Byte 3: Setpoint Control
		// Byte 4: Setpoint Value
		if ac.Setpoint != nil {
			buf[offset+2] = 0x40 // Change setpoint
			// Setpoint = (data+100)/10 -> data = Setpoint*10 - 100
			val := (*ac.Setpoint * 10) - 100
			if val < 0 {
				val = 0
			}
			buf[offset+3] = uint8(val)
		} else {
			buf[offset+2] = 0x00 // Keep setpoint
		}
	}
	return buf, nil
}

// UnmarshalACStatus parses the byte payload of an AC Status message
func UnmarshalACStatus(data []byte) ([]ACStatus, error) {
	if len(data) < 8 {
		return nil, ErrInvalidLength
	}

	subType := data[0]
	if subType != SubMsgTypeACStatus {
		return nil, fmt.Errorf("invalid sub type for ac status: %x", subType)
	}

	count := int(binary.BigEndian.Uint16(data[4:6]))
	repeatLen := int(binary.BigEndian.Uint16(data[6:8]))

	expectedLen := 8 + count*repeatLen
	if len(data) < expectedLen {
		return nil, ErrInvalidLength
	}

	acs := make([]ACStatus, 0, count)
	for i := 0; i < count; i++ {
		offset := 8 + i*repeatLen
		chunk := data[offset : offset+repeatLen]

		// Byte 1: Bit8-5 Power, Bit4-1 AC Num
		power := (chunk[0] >> 4) & 0x0F
		acNum := chunk[0] & 0x0F

		// Byte 2: Bit8-5 Mode, Bit4-1 Fan
		mode := (chunk[1] >> 4) & 0x0F
		fan := chunk[1] & 0x0F

		// Byte 3: Setpoint (VALUE+100)/10
		setpointVal := int(chunk[2])
		setpoint := (setpointVal + 100) / 10

		// Byte 4: Turbo, Bypass, Spill, Timer
		turbo := (chunk[3] & 0x10) != 0
		bypass := (chunk[3] & 0x08) != 0
		spill := (chunk[3] & 0x04) != 0
		timer := (chunk[3] & 0x02) != 0

		// Byte 5-6: Temperature (VALUE-500)/10
		tempVal := int(binary.BigEndian.Uint16(chunk[4:6]))
		temperature := (tempVal - 500) / 10

		// Byte 7: Error Code
		errCode := int(chunk[6])

		acs = append(acs, ACStatus{
			ACNumber:    acNum,
			Power:       int(power),
			Mode:        int(mode),
			FanSpeed:    int(fan),
			Setpoint:    setpoint,
			Temperature: temperature,
			Turbo:       turbo,
			Bypass:      bypass,
			Spill:       spill,
			Timer:       timer,
			ErrorCode:   errCode,
		})
	}
	return acs, nil
}

// ExtendedMessage represents the generic extended message structure
// Note: Extended messages usually just have data that needs specific parsing based on the subtype.

// ACAbility represents the capabilities of an AC
type ACAbility struct {
	ACNumber    uint8
	Name        string
	StartGroup  uint8
	GroupCount  uint8
	CoolMode    bool
	FanMode     bool
	DryMode     bool
	HeatMode    bool
	AutoMode    bool
	FanTurbo    bool
	FanPowerful bool
	FanHigh     bool
	FanMed      bool
	FanLow      bool
	FanQuiet    bool
	FanAuto     bool
	MinCoolSet  int
	MaxCoolSet  int
	MinHeatSet  int
	MaxHeatSet  int
}

// UnmarshalACAbility parses the AC Ability extended message
func UnmarshalACAbility(data []byte) ([]ACAbility, error) {
	// Header: FF 11 ACNum Length ...
	if len(data) < 4 {
		return nil, ErrInvalidLength
	}

	if data[0] != 0xFF || data[1] != ExtMsgTypeACAbility {
		return nil, errors.New("invalid ac ability header")
	}

	// Spec says: "If there are more than one AC, the data will be repeated"
	// But the structure is a bit weird.
	// Byte 3: AC Number (0-3) - wait, if multiple ACs, does this increment?
	// Spec Example: "2 ACs will receive 54 bytes data".
	// The example shows:
	// Byte 1: FF
	// Byte 2: 11
	// Byte 3: AC Number
	// Byte 4: Following data length (24)
	// Byte 5-28: Data
	// Then repeated?

	// Let's assume it repeats from Byte 3.

	var abilities []ACAbility
	offset := 0

	for offset < len(data) {
		if offset+4 > len(data) {
			break
		}

		// Check FF 11 again? The spec says "Data received... Byte 1 Fixed 0xFF...".
		// If multiple ACs, does it repeat FF 11?
		// Spec says: "If there are more than one AC, the data will be repeated with relevant values. E.g. 2 ACs will receive 54(2+26+26) bytes data"
		// 2 (Header FF 11) + 26 (AC0) + 26 (AC1).
		// So FF 11 is ONLY at the start.

		if offset == 0 {
			if data[0] != 0xFF || data[1] != 0x11 {
				return nil, errors.New("invalid header")
			}
			offset += 2
		}

		if offset+2 > len(data) {
			break
		}

		acNum := data[offset]
		length := int(data[offset+1])

		if offset+2+length > len(data) {
			return nil, ErrInvalidLength
		}

		chunk := data[offset+2 : offset+2+length]

		// Parse chunk (24 bytes)
		// Byte 1-16: Name
		nameBytes := chunk[0:16]
		// Find null terminator
		nameLen := 0
		for i, b := range nameBytes {
			if b == 0 {
				nameLen = i
				break
			}
			nameLen = i + 1
		}
		name := string(nameBytes[:nameLen])

		startGroup := chunk[16]
		groupCount := chunk[17]

		// Byte 19: Modes
		cool := (chunk[18] & 0x20) != 0
		fan := (chunk[18] & 0x10) != 0
		dry := (chunk[18] & 0x08) != 0
		heat := (chunk[18] & 0x04) != 0
		auto := (chunk[18] & 0x02) != 0

		// Byte 20: Fan Speeds
		fTurbo := (chunk[19] & 0x80) != 0
		fPowerful := (chunk[19] & 0x40) != 0
		fHigh := (chunk[19] & 0x20) != 0
		fMed := (chunk[19] & 0x10) != 0
		fLow := (chunk[19] & 0x08) != 0
		fQuiet := (chunk[19] & 0x04) != 0
		fAuto := (chunk[19] & 0x02) != 0

		minCool := int(chunk[20])
		maxCool := int(chunk[21])
		minHeat := int(chunk[22])
		maxHeat := int(chunk[23])

		abilities = append(abilities, ACAbility{
			ACNumber:    acNum,
			Name:        name,
			StartGroup:  startGroup,
			GroupCount:  groupCount,
			CoolMode:    cool,
			FanMode:     fan,
			DryMode:     dry,
			HeatMode:    heat,
			AutoMode:    auto,
			FanTurbo:    fTurbo,
			FanPowerful: fPowerful,
			FanHigh:     fHigh,
			FanMed:      fMed,
			FanLow:      fLow,
			FanQuiet:    fQuiet,
			FanAuto:     fAuto,
			MinCoolSet:  minCool,
			MaxCoolSet:  maxCool,
			MinHeatSet:  minHeat,
			MaxHeatSet:  maxHeat,
		})

		offset += 2 + length
	}

	return abilities, nil
}

// GroupName represents a group name
type GroupName struct {
	GroupNumber uint8
	Name        string
}

// UnmarshalGroupName parses the Group Name extended message
func UnmarshalGroupName(data []byte) ([]GroupName, error) {
	// Header: FF 12 ...
	if len(data) < 2 {
		return nil, ErrInvalidLength
	}
	if data[0] != 0xFF || data[1] != ExtMsgTypeGroupName {
		return nil, errors.New("invalid group name header")
	}

	var names []GroupName
	offset := 2

	// Spec: "2 groups will receive 20(2+9+9) bytes data"
	// So each chunk is 1 byte (Group Num) + 8 bytes (Name) = 9 bytes.

	for offset < len(data) {
		if offset+9 > len(data) {
			break
		}

		groupNum := data[offset]
		nameBytes := data[offset+1 : offset+9]

		nameLen := 0
		for i, b := range nameBytes {
			if b == 0 {
				nameLen = i
				break
			}
			nameLen = i + 1
		}
		name := string(nameBytes[:nameLen])

		names = append(names, GroupName{
			GroupNumber: groupNum,
			Name:        name,
		})

		offset += 9
	}

	return names, nil
}
