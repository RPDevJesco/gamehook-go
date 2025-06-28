package pokemon_red_blue

// ===== POKEMON RED/BLUE ENHANCED MAPPER =====

// Mapper metadata
name: "pokemon_red_blue_enhanced"
game: "Pokemon Red/Blue"
version: "3.0.0"
minGameHookVersion: "0.8.0"
author: "GameHook Team"
description: "Enhanced Pokemon Red/Blue mapper with intelligent data structures"
license: "MIT"

metadata: {
    created: "2025-01-01T00:00:00Z"
    modified: "2025-01-01T00:00:00Z"
    tags: ["pokemon", "gameboy", "rpg", "nintendo"]
    category: "RPG"
    language: "English"
    region: "US"
    revision: "1.0"
}

// ===== PLATFORM CONFIGURATION =====
platform: {
    name: "Game Boy"
    endian: "little"
    description: "Nintendo Game Boy running Pokemon Red/Blue"

    // Platform constants for addressing and calculations
    constants: {
        // Memory layout
        ramBase: 0xC000
        wramBank0Start: 0xC000
        wramBank1Start: 0xD000

        // Pokemon data structure sizes
        pokemonDataSize: 44
        pokemonNicknameSize: 11
        pokemonMoveCount: 4

        // Game limits
        maxPokemonLevel: 100
        maxPartySize: 6
        maxBoxSize: 20
        maxMoney: 999999
        maxItemQuantity: 99
        pokemonSpeciesCount: 151
        badgeCount: 8

        // Type system
        typeCount: 16
        moveCount: 165
        itemCount: 255
    }

    baseAddresses: {
        // Player data
        player: "0xD158"
        playerId: "0xD359"
        money: "0xD347"

        // Party data
        partyCount: "0xD163"
        partySpecies: "0xD164"
        partyData: "0xD16B"
        partyNicknames: "0xD2B5"

        // PC storage
        currentBox: "0xDA96"
        currentBoxNicknames: "0xDE06"
        currentBoxCount: "0xDA80"

        // Battle system
        battleType: "0xD05A"
        battlePlayerData: "0xD009"
        battleEnemyData: "0xCFD8"
        battleEffects: "0xD062"

        // Game state
        badges: "0xD356"
        pokedexSeen: "0xD30A"
        pokedexCaught: "0xD2F7"
        starterPokemon: "0xD717"

        // Items
        bagCount: "0xD31D"
        bagItems: "0xD31E"
    }

    memoryBlocks: [
        {
            name: "WRAM Bank 0"
            start: "0xC000"
            end: "0xCFFF"
            description: "Work RAM Bank 0 - System variables and temporary data"
            cacheable: true
            accessPattern: "random"
        },
        {
            name: "WRAM Bank 1"
            start: "0xD000"
            end: "0xDFFF"
            description: "Work RAM Bank 1 - Game data and save state"
            cacheable: true
            accessPattern: "sequential"
            watchable: true
        }
    ]

    capabilities: {
        maxMemorySize: 65536
        addressBusWidth: 16
        dataBusWidth: 8
        hasMemoryMapping: false
        supportsBanking: false
    }

    performance: {
        readLatency: 1
        writeLatency: 1
        batchSize: 32
    }
}

// ===== GLOBAL CHARACTER MAPS =====
characterMaps: {
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
}

