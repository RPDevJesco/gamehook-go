package pokemon

name: "pokemon_red_blue"
game: "Pokemon Red/Blue"

platform: {
	name: "Game Boy"
	endian: "little"
	memoryBlocks: [{
		name: "WRAM"
		start: "0xC000"
		end: "0xDFFF"
	}]
}

properties: {
	playerName: {
		name: "playerName"
		altName: "player_name"
		type: "string"
		address: "0xD158"
		length: 11
		description: "Player's name"
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
		type: "uint32"
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
}