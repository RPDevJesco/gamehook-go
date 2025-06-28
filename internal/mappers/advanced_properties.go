package mappers

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"gamehook/internal/memory"
	"hash/crc32"
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

// ProcessAdvancedProperty processes complex property types with enhanced capabilities
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

// ===== ENHANCED POINTER PROCESSING =====

// processPointer handles enhanced pointer dereferencing with safety checks
func (app *AdvancedPropertyProcessor) processPointer(prop *Property) (interface{}, error) {
	// Read the pointer value
	pointerValue, err := app.readUint32(prop.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to read pointer: %w", err)
	}

	// Enhanced null checking
	nullValue := uint32(0)
	if prop.Advanced != nil && prop.Advanced.NullValue != nil {
		nullValue = *prop.Advanced.NullValue
	}

	if pointerValue == nullValue {
		return map[string]interface{}{
			"is_null":     true,
			"pointer":     pointerValue,
			"target":      nil,
			"target_type": app.getTargetType(prop),
		}, nil
	}

	// Validate pointer range
	if !app.isValidPointerAddress(pointerValue) {
		return map[string]interface{}{
			"is_null":     false,
			"pointer":     pointerValue,
			"target":      nil,
			"error":       "invalid_pointer_address",
			"target_type": app.getTargetType(prop),
		}, nil
	}

	// Get target type from advanced configuration
	targetType := app.getTargetType(prop)

	// Handle maximum dereferences
	maxDeref := uint(1)
	if prop.Advanced != nil && prop.Advanced.MaxDereferences != nil {
		maxDeref = *prop.Advanced.MaxDereferences
	}

	// Read value at pointer location with type safety
	targetValue, err := app.readTypedValue(pointerValue, targetType, maxDeref)
	if err != nil {
		return map[string]interface{}{
			"is_null":     false,
			"pointer":     pointerValue,
			"target":      nil,
			"error":       err.Error(),
			"target_type": targetType,
		}, nil
	}

	return map[string]interface{}{
		"is_null":     false,
		"pointer":     pointerValue,
		"target":      targetValue,
		"target_type": targetType,
	}, nil
}

// getTargetType gets the target type for pointer dereferencing
func (app *AdvancedPropertyProcessor) getTargetType(prop *Property) PropertyType {
	if prop.Advanced != nil && prop.Advanced.TargetType != nil {
		return *prop.Advanced.TargetType
	}
	return PropertyTypeUint32 // default
}

// isValidPointerAddress validates if a pointer address is within valid memory ranges
func (app *AdvancedPropertyProcessor) isValidPointerAddress(address uint32) bool {
	// Check against platform memory blocks
	for _, block := range app.platform.MemoryBlocks {
		if address >= block.Start && address <= block.End {
			return true
		}
	}
	return false
}

// readTypedValue reads a value at the given address with type safety
func (app *AdvancedPropertyProcessor) readTypedValue(address uint32, targetType PropertyType, maxDeref uint) (interface{}, error) {
	if maxDeref == 0 {
		return nil, fmt.Errorf("maximum dereference depth reached")
	}

	switch targetType {
	case PropertyTypeUint8:
		return app.memory.ReadUint8(address)
	case PropertyTypeUint16:
		return app.memory.ReadUint16(address, app.platform.Endian == "little")
	case PropertyTypeUint32:
		return app.memory.ReadUint32(address, app.platform.Endian == "little")
	case PropertyTypeInt8:
		return app.memory.ReadInt8(address)
	case PropertyTypeInt16:
		return app.memory.ReadInt16(address, app.platform.Endian == "little")
	case PropertyTypeInt32:
		return app.memory.ReadInt32(address, app.platform.Endian == "little")
	case PropertyTypeString:
		maxLen := uint32(256) // default max string length
		return app.memory.ReadString(address, maxLen, nil)
	case PropertyTypePointer:
		// Recursive pointer dereferencing
		if maxDeref > 1 {
			nextPointer, err := app.memory.ReadUint32(address, app.platform.Endian == "little")
			if err != nil {
				return nil, err
			}
			return app.readTypedValue(nextPointer, PropertyTypeUint32, maxDeref-1)
		}
		return app.memory.ReadUint32(address, app.platform.Endian == "little")
	default:
		return app.memory.ReadUint32(address, app.platform.Endian == "little")
	}
}

// ===== ENHANCED ARRAY PROCESSING =====