// ===== REFERENCE TYPE DEFINITIONS =====
references: {
    pokemonSpecies: {
        type: "enum"
        advanced: {
            enumValues: {
                "0": {value: 0, description: "MissingNo", color: "#808080"}
                "1": {value: 1, description: "Bulbasaur", color: "#78C850", type1: "Grass", type2: "Poison"}
                "4": {value: 4, description: "Charmander", color: "#F08030", type1: "Fire"}
                "7": {value: 7, description: "Squirtle", color: "#6890F0", type1: "Water"}
                "25": {value: 25, description: "Pikachu", color: "#F8D030", type1: "Electric"}
                "150": {value: 150, description: "Mewtwo", color: "#A040A0", type1: "Psychic"}
                "151": {value: 151, description: "Mew", color: "#FF1493", type1: "Psychic"}
                // More species can be added as needed
            }
            allowUnknownValues: true
            defaultValue: 0
        }
    }

    pokemonTypes: {
        type: "enum"
        advanced: {
            enumValues: {
                "0": {value: 0, description: "Normal", color: "#A8A878"}
                "1": {value: 1, description: "Fighting", color: "#C03028"}
                "2": {value: 2, description: "Flying", color: "#A890F0"}
                "3": {value: 3, description: "Poison", color: "#A040A0"}
                "4": {value: 4, description: "Ground", color: "#E0C068"}
                "5": {value: 5, description: "Rock", color: "#B8A038"}
                "6": {value: 6, description: "Bug", color: "#A8B820"}
                "7": {value: 7, description: "Ghost", color: "#705898"}
                "8": {value: 8, description: "Fire", color: "#F08030"}
                "9": {value: 9, description: "Water", color: "#6890F0"}
                "10": {value: 10, description: "Grass", color: "#78C850"}
                "11": {value: 11, description: "Electric", color: "#F8D030"}
                "12": {value: 12, description: "Psychic", color: "#F85888"}
                "13": {value: 13, description: "Ice", color: "#98D8D8"}
                "14": {value: 14, description: "Dragon", color: "#7038F8"}
            }
        }
    }

    statusConditions: {
        type: "enum"
        advanced: {
            enumValues: {
                "0": {value: 0, description: "None", color: "#00FF00"}
                "2": {value: 2, description: "Sleep", color: "#6F42C1"}
                "4": {value: 4, description: "Poison", color: "#A040A0"}
                "8": {value: 8, description: "Burn", color: "#F08030"}
                "16": {value: 16, description: "Freeze", color: "#98D8D8"}
                "32": {value: 32, description: "Paralysis", color: "#F8D030"}
            }
        }
    }
}

// ===== CORE PROPERTY DEFINITIONS =====

