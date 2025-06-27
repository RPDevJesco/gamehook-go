package mappers

import (
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"encoding/binary"
	"fmt"
	"gamehook/internal/drivers"
	"gamehook/internal/memory"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// PropertyType represents the enhanced property types
type PropertyType string

const (
	// Basic types
	PropertyTypeUint8    PropertyType = "uint8"
	PropertyTypeUint16   PropertyType = "uint16"
	PropertyTypeUint32   PropertyType = "uint32"
	PropertyTypeInt8     PropertyType = "int8"
	PropertyTypeInt16    PropertyType = "int16"
	PropertyTypeInt32    PropertyType = "int32"
	PropertyTypeString   PropertyType = "string"
	PropertyTypeBool     PropertyType = "bool"
	PropertyTypeBitfield PropertyType = "bitfield"
	PropertyTypeBCD      PropertyType = "bcd"

	// Advanced types
	PropertyTypeBit        PropertyType = "bit"
	PropertyTypeNibble     PropertyType = "nibble"
	PropertyTypeFloat32    PropertyType = "float32"
	PropertyTypeFloat64    PropertyType = "float64"
	PropertyTypePointer    PropertyType = "pointer"
	PropertyTypeArray      PropertyType = "array"
	PropertyTypeStruct     PropertyType = "struct"
	PropertyTypeEnum       PropertyType = "enum"
	PropertyTypeFlags      PropertyType = "flags"
	PropertyTypeTime       PropertyType = "time"
	PropertyTypeVersion    PropertyType = "version"
	PropertyTypeChecksum   PropertyType = "checksum"
	PropertyTypeCoordinate PropertyType = "coordinate"
	PropertyTypeColor      PropertyType = "color"
	PropertyTypePercentage PropertyType = "percentage"
)

// PropertyValidation represents validation constraints for a property
type PropertyValidation struct {
	MinValue      *float64      `json:"min_value,omitempty"`
	MaxValue      *float64      `json:"max_value,omitempty"`
	AllowedValues []interface{} `json:"allowed_values,omitempty"`
	Pattern       string        `json:"pattern,omitempty"`
	Required      bool          `json:"required"`
	Constraint    string        `json:"constraint,omitempty"` // CUE expression
}

// RangeTransform represents value range mapping
type RangeTransform struct {
	InputMin  float64 `json:"input_min"`
	InputMax  float64 `json:"input_max"`
	OutputMin float64 `json:"output_min"`
	OutputMax float64 `json:"output_max"`
	Clamp     bool    `json:"clamp"`
}

// ConditionalTransform represents conditional value transformation
type ConditionalTransform struct {
	If   string      `json:"if"`             // CUE condition
	Then interface{} `json:"then"`           // value if condition is true
	Else interface{} `json:"else,omitempty"` // value if condition is false
}

// StringOperations represents string transformation operations
type StringOperations struct {
	Trim      bool              `json:"trim"`
	Uppercase bool              `json:"uppercase"`
	Lowercase bool              `json:"lowercase"`
	Replace   map[string]string `json:"replace"`
}

// Transform represents enhanced value transformation rules
type Transform struct {
	// Simple arithmetic
	Multiply *float64 `json:"multiply,omitempty"`
	Add      *float64 `json:"add,omitempty"`
	Divide   *float64 `json:"divide,omitempty"`
	Subtract *float64 `json:"subtract,omitempty"`

	// CUE expressions
	Expression string `json:"expression,omitempty"`

	// Conditional transformations
	Conditions []ConditionalTransform `json:"conditions,omitempty"`

	// Lookup tables
	Lookup map[string]string `json:"lookup,omitempty"`

	// Range mapping
	Range *RangeTransform `json:"range,omitempty"`

	// String operations
	StringOps *StringOperations `json:"string_ops,omitempty"`
}

// ComputedProperty represents a property derived from other properties
type ComputedProperty struct {
	Expression   string       `json:"expression"`
	Dependencies []string     `json:"dependencies"`
	Type         PropertyType `json:"type,omitempty"`
}

// Property represents an enhanced parsed property definition
type Property struct {
	Name        string
	Type        PropertyType
	Address     uint32
	Length      uint32
	Position    *uint32 // for bit/nibble properties
	Size        *uint32 // element size for arrays/structs
	Endian      string
	Description string
	ReadOnly    bool
	Transform   *Transform
	Validation  *PropertyValidation
	CharMap     map[uint8]string

	// Freezing support
	Freezable     bool
	DefaultFrozen bool
	Frozen        bool
	FrozenData    []byte

	// Computed properties
	DependsOn []string
	Computed  *ComputedProperty

	// Custom expressions
	ReadExpression  string
	WriteExpression string
}

// PropertyGroup represents a group of related properties for UI organization
type PropertyGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Icon        string   `json:"icon,omitempty"`
	Properties  []string `json:"properties"`
	Collapsed   bool     `json:"collapsed"`
}