// processArray handles enhanced array processing with dynamic sizing and advanced element types
func (app *AdvancedPropertyProcessor) processArray(prop *Property) (interface{}, error) {
	// Determine array length
	length, err := app.getArrayLength(prop)
	if err != nil {
		return nil, fmt.Errorf("failed to determine array length: %w", err)
	}

	if length == 0 {
		return []interface{}{}, nil
	}

	// Get element configuration
	elementType := PropertyTypeUint8 // default
	elementSize := uint32(1)
	indexOffset := uint32(0)
	stride := uint32(1)

	if prop.Advanced != nil {
		if prop.Advanced.ElementType != nil {
			elementType = *prop.Advanced.ElementType
		}
		if prop.Advanced.ElementSize != nil {
			elementSize = uint32(*prop.Advanced.ElementSize)
		}
		if prop.Advanced.IndexOffset != nil {
			indexOffset = uint32(*prop.Advanced.IndexOffset)
		}
		if prop.Advanced.Stride != nil {
			stride = uint32(*prop.Advanced.Stride)
		} else {
			stride = elementSize
		}
	}

	// Validate maximum elements
	if prop.Advanced != nil && prop.Advanced.MaxElements != nil {
		maxElements := uint32(*prop.Advanced.MaxElements)
		if length > maxElements {
			length = maxElements
		}
	}

	result := make([]interface{}, 0, length)

	for i := uint32(0); i < length; i++ {
		elementAddress := prop.Address + indexOffset + (i * stride)

		var value interface{}
		var readErr error

		switch elementType {
		case PropertyTypeUint8:
			value, readErr = app.memory.ReadUint8(elementAddress)
		case PropertyTypeUint16:
			value, readErr = app.memory.ReadUint16(elementAddress, app.platform.Endian == "little")
		case PropertyTypeUint32:
			value, readErr = app.memory.ReadUint32(elementAddress, app.platform.Endian == "little")
		case PropertyTypeInt8:
			value, readErr = app.memory.ReadInt8(elementAddress)
		case PropertyTypeInt16:
			value, readErr = app.memory.ReadInt16(elementAddress, app.platform.Endian == "little")
		case PropertyTypeInt32:
			value, readErr = app.memory.ReadInt32(elementAddress, app.platform.Endian == "little")
		case PropertyTypeFloat32:
			value, readErr = app.memory.ReadFloat32(elementAddress, app.platform.Endian == "little")
		case PropertyTypeFloat64:
			value, readErr = app.memory.ReadFloat64(elementAddress, app.platform.Endian == "little")
		case PropertyTypeString:
			value, readErr = app.memory.ReadString(elementAddress, elementSize, nil)
		case PropertyTypeStruct:
			value, readErr = app.processStructAtAddress(elementAddress, elementSize, prop)
		default:
			value, readErr = app.memory.ReadUint8(elementAddress)
		}

		if readErr != nil {
			// For non-critical errors, include error information but continue
			value = map[string]interface{}{
				"error":   readErr.Error(),
				"address": elementAddress,
				"index":   i,
			}
		}

		result = append(result, value)
	}

	return map[string]interface{}{
		"elements":     result,
		"length":       length,
		"element_type": elementType,
		"element_size": elementSize,
		"stride":       stride,
	}, nil
}

// getArrayLength determines the length of an array
func (app *AdvancedPropertyProcessor) getArrayLength(prop *Property) (uint32, error) {
	// Check if dynamic length is enabled
	if prop.Advanced != nil && prop.Advanced.DynamicLength != nil && *prop.Advanced.DynamicLength {
		if prop.Advanced.LengthProperty != "" {
			// TODO: Get length from another property
			// For now, return the static length
			return prop.Length, nil
		}
	}

	return prop.Length, nil
}

// ===== ENHANCED STRUCT PROCESSING =====

// processStruct handles enhanced struct processing with inheritance and validation
func (app *AdvancedPropertyProcessor) processStruct(prop *Property) (interface{}, error) {
	return app.processStructAtAddress(prop.Address, prop.Length, prop)
}

