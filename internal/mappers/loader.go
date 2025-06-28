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
	"gamehook/internal/types"
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

// ===== ENHANCED VALIDATION SYSTEM =====

// PropertyValidation represents enhanced validation constraints for a property
type PropertyValidation struct {
	MinValue        *float64          `json:"min_value,omitempty"`
	MaxValue        *float64          `json:"max_value,omitempty"`
	AllowedValues   []interface{}     `json:"allowed_values,omitempty"`
	Pattern         string            `json:"pattern,omitempty"`
	Required        bool              `json:"required"`
	Constraint      string            `json:"constraint,omitempty"` // CUE expression
	DependsOn       []string          `json:"depends_on,omitempty"`
	CrossValidation string            `json:"cross_validation,omitempty"`
	Messages        map[string]string `json:"messages,omitempty"`
}

// ===== ENHANCED TRANSFORMATION SYSTEM =====

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

// PadOperation represents string padding configuration
type PadOperation struct {
	Length uint   `json:"length"`
	Char   string `json:"char"`
}

// StringOperations represents enhanced string transformation operations
type StringOperations struct {
	Trim      bool              `json:"trim"`
	Uppercase bool              `json:"uppercase"`
	Lowercase bool              `json:"lowercase"`
	Replace   map[string]string `json:"replace"`
	Truncate  *uint             `json:"truncate,omitempty"`
	PadLeft   *PadOperation     `json:"pad_left,omitempty"`
	PadRight  *PadOperation     `json:"pad_right,omitempty"`
}

// Transform represents enhanced value transformation rules
type Transform struct {
	// Simple arithmetic
	Multiply *float64 `json:"multiply,omitempty"`
	Add      *float64 `json:"add,omitempty"`
	Divide   *float64 `json:"divide,omitempty"`
	Subtract *float64 `json:"subtract,omitempty"`
	Modulo   *float64 `json:"modulo,omitempty"`

	// Bitwise operations
	BitwiseAnd *uint32 `json:"bitwise_and,omitempty"`
	BitwiseOr  *uint32 `json:"bitwise_or,omitempty"`
	BitwiseXor *uint32 `json:"bitwise_xor,omitempty"`
	LeftShift  *uint32 `json:"left_shift,omitempty"`
	RightShift *uint32 `json:"right_shift,omitempty"`

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

	// Custom functions
	CustomFunction string `json:"custom_function,omitempty"`
}

// ===== UI SYSTEM =====

// UIHints represents enhanced UI presentation hints
type UIHints struct {
	DisplayFormat string `json:"display_format,omitempty"` // "hex", "decimal", "binary", "percentage", "currency", "time", "custom"
	Unit          string `json:"unit,omitempty"`           // "bytes", "seconds", "pixels", etc.
	Precision     *uint  `json:"precision,omitempty"`      // decimal places for floats
	ShowInList    *bool  `json:"show_in_list,omitempty"`   // show in main property list
	Category      string `json:"category,omitempty"`       // custom category for grouping
	Priority      *uint  `json:"priority,omitempty"`       // display priority
	Tooltip       string `json:"tooltip,omitempty"`        // tooltip text
	Color         string `json:"color,omitempty"`          // hex color for value display
	Icon          string `json:"icon,omitempty"`           // icon for property
	Badge         string `json:"badge,omitempty"`          // badge text
	Editable      *bool  `json:"editable,omitempty"`       // can be edited in UI
	Copyable      *bool  `json:"copyable,omitempty"`       // can be copied to clipboard
	Watchable     *bool  `json:"watchable,omitempty"`      // can be watched for changes
	Chartable     *bool  `json:"chartable,omitempty"`      // can be displayed in charts
	ChartType     string `json:"chart_type,omitempty"`     // "line", "bar", "pie", "gauge"
	ChartColor    string `json:"chart_color,omitempty"`    // color for charts
}

// ===== ADVANCED CONFIGURATION =====

// AdvancedConfig represents type-specific advanced configuration
type AdvancedConfig struct {
	// Pointer types
	TargetType      *PropertyType `json:"target_type,omitempty"`
	MaxDereferences *uint         `json:"max_dereferences,omitempty"`
	NullValue       *uint32       `json:"null_value,omitempty"`

	// Array types
	ElementType    *PropertyType `json:"element_type,omitempty"`
	ElementSize    *uint         `json:"element_size,omitempty"`
	DynamicLength  *bool         `json:"dynamic_length,omitempty"`
	LengthProperty string        `json:"length_property,omitempty"`
	MaxElements    *uint         `json:"max_elements,omitempty"`
	IndexOffset    *uint         `json:"index_offset,omitempty"`
	Stride         *uint         `json:"stride,omitempty"`

	// Struct types
	Fields  map[string]*StructField `json:"fields,omitempty"`
	Extends string                  `json:"extends,omitempty"`

	// Enum types
	EnumValues         map[string]*EnumValue `json:"enum_values,omitempty"`
	AllowUnknownValues *bool                 `json:"allow_unknown_values,omitempty"`
	DefaultValue       *uint32               `json:"default_value,omitempty"`

	// Flags/bitfield types
	FlagDefinitions map[string]*FlagDefinition `json:"flag_definitions,omitempty"`

	// Time types
	TimeFormat string   `json:"time_format,omitempty"` // "frames", "milliseconds", "seconds", "unix", "bcd"
	FrameRate  *float64 `json:"frame_rate,omitempty"`
	Epoch      string   `json:"epoch,omitempty"`

	// Coordinate types
	CoordinateSystem string `json:"coordinate_system,omitempty"` // "cartesian", "screen", "polar", "geographic"
	Dimensions       *uint  `json:"dimensions,omitempty"`        // 2D, 3D, etc.
	Units            string `json:"units,omitempty"`             // "pixels", "meters", "degrees"

	// Color types
	ColorFormat  string `json:"color_format,omitempty"` // "rgb565", "argb8888", "rgb888", etc.
	AlphaChannel *bool  `json:"alpha_channel,omitempty"`
	PaletteRef   string `json:"palette_ref,omitempty"`

	// Percentage types
	MaxValue  *float64 `json:"max_value,omitempty"`
	Precision *uint    `json:"precision,omitempty"`

	// Version types
	VersionFormat string `json:"version_format,omitempty"` // "major.minor.patch", "bcd", "packed", "string"

	// Checksum types
	ChecksumAlgorithm string         `json:"checksum_algorithm,omitempty"` // "crc16", "crc32", "md5", "sha1", "custom"
	ChecksumRange     *ChecksumRange `json:"checksum_range,omitempty"`
}

// StructField represents a field in a struct property
type StructField struct {
	Type        PropertyType        `json:"type"`
	Offset      uint                `json:"offset"`
	Size        *uint               `json:"size,omitempty"`
	Transform   *Transform          `json:"transform,omitempty"`
	Validation  *PropertyValidation `json:"validation,omitempty"`
	Description string              `json:"description,omitempty"`
	Computed    *ComputedProperty   `json:"computed,omitempty"`
}

// EnumValue represents a value in an enumeration
type EnumValue struct {
	Value       uint32 `json:"value"`
	Description string `json:"description"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Deprecated  *bool  `json:"deprecated,omitempty"`
}

// FlagDefinition represents a bit flag definition
type FlagDefinition struct {
	Bit               uint     `json:"bit"`
	Description       string   `json:"description"`
	InvertLogic       *bool    `json:"invert_logic,omitempty"`
	Group             string   `json:"group,omitempty"`
	MutuallyExclusive []string `json:"mutually_exclusive,omitempty"`
}

// ChecksumRange represents the range for checksum calculation
type ChecksumRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ===== PERFORMANCE SYSTEM =====

// PerformanceHints represents performance optimization hints
type PerformanceHints struct {
	Cacheable     *bool  `json:"cacheable,omitempty"`
	CacheTimeout  *uint  `json:"cache_timeout,omitempty"`
	ReadFrequency string `json:"read_frequency,omitempty"` // "high", "medium", "low"
	Critical      *bool  `json:"critical,omitempty"`
	Batchable     *bool  `json:"batchable,omitempty"`
	BatchGroup    string `json:"batch_group,omitempty"`
}

// ===== DEBUG SYSTEM =====

// DebugConfig represents debug and development features
type DebugConfig struct {
	LogReads        *bool  `json:"log_reads,omitempty"`
	LogWrites       *bool  `json:"log_writes,omitempty"`
	BreakOnRead     *bool  `json:"break_on_read,omitempty"`
	BreakOnWrite    *bool  `json:"break_on_write,omitempty"`
	WatchExpression string `json:"watch_expression,omitempty"`
}

// ===== COMPUTED PROPERTIES =====

// ComputedProperty represents a property derived from other properties
type ComputedProperty struct {
	Expression        string       `json:"expression"`
	Dependencies      []string     `json:"dependencies"`
	Type              PropertyType `json:"type,omitempty"`
	Cached            *bool        `json:"cached,omitempty"`
	CacheInvalidation []string     `json:"cache_invalidation,omitempty"`
}

// ===== ENHANCED PROPERTY =====

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

	// Enhanced features
	UIHints     *UIHints          `json:"ui_hints,omitempty"`
	Advanced    *AdvancedConfig   `json:"advanced,omitempty"`
	Performance *PerformanceHints `json:"performance,omitempty"`
	Debug       *DebugConfig      `json:"debug,omitempty"`

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

// ===== ENHANCED PROPERTY GROUPS =====

// PropertyGroup represents enhanced property groups for UI organization
type PropertyGroup struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description,omitempty"`
	Icon        string                    `json:"icon,omitempty"`
	Properties  []string                  `json:"properties"`
	Collapsed   bool                      `json:"collapsed"`
	Color       string                    `json:"color,omitempty"`
	DisplayMode string                    `json:"display_mode,omitempty"` // "table", "cards", "tree", "custom"
	SortBy      string                    `json:"sort_by,omitempty"`
	FilterBy    string                    `json:"filter_by,omitempty"`
	Conditional *ConditionalDisplay       `json:"conditional_display,omitempty"`
	Subgroups   map[string]*PropertyGroup `json:"subgroups,omitempty"`
	Renderer    string                    `json:"custom_renderer,omitempty"`
	Priority    *uint                     `json:"priority,omitempty"`
}

// ConditionalDisplay represents conditional group display
type ConditionalDisplay struct {
	Expression   string   `json:"expression"`
	Dependencies []string `json:"dependencies"`
}

// ===== ENHANCED PLATFORM =====

// Platform represents enhanced platform configuration
type Platform struct {
	Name          string
	Endian        string
	MemoryBlocks  []types.MemoryBlock
	Constants     map[string]interface{} // Platform-specific constants
	BaseAddresses map[string]string      // Named base addresses
	Description   string                 `json:"description,omitempty"`
	Version       string                 `json:"version,omitempty"`
	Manufacturer  string                 `json:"manufacturer,omitempty"`
	ReleaseYear   *uint                  `json:"release_year,omitempty"`
	Capabilities  *PlatformCapabilities  `json:"capabilities,omitempty"`
	Performance   *PlatformPerformance   `json:"performance,omitempty"`
}

