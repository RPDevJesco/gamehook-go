package mappers

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"fmt"
	"gamehook/internal/drivers"
	"gamehook/internal/memory"
	"io/fs" // Add this
	"os"    // Add this
	"path/filepath"
	"strconv"
	"strings"
)

// Loader handles loading and parsing CUE mapper files
type Loader struct {
	mappersDir string
	mappers    map[string]*Mapper
}

// Property represents a parsed property definition
type Property struct {
	Name        string
	Type        string
	Address     uint32
	Length      uint32
	Endian      string
	Description string
	Transform   *Transform
	CharMap     map[uint8]string
}

// Transform represents value transformation rules
type Transform struct {
	Multiply *float64
	Add      *float64
	Lookup   map[string]string
}

// Platform represents platform configuration
type Platform struct {
	Name         string
	Endian       string
	MemoryBlocks []drivers.MemoryBlock
}

// Mapper represents a complete mapper with properties
type Mapper struct {
	Name       string
	Game       string
	Platform   Platform
	Properties map[string]*Property
}

// NewLoader creates a new mapper loader
func NewLoader(mappersDir string) *Loader {
	return &Loader{
		mappersDir: mappersDir,
		mappers:    make(map[string]*Mapper),
	}
}

// List returns available mapper names
func (l *Loader) List() []string {
	names := make([]string, 0)

	// Walk through all subdirectories to find .cue files
	err := filepath.WalkDir(l.mappersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors, just continue
		}

		// Check if it's a .cue file
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".cue") {
			// Get relative path from mappers directory
			relPath, err := filepath.Rel(l.mappersDir, path)
			if err != nil {
				return nil
			}

			// Convert path to mapper name (remove .cue extension and normalize separators)
			mapperName := strings.TrimSuffix(relPath, ".cue")
			mapperName = strings.ReplaceAll(mapperName, "\\", "/") // Normalize to forward slashes

			names = append(names, mapperName)
		}

		return nil
	})

	if err != nil {
		// If we can't walk the directory, return empty list
		return []string{}
	}

	return names
}

// Load loads a mapper by name
func (l *Loader) Load(name string) (*Mapper, error) {
	// Check if already loaded
	if mapper, exists := l.mappers[name]; exists {
		return mapper, nil
	}

	// Convert name to file path (handle subdirectories)
	// Examples:
	//   "nes/super_mario_bros" -> "mappers/nes/super_mario_bros.cue"
	//   "super_mario_bros" -> "mappers/super_mario_bros.cue"
	filePath := filepath.Join(l.mappersDir, name+".cue")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mapper file not found: %s", filePath)
	}

	// Load and parse CUE file
	mapper, err := l.loadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Set the name from the file path if not specified in CUE
	if mapper.Name == "" {
		mapper.Name = name
	}

	// Cache mapper
	l.mappers[name] = mapper
	return mapper, nil
}

// loadFromFile loads a mapper from a CUE file
func (l *Loader) loadFromFile(filePath string) (*Mapper, error) {
	// Load CUE instance
	ctx := cuecontext.New()

	// Load the file
	buildInstances := load.Instances([]string{filePath}, &load.Config{})
	if len(buildInstances) == 0 {
		return nil, fmt.Errorf("no CUE instances found in %s", filePath)
	}

	inst := buildInstances[0]
	if inst.Err != nil {
		return nil, fmt.Errorf("CUE load error: %w", inst.Err)
	}

	value := ctx.BuildInstance(inst)
	if value.Err() != nil {
		return nil, fmt.Errorf("CUE build error: %w", value.Err())
	}

	// Parse the mapper
	return l.parseMapper(value)
}

// parseMapper parses a CUE value into a Mapper struct
func (l *Loader) parseMapper(value cue.Value) (*Mapper, error) {
	mapper := &Mapper{
		Properties: make(map[string]*Property),
	}

	// Parse mapper name and game
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		mapper.Name = name
	}

	if game, err := value.LookupPath(cue.ParsePath("game")).String(); err == nil {
		mapper.Game = game
	}

	// Parse platform
	platformValue := value.LookupPath(cue.ParsePath("platform"))
	if platformValue.Exists() {
		platform, err := l.parsePlatform(platformValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse platform: %w", err)
		}
		mapper.Platform = platform
	}

	// Parse properties
	propertiesValue := value.LookupPath(cue.ParsePath("properties"))
	if propertiesValue.Exists() {
		if err := l.parseProperties(propertiesValue, mapper.Properties); err != nil {
			return nil, fmt.Errorf("failed to parse properties: %w", err)
		}
	}

	return mapper, nil
}

