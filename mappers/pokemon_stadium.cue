// Pokemon Stadium N64 Enhanced Mapper
name: "pokemon_stadium_n64_enhanced"
game: "Pokemon Stadium"
version: "2.0.0"
minGameHookVersion: "0.8.0"
author: "GameHook Team"
description: "Enhanced Pokemon Stadium N64 mapper with 3D battle system support"
license: "MIT"

metadata: {
	created: "2025-01-01T00:00:00Z"
	modified: "2025-01-01T00:00:00Z"
	tags: ["pokemon", "n64", "battle", "nintendo", "3d"]
	category: "Battle Arena"
	language: "English"
	region: "US"
	revision: "1.0"
}

platform: {
	name: "Nintendo 64"
	endian: "big"
	description: "Nintendo 64 running Pokemon Stadium"

	constants: {
		// N64 Memory layout
		rdramBase: 0x80000000
		rdramSize: 0x800000        // 8MB RDRAM
		cartridgeBase: 0x90000000

		// Pokemon Stadium specific
		maxRentalPokemon: 150
		maxCustomPokemon: 6
		battleTeamSize: 6
		cupCount: 8
		battleModeCount: 4

		// Pokemon data sizes
		stadiumPokemonSize: 48
		moveDataSize: 12
		trainerDataSize: 32

		// Battle system
		maxBattleParticipants: 12
		moveAnimationCount: 200
		damageCalculationPrecision: 256

		// Stadium cups and modes
		pokeCupStages: 4
		primeCupStages: 4
		ultraCupStages: 4
		masterBallStages: 4
	}

	baseAddresses: {
		// Game state
		gameMode: "0x80100000"
		currentMenu: "0x80100004"
		battleState: "0x80100010"

		// Player data
		playerProfile: "0x80200000"
		playerName: "0x80200010"
		playerProgress: "0x80200100"

		// Battle system
		battleMode: "0x80300000"
		currentBattle: "0x80300010"
		battleParticipants: "0x80300100"
		battleField: "0x80300200"

		// Pokemon data
		rentalPokemon: "0x80400000"
		customTeam: "0x80450000"
		enemyTeam: "0x80460000"

		// Stadium progress
		pokeCup: "0x80500000"
		primeCup: "0x80500100"
		ultraCup: "0x80500200"
		masterBall: "0x80500300"

		// Mini-games
		miniGameScores: "0x80600000"
		miniGameUnlocks: "0x80600100"

		// Graphics and effects
		cameraPosition: "0x80700000"
		lightingState: "0x80700100"
		particleEffects: "0x80700200"
	}

	memoryBlocks: [
		{
			name: "RDRAM Main"
			start: "0x80000000"
			end: "0x807FFFFF"
			description: "Main RDRAM - Game data and runtime state"
			cacheable: true
			accessPattern: "random"
			watchable: true
		},
		{
			name: "Game State"
			start: "0x80100000"
			end: "0x801FFFFF"
			description: "Core game state and menu system"
			cacheable: true
			accessPattern: "sequential"
		},
		{
			name: "Battle System"
			start: "0x80300000"
			end: "0x803FFFFF"
			description: "3D battle engine and Pokemon data"
			cacheable: true
			accessPattern: "random"
			protected: false
		},
		{
			name: "Stadium Data"
			start: "0x80500000"
			end: "0x805FFFFF"
			description: "Stadium progress and cup data"
			cacheable: true
			accessPattern: "sequential"
		}
	]

	capabilities: {
		maxMemorySize: 8388608
		addressBusWidth: 32
		dataBusWidth: 64
		hasMemoryMapping: true
		supportsBanking: false
	}

	performance: {
		readLatency: 2
		writeLatency: 3
		batchSize: 64
	}
}

