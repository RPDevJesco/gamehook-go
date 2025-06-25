package mappers

import (
	"encoding/binary"
	"fmt"
	"gamehook/internal/memory"
	"math"
	"strings"
	"time"
)

// AdvancedPropertyProcessor handles complex property types and calculations
type AdvancedPropertyProcessor struct {
	memory   *memory.Manager
	platform Platform
}

// NewAdvancedPropertyProcessor creates a new advanced property processor
func NewAdvancedPropertyProcessor(memManager *memory.Manager, platform Platform) *AdvancedPropertyProcessor {
	return &AdvancedPropertyProcessor{
		memory:   memManager,
		platform: platform,
	}
}

// ProcessAdvancedProperty processes complex property types
func (app *AdvancedPropertyProcessor) ProcessAdvancedProperty(prop *Property) (interface{}, error) {
	switch prop.Type {
	case "pointer":
		return app.processPointer(prop)
	case "array":
		return app.processArray(prop)
	case "struct":
		return app.processStruct(prop)
	case "enum":
		return app.processEnum(prop)
	case "flags":
		return app.processFlags(prop)
	case "float32":
		return app.processFloat32(prop)
	case "float64":
		return app.processFloat64(prop)
	case "time":
		return app.processTime(prop)
	case "version":
		return app.processVersion(prop)
	case "checksum":
		return app.processChecksum(prop)
	case "coordinate":
		return app.processCoordinate(prop)
	case "color":
		return app.processColor(prop)
	case "percentage":
		return app.processPercentage(prop)
	default:
		return nil, fmt.Errorf("unsupported advanced property type: %s", prop.Type)
	}
}

// processPointer handles pointer dereferencing
func (app *AdvancedPropertyProcessor) processPointer(prop *Property) (interface{}, error) {
	// Read the pointer value
	pointerValue, err := app.readUint32(prop.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to read pointer: %w", err)
	}

	// Validate pointer (basic null check and range check)
	if pointerValue == 0 {
		return nil, nil // null pointer
	}

	// Get target type from metadata
	targetType := "uint32" // default
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if t, exists := prop.Transform.Lookup["target_type"]; exists {
			targetType = t
		}
	}

	// Read value at pointer location
	switch targetType {
	case "uint8":
		return app.memory.ReadUint8(pointerValue)
	case "uint16":
		return app.memory.ReadUint16(pointerValue, app.platform.Endian == "little")
	case "uint32":
		return app.memory.ReadUint32(pointerValue, app.platform.Endian == "little")
	case "string":
		maxLen := uint32(256) // default max string length
		if prop.Length > 0 {
			maxLen = prop.Length
		}
		return app.memory.ReadString(pointerValue, maxLen, prop.CharMap)
	default:
		return pointerValue, nil
	}
}

// processArray handles array processing
func (app *AdvancedPropertyProcessor) processArray(prop *Property) (interface{}, error) {
	if prop.Length == 0 {
		return nil, fmt.Errorf("array length not specified")
	}

	elementType := "uint8" // default
	elementSize := uint32(1)

	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if t, exists := prop.Transform.Lookup["element_type"]; exists {
			elementType = t
		}
		if s, exists := prop.Transform.Lookup["element_size"]; exists {
			if size, err := parseUint32(s); err == nil {
				elementSize = size
			}
		}
	}

	result := make([]interface{}, prop.Length)

	for i := uint32(0); i < prop.Length; i++ {
		address := prop.Address + (i * elementSize)

		var value interface{}
		var err error

		switch elementType {
		case "uint8":
			value, err = app.memory.ReadUint8(address)
		case "uint16":
			value, err = app.memory.ReadUint16(address, app.platform.Endian == "little")
		case "uint32":
			value, err = app.memory.ReadUint32(address, app.platform.Endian == "little")
		case "int8":
			value, err = app.memory.ReadInt8(address)
		case "int16":
			value, err = app.memory.ReadInt16(address, app.platform.Endian == "little")
		case "int32":
			value, err = app.memory.ReadInt32(address, app.platform.Endian == "little")
		default:
			value, err = app.memory.ReadUint8(address)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read array element %d: %w", i, err)
		}

		result[i] = value
	}

	return result, nil
}

