package mappers

import "strings"
import "list"

// ===== ENHANCED PROPERTY TYPE SYSTEM =====

// Core property types with rich metadata
#PropertyType: "uint8" | "uint16" | "uint32" | "int8" | "int16" | "int32" |
              "string" | "bool" | "bitfield" | "bcd" | "bit" | "nibble" |
              "float32" | "float64" | "pointer" | "array" | "struct" |
              "enum" | "flags" | "time" | "version" | "coordinate" | "color" |
              "percentage" | "checksum"

// Endianness enumeration
#Endian: "little" | "big"

// ===== REUSABLE TRANSFORM SYSTEM =====

// Common transform patterns as reusable definitions
#PercentageTransform: {
    expression: "round((value / 255) * 100, 2)"
    validation: {
        minValue: 0
        maxValue: 100
    }
    uiHints: {
        displayFormat: "percentage"
        unit: "%"
        precision: 2
    }
}

#PPTransform: {
    expression: "value % 64"
    validation: {
        minValue: 0
        maxValue: 63
    }
    uiHints: {
        displayFormat: "decimal"
        unit: "PP"
    }
}

#ExperienceTransform: {
    expression: "value & 0xFFFFFF" // Mask to 24 bits
    validation: {
        minValue: 0
        maxValue: 16777215
    }
    uiHints: {
        displayFormat: "decimal"
        unit: "EXP"
    }
}

#LevelTransform: {
    validation: {
        minValue: 1
        maxValue: 100
    }
    uiHints: {
        displayFormat: "decimal"
        unit: "Lv"
    }
}

#MoneyTransform: {
    expression: "bcdToDecimal(value)"
    validation: {
        minValue: 0
        maxValue: 999999
    }
    uiHints: {
        displayFormat: "currency"
        unit: "₽"
    }
}

// IV extraction transforms
#IVTransforms: {
    attack: {
        expression: "(value >> 12) & 0xF"
        validation: {minValue: 0, maxValue: 15}
    }
    defense: {
        expression: "(value >> 8) & 0xF"
        validation: {minValue: 0, maxValue: 15}
    }
    speed: {
        expression: "(value >> 4) & 0xF"
        validation: {minValue: 0, maxValue: 15}
    }
    special: {
        expression: "value & 0xF"
        validation: {minValue: 0, maxValue: 15}
    }
}

// ===== ENHANCED TRANSFORMATION SYSTEM =====

// Range transformation for value mapping
#RangeTransform: {
    inputMin: number
    inputMax: number
    outputMin: number
    outputMax: number
    clamp?: bool // clamp to output range
}

// Conditional transformation with multiple conditions
#ConditionalTransform: {
    if: string      // CUE condition like "value > 100"
    then: _         // value to return if condition is true
    else?: _        // optional else value
}

// String transformation operations
#StringOperations: {
    trim?: bool
    uppercase?: bool
    lowercase?: bool
    replace?: [string]: string
    truncate?: uint
    padLeft?: {
        length: uint
        char: string
    }
    padRight?: {
        length: uint
        char: string
    }
}

// Comprehensive transformation system
#Transform: {
    // Simple arithmetic
    multiply?: number
    add?: number
    divide?: number
    subtract?: number
    modulo?: number

    // Bitwise operations
    bitwiseAnd?: number
    bitwiseOr?: number
    bitwiseXor?: number
    leftShift?: number
    rightShift?: number

    // CUE expressions (evaluated at runtime)
    expression?: string // CUE expression like "value * 0.1" or "bcdToDecimal(value)"

    // Conditional transformations
    conditions?: [...#ConditionalTransform]

    // Lookup tables
    lookup?: [string]: string

    // Range mapping
    range?: #RangeTransform

    // String transformations
    stringOps?: #StringOperations

    // Custom functions
    customFunction?: string // Reference to predefined function
}

// ===== VALIDATION SYSTEM =====

// Enhanced property validation with rich constraints
#PropertyValidation: {
    minValue?: number
    maxValue?: number
    allowedValues?: [..._]
    pattern?: string // regex pattern
    required?: bool

    // CUE expression for custom validation
    constraint?: string // like "value >= 0 && value <= 255"

    // Cross-property validation
    dependsOn?: [...string]
    crossValidation?: string // CUE expression using multiple properties

    // Custom validation messages
    messages?: {
        minValue?: string
        maxValue?: string
        pattern?: string
        constraint?: string
    }
}