// Reference type definitions
references: {
	pokemonSpecies: {
		type: "enum"
		advanced: {
			enumValues: {
				"1": {value: 1, description: "Bulbasaur", color: "#78C850", type1: "Grass", type2: "Poison"}
				"4": {value: 4, description: "Charmander", color: "#F08030", type1: "Fire"}
				"7": {value: 7, description: "Squirtle", color: "#6890F0", type1: "Water"}
				"25": {value: 25, description: "Pikachu", color: "#F8D030", type1: "Electric"}
				"150": {value: 150, description: "Mewtwo", color: "#A040A0", type1: "Psychic"}
				"151": {value: 151, description: "Mew", color: "#FF1493", type1: "Psychic"}
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
				"15": {value: 15, description: "Dark", color: "#705848"}
				"16": {value: 16, description: "Steel", color: "#B8B8D0"}
			}
		}
	}

	battleModes: {
		type: "enum"
		advanced: {
			enumValues: {
				"0": {value: 0, description: "Stadium", color: "#4CAF50"}
				"1": {value: 1, description: "Gym Leader Castle", color: "#FF9800"}
				"2": {value: 2, description: "Elite Four", color: "#9C27B0"}
				"3": {value: 3, description: "Champion", color: "#FFD700"}
				"4": {value: 4, description: "Free Battle", color: "#2196F3"}
				"5": {value: 5, description: "Mini Games", color: "#E91E63"}
			}
		}
	}

	stadiumCups: {
		type: "enum"
		advanced: {
			enumValues: {
				"0": {value: 0, description: "Poke Cup", color: "#4CAF50"}
				"1": {value: 1, description: "Great Ball", color: "#2196F3"}
				"2": {value: 2, description: "Ultra Ball", color: "#FF9800"}
				"3": {value: 3, description: "Master Ball", color: "#9C27B0"}
				"4": {value: 4, description: "Prime Cup", color: "#F44336"}
				"5": {value: 5, description: "Little Cup", color: "#FFEB3B"}
				"6": {value: 6, description: "Petit Cup", color: "#8BC34A"}
				"7": {value: 7, description: "Challenge Cup", color: "#607D8B"}
			}
		}
	}

	battleStates: {
		type: "enum"
		advanced: {
			enumValues: {
				"0": {value: 0, description: "No Battle", color: "#9E9E9E"}
				"1": {value: 1, description: "Initializing", color: "#FFEB3B"}
				"2": {value: 2, description: "Pokemon Selection", color: "#2196F3"}
				"3": {value: 3, description: "Battle Active", color: "#F44336"}
				"4": {value: 4, description: "Move Selection", color: "#FF9800"}
				"5": {value: 5, description: "Animation", color: "#9C27B0"}
				"6": {value: 6, description: "Battle End", color: "#4CAF50"}
			}
		}
	}
}

