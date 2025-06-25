# GameHook Go

A modern retro game memory manipulation tool built with Go and CUE. This is a complete rewrite of the original GameHook application with a focus on simplicity, performance, and extensibility.

## Features

- 🎮 **RetroArch Integration** - Connect to RetroArch via UDP for real-time memory access
- 📝 **CUE Configuration** - Type-safe mapper definitions using CUE language
- 🌐 **REST API** - Clean HTTP API for all operations
- 📡 **WebSocket Support** - Real-time updates via WebSocket
- 🎨 **Custom UIs** - Drop-in UI system for OBS overlays and custom tools
- ⚡ **High Performance** - Sub-10ms update loops written in Go
- 🔧 **Extensible** - Easy to add new platforms and property types

## Quick Start

### Prerequisites

- Go 1.21+
- RetroArch with network commands enabled

### Build & Run

```bash
# Clone and build
git clone <repo>
cd gamehook-go
go mod tidy
go build -o gamehook cmd/gamehook/main.go

# Run with RetroArch
./gamehook --retroarch-host 127.0.0.1 --retroarch-port 55355
```

### RetroArch Setup

1. Open RetroArch
2. Go to Settings → Network → Network Commands: **ON**
3. Set Network Command Port: **55355**
4. Load a NES game (for the example mapper)

### Load a Mapper

```bash
# Load the Super Mario Bros mapper
curl -X POST http://localhost:8080/api/mappers/super_mario_bros/load
```

### View the Overlay

Open http://localhost:8080/ui/mario-overlay/ in your browser or OBS.

## Project Structure

```
gamehook/
├── cmd/gamehook/main.go           # Application entry point
├── internal/
│   ├── drivers/retroarch.go       # RetroArch UDP communication
│   ├── memory/manager.go          # Memory management & type conversion
│   ├── mappers/                   # CUE parser & property system
│   │   ├── schema.cue             # CUE schema definitions
│   │   └── loader.go              # Mapper loading & processing
│   └── server/server.go           # HTTP server & REST API
├── mappers/                       # CUE mapper definitions
│   └── super_mario_bros.cue       # Example NES mapper
└── uis/                          # User interface folders
    └── mario-overlay/             # Example OBS overlay
        └── index.html
```

## API Reference

### Mappers

```bash
# List available mappers
GET /api/mappers

# Load a specific mapper
POST /api/mappers/{name}/load

# Get current mapper info
GET /api/mapper
```

### Properties

```bash
# List all properties
GET /api/properties

# Get specific property value
GET /api/properties/{name}

# Set property value
PUT /api/properties/{name}
Content-Type: application/json
{"value": 5}

# Raw memory access
GET /api/memory/{address}/{length}
```

### Real-time Updates

Connect to `ws://localhost:8080/api/stream` for real-time property updates:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/stream');
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'property_changed') {
        console.log(`${data.property} = ${data.value}`);
    }
};
```

## Creating Mappers

Mappers are defined using CUE files in the `mappers/` directory:

```cue
// mappers/my_game.cue
package mygame

name: "My Game"
game: "My Game Title"

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
    playerLives: {
        name: "playerLives"
        type: "uint8"
        address: "0x0030"
        description: "Number of lives remaining"
    }
    
    playerHealth: {
        name: "playerHealth"
        type: "uint8"
        address: "0x0040"
        description: "Player health points"
        transform: {
            lookup: {
                "0": "Dead"
                "1": "Critical"
                "2": "Low"
                "3": "Full"
            }
        }
    }
}
```

### Supported Property Types

- `uint8`, `uint16`, `uint32` - Unsigned integers
- `int8`, `int16`, `int32` - Signed integers
- `bool` - Boolean values
- `string` - Text with character mapping
- `bitfield` - Array of boolean flags

### Transformations

```cue
transform: {
    multiply: 10        // Multiply value
    add: 1             // Add to value
    lookup: {          // Value mapping
        "0": "Option A"
        "1": "Option B"
    }
}
```

## Creating Custom UIs

Drop any folder into `uis/` and access it at `/ui/{folder-name}/`:

```
uis/
├── my-overlay/
│   ├── index.html      # Entry point
│   ├── style.css       # Custom styles
│   ├── script.js       # GameHook API integration
│   └── assets/         # Images, fonts, etc.
└── speedrun-timer/
    └── index.html
```

### Example UI Integration

```javascript
class GameHookClient {
    constructor(baseUrl = 'http://localhost:8080') {
        this.baseUrl = baseUrl;
    }
    
    async getProperty(name) {
        const response = await fetch(`${this.baseUrl}/api/properties/${name}`);
        return response.json();
    }
    
    connectWebSocket(onMessage) {
        const ws = new WebSocket(`ws://localhost:8080/api/stream`);
        ws.onmessage = (event) => onMessage(JSON.parse(event.data));
        return ws;
    }
}

// Usage
const client = new GameHookClient();
client.getProperty('playerLives').then(data => {
    console.log('Lives:', data.value);
});
```

## Command Line Options

```bash
./gamehook [options]

Options:
  --port int                    HTTP server port (default 8080)
  --retroarch-host string       RetroArch host (default "127.0.0.1")
  --retroarch-port int          RetroArch UDP port (default 55355)
  --mappers-dir string          Directory containing CUE files (default "./mappers")
  --uis-dir string              Directory containing UI folders (default "./uis")
  --update-interval duration    Memory update interval (default 5ms)
  --request-timeout duration    RetroArch request timeout (default 64ms)
```

## Development

### Adding New Property Types

1. Add the type to the CUE schema in `internal/mappers/schema.cue`
2. Implement reading/writing in `internal/memory/manager.go`
3. Add parsing logic in `internal/mappers/loader.go`

### Adding New Emulator Support

1. Implement the `Driver` interface in `internal/drivers/`
2. Add platform configuration in mapper files
3. Update the main application to support the new driver

### Testing

```bash
# Run tests
go test ./...

# Test with static memory (no emulator required)
go run cmd/gamehook/main.go --driver static

# Load test data
curl -X POST http://localhost:8080/api/test/load-sample-data
```

## Use Cases

- **OBS Overlays** - Real-time game stats for streaming
- **Speedrunning Tools** - Custom timers and route tracking
- **Game Analysis** - Memory watching for research
- **Automation** - Automated testing and TAS creation
- **Learning** - Understanding how games work internally

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

[Your license here]

## Acknowledgments

- Original GameHook project for inspiration
- RetroArch team for the network command interface
- CUE language team for the configuration system

go run ./cmd/gamehook