// processStructAtAddress processes a struct at a specific address
func (app *AdvancedPropertyProcessor) processStructAtAddress(address uint32, size uint32, prop *Property) (interface{}, error) {
	if prop.Advanced == nil || prop.Advanced.Fields == nil {
		return nil, fmt.Errorf("struct definition not found in advanced configuration")
	}

	result := make(map[string]interface{})

	// Handle struct inheritance
	if prop.Advanced.Extends != "" {
		// TODO: Implement struct inheritance
		result["extends"] = prop.Advanced.Extends
	}

	// Process each field
	for fieldName, field := range prop.Advanced.Fields {
		fieldAddress := address + uint32(field.Offset)

		// Validate field bounds
		fieldSize := uint32(1)
		if field.Size != nil {
			fieldSize = uint32(*field.Size)
		}

		if field.Offset >= uint(size) || uint32(field.Offset)+fieldSize > size {
			result[fieldName] = map[string]interface{}{
				"error":  "field_out_of_bounds",
				"offset": field.Offset,
				"size":   fieldSize,
			}
			continue
		}

		var fieldValue interface{}
		var err error

		// Handle computed fields
		if field.Computed != nil {
			// TODO: Implement computed struct fields
			fieldValue = fmt.Sprintf("computed(%s)", field.Computed.Expression)
		} else {
			// Read field value based on type
			fieldValue, err = app.readStructField(fieldAddress, field)
			if err != nil {
				fieldValue = map[string]interface{}{
					"error":   err.Error(),
					"address": fieldAddress,
					"type":    field.Type,
				}
			}
		}

		// Apply field transformations
		if field.Transform != nil {
			// TODO: Apply field-specific transformations
		}

		// Validate field value
		if field.Validation != nil {
			// TODO: Apply field-specific validation
		}

		result[fieldName] = fieldValue
	}

	return map[string]interface{}{
		"fields": result,
		"size":   size,
		"type":   "struct",
	}, nil
}

// readStructField reads a single struct field
func (app *AdvancedPropertyProcessor) readStructField(address uint32, field *StructField) (interface{}, error) {
	switch field.Type {
	case PropertyTypeUint8:
		return app.memory.ReadUint8(address)
	case PropertyTypeUint16:
		return app.memory.ReadUint16(address, app.platform.Endian == "little")
	case PropertyTypeUint32:
		return app.memory.ReadUint32(address, app.platform.Endian == "little")
	case PropertyTypeInt8:
		return app.memory.ReadInt8(address)
	case PropertyTypeInt16:
		return app.memory.ReadInt16(address, app.platform.Endian == "little")
	case PropertyTypeInt32:
		return app.memory.ReadInt32(address, app.platform.Endian == "little")
	case PropertyTypeFloat32:
		return app.memory.ReadFloat32(address, app.platform.Endian == "little")
	case PropertyTypeFloat64:
		return app.memory.ReadFloat64(address, app.platform.Endian == "little")
	case PropertyTypeBool:
		return app.memory.ReadBool(address)
	case PropertyTypeString:
		fieldSize := uint32(1)
		if field.Size != nil {
			fieldSize = uint32(*field.Size)
		}
		return app.memory.ReadString(address, fieldSize, nil)
	default:
		return app.memory.ReadUint8(address)
	}
}

// ===== ENHANCED ENUM PROCESSING =====

// processEnum handles enhanced enumeration values with metadata
func (app *AdvancedPropertyProcessor) processEnum(prop *Property) (interface{}, error) {
	// Read the raw value
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)

	// Enhanced enum processing with advanced configuration
	if prop.Advanced != nil && prop.Advanced.EnumValues != nil {
		// Look for exact match first
		for key, enumValue := range prop.Advanced.EnumValues {
			if enumValue.Value == intValue {
				result := map[string]interface{}{
					"value": intValue,
					"name":  enumValue.Description,
					"key":   key,
					"valid": true,
				}

				if enumValue.Color != "" {
					result["color"] = enumValue.Color
				}
				if enumValue.Icon != "" {
					result["icon"] = enumValue.Icon
				}
				if enumValue.Deprecated != nil && *enumValue.Deprecated {
					result["deprecated"] = true
				}

				return result, nil
			}
		}

		// Handle unknown values based on configuration
		if prop.Advanced.AllowUnknownValues != nil && !*prop.Advanced.AllowUnknownValues {
			return map[string]interface{}{
				"value": intValue,
				"name":  "INVALID_ENUM_VALUE",
				"valid": false,
				"error": "unknown_enum_value_not_allowed",
			}, nil
		}
	}

	// Return unknown value with metadata
	return map[string]interface{}{
		"value": intValue,
		"name":  fmt.Sprintf("Unknown(0x%X)", intValue),
		"valid": false,
	}, nil
}

// ===== ENHANCED FLAGS PROCESSING =====