// ===== ADVANCED TYPE DEFINITIONS =====

// Computed property definition with caching
#ComputedProperty: {
    expression: string // CUE expression using other property values
    dependencies: [...string] // properties this computation depends on
    type?: #PropertyType
    cached?: bool // whether to cache the result
    cacheInvalidation?: [...string] // properties that invalidate cache
}

// Enhanced memory block with validation
#MemoryBlock: {
    name: string
    start: string // hex address like "0x0000"
    end: string   // hex address like "0x07FF"

    // Optional CUE expression for dynamic addresses
    startExpr?: string // CUE expression that resolves to start address
    endExpr?: string   // CUE expression that resolves to end address

    // Block metadata
    description?: string
    readable?: bool    // default true
    writable?: bool    // default true
    cacheable?: bool   // default true

    // Access patterns for optimization
    accessPattern?: "sequential" | "random" | "sparse"

    // Memory protection
    protected?: bool   // prevent accidental writes
    watchable?: bool   // trigger events on changes
}

// Enhanced platform with rich configuration
#Platform: {
    name: string
    endian: #Endian
    memoryBlocks: [...#MemoryBlock]

    // Platform-specific constants accessible in expressions
    constants?: [string]: _

    // Platform base addresses for easier property definition
    baseAddresses?: [string]: string

    // Platform metadata
    description?: string
    version?: string
    manufacturer?: string
    releaseYear?: number

    // Platform capabilities
    capabilities?: {
        maxMemorySize?: number
        addressBusWidth?: number
        dataBusWidth?: number
        hasMemoryMapping?: bool
        supportsBanking?: bool
    }

    // Performance hints
    performance?: {
        readLatency?: number // milliseconds
        writeLatency?: number
        batchSize?: number   // optimal batch read size
    }
}

// ===== UI ORGANIZATION SYSTEM =====

// Property group with advanced UI features
#PropertyGroup: {
    name: string
    description?: string
    icon?: string // for UI (e.g., "🎮", "💰", "⚡")
    properties: [...string] // property names in this group
    collapsed?: bool // default UI state
    color?: string   // hex color for UI theming

    // Advanced UI features
    displayMode?: "table" | "cards" | "tree" | "custom"
    sortBy?: string // property to sort by
    filterBy?: string // CUE expression for filtering

    // Conditional display
    conditionalDisplay?: {
        expression: string // CUE condition for when to show group
        dependencies: [...string]
    }

    // Subgroups for hierarchical organization
    subgroups?: [string]: #PropertyGroup

    // Custom rendering hints
    customRenderer?: string // reference to custom UI component
}

// ===== ADVANCED PROPERTY SYSTEM =====

