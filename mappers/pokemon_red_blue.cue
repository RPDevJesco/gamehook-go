package pokemon_red_blue

name: "pokemon_red_blue"
game: "Pokemon Red/Blue"
version: "2.0.0"
minGameHookVersion: "0.6.0"
author: "GameHook Enhanced Team"
description: "Enhanced Pokemon Red/Blue mapper with advanced property management, validation, and freezing capabilities"
website: "https://github.com/gamehook/mappers"

// Enhanced platform configuration
platform: {
    name: "Game Boy"
    endian: "little"
    description: "Nintendo Game Boy running Pokemon Red/Blue"
    manufacturer: "Nintendo"
    releaseYear: 1989

    constants: {
        ramBase: 0xC000
        pokemonDataSize: 44
        maxPokemonLevel: 100
        maxPartySize: 6
        maxMoney: 999999
    }

    baseAddresses: {
        "wram": "0xC000"
        "party": "0xD163"
        "money": "0xD347"
        "player": "0xD158"
    }

    memoryBlocks: [
        {
            name: "WRAM Bank 0"
            start: "0xC000"
            end: "0xCFFF"
            description: "Work RAM Bank 0 - Contains battle data, active Pokemon"
        },
        {
            name: "WRAM Bank 1"
            start: "0xD000"
            end: "0xDFFF"
            description: "Work RAM Bank 1 - Contains party data, player data, save data"
        }
    ]
}

// Global constants accessible in expressions
constants: {
    maxLevel: 100
    maxHp: 999
    maxMoney: 999999
    pokemonSpeciesCount: 151
    badgeCount: 8
}

// Mapper configuration
config: {
    updateInterval: "16ms"      // 60fps monitoring
    enableAutoFreeze: false
    validateOnLoad: true
    enableStatistics: true
    cacheProperties: true
    logChanges: true
}

// UI configuration
ui: {
    theme: "retro"
    primaryColor: "#FFD700"     // Pokemon yellow
    layout: "grid"
    defaultGroup: "trainer"
    showAddresses: true
    showTypes: true
    compactMode: false
}