// processStruct handles struct processing
func (app *AdvancedPropertyProcessor) processStruct(prop *Property) (interface{}, error) {
	if prop.Transform == nil || prop.Transform.Lookup == nil {
		return nil, fmt.Errorf("struct definition not found in transform.lookup")
	}

	result := make(map[string]interface{})

	for fieldName, fieldDef := range prop.Transform.Lookup {
		if strings.HasPrefix(fieldName, "field_") {
			// Parse field definition: "type:offset:size"
			parts := strings.Split(fieldDef, ":")
			if len(parts) < 2 {
				continue
			}

			fieldType := parts[0]
			fieldOffset, err := parseUint32(parts[1])
			if err != nil {
				continue
			}

			fieldAddress := prop.Address + fieldOffset
			fieldName = strings.TrimPrefix(fieldName, "field_")

			var value interface{}

			switch fieldType {
			case "uint8":
				value, err = app.memory.ReadUint8(fieldAddress)
			case "uint16":
				value, err = app.memory.ReadUint16(fieldAddress, app.platform.Endian == "little")
			case "uint32":
				value, err = app.memory.ReadUint32(fieldAddress, app.platform.Endian == "little")
			default:
				value, err = app.memory.ReadUint8(fieldAddress)
			}

			if err != nil {
				return nil, fmt.Errorf("failed to read struct field %s: %w", fieldName, err)
			}

			result[fieldName] = value
		}
	}

	return result, nil
}

// processEnum handles enumeration values
func (app *AdvancedPropertyProcessor) processEnum(prop *Property) (interface{}, error) {
	// Read the raw value
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	// Convert to string for lookup
	key := fmt.Sprintf("%.0f", rawValue)

	// Look up enum value
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if enumValue, exists := prop.Transform.Lookup[key]; exists {
			return map[string]interface{}{
				"value": rawValue,
				"name":  enumValue,
			}, nil
		}
	}

	// Return raw value if no enum mapping found
	return map[string]interface{}{
		"value": rawValue,
		"name":  fmt.Sprintf("Unknown(%v)", rawValue),
	}, nil
}

// processFlags handles bit flag values
func (app *AdvancedPropertyProcessor) processFlags(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)
	flags := make(map[string]bool)
	activeFlags := make([]string, 0)

	// Check each bit flag
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		for bitStr, flagName := range prop.Transform.Lookup {
			if strings.HasPrefix(bitStr, "bit_") {
				bitNum, err := parseUint32(strings.TrimPrefix(bitStr, "bit_"))
				if err != nil {
					continue
				}

				isSet := (intValue & (1 << bitNum)) != 0
				flags[flagName] = isSet

				if isSet {
					activeFlags = append(activeFlags, flagName)
				}
			}
		}
	}

	return map[string]interface{}{
		"value":        intValue,
		"flags":        flags,
		"active_flags": activeFlags,
	}, nil
}

// processFloat32 handles 32-bit floating point values
func (app *AdvancedPropertyProcessor) processFloat32(prop *Property) (interface{}, error) {
	data, err := app.memory.ReadBytes(prop.Address, 4)
	if err != nil {
		return nil, err
	}

	var bits uint32
	if app.platform.Endian == "little" {
		bits = binary.LittleEndian.Uint32(data)
	} else {
		bits = binary.BigEndian.Uint32(data)
	}

	value := math.Float32frombits(bits)

	// Apply transformations
	if prop.Transform != nil {
		if prop.Transform.Multiply != nil {
			value *= float32(*prop.Transform.Multiply)
		}
		if prop.Transform.Add != nil {
			value += float32(*prop.Transform.Add)
		}
	}

	return value, nil
}