// Enhanced property definition with all features
#Property: {
    name: string
    type: #PropertyType
    address: string // hex address or CUE expression

    // Optional attributes
    length?: uint | string // can be number or CUE expression
    position?: uint        // for bit/nibble properties (0-7 for bits, 0-1 for nibbles)
    size?: uint           // element size for arrays/structs
    endian?: #Endian
    description?: string
    readOnly?: bool

    // Value transformation
    transform?: #Transform

    // Validation rules
    validation?: #PropertyValidation

    // Character mapping for strings
    charMap?: [string]: string

    // Freezing support
    freezable?: bool
    defaultFrozen?: bool

    // Custom read/write logic as CUE expressions
    readExpression?: string  // CUE expression to process raw bytes
    writeExpression?: string // CUE expression to convert value to bytes

    // Dependencies on other properties
    dependsOn?: [...string] // property names this depends on

    // Computed properties (derived from other properties)
    computed?: #ComputedProperty

    // UI hints
    uiHints?: {
        displayFormat?: "hex" | "decimal" | "binary" | "percentage" | "currency" | "time" | "custom"
        unit?: string          // "bytes", "seconds", "pixels", etc.
        precision?: uint       // decimal places for floats
        showInList?: bool      // show in main property list (default true)
        category?: string      // custom category for grouping
        priority?: uint        // display priority (higher = more prominent)
        tooltip?: string       // tooltip text

        // Visual styling
        color?: string         // hex color for value display
        icon?: string         // icon for property
        badge?: string        // badge text (like "NEW" or "HOT")

        // Interactive features
        editable?: bool       // can be edited in UI
        copyable?: bool       // can be copied to clipboard
        watchable?: bool      // can be watched for changes

        // Chart/graph hints
        chartable?: bool      // can be displayed in charts
        chartType?: "line" | "bar" | "pie" | "gauge"
        chartColor?: string
    }

    // Advanced type-specific configuration
    advanced?: {
        // For pointer types
        targetType?: #PropertyType
        maxDereferences?: uint
        nullValue?: number // what value represents null

        // For array types
        elementType?: #PropertyType
        elementSize?: uint
        dynamicLength?: bool
        lengthProperty?: string // property name that contains array length
        maxElements?: uint

        // Array access patterns
        indexOffset?: uint     // starting index (default 0)
        stride?: uint         // bytes between elements (default elementSize)

        // For struct types
        fields?: [string]: {
            type: #PropertyType
            offset: uint
            size?: uint
            transform?: #Transform
            validation?: #PropertyValidation
            description?: string
            computed?: #ComputedProperty
        }

        // Struct inheritance
        extends?: string      // inherit from another struct type

        // For enum types
        enumValues?: [string]: {
            value: number
            description?: string
            color?: string
            icon?: string
            deprecated?: bool
        }

        // Enum behavior
        allowUnknownValues?: bool
        defaultValue?: number

        // For flags/bitfield types
        flagDefinitions?: [string]: {
            bit: uint
            description?: string
            invertLogic?: bool // true if flag is active when bit is 0
            group?: string    // group related flags
            mutuallyExclusive?: [...string] // flags that can't be set together
        }

        // For time types
        timeFormat?: "frames" | "milliseconds" | "seconds" | "unix" | "bcd"
        frameRate?: number    // for frame-based time
        epoch?: string       // epoch for unix time

        // For coordinate types
        coordinateSystem?: "cartesian" | "screen" | "polar" | "geographic"
        dimensions?: uint     // 2D, 3D, etc.
        units?: string       // "pixels", "meters", "degrees"

        // For color types
        colorFormat?: "rgb565" | "argb8888" | "rgb888" | "rgba8888" | "palette" | "yuv"
        alphaChannel?: bool
        paletteRef?: string  // reference to palette property

        // For percentage types
        maxValue?: number    // what value represents 100%
        precision?: uint     // decimal places

        // For version types
        versionFormat?: "major.minor.patch" | "bcd" | "packed" | "string"

        // For checksum types
        checksumAlgorithm?: "crc16" | "crc32" | "md5" | "sha1" | "custom"
        checksumRange?: {
            start: string // start address
            end: string   // end address
        }
    }

    // Performance optimization hints
    performance?: {
        cacheable?: bool     // can result be cached
        cacheTimeout?: uint  // cache timeout in milliseconds
        readFrequency?: "high" | "medium" | "low" // how often this is read
        critical?: bool      // critical for performance

        // Batch reading hints
        batchable?: bool     // can be batched with other reads
        batchGroup?: string  // group for batching
    }

    // Debug and development features
    debug?: {
        logReads?: bool      // log all reads of this property
        logWrites?: bool     // log all writes of this property
        breakOnRead?: bool   // debugger break on read
        breakOnWrite?: bool  // debugger break on write
        watchExpression?: string // CUE expression for conditional watching
    }
}

// ===== REFERENCE TYPE SYSTEM =====

// Centralized reference types for consistency
#ReferenceTypes: {
    // Common game reference types
    pokemonSpecies?: #Property & {
        type: "enum"
        advanced: {
            enumValues: [string]: {
                value: number
                description: string
                color?: string
                type1?: string
                type2?: string
                baseStats?: {
                    hp: number
                    attack: number
                    defense: number
                    speed: number
                    special: number
                }
            }
        }
    }

    pokemonTypes?: #Property & {
        type: "enum"
        advanced: {
            enumValues: [string]: {
                value: number
                description: string
                color: string
                effectiveness?: [string]: number // type effectiveness chart
            }
        }
    }

    moves?: #Property & {
        type: "enum"
        advanced: {
            enumValues: [string]: {
                value: number
                description: string
                type: string
                power?: number
                accuracy?: number
                pp: number
                effect?: string
            }
        }
    }

    items?: #Property & {
        type: "enum"
        advanced: {
            enumValues: [string]: {
                value: number
                description: string
                category: "medicine" | "pokeball" | "tm" | "berry" | "key" | "misc"
                price?: number
                effect?: string
            }
        }
    }
}