// Enhanced properties with validation, freezing, and advanced types
properties: {
    // === TRAINER INFORMATION ===
    playerName: {
        name: "playerName"
        type: "string"
        address: "0xD158"
        length: 11
        description: "Player's name (trainer name)"
        charMap: #PokemonCharMap
        validation: {
            required: true
            pattern: "^[A-Z][A-Z ]*$"  // Must start with capital letter
        }
        freezable: true
        uiHints: {
            category: "trainer"
            displayFormat: "text"
            showInList: true
        }
    }

    rivalName: {
        name: "rivalName"
        type: "string"
        address: "0xD34A"
        length: 8
        description: "Rival's name"
        charMap: #PokemonCharMap
        validation: {
            required: true
        }
        freezable: true
        uiHints: {
            category: "trainer"
            displayFormat: "text"
        }
    }

    money: {
        name: "money"
        type: "bcd"
        address: "0xD347"
        length: 3
        description: "Player's money in BCD format"
        validation: {
            minValue: 0
            maxValue: 999999
        }
        transform: {
            // Money is stored as BCD, but we want to validate the decimal value
            expression: "value"
        }
        freezable: true
        defaultFrozen: false
        uiHints: {
            category: "trainer"
            displayFormat: "currency"
            unit: "₽"
            showInList: true
        }
    }

    badges: {
        name: "badges"
        type: "flags"
        address: "0xD356"
        description: "Badge collection bitfield"
        advanced: {
            flagDefinitions: {
                "boulder": {bit: 0, description: "Boulder Badge (Brock)"}
                "cascade": {bit: 1, description: "Cascade Badge (Misty)"}
                "thunder": {bit: 2, description: "Thunder Badge (Lt. Surge)"}
                "rainbow": {bit: 3, description: "Rainbow Badge (Erika)"}
                "soul": {bit: 4, description: "Soul Badge (Koga)"}
                "marsh": {bit: 5, description: "Marsh Badge (Sabrina)"}
                "volcano": {bit: 6, description: "Volcano Badge (Blaine)"}
                "earth": {bit: 7, description: "Earth Badge (Giovanni)"}
            }
        }
        validation: {
            constraint: "value >= 0 && value <= 255"
        }
        freezable: true
        uiHints: {
            category: "progression"
            displayFormat: "flag_list"
            showInList: true
        }
    }

    currentMap: {
        name: "currentMap"
        type: "enum"
        address: "0xD35E"
        description: "Current map/location ID"
        advanced: {
            enumValues: {
                "pallet_town": {value: 0, description: "Pallet Town", color: "#90EE90"}
                "viridian_city": {value: 1, description: "Viridian City", color: "#228B22"}
                "pewter_city": {value: 2, description: "Pewter City", color: "#A9A9A9"}
                "cerulean_city": {value: 3, description: "Cerulean City", color: "#87CEEB"}
                "lavender_town": {value: 4, description: "Lavender Town", color: "#E6E6FA"}
                "vermilion_city": {value: 5, description: "Vermilion City", color: "#FF6347"}
                "celadon_city": {value: 6, description: "Celadon City", color: "#98FB98"}
                "fuchsia_city": {value: 7, description: "Fuchsia City", color: "#FF69B4"}
                "cinnabar_island": {value: 8, description: "Cinnabar Island", color: "#DC143C"}
                "indigo_plateau": {value: 9, description: "Indigo Plateau", color: "#4B0082"}
                "saffron_city": {value: 10, description: "Saffron City", color: "#F0E68C"}
            }
        }
        validation: {
            minValue: 0
            maxValue: 255
        }
        freezable: true
        uiHints: {
            category: "location"
            displayFormat: "enum_dropdown"
            showInList: true
        }
    }

    // === PARTY INFORMATION ===
    teamCount: {
        name: "teamCount"
        type: "uint8"
        address: "0xD163"
        description: "Number of Pokemon in party"
        validation: {
            minValue: 0
            maxValue: 6
        }
        freezable: true
        uiHints: {
            category: "party"
            displayFormat: "decimal"
            unit: "Pokemon"
            showInList: true
        }
    }

    // Pokemon 1 (Enhanced with validation and advanced features)
    pokemon1Nickname: {
        name: "pokemon1Nickname"
        type: "string"
        address: "0xD2B5"
        length: 11
        description: "First Pokemon's nickname"
        charMap: #PokemonCharMap
        dependsOn: ["teamCount"]
        validation: {
            constraint: "teamCount > 0"  // Only valid if we have Pokemon
        }
        freezable: true
        uiHints: {
            category: "pokemon"
            displayFormat: "text"
        }
    }

    pokemon1Species: {
        name: "pokemon1Species"
        type: "enum"
        address: "0xD16B"
        description: "First Pokemon's species ID"
        advanced: {
            enumValues: {
                "bulbasaur": {value: 153, description: "Bulbasaur #001", color: "#78C850"}
                "ivysaur": {value: 9, description: "Ivysaur #002", color: "#78C850"}
                "venusaur": {value: 154, description: "Venusaur #003", color: "#78C850"}
                "charmander": {value: 176, description: "Charmander #004", color: "#F08030"}
                "charmeleon": {value: 178, description: "Charmeleon #005", color: "#F08030"}
                "charizard": {value: 180, description: "Charizard #006", color: "#F08030"}
                "squirtle": {value: 177, description: "Squirtle #007", color: "#6890F0"}
                "wartortle": {value: 179, description: "Wartortle #008", color: "#6890F0"}
                "blastoise": {value: 28, description: "Blastoise #009", color: "#6890F0"}
                "pikachu": {value: 84, description: "Pikachu #025", color: "#F8D030"}
                "raichu": {value: 85, description: "Raichu #026", color: "#F8D030"}
                "mew": {value: 21, description: "Mew #151", color: "#FB1B69"}
                "mewtwo": {value: 131, description: "Mewtwo #150", color: "#A040A0"}
            }
        }
        dependsOn: ["teamCount"]
        validation: {
            constraint: "teamCount > 0"
            minValue: 1
            maxValue: 255
        }
        freezable: true
        uiHints: {
            category: "pokemon"
            displayFormat: "enum_dropdown"
            showInList: true
        }
    }

    pokemon1Level: {
        name: "pokemon1Level"
        type: "uint8"
        address: "0xD18C"
        description: "First Pokemon's level"
        dependsOn: ["teamCount"]
        validation: {
            constraint: "teamCount > 0"
            minValue: 1
            maxValue: 100
        }
        freezable: true
        uiHints: {
            category: "pokemon"
            displayFormat: "decimal"
            unit: "Lv"
            showInList: true
        }
    }

    pokemon1Hp: {
        name: "pokemon1Hp"
        type: "uint16"
        address: "0xD16C"
        description: "First Pokemon's current HP"
        dependsOn: ["teamCount", "pokemon1MaxHp"]
        validation: {
            constraint: "teamCount > 0 && value <= pokemon1MaxHp"
            minValue: 0
            maxValue: 999
        }
        freezable: true
        uiHints: {
            category: "pokemon"
            displayFormat: "fraction"
            unit: "HP"
            showInList: true
        }
    }

    pokemon1MaxHp: {
        name: "pokemon1MaxHp"
        type: "uint16"
        address: "0xD18D"
        description: "First Pokemon's maximum HP"
        dependsOn: ["teamCount"]
        validation: {
            constraint: "teamCount > 0"
            minValue: 1
            maxValue: 999
        }
        freezable: true
        uiHints: {
            category: "pokemon"
            displayFormat: "decimal"
            unit: "HP"
        }
    }

    // === BATTLE INFORMATION ===
    battleType: {
        name: "battleType"
        type: "enum"
        address: "0xD057"
        description: "Current battle type"
        advanced: {
            enumValues: {
                "none": {value: 0, description: "Not in battle", color: "#90EE90"}
                "wild": {value: 1, description: "Wild Pokemon", color: "#FFD700"}
                "trainer": {value: 2, description: "Trainer battle", color: "#FF6347"}
                "safari": {value: 3, description: "Safari Zone", color: "#98FB98"}
                "old_man": {value: 4, description: "Old Man tutorial", color: "#DDA0DD"}
                "ghost": {value: 5, description: "Ghost battle", color: "#9370DB"}
            }
        }
        validation: {
            minValue: 0
            maxValue: 255
        }
        freezable: true
        uiHints: {
            category: "battle"
            displayFormat: "enum_badge"
            showInList: true
        }
    }

    activePokemonSlot: {
        name: "activePokemonSlot"
        type: "uint8"
        address: "0xCC2F"
        description: "Active Pokemon party slot (0-5)"
        dependsOn: ["battleType", "teamCount"]
        validation: {
            constraint: "battleType > 0 && value < teamCount"
            minValue: 0
            maxValue: 5
        }
        freezable: true
        uiHints: {
            category: "battle"
            displayFormat: "decimal"
            unit: "Slot"
        }
    }

    activePokemonSpecies: {
        name: "activePokemonSpecies"
        type: "enum"
        address: "0xD014"
        description: "Active Pokemon species in battle"
        dependsOn: ["battleType"]
        // Reuse the same enum values as pokemon1Species
        advanced: {
            enumValues: {
                "bulbasaur": {value: 153, description: "Bulbasaur #001", color: "#78C850"}
                "charmander": {value: 176, description: "Charmander #004", color: "#F08030"}
                "squirtle": {value: 177, description: "Squirtle #007", color: "#6890F0"}
                "pikachu": {value: 84, description: "Pikachu #025", color: "#F8D030"}
            }
        }
        validation: {
            constraint: "battleType > 0"
        }
        freezable: true
        uiHints: {
            category: "battle"
            displayFormat: "enum_dropdown"
        }
    }

    activePokemonLevel: {
        name: "activePokemonLevel"
        type: "uint8"
        address: "0xD022"
        description: "Active Pokemon level in battle"
        dependsOn: ["battleType"]
        validation: {
            constraint: "battleType > 0"
            minValue: 1
            maxValue: 100
        }
        freezable: true
        uiHints: {
            category: "battle"
            displayFormat: "decimal"
            unit: "Lv"
        }
    }

    // === COMPUTED PROPERTIES ===
    badgeCount: {
        name: "badgeCount"
        type: "uint8"
        computed: {
            expression: "// Count set bits in badges byte - implemented in Go"
            dependencies: ["badges"]
        }
        description: "Number of badges earned"
        uiHints: {
            category: "progression"
            displayFormat: "decimal"
            unit: "badges"
            showInList: true
        }
    }

    pokemon1HpPercentage: {
        name: "pokemon1HpPercentage"
        type: "percentage"
        computed: {
            expression: "if pokemon1MaxHp > 0 then (pokemon1Hp / pokemon1MaxHp) * 100 else 0"
            dependencies: ["pokemon1Hp", "pokemon1MaxHp"]
        }
        description: "First Pokemon's HP as percentage"
        advanced: {
            maxValue: 100
        }
        uiHints: {
            category: "pokemon"
            displayFormat: "percentage"
            precision: 1
            showInList: true
        }
    }

    partyStrength: {
        name: "partyStrength"
        type: "uint16"
        computed: {
            expression: "if teamCount > 0 then pokemon1Level * teamCount else 0"
            dependencies: ["pokemon1Level", "teamCount"]
        }
        description: "Calculated party strength"
        uiHints: {
            category: "stats"
            displayFormat: "decimal"
        }
    }

    gameCompletionPercentage: {
        name: "gameCompletionPercentage"
        type: "percentage"
        computed: {
            expression: "(badgeCount / 8) * 100"
            dependencies: ["badgeCount"]
        }
        description: "Game completion percentage based on badges"
        advanced: {
            maxValue: 100
        }
        uiHints: {
            category: "progression"
            displayFormat: "percentage"
            precision: 1
            showInList: true
        }
    }
}