// MemoryBlock represents enhanced memory block
type MemoryBlock struct {
	Name          string `json:"name"`
	Start         uint32 `json:"start"`
	End           uint32 `json:"end"`
	StartExpr     string `json:"start_expr,omitempty"`
	EndExpr       string `json:"end_expr,omitempty"`
	Description   string `json:"description,omitempty"`
	Readable      *bool  `json:"readable,omitempty"`
	Writable      *bool  `json:"writable,omitempty"`
	Cacheable     *bool  `json:"cacheable,omitempty"`
	AccessPattern string `json:"access_pattern,omitempty"` // "sequential", "random", "sparse"
	Protected     *bool  `json:"protected,omitempty"`
	Watchable     *bool  `json:"watchable,omitempty"`
}

// PlatformCapabilities represents platform capabilities
type PlatformCapabilities struct {
	MaxMemorySize    *uint32 `json:"max_memory_size,omitempty"`
	AddressBusWidth  *uint   `json:"address_bus_width,omitempty"`
	DataBusWidth     *uint   `json:"data_bus_width,omitempty"`
	HasMemoryMapping *bool   `json:"has_memory_mapping,omitempty"`
	SupportsBanking  *bool   `json:"supports_banking,omitempty"`
}

// PlatformPerformance represents platform performance hints
type PlatformPerformance struct {
	ReadLatency  *uint `json:"read_latency,omitempty"`  // milliseconds
	WriteLatency *uint `json:"write_latency,omitempty"` // milliseconds
	BatchSize    *uint `json:"batch_size,omitempty"`    // optimal batch size
}

// ===== EVENTS SYSTEM =====

// EventsConfig represents the events configuration
type EventsConfig struct {
	OnLoad            string                  `json:"on_load,omitempty"`
	OnUnload          string                  `json:"on_unload,omitempty"`
	OnPropertyChanged string                  `json:"on_property_changed,omitempty"`
	Custom            map[string]*CustomEvent `json:"custom,omitempty"`
}

// CustomEvent represents a custom event definition
type CustomEvent struct {
	Trigger      string   `json:"trigger"`
	Action       string   `json:"action"`
	Dependencies []string `json:"dependencies"`
}

// ===== VALIDATION SYSTEM =====

// GlobalValidation represents global validation rules
type GlobalValidation struct {
	MemoryLayout    *MemoryLayoutValidation `json:"memory_layout,omitempty"`
	CrossValidation []CrossValidationRule   `json:"cross_validation,omitempty"`
	Performance     *PerformanceValidation  `json:"performance,omitempty"`
}

// MemoryLayoutValidation represents memory layout validation
type MemoryLayoutValidation struct {
	CheckOverlaps  *bool `json:"check_overlaps,omitempty"`
	CheckBounds    *bool `json:"check_bounds,omitempty"`
	CheckAlignment *bool `json:"check_alignment,omitempty"`
}

// CrossValidationRule represents cross-property validation
type CrossValidationRule struct {
	Name         string   `json:"name"`
	Expression   string   `json:"expression"`
	Dependencies []string `json:"dependencies"`
	Message      string   `json:"message,omitempty"`
}

// PerformanceValidation represents performance validation
type PerformanceValidation struct {
	MaxProperties      *uint `json:"max_properties,omitempty"`
	MaxComputedDepth   *uint `json:"max_computed_depth,omitempty"`
	WarnSlowProperties *bool `json:"warn_slow_properties,omitempty"`
}

// ===== ENHANCED MAPPER =====

// Mapper represents an enhanced complete mapper with properties
type Mapper struct {
	Name        string
	Game        string
	Version     string
	MinVersion  string // Minimum GameHook version
	Author      string
	Description string
	Website     string
	License     string
	Metadata    *MapperMetadata
	Platform    Platform
	Properties  map[string]*Property
	Groups      map[string]*PropertyGroup
	Computed    map[string]*ComputedProperty
	Constants   map[string]interface{}      // Global constants
	Preprocess  []string                    // CUE expressions run before property evaluation
	Postprocess []string                    // CUE expressions run after property evaluation
	References  map[string]*Property        // Reference types
	CharMaps    map[string]map[uint8]string // Character maps
	Events      *EventsConfig               // Events configuration
	Validation  *GlobalValidation           // Global validation
	Debug       *MapperDebugConfig          // Debug configuration
}

// MapperMetadata represents mapper metadata
type MapperMetadata struct {
	Created  string   `json:"created"`
	Modified string   `json:"modified"`
	Tags     []string `json:"tags"`
	Category string   `json:"category"`
	Language string   `json:"language"`
	Region   string   `json:"region"`
	Revision string   `json:"revision,omitempty"`
}

// MapperDebugConfig represents mapper debug configuration
type MapperDebugConfig struct {
	Enabled             *bool    `json:"enabled,omitempty"`
	LogLevel            string   `json:"log_level,omitempty"`
	LogProperties       []string `json:"log_properties,omitempty"`
	BenchmarkProperties []string `json:"benchmark_properties,omitempty"`
	HotReload           *bool    `json:"hot_reload,omitempty"`
	TypeChecking        *bool    `json:"type_checking,omitempty"`
	MemoryDumps         *bool    `json:"memory_dumps,omitempty"`
}

// ===== LOADER =====

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
	log.Printf("🔍 Loading enhanced mapper from file: %s", filePath)

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

	log.Printf("✅ CUE value built successfully, parsing enhanced mapper...")
	mapper, err := l.parseEnhancedMapper(value)
	if err != nil {
		log.Printf("❌ Enhanced mapper parsing error: %v", err)
		return nil, err
	}

	log.Printf("✅ Enhanced mapper parsed: %d properties, %d groups, %d computed, %d references",
		len(mapper.Properties), len(mapper.Groups), len(mapper.Computed), len(mapper.References))

	return mapper, nil
}

// parseEnhancedMapper parses a CUE value into an enhanced Mapper struct
func (l *Loader) parseEnhancedMapper(value cue.Value) (*Mapper, error) {
	mapper := &Mapper{
		Properties: make(map[string]*Property),
		Groups:     make(map[string]*PropertyGroup),
		Computed:   make(map[string]*ComputedProperty),
		Constants:  make(map[string]interface{}),
		References: make(map[string]*Property),
		CharMaps:   make(map[string]map[uint8]string),
	}

	// Parse enhanced mapper metadata
	if err := l.parseMapperMetadata(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Parse enhanced platform
	if err := l.parseEnhancedPlatform(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse platform: %w", err)
	}

	// Parse constants
	if err := l.parseConstants(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse constants: %w", err)
	}

	// Parse character maps
	if err := l.parseCharacterMaps(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse character maps: %w", err)
	}

	// Parse reference types
	if err := l.parseReferences(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse references: %w", err)
	}

	// Parse enhanced properties
	if err := l.parseEnhancedPropertiesWithDebug(value, mapper.Properties); err != nil {
		return nil, fmt.Errorf("failed to parse properties: %w", err)
	}

	// Parse enhanced groups
	if err := l.parseEnhancedPropertyGroups(value, mapper.Groups); err != nil {
		return nil, fmt.Errorf("failed to parse groups: %w", err)
	}

	// Parse computed properties
	if err := l.parseComputedProperties(value, mapper.Computed); err != nil {
		return nil, fmt.Errorf("failed to parse computed properties: %w", err)
	}

	// Parse events
	if err := l.parseEvents(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	// Parse global validation
	if err := l.parseGlobalValidation(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse validation: %w", err)
	}

	// Parse mapper debug configuration
	if err := l.parseMapperDebugConfig(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse debug config: %w", err)
	}

	// Parse preprocessing and postprocessing
	if err := l.parseProcessingSteps(value, mapper); err != nil {
		return nil, fmt.Errorf("failed to parse processing steps: %w", err)
	}

	return mapper, nil
}

// parseMapperMetadata parses enhanced mapper metadata
func (l *Loader) parseMapperMetadata(value cue.Value, mapper *Mapper) error {
	// Basic metadata
	if name, err := value.LookupPath(cue.ParsePath("name")).String(); err == nil {
		mapper.Name = name
		log.Printf("📋 Enhanced mapper name: %s", name)
	}

	if game, err := value.LookupPath(cue.ParsePath("game")).String(); err == nil {
		mapper.Game = game
		log.Printf("🎮 Game: %s", game)
	}

	if version, err := value.LookupPath(cue.ParsePath("version")).String(); err == nil {
		mapper.Version = version
		log.Printf("🔖 Version: %s", version)
	}

	if minVersion, err := value.LookupPath(cue.ParsePath("minGameHookVersion")).String(); err == nil {
		mapper.MinVersion = minVersion
	}

	if author, err := value.LookupPath(cue.ParsePath("author")).String(); err == nil {
		mapper.Author = author
	}

	if description, err := value.LookupPath(cue.ParsePath("description")).String(); err == nil {
		mapper.Description = description
	}

	if website, err := value.LookupPath(cue.ParsePath("website")).String(); err == nil {
		mapper.Website = website
	}

	if license, err := value.LookupPath(cue.ParsePath("license")).String(); err == nil {
		mapper.License = license
	}

	// Enhanced metadata
	metadataValue := value.LookupPath(cue.ParsePath("metadata"))
	if metadataValue.Exists() {
		metadata := &MapperMetadata{}

		if created, err := metadataValue.LookupPath(cue.ParsePath("created")).String(); err == nil {
			metadata.Created = created
		}

		if modified, err := metadataValue.LookupPath(cue.ParsePath("modified")).String(); err == nil {
			metadata.Modified = modified
		}

		if category, err := metadataValue.LookupPath(cue.ParsePath("category")).String(); err == nil {
			metadata.Category = category
		}

		if language, err := metadataValue.LookupPath(cue.ParsePath("language")).String(); err == nil {
			metadata.Language = language
		}

		if region, err := metadataValue.LookupPath(cue.ParsePath("region")).String(); err == nil {
			metadata.Region = region
		}

		if revision, err := metadataValue.LookupPath(cue.ParsePath("revision")).String(); err == nil {
			metadata.Revision = revision
		}

		// Parse tags array
		tagsValue := metadataValue.LookupPath(cue.ParsePath("tags"))
		if tagsValue.Exists() {
			if list, err := tagsValue.List(); err == nil {
				for list.Next() {
					if tag, err := list.Value().String(); err == nil {
						metadata.Tags = append(metadata.Tags, tag)
					}
				}
			}
		}

		mapper.Metadata = metadata
	}

	return nil
}

// parseEnhancedPlatform parses enhanced platform configuration
func (l *Loader) parseEnhancedPlatform(value cue.Value, mapper *Mapper) error {
	platformValue := value.LookupPath(cue.ParsePath("platform"))
	if !platformValue.Exists() {
		return fmt.Errorf("platform configuration is required")
	}

	platform := Platform{
		Constants:     make(map[string]interface{}),
		BaseAddresses: make(map[string]string),
	}

	// Basic platform info
	if name, err := platformValue.LookupPath(cue.ParsePath("name")).String(); err == nil {
		platform.Name = name
	}

	if endian, err := platformValue.LookupPath(cue.ParsePath("endian")).String(); err == nil {
		platform.Endian = endian
	}

	if description, err := platformValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
		platform.Description = description
	}

	if version, err := platformValue.LookupPath(cue.ParsePath("version")).String(); err == nil {
		platform.Version = version
	}

	if manufacturer, err := platformValue.LookupPath(cue.ParsePath("manufacturer")).String(); err == nil {
		platform.Manufacturer = manufacturer
	}

	if releaseYear, err := platformValue.LookupPath(cue.ParsePath("releaseYear")).Uint64(); err == nil {
		year := uint(releaseYear)
		platform.ReleaseYear = &year
	}

	// Parse platform constants
	constantsValue := platformValue.LookupPath(cue.ParsePath("constants"))
	if constantsValue.Exists() {
		l.parseConstantsValue(constantsValue, platform.Constants)
	}

	// Parse base addresses
	baseAddrsValue := platformValue.LookupPath(cue.ParsePath("baseAddresses"))
	if baseAddrsValue.Exists() {
		fields, _ := baseAddrsValue.Fields()
		for fields.Next() {
			label := fields.Label()
			if addr, err := fields.Value().String(); err == nil {
				platform.BaseAddresses[label] = addr
			}
		}
	}

	// Parse enhanced memory blocks
	if err := l.parseEnhancedMemoryBlocks(platformValue, &platform); err != nil {
		return fmt.Errorf("failed to parse memory blocks: %w", err)
	}

	// Parse platform capabilities
	if err := l.parsePlatformCapabilities(platformValue, &platform); err != nil {
		return fmt.Errorf("failed to parse platform capabilities: %w", err)
	}

	// Parse platform performance
	if err := l.parsePlatformPerformance(platformValue, &platform); err != nil {
		return fmt.Errorf("failed to parse platform performance: %w", err)
	}

	mapper.Platform = platform
	log.Printf("✅ Enhanced platform parsed: %s with %d memory blocks", platform.Name, len(platform.MemoryBlocks))

	return nil
}

// parseEnhancedMemoryBlocks parses enhanced memory blocks
func (l *Loader) parseEnhancedMemoryBlocks(platformValue cue.Value, platform *Platform) error {
	blocksValue := platformValue.LookupPath(cue.ParsePath("memoryBlocks"))
	if !blocksValue.Exists() {
		return nil
	}

	blocksIter, _ := blocksValue.List()
	for blocksIter.Next() {
		blockValue := blocksIter.Value()

		block := MemoryBlock{}

		// Basic block info
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

		// Enhanced block properties
		if description, err := blockValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			block.Description = description
		}

		if startExpr, err := blockValue.LookupPath(cue.ParsePath("startExpr")).String(); err == nil {
			block.StartExpr = startExpr
		}

		if endExpr, err := blockValue.LookupPath(cue.ParsePath("endExpr")).String(); err == nil {
			block.EndExpr = endExpr
		}

		if readable, err := blockValue.LookupPath(cue.ParsePath("readable")).Bool(); err == nil {
			block.Readable = &readable
		}

		if writable, err := blockValue.LookupPath(cue.ParsePath("writable")).Bool(); err == nil {
			block.Writable = &writable
		}

		if cacheable, err := blockValue.LookupPath(cue.ParsePath("cacheable")).Bool(); err == nil {
			block.Cacheable = &cacheable
		}

		if pattern, err := blockValue.LookupPath(cue.ParsePath("accessPattern")).String(); err == nil {
			block.AccessPattern = pattern
		}

		if protected, err := blockValue.LookupPath(cue.ParsePath("protected")).Bool(); err == nil {
			block.Protected = &protected
		}

		if watchable, err := blockValue.LookupPath(cue.ParsePath("watchable")).Bool(); err == nil {
			block.Watchable = &watchable
		}

		// Convert to old MemoryBlock format for compatibility
		oldBlock := types.MemoryBlock{ // ← Change from drivers.MemoryBlock to types.MemoryBlock
			Name:  block.Name,
			Start: block.Start,
			End:   block.End,
		}

		platform.MemoryBlocks = append(platform.MemoryBlocks, oldBlock)
	}

	return nil
}