// ===== GLOBAL SYSTEMS =====

// Global character encoding systems
#CharacterMaps: {
    pokemon: {
        "0x50": " "
        "0x80": "A", "0x81": "B", "0x82": "C", "0x83": "D", "0x84": "E"
        "0x85": "F", "0x86": "G", "0x87": "H", "0x88": "I", "0x89": "J"
        "0x8A": "K", "0x8B": "L", "0x8C": "M", "0x8D": "N", "0x8E": "O"
        "0x8F": "P", "0x90": "Q", "0x91": "R", "0x92": "S", "0x93": "T"
        "0x94": "U", "0x95": "V", "0x96": "W", "0x97": "X", "0x98": "Y"
        "0x99": "Z"
        "0xA0": "a", "0xA1": "b", "0xA2": "c", "0xA3": "d", "0xA4": "e"
        "0xA5": "f", "0xA6": "g", "0xA7": "h", "0xA8": "i", "0xA9": "j"
        "0xAA": "k", "0xAB": "l", "0xAC": "m", "0xAD": "n", "0xAE": "o"
        "0xAF": "p", "0xB0": "q", "0xB1": "r", "0xB2": "s", "0xB3": "t"
        "0xB4": "u", "0xB5": "v", "0xB6": "w", "0xB7": "x", "0xB8": "y"
        "0xB9": "z"
        "0xFF": ""
    }

    ascii: {
        // Standard ASCII mapping
        for i in list.Range(32, 127, 1) {
            "\(i)": string.FromBytes([i])
        }
    }

    custom?: [string]: string // Allow custom character maps
}

// Reusable string type with character map
#PokemonString: #Property & {
    type: "string"
    charMap: #CharacterMaps.pokemon
    validation: {
        maxLength: 11
        pattern: "^[A-Za-z0-9 ]*$"
    }
    transform: {
        stringOps: {
            trim: true
        }
    }
}

// ===== ENHANCED MAPPER DEFINITION =====

// Complete mapper with global expressions and rich metadata
#Mapper: {
    // Metadata
    name: string
    game: string
    version?: string           // mapper version (semver recommended)
    minGameHookVersion?: string // minimum required GameHook version
    author?: string
    description?: string
    website?: string
    license?: string

    // Mapper metadata
    metadata?: {
        created: string      // ISO date
        modified: string     // ISO date
        tags: [...string]    // searchable tags
        category: string     // game category
        language: string     // primary language
        region: string       // game region
        revision?: string    // game revision/version
    }

    // Platform configuration
    platform: #Platform

    // Global constants accessible in all property expressions
    constants?: [string]: _

    // Global character maps
    characterMaps?: #CharacterMaps

    // Reference types for this mapper
    references?: #ReferenceTypes

    // Global preprocessing expressions
    preprocess?: [...string] // CUE expressions run before property evaluation
    postprocess?: [...string] // CUE expressions run after property evaluation

    // Property definitions
    properties: [string]: #Property

    // Property groups for organization
    groups?: [string]: #PropertyGroup

    // Computed values derived from multiple properties
    computed?: [string]: #ComputedProperty

    // Global validation rules
    globalValidation?: {
        // Memory layout validation
        memoryLayout?: {
            checkOverlaps?: bool      // verify no overlapping properties
            checkBounds?: bool        // verify all addresses are in valid ranges
            checkAlignment?: bool     // verify proper alignment
        }

        // Cross-property validation
        crossValidation?: [...{
            name: string
            expression: string      // CUE expression
            dependencies: [...string]
            message?: string
        }]

        // Performance validation
        performance?: {
            maxProperties?: uint     // maximum number of properties
            maxComputedDepth?: uint  // maximum dependency depth
            warnSlowProperties?: bool // warn about potentially slow properties
        }
    }

    // Event system
    events?: {
        onLoad?: string          // CUE expression run when mapper loads
        onUnload?: string        // CUE expression run when mapper unloads
        onPropertyChanged?: string // CUE expression run when any property changes

        // Custom events
        custom?: [string]: {
            trigger: string      // CUE expression for when to trigger
            action: string       // CUE expression for what to do
            dependencies: [...string]
        }
    }

    // Debugging and development features
    debug?: {
        enabled?: bool
        logLevel?: "trace" | "debug" | "info" | "warn" | "error"
        logProperties?: [...string] // properties to log
        benchmarkProperties?: [...string] // properties to benchmark

        // Development tools
        hotReload?: bool         // enable hot reloading
        typeChecking?: bool      // strict type checking
        memoryDumps?: bool       // enable memory dumps
    }
}

