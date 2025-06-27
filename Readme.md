# Enhanced GameHook Go

A next-generation retro game memory manipulation tool with advanced property management, real-time monitoring, and validation capabilities.

## 🚀 Enhanced Features

### ⭐ Core Enhancements
- **Property Freezing/Unfreezing** - Lock property values to prevent changes
- **Advanced Property Types** - Support for enums, flags, coordinates, colors, and more
- **Real-time Monitoring** - 60fps property change detection and streaming
- **Property Validation** - Enforce constraints and rules on property values
- **Batch Operations** - Update multiple properties atomically
- **State Tracking** - Monitor read/write counts and change history
- **Enhanced WebSocket API** - Feature-rich real-time communication

### 🎯 Advanced Property Types
- `enum` - Named enumeration values with colors and descriptions
- `flags` - Bitfield properties with individual flag definitions
- `coordinate` - 2D/3D position data with coordinate system support
- `color` - Color values with various format support (RGB565, ARGB8888, etc.)
- `percentage` - Percentage values with custom max values
- `time` - Time-based values with frame/second conversion
- `version` - Version numbers with different encoding formats
- `checksum` - Checksum values with validation
- `pointer` - Memory pointer dereferencing
- `array` - Dynamic arrays with type-safe elements
- `struct` - Complex structured data

### 🛡️ Validation & Constraints
- Min/max value constraints
- Pattern matching (regex)
- Allowed values lists
- Custom CUE expressions
- Dependency validation
- Real-time validation feedback

### 📊 Property Management
- **Property Groups** - Organize properties by category with icons and colors
- **Computed Properties** - Derived values using CUE expressions
- **Property Dependencies** - Track relationships between properties
- **State Statistics** - Access counts, change frequency, and history
- **UI Hints** - Display formatting, units, and categorization

## 🎮 Quick Start

### Prerequisites

- Go 1.21+
- RetroArch with network commands enabled

### Installation

```bash
# Clone the enhanced repository
git clone <enhanced-repo>
cd gamehook-enhanced
go mod tidy

# Build the enhanced version
go build -o gamehook-enhanced cmd/gamehook/main.go

# Run with enhanced features
./gamehook-enhanced --port 8080 --update-interval 16ms
```

### RetroArch Setup

1. Open RetroArch
2. Go to Settings → Network → Network Commands: **ON**
3. Set Network Command Port: **55355**
4. Load a compatible game

### Load Enhanced Mapper

```bash
# Load the enhanced Pokemon Red/Blue mapper
curl -X POST http://localhost:8080/api/mappers/pokemon_red_blue_enhanced/load
```

### Access Enhanced UI

Open http://localhost:8080/ui/enhanced-admin/ for the full-featured admin panel.

## 📁 Enhanced Project Structure

```
gamehook-enhanced/
├── cmd/gamehook/main.go              # Enhanced application entry point
├── internal/
│   ├── drivers/
│   │   ├── retroarch.go              # Adaptive RetroArch driver
│   │   └── drivers.go                # Driver interface
│   ├── memory/
│   │   ├── manager.go                # Enhanced memory manager with freezing
│   │   └── advanced_properties.go   # Advanced property processors
│   ├── mappers/
│   │   ├── loader.go                 # Enhanced CUE parser & loader
│   │   ├── schema.cue                # Enhanced CUE schema
│   │   └── validation.go             # Property validation engine
│   ├── server/
│   │   ├── server.go                 # Enhanced HTTP server
│   │   ├── websocket.go              # Advanced WebSocket management
│   │   └── middleware.go             # Security and validation middleware
│   └── config/
│       └── config.go                 # Enhanced configuration management
├── mappers/                          # Enhanced mapper definitions
│   ├── pokemon_red_blue_enhanced.cue # Enhanced Pokemon mapper
│   ├── nes/
│   │   └── super_mario_bros.cue      # Enhanced Mario mapper
│   └── gameboy/
│       └── tetris.cue                # Enhanced Tetris mapper
├── uis/                              # Enhanced user interfaces
│   ├── enhanced-admin/               # Full-featured admin panel
│   ├── pokemon-overlay/              # Enhanced Pokemon overlay
│   └── property-monitor/             # Real-time property monitor
├── config/
│   ├── gamehook.yml                  # Default configuration
│   └── production.yml                # Production configuration
├── docker/
│   ├── Dockerfile                    # Container build
│   └── docker-compose.yml            # Multi-service setup
└── docs/                             # Enhanced documentation
    ├── MAPPER_GUIDE.md              # Mapper creation guide
    ├── API_REFERENCE.md             # Complete API documentation
    └── CONFIGURATION.md             # Configuration reference
```

## 🔧 Enhanced Configuration