// parsePlatformCapabilities parses platform capabilities
func (l *Loader) parsePlatformCapabilities(platformValue cue.Value, platform *Platform) error {
	capValue := platformValue.LookupPath(cue.ParsePath("capabilities"))
	if !capValue.Exists() {
		return nil
	}

	capabilities := &PlatformCapabilities{}

	if maxMemSize, err := capValue.LookupPath(cue.ParsePath("maxMemorySize")).Uint64(); err == nil {
		size := uint32(maxMemSize)
		capabilities.MaxMemorySize = &size
	}

	if addrBusWidth, err := capValue.LookupPath(cue.ParsePath("addressBusWidth")).Uint64(); err == nil {
		width := uint(addrBusWidth)
		capabilities.AddressBusWidth = &width
	}

	if dataBusWidth, err := capValue.LookupPath(cue.ParsePath("dataBusWidth")).Uint64(); err == nil {
		width := uint(dataBusWidth)
		capabilities.DataBusWidth = &width
	}

	if hasMemMapping, err := capValue.LookupPath(cue.ParsePath("hasMemoryMapping")).Bool(); err == nil {
		capabilities.HasMemoryMapping = &hasMemMapping
	}

	if supportsBanking, err := capValue.LookupPath(cue.ParsePath("supportsBanking")).Bool(); err == nil {
		capabilities.SupportsBanking = &supportsBanking
	}

	platform.Capabilities = capabilities
	return nil
}

// parsePlatformPerformance parses platform performance hints
func (l *Loader) parsePlatformPerformance(platformValue cue.Value, platform *Platform) error {
	perfValue := platformValue.LookupPath(cue.ParsePath("performance"))
	if !perfValue.Exists() {
		return nil
	}

	performance := &PlatformPerformance{}

	if readLatency, err := perfValue.LookupPath(cue.ParsePath("readLatency")).Uint64(); err == nil {
		latency := uint(readLatency)
		performance.ReadLatency = &latency
	}

	if writeLatency, err := perfValue.LookupPath(cue.ParsePath("writeLatency")).Uint64(); err == nil {
		latency := uint(writeLatency)
		performance.WriteLatency = &latency
	}

	if batchSize, err := perfValue.LookupPath(cue.ParsePath("batchSize")).Uint64(); err == nil {
		size := uint(batchSize)
		performance.BatchSize = &size
	}

	platform.Performance = performance
	return nil
}

// parseConstants parses constants from CUE value
func (l *Loader) parseConstants(value cue.Value, mapper *Mapper) error {
	constantsValue := value.LookupPath(cue.ParsePath("constants"))
	if constantsValue.Exists() {
		l.parseConstantsValue(constantsValue, mapper.Constants)
	}
	return nil
}

// parseConstantsValue parses constants from a CUE value into a map
func (l *Loader) parseConstantsValue(value cue.Value, constants map[string]interface{}) {
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
		} else if i, err := fieldValue.Int64(); err == nil {
			constants[label] = i
		}
	}
}

// parseCharacterMaps parses character maps
func (l *Loader) parseCharacterMaps(value cue.Value, mapper *Mapper) error {
	charMapsValue := value.LookupPath(cue.ParsePath("characterMaps"))
	if !charMapsValue.Exists() {
		return nil
	}

	fields, _ := charMapsValue.Fields()
	for fields.Next() {
		mapName := fields.Label()
		mapValue := fields.Value()

		charMap := make(map[uint8]string)
		mapFields, _ := mapValue.Fields()
		for mapFields.Next() {
			keyStr := mapFields.Label()
			value := mapFields.Value()

			// Parse key
			var key uint64
			var err error
			if strings.HasPrefix(keyStr, "0x") || strings.HasPrefix(keyStr, "0X") {
				key, err = strconv.ParseUint(keyStr[2:], 16, 8)
			} else {
				key, err = strconv.ParseUint(keyStr, 10, 8)
			}

			if err != nil {
				continue
			}

			if char, err := value.String(); err == nil {
				charMap[uint8(key)] = char
			}
		}

		mapper.CharMaps[mapName] = charMap
	}

	log.Printf("📝 Parsed %d character maps", len(mapper.CharMaps))
	return nil
}

// parseReferences parses reference types
func (l *Loader) parseReferences(value cue.Value, mapper *Mapper) error {
	referencesValue := value.LookupPath(cue.ParsePath("references"))
	if !referencesValue.Exists() {
		return nil
	}

	fields, _ := referencesValue.Fields()
	for fields.Next() {
		refName := fields.Label()
		refValue := fields.Value()

		property, err := l.parseEnhancedProperty(refValue)
		if err != nil {
			return fmt.Errorf("failed to parse reference %s: %w", refName, err)
		}

		property.Name = refName
		mapper.References[refName] = property
	}

	log.Printf("🔗 Parsed %d reference types", len(mapper.References))
	return nil
}

// parseEnhancedPropertiesWithDebug parses enhanced properties with debug output
func (l *Loader) parseEnhancedPropertiesWithDebug(value cue.Value, properties map[string]*Property) error {
	propertiesValue := value.LookupPath(cue.ParsePath("properties"))
	if !propertiesValue.Exists() {
		log.Printf("⚠️  No properties section found")
		return nil
	}

	fields, _ := propertiesValue.Fields()
	propertyCount := 0

	for fields.Next() {
		label := fields.Label()
		propertyValue := fields.Value()

		log.Printf("   🔍 Parsing enhanced property: %s", label)

		property, err := l.parseEnhancedProperty(propertyValue)
		if err != nil {
			log.Printf("   ❌ Failed to parse property %s: %v", label, err)
			return fmt.Errorf("failed to parse property %s: %w", label, err)
		}

		property.Name = label
		properties[label] = property
		propertyCount++

		log.Printf("   ✅ Property %s: type=%s, address=%s",
			label, property.Type, fmt.Sprintf("0x%X", property.Address))
	}

	log.Printf("📊 Total enhanced properties parsed: %d", propertyCount)
	return nil
}

// parseEnhancedProperty parses a single enhanced property definition
func (l *Loader) parseEnhancedProperty(value cue.Value) (*Property, error) {
	property := &Property{}

	// Basic property fields
	if err := l.parseBasicPropertyFields(value, property); err != nil {
		return nil, err
	}

	// Enhanced features
	if err := l.parseUIHints(value, property); err != nil {
		return nil, err
	}

	if err := l.parseAdvancedConfig(value, property); err != nil {
		return nil, err
	}

	if err := l.parsePerformanceHints(value, property); err != nil {
		return nil, err
	}

	if err := l.parseDebugConfig(value, property); err != nil {
		return nil, err
	}

	// Transform and validation
	if err := l.parseEnhancedTransform(value, property); err != nil {
		return nil, err
	}

	if err := l.parseEnhancedValidation(value, property); err != nil {
		return nil, err
	}

	// Computed properties
	if err := l.parsePropertyComputed(value, property); err != nil {
		return nil, err
	}

	// Character map
	if err := l.parsePropertyCharMap(value, property); err != nil {
		return nil, err
	}

	return property, nil
}