// Property definitions
properties: {
	// Game State
	gameMode: {
		name: "gameMode"
		type: "uint32"
		address: "0x80100000"
		description: "Current game mode"
		advanced: {
			enumValues: references.battleModes.advanced.enumValues
		}
		uiHints: {
			priority: 10
			icon: "🎮"
			displayFormat: "custom"
		}
	}

	currentMenu: {
		name: "currentMenu"
		type: "uint32"
		address: "0x80100004"
		description: "Current menu/screen ID"
		uiHints: {
			displayFormat: "hex"
			icon: "📋"
		}
	}

	battleState: {
		name: "battleState"
		type: "uint32"
		address: "0x80100010"
		description: "Current battle state"
		advanced: {
			enumValues: references.battleStates.advanced.enumValues
		}
		uiHints: {
			priority: 9
			icon: "⚔️"
			color: "#F44336"
		}
	}

	// Player Profile
	playerName: {
		name: "playerName"
		type: "string"
		address: "0x80200010"
		length: 16
		description: "Player's name"
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

	// Battle System
	battleMode: {
		name: "battleMode"
		type: "uint32"
		address: "0x80300000"
		description: "Current battle mode"
		advanced: {
			enumValues: references.battleModes.advanced.enumValues
		}
		uiHints: {
			priority: 8
			icon: "🏟️"
		}
	}

	currentCup: {
		name: "currentCup"
		type: "uint32"
		address: "0x80300004"
		description: "Currently selected cup"
		advanced: {
			enumValues: references.stadiumCups.advanced.enumValues
		}
		uiHints: {
			priority: 8
			icon: "🏆"
		}
	}

	battleTurn: {
		name: "battleTurn"
		type: "uint32"
		address: "0x80300008"
		description: "Current battle turn number"
		validation: {
			minValue: 0
			maxValue: 999
		}
		uiHints: {
			displayFormat: "decimal"
			unit: "Turn"
			priority: 7
		}
	}

	// Player Team (First 3 Pokemon)
	playerPokemon1Species: {
		name: "playerPokemon1Species"
		type: "uint16"
		address: "0x80450000"
		description: "Player's first Pokemon species"
		advanced: {
			enumValues: references.pokemonSpecies.advanced.enumValues
		}
		freezable: true
		uiHints: {
			priority: 9
			icon: "⭐"
		}
	}

	playerPokemon1Level: {
		name: "playerPokemon1Level"
		type: "uint8"
		address: "0x80450002"
		description: "Player's first Pokemon level"
		validation: {
			minValue: 1
			maxValue: 100
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "Lv"
			priority: 8
		}
	}

	playerPokemon1HP: {
		name: "playerPokemon1HP"
		type: "uint16"
		address: "0x80450004"
		description: "Player's first Pokemon current HP"
		freezable: true
		uiHints: {
			priority: 8
			color: "#FF0000"
		}
	}

	playerPokemon1MaxHP: {
		name: "playerPokemon1MaxHP"
		type: "uint16"
		address: "0x80450006"
		description: "Player's first Pokemon max HP"
		uiHints: {
			priority: 7
		}
	}

	playerPokemon2Species: {
		name: "playerPokemon2Species"
		type: "uint16"
		address: "0x80450030"
		description: "Player's second Pokemon species"
		advanced: {
			enumValues: references.pokemonSpecies.advanced.enumValues
		}
		freezable: true
	}

	playerPokemon2Level: {
		name: "playerPokemon2Level"
		type: "uint8"
		address: "0x80450032"
		description: "Player's second Pokemon level"
		validation: {
			minValue: 1
			maxValue: 100
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "Lv"
		}
	}

	playerPokemon3Species: {
		name: "playerPokemon3Species"
		type: "uint16"
		address: "0x80450060"
		description: "Player's third Pokemon species"
		advanced: {
			enumValues: references.pokemonSpecies.advanced.enumValues
		}
		freezable: true
	}

	playerPokemon3Level: {
		name: "playerPokemon3Level"
		type: "uint8"
		address: "0x80450062"
		description: "Player's third Pokemon level"
		validation: {
			minValue: 1
			maxValue: 100
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "Lv"
		}
	}

	// Enemy Team
	enemyPokemon1Species: {
		name: "enemyPokemon1Species"
		type: "uint16"
		address: "0x80460000"
		description: "Enemy's first Pokemon species"
		advanced: {
			enumValues: references.pokemonSpecies.advanced.enumValues
		}
		uiHints: {
			icon: "👹"
		}
	}

	enemyPokemon1Level: {
		name: "enemyPokemon1Level"
		type: "uint8"
		address: "0x80460002"
		description: "Enemy's first Pokemon level"
		uiHints: {
			displayFormat: "decimal"
			unit: "Lv"
		}
	}

	enemyPokemon1HP: {
		name: "enemyPokemon1HP"
		type: "uint16"
		address: "0x80460004"
		description: "Enemy's first Pokemon current HP"
		uiHints: {
			color: "#FF6600"
		}
	}

	// Stadium Progress
	pokeCupProgress: {
		name: "pokeCupProgress"
		type: "uint32"
		address: "0x80500000"
		description: "Poke Cup completion progress"
		validation: {
			minValue: 0
			maxValue: 4
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "/4"
			icon: "🏆"
			priority: 6
		}
	}

	primeCupProgress: {
		name: "primeCupProgress"
		type: "uint32"
		address: "0x80500100"
		description: "Prime Cup completion progress"
		validation: {
			minValue: 0
			maxValue: 4
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "/4"
			icon: "👑"
		}
	}

	ultraCupProgress: {
		name: "ultraCupProgress"
		type: "uint32"
		address: "0x80500200"
		description: "Ultra Cup completion progress"
		validation: {
			minValue: 0
			maxValue: 4
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "/4"
			icon: "⚡"
		}
	}

	masterBallProgress: {
		name: "masterBallProgress"
		type: "uint32"
		address: "0x80500300"
		description: "Master Ball Cup completion progress"
		validation: {
			minValue: 0
			maxValue: 4
		}
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			unit: "/4"
			icon: "🔮"
		}
	}

	// Mini-Games
	miniGameHighScore: {
		name: "miniGameHighScore"
		type: "uint32"
		address: "0x80600000"
		description: "Highest mini-game score"
		freezable: true
		uiHints: {
			displayFormat: "decimal"
			icon: "🎯"
			priority: 5
		}
	}

	// 3D Graphics State
	cameraX: {
		name: "cameraX"
		type: "float32"
		address: "0x80700000"
		description: "3D camera X position"
		uiHints: {
			displayFormat: "decimal"
			precision: 2
			unit: "units"
			chartable: true
			chartType: "line"
		}
	}

	cameraY: {
		name: "cameraY"
		type: "float32"
		address: "0x80700004"
		description: "3D camera Y position"
		uiHints: {
			displayFormat: "decimal"
			precision: 2
			unit: "units"
			chartable: true
			chartType: "line"
		}
	}

	cameraZ: {
		name: "cameraZ"
		type: "float32"
		address: "0x80700008"
		description: "3D camera Z position"
		uiHints: {
			displayFormat: "decimal"
			precision: 2
			unit: "units"
			chartable: true
			chartType: "line"
		}
	}
}

