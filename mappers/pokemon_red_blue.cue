package pokemon

name: "pokemon_red_blue"
game: "Pokemon Red/Blue"

platform: {
    name: "Game Boy"
    endian: "little"
    memoryBlocks: [
        // Priority 1: WRAM - All Pokemon data is here
        {
            name: "WRAM Bank 0"
            start: "0xC000"
            end: "0xCFFF"
        },
        {
            name: "WRAM Bank 1"
            start: "0xD000"
            end: "0xDFFF"
        },
        // Priority 2: Other areas (may not be accessible)
        // Comment these out initially to test:
        // {
        //     name: "VRAM"
        //     start: "0x8000"
        //     end: "0x9FFF"
        // },
        // {
        //     name: "External RAM"
        //     start: "0xA000"
        //     end: "0xBFFF"
        // }
    ]
}

properties: {
	playerName: {
		name: "playerName"
		altName: "player_name"
		type: "string"
		address: "0xD158"
		length: 11
		description: "Player's name"
		charMap: #PokemonCharMap
	}

	teamCount: {
		name: "teamCount"
		type: "uint8"
		address: "0xD163"
		description: "Number of Pokemon in party"
	}

	pokemon0Nickname: {
		name: "pokemon0Nickname"
		type: "string"
		address: "0xD2B5"
		length: 11
		description: "Pokemon 0 nickname"
	}

	pokemon0Species: {
		name: "pokemon0Species"
		type: "uint8"
		address: "0xD16B"
		description: "Pokemon 0 species"
		transform: {
			lookup: {
				"1": "Rhydon"
				"9": "Ivysaur"
				"21": "Mew"
				"28": "Blastoise"
				"84": "Pikachu"
				"85": "Raichu"
				"131": "Mewtwo"
				"132": "Snorlax"
				"153": "Bulbasaur"
				"154": "Venusaur"
				"176": "Charmander"
				"177": "Squirtle"
				"178": "Charmeleon"
				"179": "Wartortle"
				"180": "Charizard"
			}
		}
	}

	pokemon0Level: {
		name: "pokemon0Level"
		type: "uint8"
		address: "0xD18C"
		description: "Pokemon 0 level"
	}

	pokemon0Hp: {
		name: "pokemon0Hp"
		type: "uint16"
		address: "0xD16C"
		description: "Pokemon 0 current HP"
	}

	pokemon0MaxHp: {
		name: "pokemon0MaxHp"
		type: "uint16"
		address: "0xD18D"
		description: "Pokemon 0 max HP"
	}

	pokemon1Nickname: {
		name: "pokemon1Nickname"
		type: "string"
		address: "0xD2C0"
		length: 11
		description: "Pokemon 1 nickname"
	}

	pokemon1Species: {
		name: "pokemon1Species"
		type: "uint8"
		address: "0xD197"
		description: "Pokemon 1 species"
		transform: {
			lookup: {
				"1": "Rhydon"
				"9": "Ivysaur"
				"21": "Mew"
				"28": "Blastoise"
				"84": "Pikachu"
				"85": "Raichu"
				"131": "Mewtwo"
				"132": "Snorlax"
				"153": "Bulbasaur"
				"154": "Venusaur"
				"176": "Charmander"
				"177": "Squirtle"
				"178": "Charmeleon"
				"179": "Wartortle"
				"180": "Charizard"
			}
		}
	}

	pokemon1Level: {
		name: "pokemon1Level"
		type: "uint8"
		address: "0xD1B8"
		description: "Pokemon 1 level"
	}

	pokemon1Hp: {
		name: "pokemon1Hp"
		type: "uint16"
		address: "0xD198"
		description: "Pokemon 1 current HP"
	}

	pokemon1MaxHp: {
		name: "pokemon1MaxHp"
		type: "uint16"
		address: "0xD1B9"
		description: "Pokemon 1 max HP"
	}

	battleType: {
		name: "battleType"
		type: "uint8"
		address: "0xD057"
		description: "Battle type"
		transform: {
			lookup: {
				"0": "None"
				"1": "Wild"
				"2": "Trainer"
				"255": "Lost Battle"
			}
		}
	}

	activePokemonSlot: {
		name: "activePokemonSlot"
		type: "uint8"
		address: "0xCC2F"
		description: "Active Pokemon slot"
	}

	activePokemonSpecies: {
		name: "activePokemonSpecies"
		type: "uint8"
		address: "0xD014"
		description: "Active Pokemon species"
		transform: {
			lookup: {
				"1": "Rhydon"
				"9": "Ivysaur"
				"21": "Mew"
				"28": "Blastoise"
				"84": "Pikachu"
				"85": "Raichu"
				"131": "Mewtwo"
				"132": "Snorlax"
				"153": "Bulbasaur"
				"154": "Venusaur"
				"176": "Charmander"
				"177": "Squirtle"
				"178": "Charmeleon"
				"179": "Wartortle"
				"180": "Charizard"
			}
		}
	}

	activePokemonLevel: {
		name: "activePokemonLevel"
		type: "uint8"
		address: "0xD022"
		description: "Active Pokemon level"
	}

	money: {
		name: "money"
		type: "bcd"
		address: "0xD347"
		length: 3
		description: "Player's money"
	}

	rivalName: {
		name: "rivalName"
		type: "string"
		address: "0xD34A"
		length: 8
		description: "Rival's name"
	}

	currentMap: {
		name: "currentMap"
		type: "uint8"
		address: "0xD35E"
		description: "Current map ID"
		transform: {
			lookup: {
				"0": "Pallet Town"
				"1": "Viridian City"
				"2": "Pewter City"
				"3": "Cerulean City"
				"4": "Lavender Town"
				"5": "Vermilion City"
				"6": "Celadon City"
				"7": "Fuchsia City"
				"8": "Cinnabar Island"
				"9": "Indigo Plateau"
				"10": "Saffron City"
			}
		}
	}

	badges: {
		name: "badges"
		type: "uint8"
		address: "0xD356"
		description: "Badge bitfield"
	}

	#PokemonCharMap: {
    "0x50": " "
    "0x80": "A"
    "0x81": "B"
    "0x82": "C"
    "0x83": "D"
    "0x84": "E"
    "0x85": "F"
    "0x86": "G"
    "0x87": "H"
    "0x88": "I"
    "0x89": "J"
    "0x8A": "K"
    "0x8B": "L"
    "0x8C": "M"
    "0x8D": "N"
    "0x8E": "O"
    "0x8F": "P"
    "0x90": "Q"
    "0x91": "R"
    "0x92": "S"
    "0x93": "T"
    "0x94": "U"
    "0x95": "V"
    "0x96": "W"
    "0x97": "X"
    "0x98": "Y"
    "0x99": "Z"
    "0x9A": "("
    "0x9B": ")"
    "0x9C": ":"
    "0x9D": ";"
    "0xA0": "a"
    "0xA1": "b"
    "0xA2": "c"
    "0xA3": "d"
    "0xA4": "e"
    "0xA5": "f"
    "0xA6": "g"
    "0xA7": "h"
    "0xA8": "i"
    "0xA9": "j"
    "0xAA": "k"
    "0xAB": "l"
    "0xAC": "m"
    "0xAD": "n"
    "0xAE": "o"
    "0xAF": "p"
    "0xB0": "q"
    "0xB1": "r"
    "0xB2": "s"
    "0xB3": "t"
    "0xB4": "u"
    "0xB5": "v"
    "0xB6": "w"
    "0xB7": "x"
    "0xB8": "y"
    "0xB9": "z"
    "0xE1": "P"  // Pokemon abbreviation
    "0xE2": "K"  // Pokemon abbreviation
    "0xE3": "M"  // Pokemon abbreviation
    "0xE4": "N"  // Pokemon abbreviation
    "0xE6": "r"  // Pokemon abbreviation
    "0xE7": "m"  // Pokemon abbreviation
    "0xF7": "♂"  // Male symbol
    "0xF8": "♀"  // Female symbol
    "0xFF": "?"  // Terminator/Unknown
}
}