// parseBasicPropertyFields parses basic property fields
func (l *Loader) parseBasicPropertyFields(value cue.Value, property *Property) error {
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
			return fmt.Errorf("invalid address %s: %w", addressStr, err)
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

	return nil
}

// parseUIHints parses UI hints for enhanced presentation
func (l *Loader) parseUIHints(value cue.Value, property *Property) error {
	uiValue := value.LookupPath(cue.ParsePath("uiHints"))
	if !uiValue.Exists() {
		return nil
	}

	hints := &UIHints{}

	if displayFormat, err := uiValue.LookupPath(cue.ParsePath("displayFormat")).String(); err == nil {
		hints.DisplayFormat = displayFormat
	}

	if unit, err := uiValue.LookupPath(cue.ParsePath("unit")).String(); err == nil {
		hints.Unit = unit
	}

	if precision, err := uiValue.LookupPath(cue.ParsePath("precision")).Uint64(); err == nil {
		p := uint(precision)
		hints.Precision = &p
	}

	if showInList, err := uiValue.LookupPath(cue.ParsePath("showInList")).Bool(); err == nil {
		hints.ShowInList = &showInList
	}

	if category, err := uiValue.LookupPath(cue.ParsePath("category")).String(); err == nil {
		hints.Category = category
	}

	if priority, err := uiValue.LookupPath(cue.ParsePath("priority")).Uint64(); err == nil {
		p := uint(priority)
		hints.Priority = &p
	}

	if tooltip, err := uiValue.LookupPath(cue.ParsePath("tooltip")).String(); err == nil {
		hints.Tooltip = tooltip
	}

	if color, err := uiValue.LookupPath(cue.ParsePath("color")).String(); err == nil {
		hints.Color = color
	}

	if icon, err := uiValue.LookupPath(cue.ParsePath("icon")).String(); err == nil {
		hints.Icon = icon
	}

	if badge, err := uiValue.LookupPath(cue.ParsePath("badge")).String(); err == nil {
		hints.Badge = badge
	}

	if editable, err := uiValue.LookupPath(cue.ParsePath("editable")).Bool(); err == nil {
		hints.Editable = &editable
	}

	if copyable, err := uiValue.LookupPath(cue.ParsePath("copyable")).Bool(); err == nil {
		hints.Copyable = &copyable
	}

	if watchable, err := uiValue.LookupPath(cue.ParsePath("watchable")).Bool(); err == nil {
		hints.Watchable = &watchable
	}

	if chartable, err := uiValue.LookupPath(cue.ParsePath("chartable")).Bool(); err == nil {
		hints.Chartable = &chartable
	}

	if chartType, err := uiValue.LookupPath(cue.ParsePath("chartType")).String(); err == nil {
		hints.ChartType = chartType
	}

	if chartColor, err := uiValue.LookupPath(cue.ParsePath("chartColor")).String(); err == nil {
		hints.ChartColor = chartColor
	}

	property.UIHints = hints
	return nil
}

// parseAdvancedConfig parses advanced type-specific configuration
func (l *Loader) parseAdvancedConfig(value cue.Value, property *Property) error {
	advValue := value.LookupPath(cue.ParsePath("advanced"))
	if !advValue.Exists() {
		return nil
	}

	advanced := &AdvancedConfig{}

	// Pointer configuration
	if targetType, err := advValue.LookupPath(cue.ParsePath("targetType")).String(); err == nil {
		t := PropertyType(targetType)
		advanced.TargetType = &t
	}

	if maxDeref, err := advValue.LookupPath(cue.ParsePath("maxDereferences")).Uint64(); err == nil {
		m := uint(maxDeref)
		advanced.MaxDereferences = &m
	}

	if nullValue, err := advValue.LookupPath(cue.ParsePath("nullValue")).Uint64(); err == nil {
		n := uint32(nullValue)
		advanced.NullValue = &n
	}

	// Array configuration
	if elementType, err := advValue.LookupPath(cue.ParsePath("elementType")).String(); err == nil {
		t := PropertyType(elementType)
		advanced.ElementType = &t
	}

	if elementSize, err := advValue.LookupPath(cue.ParsePath("elementSize")).Uint64(); err == nil {
		s := uint(elementSize)
		advanced.ElementSize = &s
	}

	if dynLength, err := advValue.LookupPath(cue.ParsePath("dynamicLength")).Bool(); err == nil {
		advanced.DynamicLength = &dynLength
	}

	if lengthProp, err := advValue.LookupPath(cue.ParsePath("lengthProperty")).String(); err == nil {
		advanced.LengthProperty = lengthProp
	}

	if maxElements, err := advValue.LookupPath(cue.ParsePath("maxElements")).Uint64(); err == nil {
		m := uint(maxElements)
		advanced.MaxElements = &m
	}

	if indexOffset, err := advValue.LookupPath(cue.ParsePath("indexOffset")).Uint64(); err == nil {
		i := uint(indexOffset)
		advanced.IndexOffset = &i
	}

	if stride, err := advValue.LookupPath(cue.ParsePath("stride")).Uint64(); err == nil {
		s := uint(stride)
		advanced.Stride = &s
	}

	// Struct configuration
	if err := l.parseStructFields(advValue, advanced); err != nil {
		return err
	}

	if extends, err := advValue.LookupPath(cue.ParsePath("extends")).String(); err == nil {
		advanced.Extends = extends
	}

	// Enum configuration
	if err := l.parseEnumValues(advValue, advanced); err != nil {
		return err
	}

	if allowUnknown, err := advValue.LookupPath(cue.ParsePath("allowUnknownValues")).Bool(); err == nil {
		advanced.AllowUnknownValues = &allowUnknown
	}

	if defaultValue, err := advValue.LookupPath(cue.ParsePath("defaultValue")).Uint64(); err == nil {
		d := uint32(defaultValue)
		advanced.DefaultValue = &d
	}

	// Flags configuration
	if err := l.parseFlagDefinitions(advValue, advanced); err != nil {
		return err
	}

	// Time configuration
	if timeFormat, err := advValue.LookupPath(cue.ParsePath("timeFormat")).String(); err == nil {
		advanced.TimeFormat = timeFormat
	}

	if frameRate, err := advValue.LookupPath(cue.ParsePath("frameRate")).Float64(); err == nil {
		advanced.FrameRate = &frameRate
	}

	if epoch, err := advValue.LookupPath(cue.ParsePath("epoch")).String(); err == nil {
		advanced.Epoch = epoch
	}

	// Coordinate configuration
	if coordSystem, err := advValue.LookupPath(cue.ParsePath("coordinateSystem")).String(); err == nil {
		advanced.CoordinateSystem = coordSystem
	}

	if dimensions, err := advValue.LookupPath(cue.ParsePath("dimensions")).Uint64(); err == nil {
		d := uint(dimensions)
		advanced.Dimensions = &d
	}

	if units, err := advValue.LookupPath(cue.ParsePath("units")).String(); err == nil {
		advanced.Units = units
	}

	// Color configuration
	if colorFormat, err := advValue.LookupPath(cue.ParsePath("colorFormat")).String(); err == nil {
		advanced.ColorFormat = colorFormat
	}

	if alphaChannel, err := advValue.LookupPath(cue.ParsePath("alphaChannel")).Bool(); err == nil {
		advanced.AlphaChannel = &alphaChannel
	}

	if paletteRef, err := advValue.LookupPath(cue.ParsePath("paletteRef")).String(); err == nil {
		advanced.PaletteRef = paletteRef
	}

	// Percentage configuration
	if maxValue, err := advValue.LookupPath(cue.ParsePath("maxValue")).Float64(); err == nil {
		advanced.MaxValue = &maxValue
	}

	if precision, err := advValue.LookupPath(cue.ParsePath("precision")).Uint64(); err == nil {
		p := uint(precision)
		advanced.Precision = &p
	}

	// Version configuration
	if versionFormat, err := advValue.LookupPath(cue.ParsePath("versionFormat")).String(); err == nil {
		advanced.VersionFormat = versionFormat
	}

	// Checksum configuration
	if checksumAlgo, err := advValue.LookupPath(cue.ParsePath("checksumAlgorithm")).String(); err == nil {
		advanced.ChecksumAlgorithm = checksumAlgo
	}

	if err := l.parseChecksumRange(advValue, advanced); err != nil {
		return err
	}

	property.Advanced = advanced
	return nil
}

// parseStructFields parses struct field definitions
func (l *Loader) parseStructFields(advValue cue.Value, advanced *AdvancedConfig) error {
	fieldsValue := advValue.LookupPath(cue.ParsePath("fields"))
	if !fieldsValue.Exists() {
		return nil
	}

	fields := make(map[string]*StructField)
	fieldIter, _ := fieldsValue.Fields()
	for fieldIter.Next() {
		fieldName := fieldIter.Label()
		fieldValue := fieldIter.Value()

		field := &StructField{}

		if fieldType, err := fieldValue.LookupPath(cue.ParsePath("type")).String(); err == nil {
			field.Type = PropertyType(fieldType)
		}

		if offset, err := fieldValue.LookupPath(cue.ParsePath("offset")).Uint64(); err == nil {
			field.Offset = uint(offset)
		}

		if size, err := fieldValue.LookupPath(cue.ParsePath("size")).Uint64(); err == nil {
			s := uint(size)
			field.Size = &s
		}

		if description, err := fieldValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			field.Description = description
		}

		// Parse field transform
		if transformValue := fieldValue.LookupPath(cue.ParsePath("transform")); transformValue.Exists() {
			if transform, err := l.parseTransformValue(transformValue); err == nil {
				field.Transform = transform
			}
		}

		// Parse field validation
		if validationValue := fieldValue.LookupPath(cue.ParsePath("validation")); validationValue.Exists() {
			if validation, err := l.parseValidationValue(validationValue); err == nil {
				field.Validation = validation
			}
		}

		// Parse field computed
		if computedValue := fieldValue.LookupPath(cue.ParsePath("computed")); computedValue.Exists() {
			if computed, err := l.parseComputedValue(computedValue); err == nil {
				field.Computed = computed
			}
		}

		fields[fieldName] = field
	}

	if len(fields) > 0 {
		advanced.Fields = fields
	}

	return nil
}