// ===== VALIDATION CONSTRAINTS =====

// Structured validation result
#ValidationResult: {
    isValid: bool
    message: string
}

// Global validation constraints
#MapperConstraints: {
    // Ensure all property names are unique
    #uniquePropertyNames: #ValidationResult & {
        isValid: bool & (len([for name, _ in properties { name }]) ==
                len(list.Unique([for name, _ in properties { name }])))
        message: string = "Property names must be unique"
    }

    // Ensure all group properties exist
    #validGroupProperties: #ValidationResult & {
        isValid: bool & (groups == null || list.All([
            for groupName, group in groups {
                list.All([
                    for propName in group.properties {
                        list.Contains([for name, _ in properties { name }], propName)
                    }
                ])
            }
        ]))
        message: string = "All properties referenced in groups must exist in the properties section"
    }

    // Ensure computed property dependencies exist
    #validComputedDependencies: #ValidationResult & {
        isValid: bool & (computed == null || list.All([
            for compName, comp in computed {
                list.All([
                    for depName in comp.dependencies {
                        list.Contains([for name, _ in properties { name }], depName)
                    }
                ])
            }
        ]))
        message: string = "All computed property dependencies must reference existing properties"
    }

    // Ensure memory addresses don't overlap (simplified version)
    #noMemoryOverlaps: #ValidationResult & {
        isValid: bool & true // Simplified for now - can be enhanced with proper address parsing
        message: string = "Memory addresses must not overlap between properties"
    }

    // Ensure array element sizes are consistent
		#validArraySizes: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties if prop.type == "array" {
								(prop.advanced.elementSize > 0) | (prop.advanced.elementSize == null)
						}
				])
				message: string = "Array element sizes must be greater than 0"
		}


    // Ensure bit positions are valid for bit/nibble types
		#validBitPositions: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties {
								(prop.type == "bit" && prop.position >= 0 && prop.position <= 7) |
								(prop.type == "nibble" && prop.position >= 0 && prop.position <= 1) |
								(prop.type != "bit" && prop.type != "nibble") |
								(!defined(prop.position))
						}
				])
				message: string = "Bit positions must be 0-7 for bit type, 0-1 for nibble type"
		}

    // Ensure struct field offsets are valid
		#validStructOffsets: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties if prop.type == "struct" {
								!defined(prop.advanced.fields) ||
								list.All([
										for _, field in prop.advanced.fields {
												field.offset >= 0
										}
								])
						}
				])
				message: string = "Struct field offsets must be non-negative"
		}

    // Ensure enum values are unique
		#uniqueEnumValues: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties if prop.type == "enum" {
								!defined(prop.advanced.enumValues) ||
								len([for _, v in prop.advanced.enumValues { v.value }]) ==
								len(list.Unique([for _, v in prop.advanced.enumValues { v.value }]))
						}
				])
				message: string = "Enum values must be unique within each enum property"
		}

    // Ensure flag bit positions are unique and valid
		#validFlagBits: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties if prop.type == "flags" {
								!defined(prop.advanced.flagDefinitions) ||
								list.All([
										for _, flag in prop.advanced.flagDefinitions {
												flag.bit >= 0 && flag.bit < ((prop.length | *1) * 8)
										}
								]) &&
								len([for _, flag in prop.advanced.flagDefinitions { flag.bit }]) ==
								len(list.Unique([for _, flag in prop.advanced.flagDefinitions { flag.bit }]))
						}
				])
				message: string = "Flag bit positions must be unique and within valid range for the property length"
		}

    // Ensure validation constraints are properly formed
		#validValidationConstraints: #ValidationResult & {
				isValid: bool & list.All([
						for _, prop in properties {
								!defined(prop.validation.minValue) ||
								!defined(prop.validation.maxValue) ||
								(prop.validation.minValue <= prop.validation.maxValue)
						}
				])
				message: string = "Validation minValue must be less than or equal to maxValue"
		}
}

// Apply constraints to mapper
#ValidatedMapper: #Mapper & #MapperConstraints