### YAML Configuration

```yaml
# config/gamehook.yml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

retroarch:
  host: "127.0.0.1"
  port: 55355
  request_timeout: "64ms"
  max_retries: 3

performance:
  update_interval: "16ms"    # 60 FPS
  max_clients: 100
  memory_buffer_size: 1048576
  websocket_buffer: 256

property_monitoring:
  update_interval: "16ms"
  enable_statistics: true
  history_size: 1000
  change_threshold: 0.01

batch_operations:
  max_batch_size: 50
  timeout: "5s"
  enable_atomic: true

validation:
  enable_strict: true
  log_validation: true
  fail_on_error: false

features:
  metrics: true
  profiling: false
  auto_mapper_reload: true
  cache_properties: true
  memory_compression: false

security:
  enable_cors: true
  allowed_origins: ["*"]
  rate_limit:
    enabled: false
    requests_per_second: 100
```

### Command Line Options

```bash
./gamehook-enhanced [options]

Enhanced Options:
  --enable-freezing          Enable property freezing (default: true)
  --enable-validation        Enable property validation (default: true)
  --enable-statistics        Enable property statistics (default: true)
  --max-batch-size int       Maximum batch operation size (default: 50)
  --update-interval duration Property update interval (default: 16ms)
  --request-timeout duration RetroArch request timeout (default: 64ms)

Standard Options:
  --port int                 HTTP server port (default: 8080)
  --retroarch-host string    RetroArch host (default: "127.0.0.1")
  --retroarch-port int       RetroArch UDP port (default: 55355)
  --mappers-dir string       Directory containing CUE files (default: "./mappers")
  --uis-dir string           Directory containing UI folders (default: "./uis")
  --config string            Config file path
```

## 📡 Enhanced API Reference

### Property Management

#### Get All Properties with Enhanced Data
```http
GET /api/properties
```

**Response:**
```json
{
  "properties": [
    {
      "name": "playerHealth",
      "value": 100,
      "type": "uint8",
      "address": "0x0040",
      "description": "Player health points",
      "frozen": false,
      "read_only": false,
      "validation": {
        "min_value": 0,
        "max_value": 100
      },
      "last_changed": "2024-01-15T10:30:45Z",
      "read_count": 1250,
      "write_count": 15
    }
  ],
  "total": 25,
  "frozen_count": 3
}
```

#### Freeze/Unfreeze Property
```http
POST /api/properties/{name}/freeze
Content-Type: application/json

{
  "freeze": true
}
```

#### Set Property Value with Validation
```http
PUT /api/properties/{name}/value
Content-Type: application/json

{
  "value": 50
}
```

#### Set Property Raw Bytes
```http
PUT /api/properties/{name}/bytes
Content-Type: application/json

{
  "bytes": [0x32, 0x00]
}
```

#### Batch Property Updates
```http
PUT /api/properties/batch
Content-Type: application/json

{
  "atomic": true,
  "properties": [
    {
      "name": "playerHealth",
      "value": 100
    },
    {
      "name": "playerMana",
      "freeze": true
    }
  ]
}
```

### Enhanced Mapper Information

#### Get Mapper Metadata
```http
GET /api/mapper/meta
```

**Response:**
```json
{
  "name": "pokemon_red_blue_enhanced",
  "game": "Pokemon Red/Blue (Enhanced)",
  "version": "2.0.0",
  "platform": {
    "name": "Game Boy",
    "endian": "little"
  },
  "property_count": 25,
  "group_count": 5,
  "computed_count": 8,
  "frozen_count": 2,
  "memory_blocks": [...]
}
```

#### Get Property Glossary
```http
GET /api/mapper/glossary
```

**Response:**
```json
{
  "properties": {
    "playerHealth": {
      "name": "playerHealth",
      "type": "uint8",
      "address": "0x0040",
      "description": "Player health points",
      "group": "player_stats",
      "freezable": true,
      "validation": {...},
      "transform": {...}
    }
  },
  "groups": {
    "player_stats": {
      "name": "Player Statistics",
      "icon": "👤",
      "properties": ["playerHealth", "playerMana"]
    }
  }
}
```

### Property State Tracking

#### Get Property State
```http
GET /api/properties/{name}/state
```

**Response:**
```json
{
  "name": "playerHealth",
  "value": 75,
  "bytes": [0x4B],
  "address": 64,
  "frozen": false,
  "last_changed": "2024-01-15T10:30:45Z",
  "last_read": "2024-01-15T10:30:50Z",
  "read_count": 1250,
  "write_count": 15
}
```

#### Get All Property States
```http
GET /api/properties/states
```

### Enhanced WebSocket API

Connect to `ws://localhost:8080/api/stream` for real-time updates.