// Platform represents enhanced platform configuration
type Platform struct {
	Name          string
	Endian        string
	MemoryBlocks  []drivers.MemoryBlock
	Constants     map[string]interface{} // Platform-specific constants
	BaseAddresses map[string]string      // Named base addresses
}

// Mapper represents an enhanced complete mapper with properties
type Mapper struct {
	Name        string
	Game        string
	Version     string
	MinVersion  string // Minimum GameHook version
	Platform    Platform
	Properties  map[string]*Property
	Groups      map[string]*PropertyGroup
	Computed    map[string]*ComputedProperty
	Constants   map[string]interface{} // Global constants
	Preprocess  []string               // CUE expressions run before property evaluation
	Postprocess []string               // CUE expressions run after property evaluation
}

// Loader handles loading and parsing enhanced CUE mapper files
type Loader struct {
	mappersDir string
	mappers    map[string]*Mapper
}

// NewLoader creates a new enhanced mapper loader
func NewLoader(mappersDir string) *Loader {
	return &Loader{
		mappersDir: mappersDir,
		mappers:    make(map[string]*Mapper),
	}
}

// List returns available mapper names
func (l *Loader) List() []string {
	names := make([]string, 0)

	err := filepath.WalkDir(l.mappersDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".cue") {
			relPath, err := filepath.Rel(l.mappersDir, path)
			if err != nil {
				return nil
			}

			mapperName := strings.TrimSuffix(relPath, ".cue")
			mapperName = strings.ReplaceAll(mapperName, "\\", "/")
			names = append(names, mapperName)
		}

		return nil
	})

	if err != nil {
		return []string{}
	}

	return names
}

// Load loads a mapper by name
func (l *Loader) Load(name string) (*Mapper, error) {
	if mapper, exists := l.mappers[name]; exists {
		return mapper, nil
	}

	filePath := filepath.Join(l.mappersDir, name+".cue")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("mapper file not found: %s", filePath)
	}

	mapper, err := l.loadFromFile(filePath)
	if err != nil {
		return nil, err
	}

	if mapper.Name == "" {
		mapper.Name = name
	}

	l.mappers[name] = mapper
	return mapper, nil
}

// loadFromFile loads a mapper from a CUE file
func (l *Loader) loadFromFile(filePath string) (*Mapper, error) {
	log.Printf("🔍 Loading mapper from file: %s", filePath)

	ctx := cuecontext.New()

	buildInstances := load.Instances([]string{filePath}, &load.Config{})
	if len(buildInstances) == 0 {
		return nil, fmt.Errorf("no CUE instances found in %s", filePath)
	}

	inst := buildInstances[0]
	if inst.Err != nil {
		log.Printf("❌ CUE load error for %s: %v", filePath, inst.Err)
		return nil, fmt.Errorf("CUE load error: %w", inst.Err)
	}

	log.Printf("✅ CUE instance loaded, building value...")
	value := ctx.BuildInstance(inst)
	if value.Err() != nil {
		log.Printf("❌ CUE build error for %s: %v", filePath, value.Err())
		// Print detailed error information
		if err := value.Validate(); err != nil {
			log.Printf("❌ CUE validation errors:")
			for _, e := range errors.Errors(err) {
				log.Printf("   • %s", e)
			}
		}
		return nil, fmt.Errorf("CUE build error: %w", value.Err())
	}

	log.Printf("✅ CUE value built successfully, parsing mapper...")
	mapper, err := l.parseMapper(value)
	if err != nil {
		log.Printf("❌ Mapper parsing error: %v", err)
		return nil, err
	}

	log.Printf("✅ Mapper parsed: %d properties, %d groups, %d computed",
		len(mapper.Properties), len(mapper.Groups), len(mapper.Computed))

	return mapper, nil
}