// processFlags handles enhanced bit flag values with grouping and mutual exclusion
func (app *AdvancedPropertyProcessor) processFlags(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)
	flags := make(map[string]bool)
	activeFlags := make([]string, 0)
	groups := make(map[string][]string)
	conflicts := make([]string, 0)

	// Enhanced flag processing with advanced configuration
	if prop.Advanced != nil && prop.Advanced.FlagDefinitions != nil {
		for flagName, flagDef := range prop.Advanced.FlagDefinitions {
			// Validate bit position
			if flagDef.Bit >= uint(prop.Length*8) {
				continue
			}

			// Check bit state
			isSet := (intValue & (1 << flagDef.Bit)) != 0

			// Handle inverted logic
			if flagDef.InvertLogic != nil && *flagDef.InvertLogic {
				isSet = !isSet
			}

			flags[flagName] = isSet

			if isSet {
				activeFlags = append(activeFlags, flagName)

				// Group active flags
				if flagDef.Group != "" {
					if groups[flagDef.Group] == nil {
						groups[flagDef.Group] = make([]string, 0)
					}
					groups[flagDef.Group] = append(groups[flagDef.Group], flagName)
				}

				// Check for mutual exclusion conflicts
				for _, exclusiveFlag := range flagDef.MutuallyExclusive {
					if otherFlagDef, exists := prop.Advanced.FlagDefinitions[exclusiveFlag]; exists {
						otherIsSet := (intValue & (1 << otherFlagDef.Bit)) != 0
						if otherFlagDef.InvertLogic != nil && *otherFlagDef.InvertLogic {
							otherIsSet = !otherIsSet
						}
						if otherIsSet {
							conflicts = append(conflicts, fmt.Sprintf("%s conflicts with %s", flagName, exclusiveFlag))
						}
					}
				}
			}
		}
	}

	result := map[string]interface{}{
		"value":        intValue,
		"binary":       fmt.Sprintf("0b%08b", intValue),
		"hex":          fmt.Sprintf("0x%X", intValue),
		"flags":        flags,
		"active_flags": activeFlags,
	}

	if len(groups) > 0 {
		result["groups"] = groups
	}

	if len(conflicts) > 0 {
		result["conflicts"] = conflicts
	}

	return result, nil
}

// ===== ENHANCED TIME PROCESSING =====

// processTime handles enhanced time-based values with multiple formats
func (app *AdvancedPropertyProcessor) processTime(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	// Determine time format and parameters
	format := "frames" // default
	frameRate := 60.0  // default
	_ = "unix"         // default epoch

	if prop.Advanced != nil {
		if prop.Advanced.TimeFormat != "" {
			format = prop.Advanced.TimeFormat
		}
		if prop.Advanced.FrameRate != nil {
			frameRate = *prop.Advanced.FrameRate
		}
		if prop.Advanced.Epoch != "" {
			_ = prop.Advanced.Epoch
		}
	}

	var duration time.Duration
	var timestamp *time.Time

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
		ts := time.Unix(int64(rawValue), 0)
		timestamp = &ts

	case "bcd":
		// Binary Coded Decimal time format
		return app.processBCDTime(uint32(rawValue))

	default:
		// Default to frame-based
		seconds := rawValue / frameRate
		duration = time.Duration(seconds * float64(time.Second))
	}

	result := map[string]interface{}{
		"raw_value": rawValue,
		"format":    format,
	}

	if timestamp != nil {
		result["timestamp"] = timestamp
		result["iso8601"] = timestamp.Format(time.RFC3339)
		result["unix"] = timestamp.Unix()
		result["readable"] = timestamp.Format("2006-01-02 15:04:05")
	} else {
		result["duration"] = duration.String()
		result["seconds"] = duration.Seconds()
		result["minutes"] = duration.Minutes()
		result["hours"] = duration.Hours()
		result["milliseconds"] = duration.Milliseconds()
	}

	if frameRate != 60.0 {
		result["frame_rate"] = frameRate
	}

	return result, nil
}

// processBCDTime processes Binary Coded Decimal time format
func (app *AdvancedPropertyProcessor) processBCDTime(value uint32) (map[string]interface{}, error) {
	// Typical BCD time format: HHMMSS
	hours := int((value >> 16) & 0xFF)
	minutes := int((value >> 8) & 0xFF)
	seconds := int(value & 0xFF)

	// Convert BCD to decimal
	hours = ((hours >> 4) * 10) + (hours & 0x0F)
	minutes = ((minutes >> 4) * 10) + (minutes & 0x0F)
	seconds = ((seconds >> 4) * 10) + (seconds & 0x0F)

	// Validate ranges
	if hours > 23 || minutes > 59 || seconds > 59 {
		return map[string]interface{}{
			"raw_value": value,
			"format":    "bcd",
			"error":     "invalid_bcd_time",
			"hours":     hours,
			"minutes":   minutes,
			"seconds":   seconds,
		}, nil
	}

	totalSeconds := hours*3600 + minutes*60 + seconds

	return map[string]interface{}{
		"raw_value":     value,
		"format":        "bcd",
		"hours":         hours,
		"minutes":       minutes,
		"seconds":       seconds,
		"total_seconds": totalSeconds,
		"readable":      fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds),
		"duration":      time.Duration(totalSeconds) * time.Second,
	}, nil
}

