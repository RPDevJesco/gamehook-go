package mappers

import "strings"

// Enhanced property type enumeration with all supported types
#PropertyType: "uint8" | "uint16" | "uint32" | "int8" | "int16" | "int32" |
              "string" | "bool" | "bitfield" | "bcd" | "bit" | "nibble" |
              "float32" | "float64" | "pointer" | "array" | "struct" |
              "enum" | "flags" | "time" | "version" | "coordinate" | "color" |
              "percentage" | "checksum"

// Endianness enumeration
#Endian: "little" | "big"

// Range transformation for value mapping
#RangeTransform: {
    inputMin: number
    inputMax: number
    outputMin: number
    outputMax: number
    clamp?: bool  // clamp to output range
}

// Conditional transformation
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
}

// Enhanced transformation with expressions and conditions
#Transform: {
    // Simple arithmetic
    multiply?: number
    add?: number
    divide?: number
    subtract?: number

    // CUE expressions (evaluated at runtime)
    expression?: string  // CUE expression like "value * 0.1" or "value + baseOffset"

    // Conditional transformations
    conditions?: [...#ConditionalTransform]

    // Lookup tables
    lookup?: [string]: string

    // Range mapping
    range?: #RangeTransform

    // String transformations
    stringOps?: #StringOperations
}

// Property validation using CUE constraints
#PropertyValidation: {
    minValue?: number
    maxValue?: number
    allowedValues?: [..._]
    pattern?: string  // regex pattern
    required?: bool

    // CUE expression for custom validation
    constraint?: string  // like "value >= 0 && value <= 255"
}

// Computed property definition
#ComputedProperty: {
    expression: string      // CUE expression using other property values
    dependencies: [...string]  // properties this computation depends on
    type?: #PropertyType
}

// Memory block definition with enhanced features
#MemoryBlock: {
    name: string
    start: string  // hex address like "0x0000"
    end: string    // hex address like "0x07FF"

    // Optional CUE expression for dynamic addresses
    startExpr?: string  // CUE expression that resolves to start address
    endExpr?: string    // CUE expression that resolves to end address

    // Block metadata
    description?: string
    readable?: bool    // default true
    writable?: bool    // default true
    cacheable?: bool   // default true
}

// Enhanced platform with constants and base addresses
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
}

// Property group for UI organization
#PropertyGroup: {
    name: string
    description?: string
    icon?: string  // for UI (e.g., "🎮", "💰", "⚡")
    properties: [...string]  // property names in this group
    collapsed?: bool  // default UI state
    color?: string   // hex color for UI theming
}

// Enhanced property definition with all features
#Property: {
    name: string
    type: #PropertyType
    address: string  // hex address or CUE expression

    // Optional attributes
    length?: uint | string      // can be number or CUE expression
    position?: uint             // for bit/nibble properties (0-7 for bits, 0-1 for nibbles)
    size?: uint                 // element size for arrays/structs
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
    readExpression?: string   // CUE expression to process raw bytes
    writeExpression?: string  // CUE expression to convert value to bytes

    // Dependencies on other properties
    dependsOn?: [...string]   // property names this depends on

    // Computed properties (derived from other properties)
    computed?: #ComputedProperty

    // UI hints
    uiHints?: {
        displayFormat?: string  // "hex", "decimal", "binary", "percentage"
        unit?: string          // "bytes", "seconds", "pixels", etc.
        precision?: uint       // decimal places for floats
        showInList?: bool      // show in main property list (default true)
        category?: string      // custom category for grouping
    }

    // Advanced type-specific configuration
    advanced?: {
        // For pointer types
        targetType?: #PropertyType
        maxDereferences?: uint

        // For array types
        elementType?: #PropertyType
        elementSize?: uint
        dynamicLength?: bool
        lengthProperty?: string  // property name that contains array length

        // For struct types
        fields?: [string]: {
            type: #PropertyType
            offset: uint
            size?: uint
        }

        // For enum types
        enumValues?: [string]: {
            value: number
            description?: string
            color?: string
        }

        // For flags/bitfield types
        flagDefinitions?: [string]: {
            bit: uint
            description?: string
            invertLogic?: bool  // true if flag is active when bit is 0
        }

        // For time types
        timeFormat?: "frames" | "milliseconds" | "seconds" | "unix"
        frameRate?: number    // for frame-based time

        // For coordinate types
        coordinateSystem?: "cartesian" | "screen" | "polar"
        dimensions?: uint     // 2D, 3D, etc.

        // For color types
        colorFormat?: "rgb565" | "argb8888" | "rgb888" | "palette"
        alphaChannel?: bool

        // For percentage types
        maxValue?: number     // what value represents 100%

        // For version types
        versionFormat?: "major.minor.patch" | "bcd" | "packed"
    }
}

// Enhanced mapper with global expressions and metadata
#Mapper: {
    // Metadata
    name: string
    game: string
    version?: string           // mapper version (semver recommended)
    minGameHookVersion?: string // minimum required GameHook version
    author?: string
    description?: string
    website?: string

    // Platform configuration
    platform: #Platform

    // Global constants accessible in all property expressions
    constants?: [string]: _

    // Global preprocessing expressions
    preprocess?: [...string]   // CUE expressions run before property evaluation
    postprocess?: [...string]  // CUE expressions run after property evaluation

    // Property definitions
    properties: [string]: #Property

    // Property groups for organization
    groups?: [string]: #PropertyGroup

    // Computed values derived from multiple properties
    computed?: [string]: #ComputedProperty

    // Mapper configuration
    config?: {
        updateInterval?: string      // override default update interval
        enableAutoFreeze?: bool      // auto-freeze changed properties
        validateOnLoad?: bool        // validate all properties on mapper load
        enableStatistics?: bool      // track property access statistics
        cacheProperties?: bool       // cache property values
        logChanges?: bool           // log property changes
    }

    // UI configuration
    ui?: {
        theme?: string              // "dark", "light", "retro"
        primaryColor?: string       // hex color
        layout?: "grid" | "list" | "tree"
        defaultGroup?: string       // group to show by default
        showAddresses?: bool        // show memory addresses in UI
        showTypes?: bool           // show property types in UI
        compactMode?: bool         // use compact display
    }

    // Mapper-specific constants and helper functions
    helpers?: [string]: string     // CUE helper functions

    // Version history and changelog
    changelog?: [...{
        version: string
        date?: string
        changes: [...string]
    }]
}