// Computed properties
computed: {
	// Player team status
	playerPokemon1HPPercentage: {
		expression: "playerPokemon1MaxHP > 0 ? (playerPokemon1HP / playerPokemon1MaxHP) * 100 : 0"
		dependencies: ["playerPokemon1HP", "playerPokemon1MaxHP"]
		type: "percentage"
		cached: true
	}

	playerTeamSize: {
		expression: """
		(playerPokemon1Species > 0 ? 1 : 0) +
		(playerPokemon2Species > 0 ? 1 : 0) +
		(playerPokemon3Species > 0 ? 1 : 0)
		"""
		dependencies: ["playerPokemon1Species", "playerPokemon2Species", "playerPokemon3Species"]
		cached: true
	}

	averagePlayerLevel: {
		expression: """
		playerTeamSize > 0 ?
		((playerPokemon1Species > 0 ? playerPokemon1Level : 0) +
		 (playerPokemon2Species > 0 ? playerPokemon2Level : 0) +
		 (playerPokemon3Species > 0 ? playerPokemon3Level : 0)) / playerTeamSize : 0
		"""
		dependencies: ["playerPokemon1Level", "playerPokemon2Level", "playerPokemon3Level", "playerTeamSize"]
		cached: true
	}

	// Stadium progress
	totalCupProgress: {
		expression: "pokeCupProgress + primeCupProgress + ultraCupProgress + masterBallProgress"
		dependencies: ["pokeCupProgress", "primeCupProgress", "ultraCupProgress", "masterBallProgress"]
		cached: true
	}

	stadiumCompletionPercentage: {
		expression: "totalCupProgress / 16 * 100"
		dependencies: ["totalCupProgress"]
		type: "percentage"
		cached: true
	}

	// Battle status
	inBattle: {
		expression: "battleState >= 2 && battleState <= 5"
		dependencies: ["battleState"]
		type: "bool"
	}

	// 3D camera position
	cameraDistance: {
		expression: "sqrt((cameraX * cameraX) + (cameraY * cameraY) + (cameraZ * cameraZ))"
		dependencies: ["cameraX", "cameraY", "cameraZ"]
		type: "float32"
		cached: true
	}
}