// parseEnumValues parses enum value definitions
func (l *Loader) parseEnumValues(advValue cue.Value, advanced *AdvancedConfig) error {
	enumValue := advValue.LookupPath(cue.ParsePath("enumValues"))
	if !enumValue.Exists() {
		return nil
	}

	enumValues := make(map[string]*EnumValue)
	enumIter, _ := enumValue.Fields()
	for enumIter.Next() {
		enumKey := enumIter.Label()
		enumItemValue := enumIter.Value()

		enumItem := &EnumValue{}

		if value, err := enumItemValue.LookupPath(cue.ParsePath("value")).Uint64(); err == nil {
			enumItem.Value = uint32(value)
		}

		if description, err := enumItemValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			enumItem.Description = description
		}

		if color, err := enumItemValue.LookupPath(cue.ParsePath("color")).String(); err == nil {
			enumItem.Color = color
		}

		if icon, err := enumItemValue.LookupPath(cue.ParsePath("icon")).String(); err == nil {
			enumItem.Icon = icon
		}

		if deprecated, err := enumItemValue.LookupPath(cue.ParsePath("deprecated")).Bool(); err == nil {
			enumItem.Deprecated = &deprecated
		}

		enumValues[enumKey] = enumItem
	}

	if len(enumValues) > 0 {
		advanced.EnumValues = enumValues
	}

	return nil
}

// parseFlagDefinitions parses flag definitions
func (l *Loader) parseFlagDefinitions(advValue cue.Value, advanced *AdvancedConfig) error {
	flagsValue := advValue.LookupPath(cue.ParsePath("flagDefinitions"))
	if !flagsValue.Exists() {
		return nil
	}

	flagDefs := make(map[string]*FlagDefinition)
	flagIter, _ := flagsValue.Fields()
	for flagIter.Next() {
		flagName := flagIter.Label()
		flagValue := flagIter.Value()

		flag := &FlagDefinition{}

		if bit, err := flagValue.LookupPath(cue.ParsePath("bit")).Uint64(); err == nil {
			flag.Bit = uint(bit)
		}

		if description, err := flagValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			flag.Description = description
		}

		if invertLogic, err := flagValue.LookupPath(cue.ParsePath("invertLogic")).Bool(); err == nil {
			flag.InvertLogic = &invertLogic
		}

		if group, err := flagValue.LookupPath(cue.ParsePath("group")).String(); err == nil {
			flag.Group = group
		}

		// Parse mutually exclusive flags
		mutuallyExclusiveValue := flagValue.LookupPath(cue.ParsePath("mutuallyExclusive"))
		if mutuallyExclusiveValue.Exists() {
			if list, err := mutuallyExclusiveValue.List(); err == nil {
				for list.Next() {
					if exclusive, err := list.Value().String(); err == nil {
						flag.MutuallyExclusive = append(flag.MutuallyExclusive, exclusive)
					}
				}
			}
		}

		flagDefs[flagName] = flag
	}

	if len(flagDefs) > 0 {
		advanced.FlagDefinitions = flagDefs
	}

	return nil
}

// parseChecksumRange parses checksum range
func (l *Loader) parseChecksumRange(advValue cue.Value, advanced *AdvancedConfig) error {
	rangeValue := advValue.LookupPath(cue.ParsePath("checksumRange"))
	if !rangeValue.Exists() {
		return nil
	}

	checksumRange := &ChecksumRange{}

	if start, err := rangeValue.LookupPath(cue.ParsePath("start")).String(); err == nil {
		checksumRange.Start = start
	}

	if end, err := rangeValue.LookupPath(cue.ParsePath("end")).String(); err == nil {
		checksumRange.End = end
	}

	advanced.ChecksumRange = checksumRange
	return nil
}

// parsePerformanceHints parses performance hints
func (l *Loader) parsePerformanceHints(value cue.Value, property *Property) error {
	perfValue := value.LookupPath(cue.ParsePath("performance"))
	if !perfValue.Exists() {
		return nil
	}

	hints := &PerformanceHints{}

	if cacheable, err := perfValue.LookupPath(cue.ParsePath("cacheable")).Bool(); err == nil {
		hints.Cacheable = &cacheable
	}

	if cacheTimeout, err := perfValue.LookupPath(cue.ParsePath("cacheTimeout")).Uint64(); err == nil {
		timeout := uint(cacheTimeout)
		hints.CacheTimeout = &timeout
	}

	if readFrequency, err := perfValue.LookupPath(cue.ParsePath("readFrequency")).String(); err == nil {
		hints.ReadFrequency = readFrequency
	}

	if critical, err := perfValue.LookupPath(cue.ParsePath("critical")).Bool(); err == nil {
		hints.Critical = &critical
	}

	if batchable, err := perfValue.LookupPath(cue.ParsePath("batchable")).Bool(); err == nil {
		hints.Batchable = &batchable
	}

	if batchGroup, err := perfValue.LookupPath(cue.ParsePath("batchGroup")).String(); err == nil {
		hints.BatchGroup = batchGroup
	}

	property.Performance = hints
	return nil
}

// parseDebugConfig parses debug configuration
func (l *Loader) parseDebugConfig(value cue.Value, property *Property) error {
	debugValue := value.LookupPath(cue.ParsePath("debug"))
	if !debugValue.Exists() {
		return nil
	}

	debug := &DebugConfig{}

	if logReads, err := debugValue.LookupPath(cue.ParsePath("logReads")).Bool(); err == nil {
		debug.LogReads = &logReads
	}

	if logWrites, err := debugValue.LookupPath(cue.ParsePath("logWrites")).Bool(); err == nil {
		debug.LogWrites = &logWrites
	}

	if breakOnRead, err := debugValue.LookupPath(cue.ParsePath("breakOnRead")).Bool(); err == nil {
		debug.BreakOnRead = &breakOnRead
	}

	if breakOnWrite, err := debugValue.LookupPath(cue.ParsePath("breakOnWrite")).Bool(); err == nil {
		debug.BreakOnWrite = &breakOnWrite
	}

	if watchExpression, err := debugValue.LookupPath(cue.ParsePath("watchExpression")).String(); err == nil {
		debug.WatchExpression = watchExpression
	}

	property.Debug = debug
	return nil
}

// parseEnhancedTransform parses enhanced transformation rules
func (l *Loader) parseEnhancedTransform(value cue.Value, property *Property) error {
	transformValue := value.LookupPath(cue.ParsePath("transform"))
	if !transformValue.Exists() {
		return nil
	}

	transform, err := l.parseTransformValue(transformValue)
	if err != nil {
		return err
	}

	property.Transform = transform
	return nil
}

// parseMapperDebugConfig parses mapper-level debug configuration
func (l *Loader) parseMapperDebugConfig(value cue.Value, mapper *Mapper) error {
	debugValue := value.LookupPath(cue.ParsePath("debug"))
	if !debugValue.Exists() {
		return nil
	}

	debug := &MapperDebugConfig{}

	if enabled, err := debugValue.LookupPath(cue.ParsePath("enabled")).Bool(); err == nil {
		debug.Enabled = &enabled
	}

	if logLevel, err := debugValue.LookupPath(cue.ParsePath("logLevel")).String(); err == nil {
		debug.LogLevel = logLevel
	}

	if hotReload, err := debugValue.LookupPath(cue.ParsePath("hotReload")).Bool(); err == nil {
		debug.HotReload = &hotReload
	}

	if typeChecking, err := debugValue.LookupPath(cue.ParsePath("typeChecking")).Bool(); err == nil {
		debug.TypeChecking = &typeChecking
	}

	if memoryDumps, err := debugValue.LookupPath(cue.ParsePath("memoryDumps")).Bool(); err == nil {
		debug.MemoryDumps = &memoryDumps
	}

	// Parse log properties array
	logPropsValue := debugValue.LookupPath(cue.ParsePath("logProperties"))
	if logPropsValue.Exists() {
		if list, err := logPropsValue.List(); err == nil {
			for list.Next() {
				if prop, err := list.Value().String(); err == nil {
					debug.LogProperties = append(debug.LogProperties, prop)
				}
			}
		}
	}

	// Parse benchmark properties array
	benchmarkPropsValue := debugValue.LookupPath(cue.ParsePath("benchmarkProperties"))
	if benchmarkPropsValue.Exists() {
		if list, err := benchmarkPropsValue.List(); err == nil {
			for list.Next() {
				if prop, err := list.Value().String(); err == nil {
					debug.BenchmarkProperties = append(debug.BenchmarkProperties, prop)
				}
			}
		}
	}

	mapper.Debug = debug
	log.Printf("🐛 Parsed mapper debug configuration")
	return nil
}

// parseTransformValue parses a transform value
func (l *Loader) parseTransformValue(transformValue cue.Value) (*Transform, error) {
	transform := &Transform{}

	// Simple arithmetic
	if multiply, err := transformValue.LookupPath(cue.ParsePath("multiply")).Float64(); err == nil {
		transform.Multiply = &multiply
	}

	if add, err := transformValue.LookupPath(cue.ParsePath("add")).Float64(); err == nil {
		transform.Add = &add
	}

	if divide, err := transformValue.LookupPath(cue.ParsePath("divide")).Float64(); err == nil {
		transform.Divide = &divide
	}

	if subtract, err := transformValue.LookupPath(cue.ParsePath("subtract")).Float64(); err == nil {
		transform.Subtract = &subtract
	}

	if modulo, err := transformValue.LookupPath(cue.ParsePath("modulo")).Float64(); err == nil {
		transform.Modulo = &modulo
	}

	// Bitwise operations
	if bitwiseAnd, err := transformValue.LookupPath(cue.ParsePath("bitwiseAnd")).Uint64(); err == nil {
		and := uint32(bitwiseAnd)
		transform.BitwiseAnd = &and
	}

	if bitwiseOr, err := transformValue.LookupPath(cue.ParsePath("bitwiseOr")).Uint64(); err == nil {
		or := uint32(bitwiseOr)
		transform.BitwiseOr = &or
	}

	if bitwiseXor, err := transformValue.LookupPath(cue.ParsePath("bitwiseXor")).Uint64(); err == nil {
		xor := uint32(bitwiseXor)
		transform.BitwiseXor = &xor
	}

	if leftShift, err := transformValue.LookupPath(cue.ParsePath("leftShift")).Uint64(); err == nil {
		shift := uint32(leftShift)
		transform.LeftShift = &shift
	}

	if rightShift, err := transformValue.LookupPath(cue.ParsePath("rightShift")).Uint64(); err == nil {
		shift := uint32(rightShift)
		transform.RightShift = &shift
	}

	// Expression
	if expression, err := transformValue.LookupPath(cue.ParsePath("expression")).String(); err == nil {
		transform.Expression = expression
	}

	// Custom function
	if customFunction, err := transformValue.LookupPath(cue.ParsePath("customFunction")).String(); err == nil {
		transform.CustomFunction = customFunction
	}

	// Conditional transformations
	if err := l.parseConditionalTransforms(transformValue, transform); err != nil {
		return nil, err
	}

	// Lookup table
	if err := l.parseLookupTable(transformValue, transform); err != nil {
		return nil, err
	}

	// Range transformation
	if err := l.parseRangeTransform(transformValue, transform); err != nil {
		return nil, err
	}

	// String operations
	if err := l.parseStringOperations(transformValue, transform); err != nil {
		return nil, err
	}

	return transform, nil
}