// parseMapper parses a CUE value into an enhanced Mapper struct
func (l *Loader) parseMapper(value cue.Value) (*Mapper, error) {
	mapper := &Mapper{
		Properties: make(map[string]*Property),
		Groups:     make(map[string]*PropertyGroup),
		Computed:   make(map[string]*ComputedProperty),
		Constants:  make(map[string]interface{}),
	}

	// Parse mapper metadata with debug info
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		mapper.Name = name
		log.Printf("📋 Mapper name: %s", name)
	} else {
		log.Printf("⚠️  Could not parse mapper name: %v", err)
	}

	if game, err := value.LookupPath(cue.ParsePath("game")).String(); err == nil {
		mapper.Game = game
		log.Printf("🎮 Game: %s", game)
	}

	if version, err := value.LookupPath(cue.ParsePath("version")).String(); err == nil {
		mapper.Version = version
		log.Printf("🔖 Version: %s", version)
	}

	// Parse platform with debug info
	platformValue := value.LookupPath(cue.ParsePath("platform"))
	if platformValue.Exists() {
		log.Printf("🔧 Parsing platform...")
		platform, err := l.parseEnhancedPlatform(platformValue)
		if err != nil {
			log.Printf("❌ Platform parsing error: %v", err)
			return nil, fmt.Errorf("failed to parse platform: %w", err)
		}
		mapper.Platform = platform
		log.Printf("✅ Platform parsed: %s with %d memory blocks", platform.Name, len(platform.MemoryBlocks))
	} else {
		log.Printf("⚠️  No platform section found")
	}

	// Parse properties with detailed debug info
	propertiesValue := value.LookupPath(cue.ParsePath("properties"))
	if propertiesValue.Exists() {
		log.Printf("📊 Parsing properties...")
		if err := l.parseEnhancedPropertiesWithDebug(propertiesValue, mapper.Properties); err != nil {
			log.Printf("❌ Properties parsing error: %v", err)
			return nil, fmt.Errorf("failed to parse properties: %w", err)
		}
		log.Printf("✅ Parsed %d properties", len(mapper.Properties))
	} else {
		log.Printf("⚠️  No properties section found")
	}

	// Parse groups with debug info
	groupsValue := value.LookupPath(cue.ParsePath("groups"))
	if groupsValue.Exists() {
		log.Printf("📂 Parsing groups...")
		if err := l.parsePropertyGroups(groupsValue, mapper.Groups); err != nil {
			log.Printf("❌ Groups parsing error: %v", err)
			return nil, fmt.Errorf("failed to parse groups: %w", err)
		}
		log.Printf("✅ Parsed %d groups", len(mapper.Groups))
	} else {
		log.Printf("⚠️  No groups section found")
	}

	// Parse computed properties with debug info
	computedValue := value.LookupPath(cue.ParsePath("computed"))
	if computedValue.Exists() {
		log.Printf("🧮 Parsing computed properties...")
		if err := l.parseComputedProperties(computedValue, mapper.Computed); err != nil {
			log.Printf("❌ Computed properties parsing error: %v", err)
			return nil, fmt.Errorf("failed to parse computed properties: %w", err)
		}
		log.Printf("✅ Parsed %d computed properties", len(mapper.Computed))
	} else {
		log.Printf("⚠️  No computed section found")
	}

	return mapper, nil
}

func (l *Loader) parseEnhancedPropertiesWithDebug(value cue.Value, properties map[string]*Property) error {
	fields, _ := value.Fields()
	propertyCount := 0

	for fields.Next() {
		label := fields.Label()
		propertyValue := fields.Value()

		log.Printf("   🔍 Parsing property: %s", label)

		property, err := l.parseEnhancedProperty(propertyValue)
		if err != nil {
			log.Printf("   ❌ Failed to parse property %s: %v", label, err)
			return fmt.Errorf("failed to parse property %s: %w", label, err)
		}

		property.Name = label
		properties[label] = property
		propertyCount++

		log.Printf("   ✅ Property %s: type=%s, address=%s", label, property.Type, fmt.Sprintf("0x%X", property.Address))
	}

	log.Printf("📊 Total properties parsed: %d", propertyCount)
	return nil
}

// parseConstants parses constants from CUE value
func (l *Loader) parseConstants(value cue.Value, constants map[string]interface{}) {
	fields, _ := value.Fields()
	for fields.Next() {
		label := fields.Label()
		fieldValue := fields.Value()

		// Try different types
		if str, err := fieldValue.String(); err == nil {
			constants[label] = str
		} else if num, err := fieldValue.Float64(); err == nil {
			constants[label] = num
		} else if b, err := fieldValue.Bool(); err == nil {
			constants[label] = b
		}
	}
}

