package mario

// Import the base schema (this would reference the schema file)
// For now, we'll define the types inline

// Super Mario Bros NES Mapper
name: "Super Mario Bros"
game: "Super Mario Bros."

platform: {
	name: "NES"
	endian: "little"
	memoryBlocks: [{
		name: "RAM"
		start: "0x0000"
		end: "0x07FF"
	}]
}

properties: {
	// Player lives
	marioLives: {
		name: "marioLives"
		type: "uint8"
		address: "0x075A"
		description: "Number of lives remaining"
	}

	// Coin count
	coinCount: {
		name: "coinCount"
		type: "uint8"
		address: "0x075E"
		description: "Number of coins collected"
		transform: {
			lookup: {
				"99": "MAX"
			}
		}
	}

	// Current score (3 bytes, little endian)
	score: {
		name: "score"
		type: "uint32"
		address: "0x07DD"
		length: 3
		description: "Current score"
		transform: {
			multiply: 10  // Score is stored divided by 10
		}
	}

	// Mario's power-up state
	powerupState: {
		name: "powerupState"
		type: "uint8"
		address: "0x0756"
		description: "Mario's current power-up state"
		transform: {
			lookup: {
				"0": "Small Mario"
				"1": "Super Mario"
				"2": "Fire Mario"
			}
		}
	}

	// World and level
	world: {
		name: "world"
		type: "uint8"
		address: "0x075F"
		description: "Current world (1-based)"
		transform: {
			add: 1  // Game stores 0-based, display 1-based
		}
	}

	level: {
		name: "level"
		type: "uint8"
		address: "0x075C"
		description: "Current level (1-based)"
		transform: {
			add: 1  // Game stores 0-based, display 1-based
		}
	}

	// Player position (16-bit X coordinate)
	playerX: {
		name: "playerX"
		type: "uint16"
		address: "0x0086"
		description: "Mario's X position in level"
	}

	playerY: {
		name: "playerY"
		type: "uint8"
		address: "0x00CE"
		description: "Mario's Y position in level"
	}

	// Timer (3 digits, BCD format would need special handling)
	timeHundreds: {
		name: "timeHundreds"
		type: "uint8"
		address: "0x07F8"
		description: "Timer hundreds digit"
	}

	timeTens: {
		name: "timeTens"
		type: "uint8"
		address: "0x07F9"
		description: "Timer tens digit"
	}

	timeOnes: {
		name: "timeOnes"
		type: "uint8"
		address: "0x07FA"
		description: "Timer ones digit"
	}

	// Game state flags
	gameState: {
		name: "gameState"
		type: "uint8"
		address: "0x000E"
		description: "Game state"
		transform: {
			lookup: {
				"0": "Title Screen"
				"1": "Game Over"
				"2": "Loading"
				"3": "Playing"
				"4": "Paused"
			}
		}
	}

	// Player input (bit field)
	controllerInput: {
		name: "controllerInput"
		type: "bitfield"
		address: "0x00F7"
		length: 1
		description: "Controller 1 input state (A, B, Select, Start, Up, Down, Left, Right)"
	}
}

// Property aliases for convenience
lives: properties.marioLives
coins: properties.coinCount
power: properties.powerupState
currentWorld: properties.world
currentLevel: properties.level