// processFloat64 handles 64-bit floating point values
func (app *AdvancedPropertyProcessor) processFloat64(prop *Property) (interface{}, error) {
	data, err := app.memory.ReadBytes(prop.Address, 8)
	if err != nil {
		return nil, err
	}

	var bits uint64
	if app.platform.Endian == "little" {
		bits = binary.LittleEndian.Uint64(data)
	} else {
		bits = binary.BigEndian.Uint64(data)
	}

	value := math.Float64frombits(bits)

	// Apply transformations
	if prop.Transform != nil {
		if prop.Transform.Multiply != nil {
			value *= *prop.Transform.Multiply
		}
		if prop.Transform.Add != nil {
			value += *prop.Transform.Add
		}
	}

	return value, nil
}

// processTime handles time-based values
func (app *AdvancedPropertyProcessor) processTime(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	// Determine time format from transform
	format := "frames" // default
	frameRate := 60.0  // default

	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if f, exists := prop.Transform.Lookup["format"]; exists {
			format = f
		}
		if fr, exists := prop.Transform.Lookup["frame_rate"]; exists {
			if rate, err := parseFloat64(fr); err == nil {
				frameRate = rate
			}
		}
	}

	var duration time.Duration

	switch format {
	case "frames":
		seconds := rawValue / frameRate
		duration = time.Duration(seconds * float64(time.Second))
	case "milliseconds":
		duration = time.Duration(rawValue * float64(time.Millisecond))
	case "seconds":
		duration = time.Duration(rawValue * float64(time.Second))
	case "unix":
		// Unix timestamp
		return time.Unix(int64(rawValue), 0), nil
	}

	return map[string]interface{}{
		"raw_value": rawValue,
		"duration":  duration.String(),
		"seconds":   duration.Seconds(),
		"minutes":   duration.Minutes(),
		"hours":     duration.Hours(),
	}, nil
}

// processVersion handles version numbers
func (app *AdvancedPropertyProcessor) processVersion(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)

	// Default version parsing: major.minor.patch
	major := (intValue >> 16) & 0xFF
	minor := (intValue >> 8) & 0xFF
	patch := intValue & 0xFF

	// Check for custom version format in transform
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if format, exists := prop.Transform.Lookup["format"]; exists {
			switch format {
			case "bcd":
				// Binary Coded Decimal
				major = ((intValue>>20)&0xF)*10 + ((intValue >> 16) & 0xF)
				minor = ((intValue>>12)&0xF)*10 + ((intValue >> 8) & 0xF)
				patch = ((intValue>>4)&0xF)*10 + (intValue & 0xF)
			}
		}
	}

	return map[string]interface{}{
		"raw_value": intValue,
		"major":     major,
		"minor":     minor,
		"patch":     patch,
		"string":    fmt.Sprintf("%d.%d.%d", major, minor, patch),
	}, nil
}

// processChecksum handles checksum values
func (app *AdvancedPropertyProcessor) processChecksum(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"value": uint32(rawValue),
		"hex":   fmt.Sprintf("0x%08X", uint32(rawValue)),
	}, nil
}

// processCoordinate handles 2D/3D coordinates
func (app *AdvancedPropertyProcessor) processCoordinate(prop *Property) (interface{}, error) {
	if prop.Length < 2 {
		return nil, fmt.Errorf("coordinate requires at least 2 components")
	}

	// Read X and Y components
	x, err := app.readValueAtOffset(prop, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read X coordinate: %w", err)
	}

	y, err := app.readValueAtOffset(prop, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to read Y coordinate: %w", err)
	}

	result := map[string]interface{}{
		"x": x,
		"y": y,
	}

	// Read Z component if available
	if prop.Length >= 3 {
		z, err := app.readValueAtOffset(prop, 2)
		if err == nil {
			result["z"] = z
		}
	}

	return result, nil
}