// Common character maps that can be reused
#PokemonCharMap: {
    "0x50": " "      // Space
    "0x80": "A", "0x81": "B", "0x82": "C", "0x83": "D", "0x84": "E"
    "0x85": "F", "0x86": "G", "0x87": "H", "0x88": "I", "0x89": "J"
    "0x8A": "K", "0x8B": "L", "0x8C": "M", "0x8D": "N", "0x8E": "O"
    "0x8F": "P", "0x90": "Q", "0x91": "R", "0x92": "S", "0x93": "T"
    "0x94": "U", "0x95": "V", "0x96": "W", "0x97": "X", "0x98": "Y"
    "0x99": "Z"
    "0x9A": "(", "0x9B": ")", "0x9C": ":", "0x9D": ";"
    "0xA0": "a", "0xA1": "b", "0xA2": "c", "0xA3": "d", "0xA4": "e"
    "0xA5": "f", "0xA6": "g", "0xA7": "h", "0xA8": "i", "0xA9": "j"
    "0xAA": "k", "0xAB": "l", "0xAC": "m", "0xAD": "n", "0xAE": "o"
    "0xAF": "p", "0xB0": "q", "0xB1": "r", "0xB2": "s", "0xB3": "t"
    "0xB4": "u", "0xB5": "v", "0xB6": "w", "0xB7": "x", "0xB8": "y"
    "0xB9": "z"
    "0xF7": "♂"      // Male symbol
    "0xF8": "♀"      // Female symbol
    "0xFF": ""       // Terminator
}

#ASCIICharMap: {
    "0x20": " ", "0x21": "!", "0x22": "\"", "0x23": "#", "0x24": "$"
    "0x25": "%", "0x26": "&", "0x27": "'", "0x28": "(", "0x29": ")"
    "0x2A": "*", "0x2B": "+", "0x2C": ",", "0x2D": "-", "0x2E": "."
    "0x2F": "/", "0x30": "0", "0x31": "1", "0x32": "2", "0x33": "3"
    "0x34": "4", "0x35": "5", "0x36": "6", "0x37": "7", "0x38": "8"
    "0x39": "9", "0x3A": ":", "0x3B": ";", "0x3C": "<", "0x3D": "="
    "0x3E": ">", "0x3F": "?", "0x40": "@"
    "0x41": "A", "0x42": "B", "0x43": "C", "0x44": "D", "0x45": "E"
    "0x46": "F", "0x47": "G", "0x48": "H", "0x49": "I", "0x4A": "J"
    "0x4B": "K", "0x4C": "L", "0x4D": "M", "0x4E": "N", "0x4F": "O"
    "0x50": "P", "0x51": "Q", "0x52": "R", "0x53": "S", "0x54": "T"
    "0x55": "U", "0x56": "V", "0x57": "W", "0x58": "X", "0x59": "Y"
    "0x5A": "Z"
    "0x61": "a", "0x62": "b", "0x63": "c", "0x64": "d", "0x65": "e"
    "0x66": "f", "0x67": "g", "0x68": "h", "0x69": "i", "0x6A": "j"
    "0x6B": "k", "0x6C": "l", "0x6D": "m", "0x6E": "n", "0x6F": "o"
    "0x70": "p", "0x71": "q", "0x72": "r", "0x73": "s", "0x74": "t"
    "0x75": "u", "0x76": "v", "0x77": "w", "0x78": "x", "0x79": "y"
    "0x7A": "z"
}

// Common platform definitions
#GameBoyPlatform: #Platform & {
    name: "Game Boy"
    endian: "little"
    description: "Nintendo Game Boy (1989)"
    manufacturer: "Nintendo"
    releaseYear: 1989

    constants: {
        ramBase: 0xC000
        vramBase: 0x8000
        oamBase: 0xFE00
        ioBase: 0xFF00
    }

    baseAddresses: {
        "wram": "0xC000"
        "vram": "0x8000"
        "oam": "0xFE00"
        "io": "0xFF00"
    }

    memoryBlocks: [
        {
            name: "WRAM Bank 0"
            start: "0xC000"
            end: "0xCFFF"
            description: "Work RAM Bank 0"
        },
        {
            name: "WRAM Bank 1"
            start: "0xD000"
            end: "0xDFFF"
            description: "Work RAM Bank 1"
        }
    ]
}

#NESPlatform: #Platform & {
    name: "NES"
    endian: "little"
    description: "Nintendo Entertainment System (1985)"
    manufacturer: "Nintendo"
    releaseYear: 1985

    constants: {
        ramBase: 0x0000
        ppuBase: 0x2000
        apuBase: 0x4000
    }

    baseAddresses: {
        "ram": "0x0000"
        "ppu": "0x2000"
        "apu": "0x4000"
    }

    memoryBlocks: [
        {
            name: "RAM"
            start: "0x0000"
            end: "0x07FF"
            description: "Internal RAM"
        },
        {
            name: "PPU Registers"
            start: "0x2000"
            end: "0x2007"
            description: "PPU Registers"
            cacheable: false
        }
    ]
}