// UI organization
groups: {
	game: {
		name: "Game State"
		icon: "🎮"
		properties: ["gameMode", "currentMenu", "battleState"]
		color: "#2196F3"
		priority: 10
	}

	player: {
		name: "Player Profile"
		icon: "👤"
		properties: ["playerName"]
		color: "#4CAF50"
		priority: 9
	}

	battle: {
		name: "Battle System"
		icon: "⚔️"
		properties: ["battleMode", "currentCup", "battleTurn", "inBattle"]
		color: "#F44336"
		priority: 8

		conditionalDisplay: {
			expression: "inBattle == true"
			dependencies: ["inBattle"]
		}
	}

	playerTeam: {
		name: "Player Team"
		icon: "⭐"
		properties: [
			"playerPokemon1Species", "playerPokemon1Level", "playerPokemon1HP", "playerPokemon1MaxHP",
			"playerPokemon2Species", "playerPokemon2Level",
			"playerPokemon3Species", "playerPokemon3Level"
		]
		color: "#4CAF50"
		priority: 7

		subgroups: {
			team1: {
				name: "Pokemon #1"
				properties: ["playerPokemon1Species", "playerPokemon1Level", "playerPokemon1HP", "playerPokemon1MaxHP"]
				collapsed: false
			}
			team23: {
				name: "Pokemon #2-3"
				properties: ["playerPokemon2Species", "playerPokemon2Level", "playerPokemon3Species", "playerPokemon3Level"]
				collapsed: true
			}
		}
	}

	enemyTeam: {
		name: "Enemy Team"
		icon: "👹"
		properties: ["enemyPokemon1Species", "enemyPokemon1Level", "enemyPokemon1HP"]
		color: "#FF9800"
		collapsed: true

		conditionalDisplay: {
			expression: "inBattle == true"
			dependencies: ["inBattle"]
		}
	}

	progress: {
		name: "Stadium Progress"
		icon: "🏆"
		properties: ["pokeCupProgress", "primeCupProgress", "ultraCupProgress", "masterBallProgress"]
		color: "#9C27B0"
		priority: 6
	}

	minigames: {
		name: "Mini-Games"
		icon: "🎯"
		properties: ["miniGameHighScore"]
		color: "#E91E63"
		collapsed: true
	}

	graphics: {
		name: "3D Graphics"
		icon: "📹"
		properties: ["cameraX", "cameraY", "cameraZ", "cameraDistance"]
		color: "#607D8B"
		collapsed: true
	}

	analysis: {
		name: "Team Analysis"
		icon: "📊"
		properties: [
			"playerPokemon1HPPercentage", "playerTeamSize", "averagePlayerLevel",
			"totalCupProgress", "stadiumCompletionPercentage"
		]
		color: "#795548"
		priority: 5
	}
}

// Global validation rules
globalValidation: {
	memoryLayout: {
		checkOverlaps: true
		checkBounds: true
		checkAlignment: true
	}

	crossValidation: [
		{
			name: "hp_bounds_check"
			expression: "playerPokemon1HP <= playerPokemon1MaxHP"
			dependencies: ["playerPokemon1HP", "playerPokemon1MaxHP"]
			message: "Current HP cannot exceed maximum HP"
		},
		{
			name: "level_bounds_check"
			expression: "playerPokemon1Level >= 1 && playerPokemon1Level <= 100"
			dependencies: ["playerPokemon1Level"]
			message: "Pokemon level must be between 1 and 100"
		},
		{
			name: "cup_progress_check"
			expression: "pokeCupProgress <= 4 && primeCupProgress <= 4 && ultraCupProgress <= 4 && masterBallProgress <= 4"
			dependencies: ["pokeCupProgress", "primeCupProgress", "ultraCupProgress", "masterBallProgress"]
			message: "Cup progress cannot exceed 4 stages"
		}
	]

	performance: {
		maxProperties: 1000
		maxComputedDepth: 5
		warnSlowProperties: true
	}
}

// Event system
events: {
	onLoad: "log('Pokemon Stadium N64 Enhanced Mapper loaded successfully')"

	onPropertyChanged: """
	if (property.name == "battleState") {
		log("Battle state changed to: " + property.value)
	}
	"""

	custom: {
		battle_started: {
			trigger: "battleState == 3"
			action: "log('Battle has started!')"
			dependencies: ["battleState"]
		}

		pokemon_fainted: {
			trigger: "playerPokemon1HP == 0 && playerPokemon1MaxHP > 0"
			action: "log('Warning: Player Pokemon has fainted!')"
			dependencies: ["playerPokemon1HP", "playerPokemon1MaxHP"]
		}

		cup_completed: {
			trigger: "totalCupProgress > 0"
			action: "log('Stadium cup progress: ' + totalCupProgress + '/16')"
			dependencies: ["totalCupProgress"]
		}

		high_level_team: {
			trigger: "averagePlayerLevel > 50"
			action: "log('High level team detected! Average level: ' + averagePlayerLevel)"
			dependencies: ["averagePlayerLevel"]
		}

		camera_moved: {
			trigger: "cameraDistance > 100"
			action: "log('Camera moved far from center: ' + cameraDistance + ' units')"
			dependencies: ["cameraDistance"]
		}
	}
}

// Debug configuration
debug: {
	enabled: false
	logLevel: "info"
	logProperties: ["battleState", "playerPokemon1Species", "currentCup"]
	benchmarkProperties: ["cameraX", "cameraY", "cameraZ"]

	hotReload: true
	typeChecking: true
	memoryDumps: false
}