// ===== ENHANCED VERSION PROCESSING =====

// processVersion handles enhanced version numbers with multiple formats
func (app *AdvancedPropertyProcessor) processVersion(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)

	// Determine version format
	format := "major.minor.patch" // default
	if prop.Advanced != nil && prop.Advanced.VersionFormat != "" {
		format = prop.Advanced.VersionFormat
	}

	switch format {
	case "bcd":
		return app.processBCDVersion(intValue)
	case "packed":
		return app.processPackedVersion(intValue)
	case "string":
		return app.processStringVersion(intValue)
	default:
		return app.processStandardVersion(intValue)
	}
}

// processStandardVersion processes standard major.minor.patch version format
func (app *AdvancedPropertyProcessor) processStandardVersion(value uint32) (map[string]interface{}, error) {
	major := (value >> 16) & 0xFF
	minor := (value >> 8) & 0xFF
	patch := value & 0xFF

	return map[string]interface{}{
		"raw_value": value,
		"format":    "major.minor.patch",
		"major":     major,
		"minor":     minor,
		"patch":     patch,
		"string":    fmt.Sprintf("%d.%d.%d", major, minor, patch),
		"sortable":  major*10000 + minor*100 + patch,
	}, nil
}

// processBCDVersion processes Binary Coded Decimal version format
func (app *AdvancedPropertyProcessor) processBCDVersion(value uint32) (map[string]interface{}, error) {
	major := ((value>>20)&0xF)*10 + ((value >> 16) & 0xF)
	minor := ((value>>12)&0xF)*10 + ((value >> 8) & 0xF)
	patch := ((value>>4)&0xF)*10 + (value & 0xF)

	return map[string]interface{}{
		"raw_value": value,
		"format":    "bcd",
		"major":     major,
		"minor":     minor,
		"patch":     patch,
		"string":    fmt.Sprintf("%d.%d.%d", major, minor, patch),
		"sortable":  major*10000 + minor*100 + patch,
	}, nil
}

// processPackedVersion processes packed version format with custom bit allocation
func (app *AdvancedPropertyProcessor) processPackedVersion(value uint32) (map[string]interface{}, error) {
	// Custom packed format: 12 bits major, 10 bits minor, 10 bits patch
	major := (value >> 20) & 0xFFF
	minor := (value >> 10) & 0x3FF
	patch := value & 0x3FF

	return map[string]interface{}{
		"raw_value": value,
		"format":    "packed",
		"major":     major,
		"minor":     minor,
		"patch":     patch,
		"string":    fmt.Sprintf("%d.%d.%d", major, minor, patch),
		"sortable":  major*1000000 + minor*1000 + patch,
	}, nil
}

// processStringVersion processes version stored as string
func (app *AdvancedPropertyProcessor) processStringVersion(value uint32) (map[string]interface{}, error) {
	// Convert uint32 to 4-character string
	versionStr := string([]byte{
		byte((value >> 24) & 0xFF),
		byte((value >> 16) & 0xFF),
		byte((value >> 8) & 0xFF),
		byte(value & 0xFF),
	})

	return map[string]interface{}{
		"raw_value": value,
		"format":    "string",
		"string":    strings.TrimSpace(versionStr),
		"bytes":     []byte(versionStr),
	}, nil
}

// ===== ENHANCED CHECKSUM PROCESSING =====

// processChecksum handles enhanced checksum values with validation
func (app *AdvancedPropertyProcessor) processChecksum(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	checksumValue := uint32(rawValue)

	result := map[string]interface{}{
		"value": checksumValue,
		"hex":   fmt.Sprintf("0x%08X", checksumValue),
	}

	// Enhanced checksum processing with validation
	if prop.Advanced != nil {
		algorithm := "unknown"
		if prop.Advanced.ChecksumAlgorithm != "" {
			algorithm = prop.Advanced.ChecksumAlgorithm
		}

		result["algorithm"] = algorithm

		// Validate checksum if range is specified
		if prop.Advanced.ChecksumRange != nil {
			validation, err := app.validateChecksum(checksumValue, algorithm, prop.Advanced.ChecksumRange)
			if err != nil {
				result["validation_error"] = err.Error()
			} else {
				result["validation"] = validation
			}
		}
	}

	return result, nil
}