// Property groups for better organization
groups: {
    trainer: {
        name: "Trainer Info"
        description: "Player and rival information"
        icon: "👤"
        properties: ["playerName", "rivalName", "money"]
        collapsed: false
        color: "#4CAF50"
    }

    party: {
        name: "Pokemon Party"
        description: "Party Pokemon information"
        icon: "⚡"
        properties: ["teamCount", "pokemon1Nickname", "pokemon1Species", "pokemon1Level", "pokemon1Hp", "pokemon1MaxHp"]
        collapsed: false
        color: "#FFD700"
    }

    battle: {
        name: "Battle Status"
        description: "Current battle information"
        icon: "⚔️"
        properties: ["battleType", "activePokemonSlot", "activePokemonSpecies", "activePokemonLevel"]
        collapsed: false
        color: "#FF6347"
    }

    progression: {
        name: "Game Progress"
        description: "Badges, location, and completion"
        icon: "🏆"
        properties: ["badges", "badgeCount", "currentMap", "gameCompletionPercentage"]
        collapsed: false
        color: "#9C27B0"
    }

    stats: {
        name: "Statistics"
        description: "Computed stats and analytics"
        icon: "📊"
        properties: ["pokemon1HpPercentage", "partyStrength"]
        collapsed: true
        color: "#607D8B"
    }
}