// parseConditionalTransforms parses conditional transformations
func (l *Loader) parseConditionalTransforms(transformValue cue.Value, transform *Transform) error {
	conditionsValue := transformValue.LookupPath(cue.ParsePath("conditions"))
	if !conditionsValue.Exists() {
		return nil
	}

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
				condition.Then = l.parseAnyValue(thenValue)
			}

			// Parse else value (optional)
			elseValue := condValue.LookupPath(cue.ParsePath("else"))
			if elseValue.Exists() {
				condition.Else = l.parseAnyValue(elseValue)
			}

			transform.Conditions = append(transform.Conditions, condition)
		}
	}

	return nil
}

// parseLookupTable parses lookup table
func (l *Loader) parseLookupTable(transformValue cue.Value, transform *Transform) error {
	lookupValue := transformValue.LookupPath(cue.ParsePath("lookup"))
	if !lookupValue.Exists() {
		return nil
	}

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

	return nil
}

// parseRangeTransform parses range transformation
func (l *Loader) parseRangeTransform(transformValue cue.Value, transform *Transform) error {
	rangeValue := transformValue.LookupPath(cue.ParsePath("range"))
	if !rangeValue.Exists() {
		return nil
	}

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
	return nil
}

// parseStringOperations parses string operations
func (l *Loader) parseStringOperations(transformValue cue.Value, transform *Transform) error {
	stringOpsValue := transformValue.LookupPath(cue.ParsePath("stringOps"))
	if !stringOpsValue.Exists() {
		return nil
	}

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

	if truncate, err := stringOpsValue.LookupPath(cue.ParsePath("truncate")).Uint64(); err == nil {
		t := uint(truncate)
		stringOps.Truncate = &t
	}

	// Parse padding operations
	if padLeftValue := stringOpsValue.LookupPath(cue.ParsePath("padLeft")); padLeftValue.Exists() {
		padLeft := &PadOperation{}
		if length, err := padLeftValue.LookupPath(cue.ParsePath("length")).Uint64(); err == nil {
			padLeft.Length = uint(length)
		}
		if char, err := padLeftValue.LookupPath(cue.ParsePath("char")).String(); err == nil {
			padLeft.Char = char
		}
		stringOps.PadLeft = padLeft
	}

	if padRightValue := stringOpsValue.LookupPath(cue.ParsePath("padRight")); padRightValue.Exists() {
		padRight := &PadOperation{}
		if length, err := padRightValue.LookupPath(cue.ParsePath("length")).Uint64(); err == nil {
			padRight.Length = uint(length)
		}
		if char, err := padRightValue.LookupPath(cue.ParsePath("char")).String(); err == nil {
			padRight.Char = char
		}
		stringOps.PadRight = padRight
	}

	// Parse replace map
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
	return nil
}

// parseEnhancedValidation parses enhanced validation constraints
func (l *Loader) parseEnhancedValidation(value cue.Value, property *Property) error {
	validationValue := value.LookupPath(cue.ParsePath("validation"))
	if !validationValue.Exists() {
		return nil
	}

	validation, err := l.parseValidationValue(validationValue)
	if err != nil {
		return err
	}

	property.Validation = validation
	return nil
}

// parseValidationValue parses a validation value
func (l *Loader) parseValidationValue(validationValue cue.Value) (*PropertyValidation, error) {
	validation := &PropertyValidation{}

	if minValue, err := validationValue.LookupPath(cue.ParsePath("minValue")).Float64(); err == nil {
		validation.MinValue = &minValue
	}

	if maxValue, err := validationValue.LookupPath(cue.ParsePath("maxValue")).Float64(); err == nil {
		validation.MaxValue = &maxValue
	}

	if pattern, err := validationValue.LookupPath(cue.ParsePath("pattern")).String(); err == nil {
		validation.Pattern = pattern
	}

	if required, err := validationValue.LookupPath(cue.ParsePath("required")).Bool(); err == nil {
		validation.Required = required
	}

	if constraint, err := validationValue.LookupPath(cue.ParsePath("constraint")).String(); err == nil {
		validation.Constraint = constraint
	}

	if crossValidation, err := validationValue.LookupPath(cue.ParsePath("crossValidation")).String(); err == nil {
		validation.CrossValidation = crossValidation
	}

	// Parse depends on
	dependsOnValue := validationValue.LookupPath(cue.ParsePath("dependsOn"))
	if dependsOnValue.Exists() {
		if list, err := dependsOnValue.List(); err == nil {
			for list.Next() {
				if dep, err := list.Value().String(); err == nil {
					validation.DependsOn = append(validation.DependsOn, dep)
				}
			}
		}
	}

	// Parse custom messages
	messagesValue := validationValue.LookupPath(cue.ParsePath("messages"))
	if messagesValue.Exists() {
		messages := make(map[string]string)
		fields, _ := messagesValue.Fields()
		for fields.Next() {
			key := fields.Label()
			if msg, err := fields.Value().String(); err == nil {
				messages[key] = msg
			}
		}
		if len(messages) > 0 {
			validation.Messages = messages
		}
	}

	// Parse allowed values
	allowedValue := validationValue.LookupPath(cue.ParsePath("allowedValues"))
	if allowedValue.Exists() {
		if list, err := allowedValue.List(); err == nil {
			for list.Next() {
				validation.AllowedValues = append(validation.AllowedValues, l.parseAnyValue(list.Value()))
			}
		}
	}

	return validation, nil
}

// parsePropertyComputed parses computed property
func (l *Loader) parsePropertyComputed(value cue.Value, property *Property) error {
	computedValue := value.LookupPath(cue.ParsePath("computed"))
	if !computedValue.Exists() {
		return nil
	}

	computed, err := l.parseComputedValue(computedValue)
	if err != nil {
		return err
	}

	property.Computed = computed
	return nil
}

// parseComputedValue parses a computed value
func (l *Loader) parseComputedValue(computedValue cue.Value) (*ComputedProperty, error) {
	computed := &ComputedProperty{}

	if expression, err := computedValue.LookupPath(cue.ParsePath("expression")).String(); err == nil {
		computed.Expression = expression
	}

	if propType, err := computedValue.LookupPath(cue.ParsePath("type")).String(); err == nil {
		computed.Type = PropertyType(propType)
	}

	if cached, err := computedValue.LookupPath(cue.ParsePath("cached")).Bool(); err == nil {
		computed.Cached = &cached
	}

	// Parse dependencies
	depsValue := computedValue.LookupPath(cue.ParsePath("dependencies"))
	if depsValue.Exists() {
		if list, err := depsValue.List(); err == nil {
			for list.Next() {
				if dep, err := list.Value().String(); err == nil {
					computed.Dependencies = append(computed.Dependencies, dep)
				}
			}
		}
	}

	// Parse cache invalidation
	cacheInvalidationValue := computedValue.LookupPath(cue.ParsePath("cacheInvalidation"))
	if cacheInvalidationValue.Exists() {
		if list, err := cacheInvalidationValue.List(); err == nil {
			for list.Next() {
				if inv, err := list.Value().String(); err == nil {
					computed.CacheInvalidation = append(computed.CacheInvalidation, inv)
				}
			}
		}
	}

	return computed, nil
}

// parsePropertyCharMap parses character map for strings
func (l *Loader) parsePropertyCharMap(value cue.Value, property *Property) error {
	charMapValue := value.LookupPath(cue.ParsePath("charMap"))
	if !charMapValue.Exists() {
		return nil
	}

	charMap, err := l.parseCharMapValue(charMapValue)
	if err != nil {
		return err
	}

	property.CharMap = charMap
	return nil
}