#### Message Types

**Property Change**
```json
{
  "type": "property_changed",
  "property": "playerHealth",
  "value": 75,
  "old_value": 100,
  "timestamp": "2024-01-15T10:30:45Z",
  "source": "game_update"
}
```

**Freeze State Change**
```json
{
  "type": "property_freeze_changed",
  "property": "playerHealth",
  "frozen": true,
  "timestamp": "2024-01-15T10:30:45Z"
}
```

**Batch Update Complete**
```json
{
  "type": "batch_update_completed",
  "results": [...],
  "success_count": 5,
  "total": 5,
  "timestamp": "2024-01-15T10:30:45Z"
}
```

**Validation Error**
```json
{
  "type": "validation_error",
  "property": "playerHealth",
  "error": "Value 150 exceeds maximum 100",
  "value": 150,
  "timestamp": "2024-01-15T10:30:45Z"
}
```

## 🗺️ Creating Enhanced Mappers

### Enhanced CUE Schema

```cue
package my_enhanced_game

name: "my_enhanced_game"
game: "My Enhanced Game"
version: "1.0.0"
author: "Your Name"

// Enhanced platform with constants
platform: {
    name: "NES"
    endian: "little"
    constants: {
        ramBase: 0x0000
        maxLevel: 99
    }
    memoryBlocks: [...]
}

// Property groups for organization
groups: {
    player: {
        name: "Player Stats"
        icon: "👤"
        properties: ["playerHealth", "playerLevel"]
        color: "#4CAF50"
    }
}

properties: {
    // Enhanced enum property
    powerupState: {
        name: "powerupState"
        type: "enum"
        address: "0x0756"
        description: "Current powerup state"
        advanced: {
            enumValues: {
                "normal": {value: 0, description: "Normal state", color: "#FFF"}
                "super": {value: 1, description: "Super state", color: "#4CAF50"}
                "fire": {value: 2, description: "Fire state", color: "#FF5722"}
            }
        }
        validation: {
            allowedValues: [0, 1, 2]
        }
        freezable: true
        uiHints: {
            displayFormat: "enum_dropdown"
            category: "powerups"
        }
    }

    // Enhanced flags property
    gameFlags: {
        name: "gameFlags"
        type: "flags"
        address: "0x0700"
        description: "Game state flags"
        advanced: {
            flagDefinitions: {
                "intro_seen": {bit: 0, description: "Intro watched"}
                "boss_defeated": {bit: 1, description: "Boss defeated"}
            }
        }
        freezable: true
        uiHints: {
            displayFormat: "flag_list"
        }
    }

    // Computed property
    totalScore: {
        name: "totalScore"
        type: "uint32"
        computed: {
            expression: "coins * 100 + lives * 1000"
            dependencies: ["coins", "lives"]
        }
        description: "Calculated total score"
        uiHints: {
            displayFormat: "decimal"
            category: "statistics"
        }
    }

    // Enhanced coordinate property
    playerPosition: {
        name: "playerPosition"
        type: "coordinate"
        address: "0x0086"
        length: 4
        description: "Player X,Y position"
        advanced: {
            coordinateSystem: "screen"
            dimensions: 2
        }
        validation: {
            constraint: "x >= 0 && x <= 255 && y >= 0 && y <= 240"
        }
        uiHints: {
            displayFormat: "coordinate"
            unit: "pixels"
        }
    }
}

// Computed properties at mapper level
computed: {
    completionPercentage: {
        expression: "(flags_collected / total_flags) * 100"
        dependencies: ["flags_collected", "total_flags"]
        type: "percentage"
    }
}
```

### Property Type Examples

#### Enum Properties
```cue
weaponType: {
    type: "enum"
    address: "0x0050"
    advanced: {
        enumValues: {
            "sword": {value: 0, description: "Basic Sword", color: "#C0C0C0"}
            "magic_sword": {value: 1, description: "Magic Sword", color: "#4169E1"}
            "flame_sword": {value: 2, description: "Flame Sword", color: "#FF4500"}
        }
    }
}
```

#### Flag Properties
```cue
abilities: {
    type: "flags"
    address: "0x0060"
    advanced: {
        flagDefinitions: {
            "double_jump": {bit: 0, description: "Can double jump"}
            "wall_climb": {bit: 1, description: "Can climb walls"}
            "fire_immunity": {bit: 2, description: "Immune to fire"}
        }
    }
}
```

#### Color Properties
```cue
backgroundColor: {
    type: "color"
    address: "0x3F00"
    length: 2
    advanced: {
        colorFormat: "rgb565"
        alphaChannel: false
    }
}
```