// parsePlatform parses platform configuration
func (l *Loader) parsePlatform(value cue.Value) (Platform, error) {
	platform := Platform{}

	// Parse platform name
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		platform.Name = name
	}

	// Parse endianness
	if endian, err := value.LookupPath(cue.ParsePath("endian")).String(); err == nil {
		platform.Endian = endian
	}

	// Parse memory blocks
	blocksValue := value.LookupPath(cue.ParsePath("memoryBlocks"))
	if blocksValue.Exists() {
		blocksIter, _ := blocksValue.List()
		for blocksIter.Next() {
			blockValue := blocksIter.Value()

			var block drivers.MemoryBlock

			if name, err := blockValue.LookupPath(cue.ParsePath("name")).String(); err == nil {
				block.Name = name
			}

			if startStr, err := blockValue.LookupPath(cue.ParsePath("start")).String(); err == nil {
				if start, err := parseAddress(startStr); err == nil {
					block.Start = start
				}
			}

			if endStr, err := blockValue.LookupPath(cue.ParsePath("end")).String(); err == nil {
				if end, err := parseAddress(endStr); err == nil {
					block.End = end
				}
			}

			platform.MemoryBlocks = append(platform.MemoryBlocks, block)
		}
	}

	return platform, nil
}

// parseProperties parses property definitions
func (l *Loader) parseProperties(value cue.Value, properties map[string]*Property) error {
	fields, _ := value.Fields()
	for fields.Next() {
		label := fields.Label()
		propertyValue := fields.Value()

		property, err := l.parseProperty(propertyValue)
		if err != nil {
			return fmt.Errorf("failed to parse property %s: %w", label, err)
		}

		properties[label] = property
	}

	return nil
}

// parseProperty parses a single property definition
func (l *Loader) parseProperty(value cue.Value) (*Property, error) {
	property := &Property{}

	// Parse name
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		property.Name = name
	}

	// Parse type
	if propType, err := value.LookupPath(cue.ParsePath("type")).String(); err == nil {
		property.Type = propType
	}

	// Parse address
	if addressStr, err := value.LookupPath(cue.ParsePath("address")).String(); err == nil {
		if address, err := parseAddress(addressStr); err == nil {
			property.Address = address
		} else {
			return nil, fmt.Errorf("invalid address %s: %w", addressStr, err)
		}
	}

	// Parse optional length
	if length, err := value.LookupPath(cue.ParsePath("length")).Uint64(); err == nil {
		property.Length = uint32(length)
	} else {
		property.Length = 1 // default length
	}

	// Parse optional endian
	if endian, err := value.LookupPath(cue.ParsePath("endian")).String(); err == nil {
		property.Endian = endian
	}

	// Parse optional description
	if desc, err := value.LookupPath(cue.ParsePath("description")).String(); err == nil {
		property.Description = desc
	}

	// Parse transform
	transformValue := value.LookupPath(cue.ParsePath("transform"))
	if transformValue.Exists() {
		transform, err := l.parseTransform(transformValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transform: %w", err)
		}
		property.Transform = transform
	}

	// Parse character map for strings
	charMapValue := value.LookupPath(cue.ParsePath("charMap"))
	if charMapValue.Exists() {
		charMap, err := l.parseCharMap(charMapValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse charMap: %w", err)
		}
		property.CharMap = charMap
	}

	return property, nil
}

// parseTransform parses transformation rules
func (l *Loader) parseTransform(value cue.Value) (*Transform, error) {
	transform := &Transform{}

	// Parse multiply
	if multiply, err := value.LookupPath(cue.ParsePath("multiply")).Float64(); err == nil {
		transform.Multiply = &multiply
	}

	// Parse add
	if add, err := value.LookupPath(cue.ParsePath("add")).Float64(); err == nil {
		transform.Add = &add
	}

	// Parse lookup table
	lookupValue := value.LookupPath(cue.ParsePath("lookup"))
	if lookupValue.Exists() {
		lookup := make(map[string]string)
		fields, _ := lookupValue.Fields()
		for fields.Next() {
			key := fields.Label()
			if val, err := fields.Value().String(); err == nil {
				lookup[key] = val
			}
		}
		if len(lookup) > 0 {
			transform.Lookup = lookup
		}
	}

	return transform, nil
}

// parseCharMap parses character mapping for strings
func (l *Loader) parseCharMap(value cue.Value) (map[uint8]string, error) {
	charMap := make(map[uint8]string)

	fields, _ := value.Fields()
	for fields.Next() {
		keyStr := fields.Label()

		// Parse hex key to uint8
		var key uint64
		var err error
		if strings.HasPrefix(keyStr, "0x") || strings.HasPrefix(keyStr, "0X") {
			key, err = strconv.ParseUint(keyStr[2:], 16, 8)
		} else {
			key, err = strconv.ParseUint(keyStr, 10, 8)
		}

		if err != nil {
			return nil, fmt.Errorf("invalid character key %s: %w", keyStr, err)
		}

		if val, err := fields.Value().String(); err == nil {
			charMap[uint8(key)] = val
		}
	}

	return charMap, nil
}