// parseEnhancedPlatform parses enhanced platform configuration
func (l *Loader) parseEnhancedPlatform(value cue.Value) (Platform, error) {
	platform := Platform{
		Constants:     make(map[string]interface{}),
		BaseAddresses: make(map[string]string),
	}

	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		platform.Name = name
	}

	if endian, err := value.LookupPath(cue.ParsePath("endian")).String(); err == nil {
		platform.Endian = endian
	}

	// Parse constants
	constantsValue := value.LookupPath(cue.ParsePath("constants"))
	if constantsValue.Exists() {
		l.parseConstants(constantsValue, platform.Constants)
	}

	// Parse base addresses
	baseAddrsValue := value.LookupPath(cue.ParsePath("baseAddresses"))
	if baseAddrsValue.Exists() {
		fields, _ := baseAddrsValue.Fields()
		for fields.Next() {
			label := fields.Label()
			if addr, err := fields.Value().String(); err == nil {
				platform.BaseAddresses[label] = addr
			}
		}
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

// parseEnhancedProperties parses enhanced property definitions
func (l *Loader) parseEnhancedProperties(value cue.Value, properties map[string]*Property) error {
	fields, _ := value.Fields()
	for fields.Next() {
		label := fields.Label()
		propertyValue := fields.Value()

		property, err := l.parseEnhancedProperty(propertyValue)
		if err != nil {
			return fmt.Errorf("failed to parse property %s: %w", label, err)
		}

		property.Name = label
		properties[label] = property
	}

	return nil
}

// parseEnhancedProperty parses a single enhanced property definition
func (l *Loader) parseEnhancedProperty(value cue.Value) (*Property, error) {
	property := &Property{}

	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		property.Name = name
	}

	if propType, err := value.LookupPath(cue.ParsePath("type")).String(); err == nil {
		property.Type = PropertyType(propType)
	}

	if addressStr, err := value.LookupPath(cue.ParsePath("address")).String(); err == nil {
		if address, err := parseAddress(addressStr); err == nil {
			property.Address = address
		} else {
			return nil, fmt.Errorf("invalid address %s: %w", addressStr, err)
		}
	}

	// Parse optional fields
	if length, err := value.LookupPath(cue.ParsePath("length")).Uint64(); err == nil {
		property.Length = uint32(length)
	} else {
		property.Length = 1
	}

	if position, err := value.LookupPath(cue.ParsePath("position")).Uint64(); err == nil {
		pos := uint32(position)
		property.Position = &pos
	}

	if size, err := value.LookupPath(cue.ParsePath("size")).Uint64(); err == nil {
		sz := uint32(size)
		property.Size = &sz
	}

	if endian, err := value.LookupPath(cue.ParsePath("endian")).String(); err == nil {
		property.Endian = endian
	}

	if desc, err := value.LookupPath(cue.ParsePath("description")).String(); err == nil {
		property.Description = desc
	}

	if readOnly, err := value.LookupPath(cue.ParsePath("readOnly")).Bool(); err == nil {
		property.ReadOnly = readOnly
	}

	if freezable, err := value.LookupPath(cue.ParsePath("freezable")).Bool(); err == nil {
		property.Freezable = freezable
	}

	if defaultFrozen, err := value.LookupPath(cue.ParsePath("defaultFrozen")).Bool(); err == nil {
		property.DefaultFrozen = defaultFrozen
	}

	if readExpr, err := value.LookupPath(cue.ParsePath("readExpression")).String(); err == nil {
		property.ReadExpression = readExpr
	}

	if writeExpr, err := value.LookupPath(cue.ParsePath("writeExpression")).String(); err == nil {
		property.WriteExpression = writeExpr
	}

	// Parse dependencies
	dependsOnValue := value.LookupPath(cue.ParsePath("dependsOn"))
	if dependsOnValue.Exists() {
		if list, err := dependsOnValue.List(); err == nil {
			for list.Next() {
				if dep, err := list.Value().String(); err == nil {
					property.DependsOn = append(property.DependsOn, dep)
				}
			}
		}
	}

	// Parse transform
	transformValue := value.LookupPath(cue.ParsePath("transform"))
	if transformValue.Exists() {
		transform, err := l.parseEnhancedTransform(transformValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse transform: %w", err)
		}
		property.Transform = transform
	}

	// Parse validation
	validationValue := value.LookupPath(cue.ParsePath("validation"))
	if validationValue.Exists() {
		validation, err := l.parseValidation(validationValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse validation: %w", err)
		}
		property.Validation = validation
	}

	// Parse computed property
	computedValue := value.LookupPath(cue.ParsePath("computed"))
	if computedValue.Exists() {
		computed, err := l.parseComputedProperty(computedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to parse computed property: %w", err)
		}
		property.Computed = computed
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

// parseEnhancedTransform parses enhanced transformation rules
func (l *Loader) parseEnhancedTransform(value cue.Value) (*Transform, error) {
	transform := &Transform{}

	// Parse simple arithmetic
	if multiply, err := value.LookupPath(cue.ParsePath("multiply")).Float64(); err == nil {
		transform.Multiply = &multiply
	}

	if add, err := value.LookupPath(cue.ParsePath("add")).Float64(); err == nil {
		transform.Add = &add
	}

	if divide, err := value.LookupPath(cue.ParsePath("divide")).Float64(); err == nil {
		transform.Divide = &divide
	}

	if subtract, err := value.LookupPath(cue.ParsePath("subtract")).Float64(); err == nil {
		transform.Subtract = &subtract
	}

	// Parse expression
	if expression, err := value.LookupPath(cue.ParsePath("expression")).String(); err == nil {
		transform.Expression = expression
	}

	// Parse conditions
	conditionsValue := value.LookupPath(cue.ParsePath("conditions"))
	if conditionsValue.Exists() {
		if list, err := conditionsValue.List(); err == nil {
			for list.Next() {
				condValue := list.Value()
				condition := ConditionalTransform{}

				if ifExpr, err := condValue.LookupPath(cue.ParsePath("if")).String(); err == nil {
					condition.If = ifExpr
				}

				// Parse then value (can be any type)
				thenValue := condValue.LookupPath(cue.ParsePath("then"))
				if thenValue.Exists() {
					condition.Then = l.parseValue(thenValue)
				}

				// Parse else value (optional)
				elseValue := condValue.LookupPath(cue.ParsePath("else"))
				if elseValue.Exists() {
					condition.Else = l.parseValue(elseValue)
				}

				transform.Conditions = append(transform.Conditions, condition)
			}
		}
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

	// Parse range transformation
	rangeValue := value.LookupPath(cue.ParsePath("range"))
	if rangeValue.Exists() {
		rangeTransform := &RangeTransform{}

		if inputMin, err := rangeValue.LookupPath(cue.ParsePath("inputMin")).Float64(); err == nil {
			rangeTransform.InputMin = inputMin
		}

		if inputMax, err := rangeValue.LookupPath(cue.ParsePath("inputMax")).Float64(); err == nil {
			rangeTransform.InputMax = inputMax
		}

		if outputMin, err := rangeValue.LookupPath(cue.ParsePath("outputMin")).Float64(); err == nil {
			rangeTransform.OutputMin = outputMin
		}

		if outputMax, err := rangeValue.LookupPath(cue.ParsePath("outputMax")).Float64(); err == nil {
			rangeTransform.OutputMax = outputMax
		}

		if clamp, err := rangeValue.LookupPath(cue.ParsePath("clamp")).Bool(); err == nil {
			rangeTransform.Clamp = clamp
		}

		transform.Range = rangeTransform
	}

	// Parse string operations
	stringOpsValue := value.LookupPath(cue.ParsePath("stringOps"))
	if stringOpsValue.Exists() {
		stringOps := &StringOperations{
			Replace: make(map[string]string),
		}

		if trim, err := stringOpsValue.LookupPath(cue.ParsePath("trim")).Bool(); err == nil {
			stringOps.Trim = trim
		}

		if uppercase, err := stringOpsValue.LookupPath(cue.ParsePath("uppercase")).Bool(); err == nil {
			stringOps.Uppercase = uppercase
		}

		if lowercase, err := stringOpsValue.LookupPath(cue.ParsePath("lowercase")).Bool(); err == nil {
			stringOps.Lowercase = lowercase
		}

		replaceValue := stringOpsValue.LookupPath(cue.ParsePath("replace"))
		if replaceValue.Exists() {
			fields, _ := replaceValue.Fields()
			for fields.Next() {
				key := fields.Label()
				if val, err := fields.Value().String(); err == nil {
					stringOps.Replace[key] = val
				}
			}
		}

		transform.StringOps = stringOps
	}

	return transform, nil
}

// parseValidation parses validation constraints
func (l *Loader) parseValidation(value cue.Value) (*PropertyValidation, error) {
	validation := &PropertyValidation{}

	if minValue, err := value.LookupPath(cue.ParsePath("minValue")).Float64(); err == nil {
		validation.MinValue = &minValue
	}

	if maxValue, err := value.LookupPath(cue.ParsePath("maxValue")).Float64(); err == nil {
		validation.MaxValue = &maxValue
	}

	if pattern, err := value.LookupPath(cue.ParsePath("pattern")).String(); err == nil {
		validation.Pattern = pattern
	}

	if required, err := value.LookupPath(cue.ParsePath("required")).Bool(); err == nil {
		validation.Required = required
	}

	if constraint, err := value.LookupPath(cue.ParsePath("constraint")).String(); err == nil {
		validation.Constraint = constraint
	}

	// Parse allowed values
	allowedValue := value.LookupPath(cue.ParsePath("allowedValues"))
	if allowedValue.Exists() {
		if list, err := allowedValue.List(); err == nil {
			for list.Next() {
				validation.AllowedValues = append(validation.AllowedValues, l.parseValue(list.Value()))
			}
		}
	}

	return validation, nil
}

// parseComputedProperty parses computed property definition
func (l *Loader) parseComputedProperty(value cue.Value) (*ComputedProperty, error) {
	computed := &ComputedProperty{}

	if expression, err := value.LookupPath(cue.ParsePath("expression")).String(); err == nil {
		computed.Expression = expression
	}

	if propType, err := value.LookupPath(cue.ParsePath("type")).String(); err == nil {
		computed.Type = PropertyType(propType)
	}

	// Parse dependencies
	depsValue := value.LookupPath(cue.ParsePath("dependencies"))
	if depsValue.Exists() {
		if list, err := depsValue.List(); err == nil {
			for list.Next() {
				if dep, err := list.Value().String(); err == nil {
					computed.Dependencies = append(computed.Dependencies, dep)
				}
			}
		}
	}

	return computed, nil
}

// parsePropertyGroups parses property groups
func (l *Loader) parsePropertyGroups(value cue.Value, groups map[string]*PropertyGroup) error {
	fields, _ := value.Fields()
	for fields.Next() {
		label := fields.Label()
		groupValue := fields.Value()

		group := &PropertyGroup{
			Name: label,
		}

		if desc, err := groupValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			group.Description = desc
		}

		if icon, err := groupValue.LookupPath(cue.ParsePath("icon")).String(); err == nil {
			group.Icon = icon
		}

		if collapsed, err := groupValue.LookupPath(cue.ParsePath("collapsed")).Bool(); err == nil {
			group.Collapsed = collapsed
		}

		// Parse properties list
		propsValue := groupValue.LookupPath(cue.ParsePath("properties"))
		if propsValue.Exists() {
			if list, err := propsValue.List(); err == nil {
				for list.Next() {
					if prop, err := list.Value().String(); err == nil {
						group.Properties = append(group.Properties, prop)
					}
				}
			}
		}

		groups[label] = group
	}

	return nil
}

// parseComputedProperties parses computed properties at mapper level
func (l *Loader) parseComputedProperties(value cue.Value, computed map[string]*ComputedProperty) error {
	fields, _ := value.Fields()
	for fields.Next() {
		label := fields.Label()
		compValue := fields.Value()

		comp, err := l.parseComputedProperty(compValue)
		if err != nil {
			return err
		}

		computed[label] = comp
	}

	return nil
}

// parseCharMap parses character mapping for strings
func (l *Loader) parseCharMap(value cue.Value) (map[uint8]string, error) {
	charMap := make(map[uint8]string)

	fields, _ := value.Fields()
	for fields.Next() {
		keyStr := fields.Label()

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

// parseValue parses a CUE value into appropriate Go type
func (l *Loader) parseValue(value cue.Value) interface{} {
	if str, err := value.String(); err == nil {
		return str
	} else if num, err := value.Float64(); err == nil {
		return num
	} else if b, err := value.Bool(); err == nil {
		return b
	}
	return nil
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

// Enhanced property operations

// FreezeProperty freezes a property at its current value
func (m *Mapper) FreezeProperty(name string, memManager *memory.Manager) error {
	prop, exists := m.Properties[name]
	if !exists {
		return fmt.Errorf("property %s not found", name)
	}

	if !prop.Freezable {
		return fmt.Errorf("property %s is not freezable", name)
	}

	// Get current value as bytes
	data, err := memManager.ReadBytes(prop.Address, prop.Length)
	if err != nil {
		return fmt.Errorf("failed to read current value: %w", err)
	}

	// Freeze in memory manager
	if err := memManager.FreezeProperty(prop.Address, data); err != nil {
		return err
	}

	// Update property state
	prop.Frozen = true
	prop.FrozenData = make([]byte, len(data))
	copy(prop.FrozenData, data)

	return nil
}

// UnfreezeProperty unfreezes a property
func (m *Mapper) UnfreezeProperty(name string, memManager *memory.Manager) error {
	prop, exists := m.Properties[name]
	if !exists {
		return fmt.Errorf("property %s not found", name)
	}

	if err := memManager.UnfreezeProperty(prop.Address); err != nil {
		return err
	}

	prop.Frozen = false
	prop.FrozenData = nil

	return nil
}

// GetProperty gets a property value from memory with enhanced processing and deadlock prevention
func (m *Mapper) GetProperty(name string, memManager *memory.Manager) (interface{}, error) {
	prop, exists := m.Properties[name]
	if !exists {
		return nil, fmt.Errorf("property %s not found", name)
	}

	if prop.Computed != nil {
		return m.evaluateComputedProperty(prop, memManager)
	}

	// Determine endianness once
	littleEndian := prop.Endian == "little" || (prop.Endian == "" && m.Platform.Endian == "little")

	// Read the raw bytes first
	rawBytes, err := memManager.ReadBytes(prop.Address, prop.Length)
	if err != nil {
		// Return reasonable defaults instead of failing completely
		return m.getDefaultValue(prop.Type), nil
	}

	// Parse the value from bytes
	var raw interface{}
	var _ error

	switch prop.Type {
	case PropertyTypeUint8:
		if len(rawBytes) >= 1 {
			raw = rawBytes[0]
		}
	case PropertyTypeUint16:
		if len(rawBytes) >= 2 {
			if littleEndian {
				raw = binary.LittleEndian.Uint16(rawBytes)
			} else {
				raw = binary.BigEndian.Uint16(rawBytes)
			}
		}
	case PropertyTypeUint32:
		if len(rawBytes) >= 4 {
			if littleEndian {
				raw = binary.LittleEndian.Uint32(rawBytes)
			} else {
				raw = binary.BigEndian.Uint32(rawBytes)
			}
		}
	case PropertyTypeString:
		raw = m.parseStringFromBytes(rawBytes, prop.CharMap)
	case PropertyTypeBCD:
		raw = m.parseBCDFromBytes(rawBytes)
	case PropertyTypeBool:
		if len(rawBytes) >= 1 {
			raw = rawBytes[0] != 0
		}
	case PropertyTypeBit:
		if len(rawBytes) >= 1 && prop.Position != nil {
			raw = (rawBytes[0] >> *prop.Position) & 1
		}
	case PropertyTypeNibble:
		if len(rawBytes) >= 1 && prop.Position != nil {
			raw = (rawBytes[0] >> (*prop.Position * 4)) & 0x0F
		}
	default:
		raw = m.getDefaultValue(prop.Type)
	}

	// Set default if parsing failed
	if raw == nil {
		raw = m.getDefaultValue(prop.Type)
	}

	// Update property state in a separate goroutine to prevent blocking
	go memManager.UpdatePropertyState(name, raw, rawBytes, prop.Address)

	// Apply transformations
	result, transformErr := m.applyTransform(raw, prop.Transform)
	if transformErr != nil {
		result = raw // Use raw value if transform fails
	}

	// Validate (but don't fail on validation errors)
	m.validateValue(result, prop.Validation)

	return result, nil
}

// Helper methods for parsing
func (m *Mapper) getDefaultValue(propType PropertyType) interface{} {
	switch propType {
	case PropertyTypeUint8, PropertyTypeUint16, PropertyTypeUint32:
		return uint32(0)
	case PropertyTypeInt8, PropertyTypeInt16, PropertyTypeInt32:
		return int32(0)
	case PropertyTypeFloat32:
		return float32(0)
	case PropertyTypeFloat64:
		return float64(0)
	case PropertyTypeBool:
		return false
	case PropertyTypeString:
		return ""
	case PropertyTypeBCD:
		return uint32(0)
	case PropertyTypeBit, PropertyTypeNibble:
		return uint8(0)
	default:
		return nil
	}
}

func (m *Mapper) parseStringFromBytes(data []byte, charMap map[uint8]string) string {
	var result string
	for _, b := range data {
		if b == 0 || b == 0xFF {
			break
		}
		if char, exists := charMap[b]; exists {
			result += char
		} else if b >= 0x20 && b <= 0x7E {
			result += string(b) // Printable ASCII
		}
	}
	return result
}

func (m *Mapper) parseBCDFromBytes(data []byte) uint32 {
	result := uint32(0)
	for _, bcd := range data {
		result *= 100
		result += uint32(10*(bcd>>4) + (bcd & 0x0F))
	}
	return result
}

// evaluateComputedProperty evaluates a computed property
func (m *Mapper) evaluateComputedProperty(prop *Property, memManager *memory.Manager) (interface{}, error) {
	if prop.Computed == nil {
		return nil, fmt.Errorf("property is not computed")
	}

	// Get dependency values
	context := make(map[string]interface{})
	for _, dep := range prop.Computed.Dependencies {
		if value, err := m.GetProperty(dep, memManager); err == nil {
			context[dep] = value
		}
	}

	// Handle special computed properties
	switch prop.Name {
	case "badgeCount":
		// Count set bits in badges byte
		if badgesValue, exists := context["badges"]; exists {
			if badges, ok := badgesValue.(uint8); ok {
				count := 0
				for i := 0; i < 8; i++ {
					if (badges & (1 << i)) != 0 {
						count++
					}
				}
				return uint8(count), nil
			}
		}
		return uint8(0), nil
	case "pokemon1HpPercentage":
		// Calculate HP percentage
		if hp, hpExists := context["pokemon1Hp"]; hpExists {
			if maxHp, maxExists := context["pokemon1MaxHp"]; maxExists {
				if hpVal, ok := hp.(uint16); ok {
					if maxHpVal, ok := maxHp.(uint16); ok && maxHpVal > 0 {
						percentage := float64(hpVal) / float64(maxHpVal) * 100
						return percentage, nil
					}
				}
			}
		}
		return 0.0, nil
	}

	// TODO: Implement full CUE expression evaluation for other computed properties
	return fmt.Sprintf("computed(%s)", prop.Computed.Expression), nil
}

// validateValue validates a value against validation constraints
func (m *Mapper) validateValue(value interface{}, validation *PropertyValidation) error {
	if validation == nil {
		return nil
	}

	// Convert value to float64 for numeric validation
	var numValue float64
	var isNumeric bool

	switch v := value.(type) {
	case uint8:
		numValue = float64(v)
		isNumeric = true
	case uint16:
		numValue = float64(v)
		isNumeric = true
	case uint32:
		numValue = float64(v)
		isNumeric = true
	case int8:
		numValue = float64(v)
		isNumeric = true
	case int16:
		numValue = float64(v)
		isNumeric = true
	case int32:
		numValue = float64(v)
		isNumeric = true
	case float32:
		numValue = float64(v)
		isNumeric = true
	case float64:
		numValue = v
		isNumeric = true
	}

	// Validate numeric constraints
	if isNumeric {
		if validation.MinValue != nil && numValue < *validation.MinValue {
			return fmt.Errorf("value %f is below minimum %f", numValue, *validation.MinValue)
		}

		if validation.MaxValue != nil && numValue > *validation.MaxValue {
			return fmt.Errorf("value %f is above maximum %f", numValue, *validation.MaxValue)
		}
	}

	// Validate allowed values
	if len(validation.AllowedValues) > 0 {
		found := false
		for _, allowed := range validation.AllowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %v is not in allowed values", value)
		}
	}

	// TODO: Implement pattern and constraint validation

	return nil
}

// SetProperty sets a property value in memory with enhanced validation
func (m *Mapper) SetProperty(name string, value interface{}, memManager *memory.Manager, driver drivers.Driver) error {
	prop, exists := m.Properties[name]
	if !exists {
		return fmt.Errorf("property %s not found", name)
	}

	if prop.ReadOnly {
		return fmt.Errorf("property %s is read-only", name)
	}

	if prop.Computed != nil {
		return fmt.Errorf("cannot set computed property %s", name)
	}

	// Validate value before setting
	if err := m.validateValue(value, prop.Validation); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Convert value to bytes based on type
	var data []byte

	switch prop.Type {
	case PropertyTypeUint8:
		if val, ok := value.(uint8); ok {
			data = memManager.WriteUint8(prop.Address, val)
		} else if val, ok := value.(float64); ok {
			data = memManager.WriteUint8(prop.Address, uint8(val))
		} else {
			return fmt.Errorf("invalid type for uint8 property")
		}
	case PropertyTypeBool:
		if val, ok := value.(bool); ok {
			data = memManager.WriteBool(prop.Address, val)
		} else {
			return fmt.Errorf("invalid type for bool property")
		}
	// Add more types as needed
	default:
		return fmt.Errorf("setting values for type %s not yet implemented", prop.Type)
	}

	// Write to emulator if not frozen
	if !prop.Frozen {
		if err := driver.WriteBytes(prop.Address, data); err != nil {
			return err
		}
	}

	return nil
}

// ProcessProperties processes all properties and updates their values
func (m *Mapper) ProcessProperties(memManager *memory.Manager) error {
	for name := range m.Properties {
		_, err := m.GetProperty(name, memManager)
		if err != nil {
			return fmt.Errorf("failed to process property %s: %w", name, err)
		}
	}
	return nil
}

// applyTransform applies enhanced transformation rules to a raw value
func (m *Mapper) applyTransform(value interface{}, transform *Transform) (interface{}, error) {
	if transform == nil {
		return value, nil
	}

	// Evaluate CUE expression if defined
	if transform.Expression != "" {
		ctx := cuecontext.New()
		expr := fmt.Sprintf("{value: %v, result: (%s)}", value, transform.Expression)
		val := ctx.CompileString(expr)
		if err := val.Err(); err != nil {
			return value, fmt.Errorf("invalid cue expression: %w", err)
		}
		result := val.LookupPath(cue.ParsePath("result"))
		if !result.Exists() {
			return value, fmt.Errorf("missing result in cue expression")
		}
		return m.parseValue(result), nil
	}

	return value, nil // fallback if no transform logic applies
}

// parseValue parses a CUE value into appropriate Go type
func (m *Mapper) parseValue(value cue.Value) interface{} {
	if str, err := value.String(); err == nil {
		return str
	} else if num, err := value.Float64(); err == nil {
		return num
	} else if i, err := value.Int64(); err == nil {
		return i
	} else if b, err := value.Bool(); err == nil {
		return b
	} else if list, err := value.List(); err == nil {
		var arr []interface{}
		for list.Next() {
			arr = append(arr, m.parseValue(list.Value()))
		}
		return arr
	} else if fields, err := value.Fields(); err == nil {
		obj := make(map[string]interface{})
		for fields.Next() {
			obj[fields.Label()] = m.parseValue(fields.Value())
		}
		return obj
	}
	return nil
}