// validateChecksum validates a checksum against data in the specified range
func (app *AdvancedPropertyProcessor) validateChecksum(expectedChecksum uint32, algorithm string, checksumRange *ChecksumRange) (map[string]interface{}, error) {
	// Parse range addresses
	startAddr, err := parseAddress(checksumRange.Start)
	if err != nil {
		return nil, fmt.Errorf("invalid start address: %w", err)
	}

	endAddr, err := parseAddress(checksumRange.End)
	if err != nil {
		return nil, fmt.Errorf("invalid end address: %w", err)
	}

	if endAddr <= startAddr {
		return nil, fmt.Errorf("invalid range: end address must be greater than start address")
	}

	// Read data in range
	dataSize := endAddr - startAddr + 1
	data, err := app.memory.ReadBytes(startAddr, dataSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read data for checksum validation: %w", err)
	}

	// Calculate checksum based on algorithm
	var calculatedChecksum uint32
	var checksumErr error

	switch strings.ToLower(algorithm) {
	case "crc32":
		calculatedChecksum = crc32.ChecksumIEEE(data)
	case "crc16":
		calculatedChecksum = uint32(app.calculateCRC16(data))
	case "md5":
		hash := md5.Sum(data)
		calculatedChecksum = binary.BigEndian.Uint32(hash[:4])
	case "sha1":
		hash := sha1.Sum(data)
		calculatedChecksum = binary.BigEndian.Uint32(hash[:4])
	case "simple":
		calculatedChecksum = app.calculateSimpleChecksum(data)
	default:
		return nil, fmt.Errorf("unsupported checksum algorithm: %s", algorithm)
	}

	if checksumErr != nil {
		return nil, checksumErr
	}

	isValid := calculatedChecksum == expectedChecksum

	return map[string]interface{}{
		"expected":    expectedChecksum,
		"calculated":  calculatedChecksum,
		"is_valid":    isValid,
		"algorithm":   algorithm,
		"range_start": startAddr,
		"range_end":   endAddr,
		"data_size":   len(data),
	}, nil
}

// calculateCRC16 calculates a simple CRC16 checksum
func (app *AdvancedPropertyProcessor) calculateCRC16(data []byte) uint16 {
	const polynomial = 0xA001 // CRC16-ANSI polynomial
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ polynomial
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// calculateSimpleChecksum calculates a simple additive checksum
func (app *AdvancedPropertyProcessor) calculateSimpleChecksum(data []byte) uint32 {
	var sum uint32
	for _, b := range data {
		sum += uint32(b)
	}
	return sum
}

// ===== ENHANCED COORDINATE PROCESSING =====

// processCoordinate handles enhanced 2D/3D coordinates with coordinate system support
func (app *AdvancedPropertyProcessor) processCoordinate(prop *Property) (interface{}, error) {
	if prop.Length < 2 {
		return nil, fmt.Errorf("coordinate requires at least 2 components")
	}

	// Determine coordinate system and parameters
	coordinateSystem := "cartesian" // default
	dimensions := uint(2)           // default
	units := "units"                // default

	if prop.Advanced != nil {
		if prop.Advanced.CoordinateSystem != "" {
			coordinateSystem = prop.Advanced.CoordinateSystem
		}
		if prop.Advanced.Dimensions != nil {
			dimensions = *prop.Advanced.Dimensions
		}
		if prop.Advanced.Units != "" {
			units = prop.Advanced.Units
		}
	}

	// Read coordinate components
	components, err := app.readCoordinateComponents(prop, dimensions)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"system":     coordinateSystem,
		"dimensions": dimensions,
		"units":      units,
		"raw":        components,
	}

	// Process based on coordinate system
	switch coordinateSystem {
	case "cartesian":
		result["x"] = components[0]
		if len(components) > 1 {
			result["y"] = components[1]
		}
		if len(components) > 2 {
			result["z"] = components[2]
		}

	case "screen":
		// Screen coordinates (typically Y-inverted)
		result["x"] = components[0]
		if len(components) > 1 {
			result["y"] = components[1]
			result["screen_y"] = components[1] // Can be inverted later if needed
		}

	case "polar":
		if len(components) >= 2 {
			r := components[0]
			theta := components[1] // in radians or degrees
			result["r"] = r
			result["theta"] = theta
			result["radius"] = r
			result["angle"] = theta

			// Convert to cartesian for convenience
			if units == "degrees" {
				theta = theta * math.Pi / 180 // convert to radians
			}
			result["x"] = r * math.Cos(theta)
			result["y"] = r * math.Sin(theta)
		}

	case "geographic":
		if len(components) >= 2 {
			result["latitude"] = components[0]
			result["longitude"] = components[1]
			result["lat"] = components[0]
			result["lon"] = components[1]
			if len(components) > 2 {
				result["altitude"] = components[2]
				result["alt"] = components[2]
			}
		}

	default:
		// Unknown coordinate system, just provide raw components
		for i, component := range components {
			result[fmt.Sprintf("component_%d", i)] = component
		}
	}

	return result, nil
}