// parseCharMapValue parses a character map value
func (l *Loader) parseCharMapValue(charMapValue cue.Value) (map[uint8]string, error) {
	charMap := make(map[uint8]string)

	fields, _ := charMapValue.Fields()
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

// parseEnhancedPropertyGroups parses enhanced property groups
func (l *Loader) parseEnhancedPropertyGroups(value cue.Value, groups map[string]*PropertyGroup) error {
	groupsValue := value.LookupPath(cue.ParsePath("groups"))
	if !groupsValue.Exists() {
		return nil
	}

	fields, _ := groupsValue.Fields()
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

		if color, err := groupValue.LookupPath(cue.ParsePath("color")).String(); err == nil {
			group.Color = color
		}

		if displayMode, err := groupValue.LookupPath(cue.ParsePath("displayMode")).String(); err == nil {
			group.DisplayMode = displayMode
		}

		if sortBy, err := groupValue.LookupPath(cue.ParsePath("sortBy")).String(); err == nil {
			group.SortBy = sortBy
		}

		if filterBy, err := groupValue.LookupPath(cue.ParsePath("filterBy")).String(); err == nil {
			group.FilterBy = filterBy
		}

		if renderer, err := groupValue.LookupPath(cue.ParsePath("customRenderer")).String(); err == nil {
			group.Renderer = renderer
		}

		if priority, err := groupValue.LookupPath(cue.ParsePath("priority")).Uint64(); err == nil {
			p := uint(priority)
			group.Priority = &p
		}

		// Parse conditional display
		if err := l.parseConditionalDisplay(groupValue, group); err != nil {
			return err
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

		// Parse subgroups
		if err := l.parseSubgroups(groupValue, group); err != nil {
			return err
		}

		groups[label] = group
	}

	log.Printf("📂 Parsed %d enhanced property groups", len(groups))
	return nil
}

// parseConditionalDisplay parses conditional display for groups
func (l *Loader) parseConditionalDisplay(groupValue cue.Value, group *PropertyGroup) error {
	condDisplayValue := groupValue.LookupPath(cue.ParsePath("conditionalDisplay"))
	if !condDisplayValue.Exists() {
		return nil
	}

	conditional := &ConditionalDisplay{}

	if expression, err := condDisplayValue.LookupPath(cue.ParsePath("expression")).String(); err == nil {
		conditional.Expression = expression
	}

	// Parse dependencies
	depsValue := condDisplayValue.LookupPath(cue.ParsePath("dependencies"))
	if depsValue.Exists() {
		if list, err := depsValue.List(); err == nil {
			for list.Next() {
				if dep, err := list.Value().String(); err == nil {
					conditional.Dependencies = append(conditional.Dependencies, dep)
				}
			}
		}
	}

	group.Conditional = conditional
	return nil
}

// parseSubgroups parses subgroups
func (l *Loader) parseSubgroups(groupValue cue.Value, group *PropertyGroup) error {
	subgroupsValue := groupValue.LookupPath(cue.ParsePath("subgroups"))
	if !subgroupsValue.Exists() {
		return nil
	}

	subgroups := make(map[string]*PropertyGroup)
	fields, _ := subgroupsValue.Fields()
	for fields.Next() {
		subgroupName := fields.Label()
		subgroupValue := fields.Value()

		subgroup := &PropertyGroup{
			Name: subgroupName,
		}

		// Parse subgroup properties (simplified version)
		if desc, err := subgroupValue.LookupPath(cue.ParsePath("description")).String(); err == nil {
			subgroup.Description = desc
		}

		if collapsed, err := subgroupValue.LookupPath(cue.ParsePath("collapsed")).Bool(); err == nil {
			subgroup.Collapsed = collapsed
		}

		// Parse subgroup properties list
		propsValue := subgroupValue.LookupPath(cue.ParsePath("properties"))
		if propsValue.Exists() {
			if list, err := propsValue.List(); err == nil {
				for list.Next() {
					if prop, err := list.Value().String(); err == nil {
						subgroup.Properties = append(subgroup.Properties, prop)
					}
				}
			}
		}

		subgroups[subgroupName] = subgroup
	}

	if len(subgroups) > 0 {
		group.Subgroups = subgroups
	}

	return nil
}

// parseComputedProperties parses computed properties at mapper level
func (l *Loader) parseComputedProperties(value cue.Value, computed map[string]*ComputedProperty) error {
	computedValue := value.LookupPath(cue.ParsePath("computed"))
	if !computedValue.Exists() {
		return nil
	}

	fields, _ := computedValue.Fields()
	for fields.Next() {
		label := fields.Label()
		compValue := fields.Value()

		comp, err := l.parseComputedValue(compValue)
		if err != nil {
			return err
		}

		computed[label] = comp
	}

	log.Printf("🧮 Parsed %d computed properties", len(computed))
	return nil
}

// parseEvents parses events configuration
func (l *Loader) parseEvents(value cue.Value, mapper *Mapper) error {
	eventsValue := value.LookupPath(cue.ParsePath("events"))
	if !eventsValue.Exists() {
		return nil
	}

	events := &EventsConfig{}

	if onLoad, err := eventsValue.LookupPath(cue.ParsePath("onLoad")).String(); err == nil {
		events.OnLoad = onLoad
	}

	if onUnload, err := eventsValue.LookupPath(cue.ParsePath("onUnload")).String(); err == nil {
		events.OnUnload = onUnload
	}

	if onPropertyChanged, err := eventsValue.LookupPath(cue.ParsePath("onPropertyChanged")).String(); err == nil {
		events.OnPropertyChanged = onPropertyChanged
	}

	// Parse custom events
	customValue := eventsValue.LookupPath(cue.ParsePath("custom"))
	if customValue.Exists() {
		customEvents := make(map[string]*CustomEvent)
		fields, _ := customValue.Fields()
		for fields.Next() {
			eventName := fields.Label()
			eventValue := fields.Value()

			customEvent := &CustomEvent{}

			if trigger, err := eventValue.LookupPath(cue.ParsePath("trigger")).String(); err == nil {
				customEvent.Trigger = trigger
			}

			if action, err := eventValue.LookupPath(cue.ParsePath("action")).String(); err == nil {
				customEvent.Action = action
			}

			// Parse dependencies
			depsValue := eventValue.LookupPath(cue.ParsePath("dependencies"))
			if depsValue.Exists() {
				if list, err := depsValue.List(); err == nil {
					for list.Next() {
						if dep, err := list.Value().String(); err == nil {
							customEvent.Dependencies = append(customEvent.Dependencies, dep)
						}
					}
				}
			}

			customEvents[eventName] = customEvent
		}

		if len(customEvents) > 0 {
			events.Custom = customEvents
		}
	}

	mapper.Events = events
	log.Printf("📅 Parsed events configuration with %d custom events", len(events.Custom))
	return nil
}

// parseGlobalValidation parses global validation rules
func (l *Loader) parseGlobalValidation(value cue.Value, mapper *Mapper) error {
	validationValue := value.LookupPath(cue.ParsePath("globalValidation"))
	if !validationValue.Exists() {
		return nil
	}

	globalValidation := &GlobalValidation{}

	// Parse memory layout validation
	memLayoutValue := validationValue.LookupPath(cue.ParsePath("memoryLayout"))
	if memLayoutValue.Exists() {
		memLayout := &MemoryLayoutValidation{}

		if checkOverlaps, err := memLayoutValue.LookupPath(cue.ParsePath("checkOverlaps")).Bool(); err == nil {
			memLayout.CheckOverlaps = &checkOverlaps
		}

		if checkBounds, err := memLayoutValue.LookupPath(cue.ParsePath("checkBounds")).Bool(); err == nil {
			memLayout.CheckBounds = &checkBounds
		}

		if checkAlignment, err := memLayoutValue.LookupPath(cue.ParsePath("checkAlignment")).Bool(); err == nil {
			memLayout.CheckAlignment = &checkAlignment
		}

		globalValidation.MemoryLayout = memLayout
	}

	// Parse cross validation rules
	crossValidationValue := validationValue.LookupPath(cue.ParsePath("crossValidation"))
	if crossValidationValue.Exists() {
		if list, err := crossValidationValue.List(); err == nil {
			for list.Next() {
				ruleValue := list.Value()
				rule := CrossValidationRule{}

				if name, err := ruleValue.LookupPath(cue.ParsePath("name")).String(); err == nil {
					rule.Name = name
				}

				if expression, err := ruleValue.LookupPath(cue.ParsePath("expression")).String(); err == nil {
					rule.Expression = expression
				}

				if message, err := ruleValue.LookupPath(cue.ParsePath("message")).String(); err == nil {
					rule.Message = message
				}

				// Parse dependencies
				depsValue := ruleValue.LookupPath(cue.ParsePath("dependencies"))
				if depsValue.Exists() {
					if depsList, err := depsValue.List(); err == nil {
						for depsList.Next() {
							if dep, err := depsList.Value().String(); err == nil {
								rule.Dependencies = append(rule.Dependencies, dep)
							}
						}
					}
				}

				globalValidation.CrossValidation = append(globalValidation.CrossValidation, rule)
			}
		}
	}

	// Parse performance validation
	perfValidationValue := validationValue.LookupPath(cue.ParsePath("performance"))
	if perfValidationValue.Exists() {
		perfValidation := &PerformanceValidation{}

		if maxProperties, err := perfValidationValue.LookupPath(cue.ParsePath("maxProperties")).Uint64(); err == nil {
			max := uint(maxProperties)
			perfValidation.MaxProperties = &max
		}

		if maxComputedDepth, err := perfValidationValue.LookupPath(cue.ParsePath("maxComputedDepth")).Uint64(); err == nil {
			max := uint(maxComputedDepth)
			perfValidation.MaxComputedDepth = &max
		}

		if warnSlowProperties, err := perfValidationValue.LookupPath(cue.ParsePath("warnSlowProperties")).Bool(); err == nil {
			perfValidation.WarnSlowProperties = &warnSlowProperties
		}

		globalValidation.Performance = perfValidation
	}

	mapper.Validation = globalValidation
	log.Printf("✅ Parsed global validation with %d cross-validation rules", len(globalValidation.CrossValidation))
	return nil
}

// parseProcessingSteps parses preprocessing and postprocessing steps
func (l *Loader) parseProcessingSteps(value cue.Value, mapper *Mapper) error {
	// Parse preprocessing
	preprocessValue := value.LookupPath(cue.ParsePath("preprocess"))
	if preprocessValue.Exists() {
		if list, err := preprocessValue.List(); err == nil {
			for list.Next() {
				if step, err := list.Value().String(); err == nil {
					mapper.Preprocess = append(mapper.Preprocess, step)
				}
			}
		}
	}

	// Parse postprocessing
	postprocessValue := value.LookupPath(cue.ParsePath("postprocess"))
	if postprocessValue.Exists() {
		if list, err := postprocessValue.List(); err == nil {
			for list.Next() {
				if step, err := list.Value().String(); err == nil {
					mapper.Postprocess = append(mapper.Postprocess, step)
				}
			}
		}
	}

	if len(mapper.Preprocess) > 0 || len(mapper.Postprocess) > 0 {
		log.Printf("⚙️  Parsed processing steps: %d preprocess, %d postprocess",
			len(mapper.Preprocess), len(mapper.Postprocess))
	}

	return nil
}

// parseAnyValue parses a CUE value into appropriate Go type
func (l *Loader) parseAnyValue(value cue.Value) interface{} {
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
			arr = append(arr, l.parseAnyValue(list.Value()))
		}
		return arr
	} else if fields, err := value.Fields(); err == nil {
		obj := make(map[string]interface{})
		for fields.Next() {
			obj[fields.Label()] = l.parseAnyValue(fields.Value())
		}
		return obj
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

// ===== ENHANCED PROPERTY OPERATIONS =====

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

// GetProperty gets a property value from memory with enhanced processing
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
	case PropertyTypeArray:
		return m.processArrayProperty(prop, rawBytes, littleEndian)
	case PropertyTypeStruct:
		return m.processStructProperty(prop, rawBytes, littleEndian)
	case PropertyTypeEnum:
		return m.processEnumProperty(prop, rawBytes, littleEndian)
	case PropertyTypeFlags:
		return m.processFlagsProperty(prop, rawBytes, littleEndian)
	case PropertyTypeCoordinate:
		return m.processCoordinateProperty(prop, rawBytes, littleEndian)
	case PropertyTypeColor:
		return m.processColorProperty(prop, rawBytes, littleEndian)
	case PropertyTypePercentage:
		return m.processPercentageProperty(prop, rawBytes, littleEndian)
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
	result, transformErr := m.applyEnhancedTransform(raw, prop.Transform)
	if transformErr != nil {
		result = raw // Use raw value if transform fails
	}

	// Validate (but don't fail on validation errors)
	m.validateValue(result, prop.Validation)

	return result, nil
}

// Helper methods for enhanced property processing

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
	case PropertyTypeArray:
		return []interface{}{}
	case PropertyTypeStruct:
		return map[string]interface{}{}
	case PropertyTypeEnum:
		return map[string]interface{}{"value": 0, "name": "Unknown"}
	case PropertyTypeFlags:
		return map[string]interface{}{"value": 0, "flags": map[string]bool{}, "active_flags": []string{}}
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

// Enhanced property type processors

func (m *Mapper) processArrayProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	if prop.Advanced == nil || prop.Advanced.ElementType == nil {
		return []interface{}{}, nil
	}

	elementType := *prop.Advanced.ElementType
	elementSize := uint(1)
	if prop.Advanced.ElementSize != nil {
		elementSize = *prop.Advanced.ElementSize
	}

	length := prop.Length
	if prop.Advanced.DynamicLength != nil && *prop.Advanced.DynamicLength {
		// TODO: Get length from another property
		// For now, use the specified length
	}

	result := make([]interface{}, 0, length)

	for i := uint32(0); i < length && uint(i)*elementSize < uint(len(rawBytes)); i++ {
		offset := uint(i) * elementSize
		if offset+elementSize > uint(len(rawBytes)) {
			break
		}

		elementBytes := rawBytes[offset : offset+elementSize]
		element := m.parseElementFromBytes(elementBytes, elementType, littleEndian)
		result = append(result, element)
	}

	return result, nil
}