// Computed properties at mapper level
computed: {
    averagePartyLevel: {
        expression: "if teamCount > 0 then pokemon1Level else 0"  // Simplified for now
        dependencies: ["teamCount", "pokemon1Level"]
        type: "uint8"
    }

    battleReadiness: {
        expression: """
            if teamCount == 0 then "No Pokemon"
            else if partyStrength < 50 then "Weak"
            else if partyStrength < 150 then "Medium"
            else "Strong"
        """
        dependencies: ["teamCount", "partyStrength"]
        type: "string"
    }

    progressStatus: {
        expression: """
            if badgeCount == 8 then "Champion!"
            else if badgeCount >= 6 then "Almost there!"
            else if badgeCount >= 4 then "Halfway done"
            else if badgeCount >= 2 then "Getting started"
            else "Just beginning"
        """
        dependencies: ["badgeCount"]
        type: "string"
    }
}

// Character map definition
#PokemonCharMap: {
    "0x50": " "
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
    "0xE1": "P", "0xE2": "K", "0xE3": "M", "0xE4": "N"
    "0xE6": "r", "0xE7": "m"
    "0xF7": "♂", "0xF8": "♀"
    "0xFF": ""
}

// Version history
changelog: [
    {
        version: "2.0.0"
        date: "2024-01-15"
        changes: [
            "Added enhanced property types (enum, flags, percentage)",
            "Implemented property validation and constraints",
            "Added property freezing capabilities",
            "Introduced computed properties",
            "Added property groups for better organization",
            "Enhanced UI hints and display formatting",
            "Added comprehensive Pokemon species definitions",
            "Implemented battle type and location enums"
        ]
    },
    {
        version: "1.0.0"
        date: "2023-12-01"
        changes: [
            "Initial enhanced mapper release",
            "Basic property definitions",
            "Character map support"
        ]
    }
]