// readCoordinateComponents reads coordinate components
func (app *AdvancedPropertyProcessor) readCoordinateComponents(prop *Property, dimensions uint) ([]float64, error) {
	elementSize := uint32(2) // default to 2 bytes per coordinate component
	if prop.Advanced != nil && prop.Advanced.ElementSize != nil {
		elementSize = uint32(*prop.Advanced.ElementSize)
	}

	maxComponents := uint(prop.Length) / uint(elementSize)
	if dimensions > maxComponents {
		dimensions = maxComponents
	}

	components := make([]float64, dimensions)

	for i := uint(0); i < dimensions; i++ {
		address := prop.Address + (uint32(i) * elementSize)

		var value float64
		var err error

		switch elementSize {
		case 1:
			val, readErr := app.memory.ReadUint8(address)
			value = float64(val)
			err = readErr
		case 2:
			val, readErr := app.memory.ReadUint16(address, app.platform.Endian == "little")
			value = float64(val)
			err = readErr
		case 4:
			// Could be float32 or uint32
			val, readErr := app.memory.ReadFloat32(address, app.platform.Endian == "little")
			value = float64(val)
			err = readErr
		case 8:
			val, readErr := app.memory.ReadFloat64(address, app.platform.Endian == "little")
			value = val
			err = readErr
		default:
			val, readErr := app.memory.ReadUint8(address)
			value = float64(val)
			err = readErr
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read coordinate component %d: %w", i, err)
		}

		components[i] = value
	}

	return components, nil
}

// ===== ENHANCED COLOR PROCESSING =====

// processColor handles enhanced color values with multiple formats and color space support
func (app *AdvancedPropertyProcessor) processColor(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	intValue := uint32(rawValue)

	// Determine color format
	format := "rgb565" // default
	hasAlpha := false

	if prop.Advanced != nil {
		if prop.Advanced.ColorFormat != "" {
			format = prop.Advanced.ColorFormat
		}
		if prop.Advanced.AlphaChannel != nil {
			hasAlpha = *prop.Advanced.AlphaChannel
		}
	}

	var r, g, b, a uint8
	var paletteIndex *uint8

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
		hasAlpha = true

	case "rgba8888":
		r = uint8((intValue >> 24) & 0xFF)
		g = uint8((intValue >> 16) & 0xFF)
		b = uint8((intValue >> 8) & 0xFF)
		a = uint8(intValue & 0xFF)
		hasAlpha = true

	case "rgb888":
		r = uint8((intValue >> 16) & 0xFF)
		g = uint8((intValue >> 8) & 0xFF)
		b = uint8(intValue & 0xFF)
		a = 255

	case "palette":
		// Palette-based color
		paletteIdx := uint8(intValue & 0xFF)
		paletteIndex = &paletteIdx

		// TODO: Look up actual RGB values from palette
		// For now, generate a default color based on index
		r = paletteIdx
		g = paletteIdx
		b = paletteIdx
		a = 255

	case "yuv":
		// YUV color space (simplified conversion)
		y := float64((intValue >> 16) & 0xFF)
		u := float64((intValue>>8)&0xFF) - 128
		v := float64(intValue&0xFF) - 128

		// Convert YUV to RGB
		rFloat := y + 1.402*v
		gFloat := y - 0.344*u - 0.714*v
		bFloat := y + 1.772*u

		// Clamp to 0-255 range
		r = uint8(math.Max(0, math.Min(255, rFloat)))
		g = uint8(math.Max(0, math.Min(255, gFloat)))
		b = uint8(math.Max(0, math.Min(255, bFloat)))
		a = 255

	default:
		// Default to RGB565
		r = uint8((intValue>>11)&0x1F) << 3
		g = uint8((intValue>>5)&0x3F) << 2
		b = uint8(intValue&0x1F) << 3
		a = 255
	}

	result := map[string]interface{}{
		"raw_value": intValue,
		"format":    format,
		"r":         r,
		"g":         g,
		"b":         b,
		"hex":       fmt.Sprintf("#%02X%02X%02X", r, g, b),
	}

	if hasAlpha {
		result["a"] = a
		result["hex_alpha"] = fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, a)
	}

	if paletteIndex != nil {
		result["palette_index"] = *paletteIndex
	}

	// Add color analysis
	result["brightness"] = app.calculateBrightness(r, g, b)
	result["hsl"] = app.rgbToHSL(r, g, b)

	// Add palette reference if specified
	if prop.Advanced != nil && prop.Advanced.PaletteRef != "" {
		result["palette_ref"] = prop.Advanced.PaletteRef
	}

	return result, nil
}

