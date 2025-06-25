package mappers

// Property type enumeration
#PropertyType: "uint8" | "uint16" | "uint32" | "int8" | "int16" | "int32" | "string" | "bool" | "bitfield"

// Endianness enumeration
#Endian: "little" | "big"

// Property transformation for value mapping and conversion
#Transform: {
multiply?: number
add?: number
lookup?: [string]: string
}

// Memory block definition for platforms
#MemoryBlock: {
name: string
start: string  // hex address like "0x0000"
end: string    // hex address like "0x07FF"
}

// Platform definition with memory layout
#Platform: {
name: string
endian: #Endian
memoryBlocks: [...#MemoryBlock]
}

// Property definition with all possible attributes
#Property: {
name: string
type: #PropertyType
address: string  // hex address like "0x075A" or expression

// Optional attributes
length?: uint      // for strings/arrays
endian?: #Endian   // override platform endian
altName?: string   // alternative name for same property
description?: string

// Value transformation
transform?: #Transform

// String-specific character mapping
charMap?: [string]: string  // "0x41": "A", "0x42": "B", etc.
}

// Complete mapper definition
#Mapper: {
// Metadata
name: string
game: string
platform: #Platform

// Property definitions
properties: [string]: #Property
}

// Example usage (not part of schema, just for reference):
//
// package mario
//
// mario: #Mapper & {
//     name: "Super Mario Bros"
//     game: "Super Mario Bros."
//     platform: {
//         name: "NES"
//         endian: "little"
//         memoryBlocks: [{
//             name: "RAM"
//             start: "0x0000"
//             end: "0x07FF"
//         }]
//     }
//     properties: {
//         marioLives: {
//             name: "marioLives"
//             type: "uint8"
//             address: "0x075A"
//             description: "Number of lives remaining"
//         }
//         coinCount: {
//             name: "coinCount"
//             type: "uint8"
//             address: "0x075E"
//             transform: {
//                 lookup: {
//                     "99": "MAX"
//                 }
//             }
//         }
//         powerupState: {
//             name: "powerupState"
//             type: "uint8"
//             address: "0x0756"
//             transform: {
//                 lookup: {
//                     "0": "Small Mario"
//                     "1": "Super Mario"
//                     "2": "Fire Mario"
//                 }
//             }
//         }
//     }
// }