properties: {
    // ===== PLAYER INFORMATION =====
    playerName: {
        name: "playerName"
        type: "string"
        address: "0xD158"
        length: 11
        description: "Player's trainer name"
        charMap: characterMaps.pokemon
        freezable: true
        validation: {
            required: true
            pattern: "^[A-Za-z0-9 ]*$"
        }
        uiHints: {
            priority: 10
            icon: "👤"
            editable: true
        }
    }

    playerId: {
        name: "playerId"
        type: "uint16"
        address: "0xD359"
        description: "Player's trainer ID"
        freezable: true
        validation: {
            minValue: 0
            maxValue: 65535
        }
        uiHints: {
            displayFormat: "hex"
            icon: "🆔"
        }
    }

    money: {
        name: "money"
        type: "uint32"
        address: "0xD347"
        length: 3
        description: "Player's money in BCD format"
        transform: {
            expression: "bcdToDecimal(value)"
            validation: {
                minValue: 0
                maxValue: 999999
            }
        }
        freezable: true
        uiHints: {
            displayFormat: "currency"
            unit: "₽"
            icon: "💰"
            priority: 8
        }
    }

    // ===== PARTY POKEMON SYSTEM =====
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
            priority: 9
            icon: "🎮"
        }
    }

    // First Pokemon detailed properties (most commonly accessed)
    pokemon1Species: {
        name: "pokemon1Species"
        type: "uint8"
        address: "0xD16B"
        description: "First Pokemon species"
        advanced: {
            enumValues: references.pokemonSpecies.advanced.enumValues
        }
        freezable: true
        uiHints: {
            priority: 10
            icon: "⭐"
        }
    }

    pokemon1Level: {
        name: "pokemon1Level"
        type: "uint8"
        address: "0xD18C"
        description: "First Pokemon level"
        validation: {
            minValue: 1
            maxValue: 100
        }
        freezable: true
        uiHints: {
            displayFormat: "decimal"
            unit: "Lv"
            priority: 9
        }
    }

    pokemon1Hp: {
        name: "pokemon1Hp"
        type: "uint16"
        address: "0xD16D"
        description: "First Pokemon current HP"
        freezable: true
        uiHints: {
            priority: 8
            color: "#FF0000"
        }
    }

    pokemon1MaxHp: {
        name: "pokemon1MaxHp"
        type: "uint16"
        address: "0xD18D"
        description: "First Pokemon max HP"
        uiHints: {
            priority: 7
        }
    }

    pokemon1Attack: {
        name: "pokemon1Attack"
        type: "uint16"
        address: "0xD18F"
        description: "First Pokemon attack stat"
    }

    pokemon1Defense: {
        name: "pokemon1Defense"
        type: "uint16"
        address: "0xD191"
        description: "First Pokemon defense stat"
    }

    pokemon1Speed: {
        name: "pokemon1Speed"
        type: "uint16"
        address: "0xD193"
        description: "First Pokemon speed stat"
    }

    pokemon1Special: {
        name: "pokemon1Special"
        type: "uint16"
        address: "0xD195"
        description: "First Pokemon special stat"
    }

    pokemon1ExpPoints: {
        name: "pokemon1ExpPoints"
        type: "uint32"
        address: "0xD179"
        length: 3
        description: "First Pokemon experience points"
        transform: {
            expression: "value & 0xFFFFFF"
        }
        freezable: true
        uiHints: {
            displayFormat: "decimal"
            unit: "EXP"
        }
    }

    pokemon1Type1: {
        name: "pokemon1Type1"
        type: "uint8"
        address: "0xD170"
        description: "First Pokemon primary type"
        advanced: {
            enumValues: references.pokemonTypes.advanced.enumValues
        }
    }

    pokemon1Type2: {
        name: "pokemon1Type2"
        type: "uint8"
        address: "0xD171"
        description: "First Pokemon secondary type"
        advanced: {
            enumValues: references.pokemonTypes.advanced.enumValues
        }
    }

    pokemon1Status: {
        name: "pokemon1Status"
        type: "uint8"
        address: "0xD16F"
        description: "First Pokemon status condition"
        advanced: {
            enumValues: references.statusConditions.advanced.enumValues
        }
        freezable: true
    }

    // Party nicknames (individual access)
    pokemon1Nickname: {
        name: "pokemon1Nickname"
        type: "string"
        address: "0xD2B5"
        length: 11
        description: "First Pokemon nickname"
        charMap: characterMaps.pokemon
        freezable: true
        uiHints: {
            editable: true
            priority: 8
        }
    }

    // ===== BATTLE SYSTEM =====
    battleType: {
        name: "battleType"
        type: "uint8"
        address: "0xD05A"
        description: "Current battle type"
        advanced: {
            enumValues: {
                "0": {value: 0, description: "No Battle"}
                "1": {value: 1, description: "Wild Pokemon"}
                "2": {value: 2, description: "Trainer Battle"}
                "3": {value: 3, description: "Safari Zone"}
            }
        }
        uiHints: {
            icon: "⚔️"
        }
    }

    // ===== GAME PROGRESS =====
    badges: {
        name: "badges"
        type: "flags"
        address: "0xD356"
        length: 1
        description: "Gym badges earned"

        advanced: {
            flagDefinitions: {
                boulder: {bit: 0, description: "Boulder Badge"}
                cascade: {bit: 1, description: "Cascade Badge"}
                thunder: {bit: 2, description: "Thunder Badge"}
                rainbow: {bit: 3, description: "Rainbow Badge"}
                soul: {bit: 4, description: "Soul Badge"}
                marsh: {bit: 5, description: "Marsh Badge"}
                volcano: {bit: 6, description: "Volcano Badge"}
                earth: {bit: 7, description: "Earth Badge"}
            }
        }

        freezable: true

        uiHints: {
            displayMode: "custom"
            icon: "🏆"
            priority: 7
        }
    }

    starterPokemon: {
        name: "starterPokemon"
        type: "uint8"
        address: "0xD717"
        description: "Starter Pokemon received from Oak"
        advanced: {
            enumValues: {
                "153": {value: 153, description: "Bulbasaur"}
                "156": {value: 156, description: "Charmander"}
                "159": {value: 159, description: "Squirtle"}
            }
        }
        uiHints: {
            icon: "🌱"
        }
    }

    // Bag system
    bagCount: {
        name: "bagCount"
        type: "uint8"
        address: "0xD31D"
        description: "Number of different items in bag"
        validation: {
            minValue: 0
            maxValue: 20
        }
        uiHints: {
            icon: "🎒"
        }
    }
}

// ===== COMPUTED PROPERTIES (Simplified) =====
computed: {
    // Pokemon HP percentage for UI
    pokemon1HpPercentage: {
        expression: "pokemon1MaxHp > 0 ? (pokemon1Hp / pokemon1MaxHp) * 100 : 0"
        dependencies: ["pokemon1Hp", "pokemon1MaxHp"]
        type: "percentage"
        cached: true
    }

    // Badge count
    badgeCount: {
        expression: """
        ((badges & 1) > 0 ? 1 : 0) +
        ((badges & 2) > 0 ? 1 : 0) +
        ((badges & 4) > 0 ? 1 : 0) +
        ((badges & 8) > 0 ? 1 : 0) +
        ((badges & 16) > 0 ? 1 : 0) +
        ((badges & 32) > 0 ? 1 : 0) +
        ((badges & 64) > 0 ? 1 : 0) +
        ((badges & 128) > 0 ? 1 : 0)
        """
        dependencies: ["badges"]
        cached: true
    }

    // Battle readiness check
    canBattle: {
        expression: "teamCount > 0 && pokemon1Hp > 0"
        dependencies: ["teamCount", "pokemon1Hp"]
        type: "bool"
    }

    // Is Pokemon at critical health
    pokemon1Critical: {
        expression: "pokemon1MaxHp > 0 && (pokemon1Hp / pokemon1MaxHp) < 0.25"
        dependencies: ["pokemon1Hp", "pokemon1MaxHp"]
        type: "bool"
    }
}