#### Coordinate Properties
```cue
cameraPosition: {
    type: "coordinate"
    address: "0x0040"
    length: 4
    advanced: {
        coordinateSystem: "world"
        dimensions: 2
    }
}
```

## 🧪 Testing Enhanced Features

### Test Property Freezing
```bash
# Test freeze functionality
./gamehook-enhanced test freeze playerHealth

# Test batch operations
./gamehook-enhanced test batch

# Validate enhanced mappers
./gamehook-enhanced validate pokemon_red_blue_enhanced
```

### Load Sample Data
```bash
# Load test data for development
curl -X POST http://localhost:8080/api/test/load-sample-data
```

## 🐳 Docker Support

### Build Container
```bash
docker build -f docker/Dockerfile -t gamehook-enhanced .
```

### Run with Docker Compose
```bash
docker-compose -f docker/docker-compose.yml up
```

### Docker Compose Configuration
```yaml
version: '3.8'
services:
  gamehook-enhanced:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./mappers:/app/mappers
      - ./uis:/app/uis
      - ./config:/app/config
    environment:
      - GAMEHOOK_SERVER_PORT=8080
      - GAMEHOOK_RETROARCH_HOST=host.docker.internal
```

## 📊 Performance & Monitoring

### Real-time Metrics
- Property read/write counts
- Memory access statistics
- WebSocket connection metrics
- Validation error rates
- Freeze operation counts

### Memory Usage
- Property state tracking
- Change history buffers
- WebSocket client management
- Memory block caching

### Benchmarks
```bash
# Run performance benchmarks
go test -bench=. -benchmem ./...

# Profile memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile
```

## 🔒 Security Features

### Input Validation
- Property value validation
- Address range checking
- Type safety enforcement
- CUE expression sandboxing

### Rate Limiting
- API request limiting
- WebSocket message throttling
- Memory operation limits

### Authentication (Optional)
- API key support
- CORS configuration
- Request origin validation

## 🛠️ Development Guide

### Adding New Property Types

1. **Define in CUE Schema**
```cue
#PropertyType: "uint8" | ... | "my_new_type"
```

2. **Implement Processor**
```go
func (app *AdvancedPropertyProcessor) processMyNewType(prop *Property) (interface{}, error) {
    // Implementation
}
```

3. **Add to Type Switch**
```go
case "my_new_type":
    return app.processMyNewType(prop)
```

4. **Update Validation**
```go
func validateMyNewType(value interface{}, validation *PropertyValidation) error {
    // Validation logic
}
```

### Adding New Validation Rules

1. **Extend PropertyValidation**
```go
type PropertyValidation struct {
    // ... existing fields
    MyNewRule *MyNewRuleConfig `json:"my_new_rule,omitempty"`
}
```

2. **Implement Validation Logic**
```go
func validateMyNewRule(value interface{}, rule *MyNewRuleConfig) error {
    // Validation logic
}
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/enhanced-xyz`)
3. Add tests for new functionality
4. Ensure all tests pass (`go test ./...`)
5. Validate enhanced mappers (`./gamehook-enhanced validate`)
6. Submit a pull request

## 🎯 Use Cases

### Enhanced OBS Overlays
- Real-time property monitoring with 60fps updates
- Color-coded property groups
- Freeze indicators and status displays
- Computed statistics and progress bars

### Advanced Speedrunning Tools
- Property freezing for practice
- Real-time validation and warnings
- Batch property manipulation
- Change history and analytics

### Game Development & Testing
- Memory state management
- Automated testing scenarios
- Property validation and constraints
- Performance monitoring

### Research & Analysis
- Property access statistics
- Memory usage patterns
- Game state correlation analysis
- Historical data tracking

## 📚 Additional Resources

- [Enhanced Mapper Creation Guide](docs/MAPPER_GUIDE.md)
- [Complete API Reference](docs/API_REFERENCE.md)
- [Configuration Reference](docs/CONFIGURATION.md)
- [Performance Tuning Guide](docs/PERFORMANCE.md)
- [Security Best Practices](docs/SECURITY.md)
- [Migration from Basic Version](docs/MIGRATION.md)

## 🤝 Community

- [GitHub Discussions](https://github.com/gamehook/enhanced/discussions)
- [Discord Server](https://discord.gg/gamehook)
- [Enhanced Mapper Repository](https://github.com/gamehook/enhanced-mappers)
- [Documentation Wiki](https://wiki.gamehook.io/enhanced)

## 📄 License

Enhanced GameHook is released under the MIT License. See [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Original GameHook project for inspiration
- RetroArch team for network command interface
- CUE language team for configuration system
- Community contributors and mapper creators
- Enhanced features inspired by modern development tools

---

**Enhanced GameHook** - Taking retro game memory manipulation to the next level! 🚀