// calculateBrightness calculates perceived brightness of a color
func (app *AdvancedPropertyProcessor) calculateBrightness(r, g, b uint8) float64 {
	// Use luminance formula: 0.299*R + 0.587*G + 0.114*B
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0
}

// rgbToHSL converts RGB to HSL color space
func (app *AdvancedPropertyProcessor) rgbToHSL(r, g, b uint8) map[string]float64 {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	var h, s, l float64
	l = (max + min) / 2

	if max == min {
		h = 0 // achromatic
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}

	return map[string]float64{
		"h": h * 360, // Convert to degrees
		"s": s * 100, // Convert to percentage
		"l": l * 100, // Convert to percentage
	}
}

// ===== ENHANCED PERCENTAGE PROCESSING =====

// processPercentage handles enhanced percentage values with custom ranges and precision
func (app *AdvancedPropertyProcessor) processPercentage(prop *Property) (interface{}, error) {
	rawValue, err := app.readBasicValue(prop)
	if err != nil {
		return nil, err
	}

	// Determine maximum value and precision
	maxValue := 100.0    // default
	precision := uint(2) // default

	if prop.Advanced != nil {
		if prop.Advanced.MaxValue != nil {
			maxValue = *prop.Advanced.MaxValue
		}
		if prop.Advanced.Precision != nil {
			precision = *prop.Advanced.Precision
		}
	}

	percentage := (rawValue / maxValue) * 100
	decimal := rawValue / maxValue

	result := map[string]interface{}{
		"raw_value":  rawValue,
		"percentage": app.roundToPlaces(percentage, precision),
		"decimal":    app.roundToPlaces(decimal, precision+2),
		"max_value":  maxValue,
		"precision":  precision,
	}

	// Add percentage category
	if percentage >= 90 {
		result["category"] = "excellent"
	} else if percentage >= 75 {
		result["category"] = "good"
	} else if percentage >= 50 {
		result["category"] = "fair"
	} else if percentage >= 25 {
		result["category"] = "poor"
	} else {
		result["category"] = "critical"
	}

	// Add progress bar representation
	barLength := 20
	filledBars := int((percentage / 100) * float64(barLength))
	emptyBars := barLength - filledBars
	progressBar := strings.Repeat("█", filledBars) + strings.Repeat("░", emptyBars)
	result["progress_bar"] = progressBar

	return result, nil
}

// roundToPlaces rounds a float64 to the specified number of decimal places
func (app *AdvancedPropertyProcessor) roundToPlaces(value float64, places uint) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}

// ===== FLOAT PROCESSING (ENHANCED) =====

// processFloat32 handles enhanced 32-bit floating point values
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
	floatValue := float64(value)
	if prop.Transform != nil {
		if prop.Transform.Multiply != nil {
			floatValue *= *prop.Transform.Multiply
		}
		if prop.Transform.Add != nil {
			floatValue += *prop.Transform.Add
		}
	}

	result := map[string]interface{}{
		"value":     float32(floatValue),
		"raw_bits":  bits,
		"raw_hex":   fmt.Sprintf("0x%08X", bits),
		"is_finite": math.IsInf(floatValue, 0) == false && math.IsNaN(floatValue) == false,
		"is_nan":    math.IsNaN(floatValue),
		"is_inf":    math.IsInf(floatValue, 0),
	}

	if math.IsInf(floatValue, 0) {
		if math.IsInf(floatValue, 1) {
			result["infinity_type"] = "positive"
		} else {
			result["infinity_type"] = "negative"
		}
	}

	return result, nil
}

// processFloat64 handles enhanced 64-bit floating point values
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

	result := map[string]interface{}{
		"value":     value,
		"raw_bits":  bits,
		"raw_hex":   fmt.Sprintf("0x%016X", bits),
		"is_finite": math.IsInf(value, 0) == false && math.IsNaN(value) == false,
		"is_nan":    math.IsNaN(value),
		"is_inf":    math.IsInf(value, 0),
	}

	if math.IsInf(value, 0) {
		if math.IsInf(value, 1) {
			result["infinity_type"] = "positive"
		} else {
			result["infinity_type"] = "negative"
		}
	}

	return result, nil
}

// ===== HELPER METHODS =====

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