// processColor handles color values
func (app *AdvancedPropertyProcessor) processColor(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)

	// Default: RGB565 or ARGB8888 based on size
	var r, g, b, a uint8

	format := "rgb565" // default
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if f, exists := prop.Transform.Lookup["format"]; exists {
			format = f
		}
	}

	switch format {
	case "rgb565":
		r = uint8((intValue>>11)&0x1F) << 3
		g = uint8((intValue>>5)&0x3F) << 2
		b = uint8(intValue&0x1F) << 3
		a = 255
	case "argb8888":
		a = uint8((intValue >> 24) & 0xFF)
		r = uint8((intValue >> 16) & 0xFF)
		g = uint8((intValue >> 8) & 0xFF)
		b = uint8(intValue & 0xFF)
	case "rgb888":
		r = uint8((intValue >> 16) & 0xFF)
		g = uint8((intValue >> 8) & 0xFF)
		b = uint8(intValue & 0xFF)
		a = 255
	}

	return map[string]interface{}{
		"raw_value": intValue,
		"r":         r,
		"g":         g,
		"b":         b,
		"a":         a,
		"hex":       fmt.Sprintf("#%02X%02X%02X", r, g, b),
		"hex_alpha": fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a),
	}, nil
}

// processPercentage handles percentage values
func (app *AdvancedPropertyProcessor) processPercentage(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	// Default: value is 0-100
	maxValue := 100.0

	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if max, exists := prop.Transform.Lookup["max_value"]; exists {
			if maxVal, err := parseFloat64(max); err == nil {
				maxValue = maxVal
			}
		}
	}

	percentage := (rawValue / maxValue) * 100

	return map[string]interface{}{
		"raw_value":  rawValue,
		"percentage": percentage,
		"decimal":    rawValue / maxValue,
	}, nil
}

// Helper methods

func (app *AdvancedPropertyProcessor) readBasicValue(prop *Property) (float64, error) {
	var value interface{}
	var err error

	littleEndian := app.platform.Endian == "little"

	switch prop.Length {
	case 1:
		value, err = app.memory.ReadUint8(prop.Address)
	case 2:
		value, err = app.memory.ReadUint16(prop.Address, littleEndian)
	case 4:
		value, err = app.memory.ReadUint32(prop.Address, littleEndian)
	default:
		value, err = app.memory.ReadUint8(prop.Address)
	}

	if err != nil {
		return 0, err
	}

	switch v := value.(type) {
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("unsupported value type: %T", v)
	}
}

func (app *AdvancedPropertyProcessor) readUint32(address uint32) (uint32, error) {
	return app.memory.ReadUint32(address, app.platform.Endian == "little")
}

func (app *AdvancedPropertyProcessor) readValueAtOffset(prop *Property, offset uint32) (float64, error) {
	elementSize := uint32(2) // default to 2 bytes per coordinate component
	if prop.Transform != nil && prop.Transform.Lookup != nil {
		if size, exists := prop.Transform.Lookup["element_size"]; exists {
			if s, err := parseUint32(size); err == nil {
				elementSize = s
			}
		}
	}

	address := prop.Address + (offset * elementSize)

	switch elementSize {
	case 1:
		val, err := app.memory.ReadUint8(address)
		return float64(val), err
	case 2:
		val, err := app.memory.ReadUint16(address, app.platform.Endian == "little")
		return float64(val), err
	case 4:
		val, err := app.memory.ReadUint32(address, app.platform.Endian == "little")
		return float64(val), err
	default:
		val, err := app.memory.ReadUint8(address)
		return float64(val), err
	}
}

// Utility functions for parsing

func parseUint32(s string) (uint32, error) {
	if strings.HasPrefix(s, "0x") {
		var val uint32
		_, err := fmt.Sscanf(s, "0x%x", &val)
		return val, err
	}

	var val uint32
	_, err := fmt.Sscanf(s, "%d", &val)
	return val, err
}

func parseFloat64(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}