func (m *Mapper) processStructProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	if prop.Advanced == nil || prop.Advanced.Fields == nil {
		return map[string]interface{}{}, nil
	}

	result := make(map[string]interface{})

	for fieldName, field := range prop.Advanced.Fields {
		if field.Offset >= uint(len(rawBytes)) {
			continue
		}

		fieldSize := uint(1)
		if field.Size != nil {
			fieldSize = *field.Size
		}

		if field.Offset+fieldSize > uint(len(rawBytes)) {
			continue
		}

		fieldBytes := rawBytes[field.Offset : field.Offset+fieldSize]
		fieldValue := m.parseElementFromBytes(fieldBytes, field.Type, littleEndian)

		// Apply field transform if present
		if field.Transform != nil {
			if transformed, err := m.applyEnhancedTransform(fieldValue, field.Transform); err == nil {
				fieldValue = transformed
			}
		}

		result[fieldName] = fieldValue
	}

	return result, nil
}

func (m *Mapper) processEnumProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	// Parse raw value
	rawValue := m.parseElementFromBytes(rawBytes, PropertyTypeUint32, littleEndian)

	if prop.Advanced != nil && prop.Advanced.EnumValues != nil {
		// Look for matching enum value
		for key, enumValue := range prop.Advanced.EnumValues {
			if enumValue.Value == rawValue.(uint32) {
				return map[string]interface{}{
					"value": rawValue,
					"name":  enumValue.Description,
					"key":   key,
					"color": enumValue.Color,
					"icon":  enumValue.Icon,
				}, nil
			}
		}
	}

	return map[string]interface{}{
		"value": rawValue,
		"name":  fmt.Sprintf("Unknown(%v)", rawValue),
	}, nil
}

func (m *Mapper) processFlagsProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	rawValue := m.parseElementFromBytes(rawBytes, PropertyTypeUint32, littleEndian)
	intValue := rawValue.(uint32)

	flags := make(map[string]bool)
	activeFlags := make([]string, 0)

	if prop.Advanced != nil && prop.Advanced.FlagDefinitions != nil {
		for flagName, flagDef := range prop.Advanced.FlagDefinitions {
			invertLogic := flagDef.InvertLogic != nil && *flagDef.InvertLogic
			isSet := (intValue & (1 << flagDef.Bit)) != 0

			if invertLogic {
				isSet = !isSet
			}

			flags[flagName] = isSet
			if isSet {
				activeFlags = append(activeFlags, flagName)
			}
		}
	}

	return map[string]interface{}{
		"value":        intValue,
		"flags":        flags,
		"active_flags": activeFlags,
	}, nil
}

func (m *Mapper) processCoordinateProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	if len(rawBytes) < 4 {
		return map[string]interface{}{"x": 0, "y": 0}, nil
	}

	x := m.parseElementFromBytes(rawBytes[0:2], PropertyTypeUint16, littleEndian)
	y := m.parseElementFromBytes(rawBytes[2:4], PropertyTypeUint16, littleEndian)

	result := map[string]interface{}{
		"x": x,
		"y": y,
	}

	// Add Z coordinate if available
	if len(rawBytes) >= 6 {
		z := m.parseElementFromBytes(rawBytes[4:6], PropertyTypeUint16, littleEndian)
		result["z"] = z
	}

	// Add coordinate system info from advanced config
	if prop.Advanced != nil {
		if prop.Advanced.CoordinateSystem != "" {
			result["system"] = prop.Advanced.CoordinateSystem
		}
		if prop.Advanced.Units != "" {
			result["units"] = prop.Advanced.Units
		}
	}

	return result, nil
}

func (m *Mapper) processColorProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	if len(rawBytes) < 2 {
		return map[string]interface{}{"r": 0, "g": 0, "b": 0, "a": 255}, nil
	}

	rawValue := m.parseElementFromBytes(rawBytes, PropertyTypeUint32, littleEndian)
	intValue := rawValue.(uint32)

	// Default format
	format := "rgb565"
	if prop.Advanced != nil && prop.Advanced.ColorFormat != "" {
		format = prop.Advanced.ColorFormat
	}

	var r, g, b, a uint8

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
	default:
		// Default to rgb565
		r = uint8((intValue>>11)&0x1F) << 3
		g = uint8((intValue>>5)&0x3F) << 2
		b = uint8(intValue&0x1F) << 3
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
		"format":    format,
	}, nil
}

func (m *Mapper) processPercentageProperty(prop *Property, rawBytes []byte, littleEndian bool) (interface{}, error) {
	rawValue := m.parseElementFromBytes(rawBytes, PropertyTypeUint32, littleEndian)

	// Default max value
	maxValue := 100.0
	if prop.Advanced != nil && prop.Advanced.MaxValue != nil {
		maxValue = *prop.Advanced.MaxValue
	}

	floatValue := float64(rawValue.(uint32))
	percentage := (floatValue / maxValue) * 100

	return map[string]interface{}{
		"raw_value":  rawValue,
		"percentage": percentage,
		"decimal":    floatValue / maxValue,
		"max_value":  maxValue,
	}, nil
}

func (m *Mapper) parseElementFromBytes(data []byte, elementType PropertyType, littleEndian bool) interface{} {
	switch elementType {
	case PropertyTypeUint8:
		if len(data) >= 1 {
			return uint32(data[0])
		}
	case PropertyTypeUint16:
		if len(data) >= 2 {
			if littleEndian {
				return uint32(binary.LittleEndian.Uint16(data))
			} else {
				return uint32(binary.BigEndian.Uint16(data))
			}
		}
	case PropertyTypeUint32:
		if len(data) >= 4 {
			if littleEndian {
				return binary.LittleEndian.Uint32(data)
			} else {
				return binary.BigEndian.Uint32(data)
			}
		}
	case PropertyTypeInt8:
		if len(data) >= 1 {
			return int32(int8(data[0]))
		}
	case PropertyTypeInt16:
		if len(data) >= 2 {
			if littleEndian {
				return int32(int16(binary.LittleEndian.Uint16(data)))
			} else {
				return int32(int16(binary.BigEndian.Uint16(data)))
			}
		}
	case PropertyTypeInt32:
		if len(data) >= 4 {
			if littleEndian {
				return int32(binary.LittleEndian.Uint32(data))
			} else {
				return int32(binary.BigEndian.Uint32(data))
			}
		}
	case PropertyTypeBool:
		if len(data) >= 1 {
			return data[0] != 0
		}
	case PropertyTypeString:
		return string(data)
	}
	return uint32(0)
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

	// Handle special computed properties that we can process without CUE
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

// validateValue validates a value against enhanced validation constraints
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
			message := fmt.Sprintf("value %f is below minimum %f", numValue, *validation.MinValue)
			if validation.Messages != nil && validation.Messages["minValue"] != "" {
				message = validation.Messages["minValue"]
			}
			return fmt.Errorf(message)
		}

		if validation.MaxValue != nil && numValue > *validation.MaxValue {
			message := fmt.Sprintf("value %f is above maximum %f", numValue, *validation.MaxValue)
			if validation.Messages != nil && validation.Messages["maxValue"] != "" {
				message = validation.Messages["maxValue"]
			}
			return fmt.Errorf(message)
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

	// TODO: Implement pattern validation (regex)
	// TODO: Implement constraint validation (CUE expressions)
	// TODO: Implement cross-property validation

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
			data = []byte{val}
		} else if val, ok := value.(float64); ok {
			data = []byte{uint8(val)}
		} else if val, ok := value.(uint32); ok {
			data = []byte{uint8(val)}
		} else {
			return fmt.Errorf("invalid type for uint8 property")
		}
	case PropertyTypeBool:
		if val, ok := value.(bool); ok {
			if val {
				data = []byte{1}
			} else {
				data = []byte{0}
			}
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

// applyEnhancedTransform applies enhanced transformation rules to a raw value
func (m *Mapper) applyEnhancedTransform(value interface{}, transform *Transform) (interface{}, error) {
	if transform == nil {
		return value, nil
	}

	// Convert value to float64 for numeric operations
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

	result := value

	// Apply simple arithmetic transforms
	if isNumeric {
		if transform.Multiply != nil {
			numValue *= *transform.Multiply
		}
		if transform.Add != nil {
			numValue += *transform.Add
		}
		if transform.Divide != nil && *transform.Divide != 0 {
			numValue /= *transform.Divide
		}
		if transform.Subtract != nil {
			numValue -= *transform.Subtract
		}
		if transform.Modulo != nil && *transform.Modulo != 0 {
			numValue = float64(int64(numValue) % int64(*transform.Modulo))
		}

		result = numValue
	}

	// Apply bitwise operations (for integer values)
	if isNumeric {
		intValue := uint32(numValue)

		if transform.BitwiseAnd != nil {
			intValue &= *transform.BitwiseAnd
		}
		if transform.BitwiseOr != nil {
			intValue |= *transform.BitwiseOr
		}
		if transform.BitwiseXor != nil {
			intValue ^= *transform.BitwiseXor
		}
		if transform.LeftShift != nil {
			intValue <<= *transform.LeftShift
		}
		if transform.RightShift != nil {
			intValue >>= *transform.RightShift
		}

		result = intValue
	}

	// Apply range transformation
	if transform.Range != nil && isNumeric {
		r := transform.Range
		inputRange := r.InputMax - r.InputMin
		outputRange := r.OutputMax - r.OutputMin

		if inputRange != 0 {
			normalized := (numValue - r.InputMin) / inputRange
			mapped := r.OutputMin + (normalized * outputRange)

			if r.Clamp {
				if mapped < r.OutputMin {
					mapped = r.OutputMin
				} else if mapped > r.OutputMax {
					mapped = r.OutputMax
				}
			}

			result = mapped
		}
	}

	// Apply lookup transformation
	if transform.Lookup != nil {
		key := fmt.Sprintf("%v", result)
		if lookupValue, exists := transform.Lookup[key]; exists {
			result = lookupValue
		}
	}

	// Apply string operations
	if transform.StringOps != nil {
		if strValue, ok := result.(string); ok {
			result = m.applyStringOperations(strValue, transform.StringOps)
		}
	}

	// TODO: Apply conditional transforms
	// TODO: Apply CUE expressions
	// TODO: Apply custom functions

	return result, nil
}

// applyStringOperations applies string transformation operations
func (m *Mapper) applyStringOperations(value string, ops *StringOperations) string {
	result := value

	if ops.Trim {
		result = strings.TrimSpace(result)
	}

	if ops.Uppercase {
		result = strings.ToUpper(result)
	}

	if ops.Lowercase {
		result = strings.ToLower(result)
	}

	// Apply replacements
	for old, new := range ops.Replace {
		result = strings.ReplaceAll(result, old, new)
	}

	// Apply truncation
	if ops.Truncate != nil && len(result) > int(*ops.Truncate) {
		result = result[:*ops.Truncate]
	}

	// Apply padding
	if ops.PadLeft != nil {
		for len(result) < int(ops.PadLeft.Length) {
			result = ops.PadLeft.Char + result
		}
	}

	if ops.PadRight != nil {
		for len(result) < int(ops.PadRight.Length) {
			result = result + ops.PadRight.Char
		}
	}

	return result
}