// ===== UI ORGANIZATION =====
groups: {
    player: {
        name: "Player Info"
        icon: "👤"
        properties: ["playerName", "playerId", "money"]
        color: "#2196F3"
        priority: 10
    }

    pokemon: {
        name: "First Pokemon"
        icon: "⭐"
        properties: [
            "pokemon1Species", "pokemon1Nickname", "pokemon1Level",
            "pokemon1Hp", "pokemon1MaxHp", "pokemon1ExpPoints",
            "pokemon1Type1", "pokemon1Type2", "pokemon1Status"
        ]
        color: "#4CAF50"
        priority: 9
    }

    stats: {
        name: "Pokemon Stats"
        icon: "📊"
        properties: [
            "pokemon1Attack", "pokemon1Defense",
            "pokemon1Speed", "pokemon1Special"
        ]
        color: "#FF9800"
        collapsed: true
    }

    party: {
        name: "Party"
        icon: "🎮"
        properties: ["teamCount", "canBattle"]
        color: "#4CAF50"
        priority: 8
    }

    battle: {
        name: "Battle System"
        icon: "⚔️"
        properties: ["battleType"]
        color: "#F44336"

        conditionalDisplay: {
            expression: "battleType > 0"
            dependencies: ["battleType"]
        }
    }

    progress: {
        name: "Game Progress"
        icon: "🏆"
        properties: ["badges", "badgeCount", "starterPokemon"]
        color: "#9C27B0"
        priority: 6
    }

    inventory: {
        name: "Inventory"
        icon: "🎒"
        properties: ["bagCount"]
        color: "#795548"
        collapsed: true
    }

    computed: {
        name: "Computed Values"
        icon: "🧮"
        properties: ["pokemon1HpPercentage", "pokemon1Critical"]
        color: "#607D8B"
        collapsed: true
    }
}

// ===== VALIDATION RULES (Simplified) =====
globalValidation: {
    memoryLayout: {
        checkOverlaps: true
        checkBounds: true
        checkAlignment: true
    }

    crossValidation: [
        {
            name: "hp_bounds_check"
            expression: "pokemon1Hp <= pokemon1MaxHp"
            dependencies: ["pokemon1Hp", "pokemon1MaxHp"]
            message: "Current HP cannot exceed maximum HP"
        },
        {
            name: "level_bounds_check"
            expression: "pokemon1Level >= 1 && pokemon1Level <= 100"
            dependencies: ["pokemon1Level"]
            message: "Pokemon level must be between 1 and 100"
        },
        {
            name: "team_size_check"
            expression: "teamCount >= 0 && teamCount <= 6"
            dependencies: ["teamCount"]
            message: "Team count must be between 0 and 6"
        }
    ]

    performance: {
        maxProperties: 1000
        maxComputedDepth: 5
        warnSlowProperties: true
    }
}

// ===== EVENT SYSTEM (Simplified) =====
events: {
    onLoad: "log('Pokemon Red/Blue Enhanced Mapper loaded successfully')"

    onPropertyChanged: """
    if (property.name == "teamCount") {
        log("Team size changed to: " + property.value)
    }
    """

    custom: {
        pokemon_fainted: {
            trigger: "pokemon1Hp == 0 && pokemon1MaxHp > 0"
            action: "log('Warning: First Pokemon has fainted!')"
            dependencies: ["pokemon1Hp", "pokemon1MaxHp"]
        }

        badge_earned: {
            trigger: "badgeCount > 0"
            action: "log('Badge progress: ' + badgeCount + '/8')"
            dependencies: ["badgeCount"]
        }

        level_up: {
            trigger: "pokemon1Level > 50"
            action: "log('High level Pokemon detected!')"
            dependencies: ["pokemon1Level"]
        }

        critical_health: {
            trigger: "pokemon1Critical == true"
            action: "log('Pokemon at critical health!')"
            dependencies: ["pokemon1Critical"]
        }
    }
}

// ===== DEBUG CONFIGURATION =====
debug: {
    enabled: false
    logLevel: "info"
    logProperties: ["teamCount", "pokemon1Species", "battleType"]
    benchmarkProperties: ["pokemon1Species", "badgeCount"]

    hotReload: true
    typeChecking: true
    memoryDumps: false
}