// parseAddress parses a hex address string to uint32
func parseAddress(addressStr string) (uint32, error) {
	if strings.HasPrefix(addressStr, "0x") || strings.HasPrefix(addressStr, "0X") {
		value, err := strconv.ParseUint(addressStr[2:], 16, 32)
		return uint32(value), err
	}

	value, err := strconv.ParseUint(addressStr, 10, 32)
	return uint32(value), err
}

// ProcessProperties processes all properties and updates their values
func (m *Mapper) ProcessProperties(memManager *memory.Manager) error {
	// This would trigger change notifications in a real implementation
	// For now, just validate that we can read all properties
	for name := range m.Properties {
		_, err := m.GetProperty(name, memManager)
		if err != nil {
			return fmt.Errorf("failed to process property %s: %w", name, err)
		}
	}
	return nil
}

// GetProperty gets a property value from memory
func (m *Mapper) GetProperty(name string, memManager *memory.Manager) (interface{}, error) {
	prop, exists := m.Properties[name]
	if !exists {
		return nil, fmt.Errorf("property %s not found", name)
	}

	// Determine endianness
	littleEndian := true
	if prop.Endian != "" {
		littleEndian = prop.Endian == "little"
	} else {
		littleEndian = m.Platform.Endian == "little"
	}

	// Read value based on type
	var rawValue interface{}
	var err error

	switch prop.Type {
	case "uint8":
		rawValue, err = memManager.ReadUint8(prop.Address)
	case "uint16":
		rawValue, err = memManager.ReadUint16(prop.Address, littleEndian)
	case "uint32":
		rawValue, err = memManager.ReadUint32(prop.Address, littleEndian)
	case "int8":
		rawValue, err = memManager.ReadInt8(prop.Address)
	case "int16":
		rawValue, err = memManager.ReadInt16(prop.Address, littleEndian)
	case "int32":
		rawValue, err = memManager.ReadInt32(prop.Address, littleEndian)
	case "bool":
		rawValue, err = memManager.ReadBool(prop.Address)
	case "string":
		rawValue, err = memManager.ReadString(prop.Address, prop.Length, prop.CharMap)
	case "bitfield":
		rawValue, err = memManager.ReadBitfield(prop.Address, prop.Length)
	default:
		return nil, fmt.Errorf("unsupported property type: %s", prop.Type)
	}

	if err != nil {
		return nil, err
	}

	// Apply transformations
	return m.applyTransform(rawValue, prop.Transform), nil
}

// SetProperty sets a property value in memory
func (m *Mapper) SetProperty(name string, value interface{}, memManager *memory.Manager, driver drivers.Driver) error {
	prop, exists := m.Properties[name]
	if !exists {
		return fmt.Errorf("property %s not found", name)
	}

	// Convert value to bytes based on type
	var data []byte

	switch prop.Type {
	case "uint8":
		if val, ok := value.(uint8); ok {
			data = memManager.WriteUint8(prop.Address, val)
		} else {
			return fmt.Errorf("invalid type for uint8 property")
		}
	case "bool":
		if val, ok := value.(bool); ok {
			data = memManager.WriteBool(prop.Address, val)
		} else {
			return fmt.Errorf("invalid type for bool property")
		}
	// Add more types as needed
	default:
		return fmt.Errorf("setting values for type %s not yet implemented", prop.Type)
	}

	// Write to emulator
	return driver.WriteBytes(prop.Address, data)
}

// applyTransform applies transformation rules to a raw value
func (m *Mapper) applyTransform(value interface{}, transform *Transform) interface{} {
	if transform == nil {
		return value
	}

	// Convert to float64 for numeric operations
	var numVal float64
	var isNumeric bool

	switch v := value.(type) {
	case uint8:
		numVal = float64(v)
		isNumeric = true
	case uint16:
		numVal = float64(v)
		isNumeric = true
	case uint32:
		numVal = float64(v)
		isNumeric = true
	case int8:
		numVal = float64(v)
		isNumeric = true
	case int16:
		numVal = float64(v)
		isNumeric = true
	case int32:
		numVal = float64(v)
		isNumeric = true
	}

	// Apply numeric transformations
	if isNumeric {
		if transform.Multiply != nil {
			numVal *= *transform.Multiply
		}
		if transform.Add != nil {
			numVal += *transform.Add
		}
	}

	// Apply lookup transformation
	if transform.Lookup != nil {
		key := fmt.Sprintf("%.0f", numVal)
		if mapped, exists := transform.Lookup[key]; exists {
			return mapped
		}
	}

	if isNumeric {
		return numVal
	}
	return value
}
