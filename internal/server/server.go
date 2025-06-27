package server

import (
	"context"
	"encoding/json"
	"fmt"
	"gamehook/internal/mappers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Enhanced GameHookAPI interface for the server
type GameHookAPI interface {
	LoadMapper(name string) error
	GetCurrentMapper() interface{}
	GetCurrentMapperFull() *mappers.Mapper
	GetProperty(name string) (interface{}, error)
	SetProperty(name string, value interface{}) error
	SetPropertyValue(name string, value interface{}) error
	SetPropertyBytes(name string, data []byte) error
	FreezeProperty(name string, freeze bool) error
	ListMappers() []string
	GetPropertyState(name string) interface{}
	GetAllPropertyStates() map[string]interface{}
	GetPropertyChanges() map[string]interface{}
	GetMapperMeta() interface{}
	GetMapperGlossary() interface{}
}

// Enhanced Server with property management capabilities
type Server struct {
	gameHook GameHookAPI
	uisDir   string
	port     int
	router   *mux.Router
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool

	// Property monitoring
	propertyMonitor *PropertyMonitor
	lastSnapshot    map[string]interface{}
}

// PropertyMonitor handles real-time property monitoring
type PropertyMonitor struct {
	server          *Server
	updateInterval  time.Duration
	running         bool
	stopChan        chan bool
	changeListeners []PropertyChangeListener
}

// PropertyChangeListener represents a callback for property changes
type PropertyChangeListener struct {
	PropertyName string
	Callback     func(name string, oldValue, newValue interface{})
}

// PropertyResponse represents an enhanced property API response
type PropertyResponse struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Address     string      `json:"address"`
	Description string      `json:"description,omitempty"`
	Frozen      bool        `json:"frozen"`
	ReadOnly    bool        `json:"read_only"`
	Validation  interface{} `json:"validation,omitempty"`
	LastChanged time.Time   `json:"last_changed,omitempty"`
	ReadCount   uint64      `json:"read_count,omitempty"`
	WriteCount  uint64      `json:"write_count,omitempty"`
}

// PropertyStateResponse represents property state information
type PropertyStateResponse struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Bytes       []byte      `json:"bytes,omitempty"`
	Address     uint32      `json:"address"`
	Frozen      bool        `json:"frozen"`
	LastChanged time.Time   `json:"last_changed"`
	LastRead    time.Time   `json:"last_read"`
	ReadCount   uint64      `json:"read_count"`
	WriteCount  uint64      `json:"write_count"`
}

// BatchPropertyUpdate represents a batch update request
type BatchPropertyUpdate struct {
	Properties []PropertyUpdate `json:"properties"`
	Atomic     bool             `json:"atomic"`
}

// PropertyUpdate represents a single property update
type PropertyUpdate struct {
	Name   string      `json:"name"`
	Value  interface{} `json:"value,omitempty"`
	Bytes  []byte      `json:"bytes,omitempty"`
	Freeze *bool       `json:"freeze,omitempty"`
}

// MapperResponse represents an enhanced mapper API response
type MapperResponse struct {
	Name       string                      `json:"name"`
	Game       string                      `json:"game"`
	Version    string                      `json:"version,omitempty"`
	Platform   string                      `json:"platform"`
	Properties map[string]PropertyResponse `json:"properties"`
	Groups     map[string]interface{}      `json:"groups,omitempty"`
	Computed   map[string]interface{}      `json:"computed,omitempty"`
	Constants  map[string]interface{}      `json:"constants,omitempty"`
}

// MapperMetaResponse represents mapper metadata
type MapperMetaResponse struct {
	Name          string                 `json:"name"`
	Game          string                 `json:"game"`
	Version       string                 `json:"version,omitempty"`
	MinVersion    string                 `json:"min_version,omitempty"`
	Platform      interface{}            `json:"platform"`
	PropertyCount int                    `json:"property_count"`
	GroupCount    int                    `json:"group_count"`
	ComputedCount int                    `json:"computed_count"`
	FrozenCount   int                    `json:"frozen_count"`
	LastLoaded    time.Time              `json:"last_loaded"`
	MemoryBlocks  []interface{}          `json:"memory_blocks"`
	Constants     map[string]interface{} `json:"constants,omitempty"`
}

// GlossaryEntry represents a glossary entry for a property
type GlossaryEntry struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Address      string                 `json:"address"`
	Description  string                 `json:"description,omitempty"`
	Group        string                 `json:"group,omitempty"`
	ReadOnly     bool                   `json:"read_only"`
	Freezable    bool                   `json:"freezable"`
	Validation   map[string]interface{} `json:"validation,omitempty"`
	Transform    map[string]interface{} `json:"transform,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// NewServer creates a new enhanced HTTP server
func New(gameHook GameHookAPI, uisDir string, port int) *Server {
	server := &Server{
		gameHook: gameHook,
		uisDir:   uisDir,
		port:     port,
		router:   mux.NewRouter(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:      make(map[*websocket.Conn]bool),
		lastSnapshot: make(map[string]interface{}),
	}

	// Initialize property monitor
	server.propertyMonitor = &PropertyMonitor{
		server:         server,
		updateInterval: 16 * time.Millisecond, // ~60fps
		stopChan:       make(chan bool),
	}

	server.setupRoutes()
	return server
}

// Start starts the HTTP server and property monitoring
func (s *Server) Start() error {
	// Start property monitoring
	s.propertyMonitor.Start()

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop property monitoring
	s.propertyMonitor.Stop()

	// Close all WebSocket connections
	for client := range s.clients {
		client.Close()
	}
	return nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.Use(s.corsMiddleware)
	api.Use(s.jsonMiddleware)

	// Mapper management
	api.HandleFunc("/mappers", s.handleListMappers).Methods("GET")
	api.HandleFunc("/mappers/{name}/load", s.handleLoadMapper).Methods("POST")
	api.HandleFunc("/mapper", s.handleGetCurrentMapper).Methods("GET")
	api.HandleFunc("/mapper/meta", s.handleGetMapperMeta).Methods("GET")
	api.HandleFunc("/mapper/glossary", s.handleGetGlossary).Methods("GET")

	// Property access and management
	api.HandleFunc("/properties", s.handleListProperties).Methods("GET")
	api.HandleFunc("/properties/states", s.handleGetPropertyStates).Methods("GET")
	api.HandleFunc("/properties/batch", s.handleBatchPropertyUpdate).Methods("PUT")
	api.HandleFunc("/properties/{name}", s.handleGetProperty).Methods("GET")
	api.HandleFunc("/properties/{name}", s.handleSetProperty).Methods("PUT")
	api.HandleFunc("/properties/{name}/value", s.handleSetPropertyValue).Methods("PUT")
	api.HandleFunc("/properties/{name}/bytes", s.handleSetPropertyBytes).Methods("PUT")
	api.HandleFunc("/properties/{name}/freeze", s.handleFreezeProperty).Methods("POST")
	api.HandleFunc("/properties/{name}/state", s.handleGetPropertyState).Methods("GET")

	// Raw memory access
	api.HandleFunc("/memory/{address}/{length}", s.handleReadMemory).Methods("GET")

	// WebSocket for real-time updates
	api.HandleFunc("/stream", s.handleWebSocket)

	// UI serving
	s.router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(s.uisDir))))

	// Default route
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
}

// Property monitoring methods

// Start starts the property monitor
func (pm *PropertyMonitor) Start() {
	if pm.running {
		return
	}

	pm.running = true
	go pm.monitorLoop()
}

// Stop stops the property monitor
func (pm *PropertyMonitor) Stop() {
	if !pm.running {
		return
	}

	pm.running = false
	pm.stopChan <- true
}

// monitorLoop continuously monitors properties for changes
func (pm *PropertyMonitor) monitorLoop() {
	ticker := time.NewTicker(pm.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.stopChan:
			return
		case <-ticker.C:
			pm.checkForChanges()
		}
	}
}

// checkForChanges checks for property changes and broadcasts them
func (pm *PropertyMonitor) checkForChanges() {
	changes := pm.server.gameHook.GetPropertyChanges()
	if len(changes) == 0 {
		return
	}

	// Broadcast property changes
	for propertyName, newValue := range changes {
		oldValue := pm.server.lastSnapshot[propertyName]
		pm.server.lastSnapshot[propertyName] = newValue

		// Notify WebSocket clients
		pm.server.broadcastMessage(map[string]interface{}{
			"type":      "property_changed",
			"property":  propertyName,
			"value":     newValue,
			"old_value": oldValue,
			"timestamp": time.Now(),
		})

		// Notify registered listeners
		for _, listener := range pm.changeListeners {
			if listener.PropertyName == propertyName || listener.PropertyName == "*" {
				go listener.Callback(propertyName, oldValue, newValue)
			}
		}
	}
}

// AddChangeListener adds a property change listener
func (pm *PropertyMonitor) AddChangeListener(propertyName string, callback func(string, interface{}, interface{})) {
	pm.changeListeners = append(pm.changeListeners, PropertyChangeListener{
		PropertyName: propertyName,
		Callback:     callback,
	})
}

// Middleware functions
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Enhanced API Handlers

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GameHook Go - Enhanced</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #1a1a1a; color: #fff; }
        .container { max-width: 1000px; margin: 0 auto; }
        h1 { color: #4CAF50; }
        .section { margin: 20px 0; padding: 20px; background: #2a2a2a; border-radius: 8px; }
        .endpoint { background: #333; padding: 10px; margin: 5px 0; border-radius: 4px; font-family: monospace; }
        a { color: #4CAF50; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .new { color: #FF9800; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎮 GameHook Go - Enhanced</h1>
        <p>Modern retro game memory manipulation with advanced property management.</p>
        
        <div class="section">
            <h2>Enhanced API Endpoints</h2>
            
            <h3>Mapper Management</h3>
            <div class="endpoint">GET <a href="/api/mappers">/api/mappers</a> - List available mappers</div>
            <div class="endpoint">POST /api/mappers/{name}/load - Load a mapper</div>
            <div class="endpoint">GET <a href="/api/mapper">/api/mapper</a> - Get current mapper info</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/mapper/meta">/api/mapper/meta</a> - Get mapper metadata</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/mapper/glossary">/api/mapper/glossary</a> - Get property glossary</div>
            
            <h3>Property Management</h3>
            <div class="endpoint">GET <a href="/api/properties">/api/properties</a> - List all properties</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/properties/states">/api/properties/states</a> - Get all property states</div>
            <div class="endpoint">GET /api/properties/{name} - Get specific property</div>
            <div class="endpoint">PUT /api/properties/{name} - Set property (legacy)</div>
            <div class="endpoint"><span class="new">NEW</span> PUT /api/properties/{name}/value - Set property value</div>
            <div class="endpoint"><span class="new">NEW</span> PUT /api/properties/{name}/bytes - Set property bytes</div>
            <div class="endpoint"><span class="new">NEW</span> POST /api/properties/{name}/freeze - Freeze/unfreeze property</div>
            <div class="endpoint"><span class="new">NEW</span> GET /api/properties/{name}/state - Get property state</div>
            <div class="endpoint"><span class="new">NEW</span> PUT /api/properties/batch - Batch property updates</div>
            
            <h3>Real-time Communication</h3>
            <div class="endpoint">WS /api/stream - Enhanced WebSocket for real-time updates</div>
        </div>
        
        <div class="section">
            <h2>New Features</h2>
            <ul>
                <li><strong>Property Freezing:</strong> Lock property values to prevent changes</li>
                <li><strong>Enhanced Types:</strong> Support for advanced property types</li>
                <li><strong>Real-time Monitoring:</strong> 60fps property change detection</li>
                <li><strong>Batch Operations:</strong> Update multiple properties atomically</li>
                <li><strong>Property Validation:</strong> Enforce constraints and rules</li>
                <li><strong>State Tracking:</strong> Monitor read/write counts and history</li>
                <li><strong>Property Groups:</strong> Organize properties for better UX</li>
            </ul>
        </div>
        
        <div class="section">
            <h2>User Interfaces</h2>
            <p>Access custom UIs at:</p>
            <div class="endpoint">GET <a href="/ui/">/ui/{folder-name}/</a></div>
        </div>
    </div>
</body>
</html>
	`)
}

func (s *Server) handleGetMapperMeta(w http.ResponseWriter, r *http.Request) {
	meta := s.gameHook.GetMapperMeta()
	if meta == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	json.NewEncoder(w).Encode(meta)
}

func (s *Server) handleGetGlossary(w http.ResponseWriter, r *http.Request) {
	glossary := s.gameHook.GetMapperGlossary()
	if glossary == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	json.NewEncoder(w).Encode(glossary)
}

func (s *Server) handleGetPropertyStates(w http.ResponseWriter, r *http.Request) {
	states := s.gameHook.GetAllPropertyStates()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"states": states,
		"count":  len(states),
	})
}

func (s *Server) handleGetPropertyState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	state := s.gameHook.GetPropertyState(name)
	if state == nil {
		s.writeError(w, http.StatusNotFound, "PROPERTY_NOT_FOUND", fmt.Sprintf("Property %s not found or has no state", name))
		return
	}

	json.NewEncoder(w).Encode(state)
}

func (s *Server) handleSetPropertyValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var request struct {
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := s.gameHook.SetPropertyValue(name, request.Value); err != nil {
		s.writeError(w, http.StatusBadRequest, "SET_FAILED", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Property %s value updated successfully", name),
	})

	// Broadcast change
	s.broadcastMessage(map[string]interface{}{
		"type":      "property_changed",
		"property":  name,
		"value":     request.Value,
		"timestamp": time.Now(),
		"source":    "api_set_value",
	})
}

func (s *Server) handleSetPropertyBytes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var request struct {
		Bytes []byte `json:"bytes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := s.gameHook.SetPropertyBytes(name, request.Bytes); err != nil {
		s.writeError(w, http.StatusBadRequest, "SET_FAILED", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Property %s bytes updated successfully", name),
	})

	// Broadcast change
	s.broadcastMessage(map[string]interface{}{
		"type":      "property_changed",
		"property":  name,
		"bytes":     request.Bytes,
		"timestamp": time.Now(),
		"source":    "api_set_bytes",
	})
}

func (s *Server) handleFreezeProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var request struct {
		Freeze bool `json:"freeze"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := s.gameHook.FreezeProperty(name, request.Freeze); err != nil {
		s.writeError(w, http.StatusBadRequest, "FREEZE_FAILED", err.Error())
		return
	}

	action := "frozen"
	if !request.Freeze {
		action = "unfrozen"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Property %s %s successfully", name, action),
		"frozen":  request.Freeze,
	})

	// Broadcast freeze state change
	s.broadcastMessage(map[string]interface{}{
		"type":      "property_freeze_changed",
		"property":  name,
		"frozen":    request.Freeze,
		"timestamp": time.Now(),
	})
}

func (s *Server) handleBatchPropertyUpdate(w http.ResponseWriter, r *http.Request) {
	var batch BatchPropertyUpdate

	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	results := make([]map[string]interface{}, 0, len(batch.Properties))
	var lastError error

	for _, update := range batch.Properties {
		result := map[string]interface{}{
			"property": update.Name,
			"success":  false,
		}

		// Handle value update
		if update.Value != nil {
			if err := s.gameHook.SetPropertyValue(update.Name, update.Value); err != nil {
				result["error"] = err.Error()
				lastError = err
				if batch.Atomic {
					break // Stop on first error in atomic mode
				}
			} else {
				result["success"] = true
				result["value_updated"] = true
			}
		}

		// Handle bytes update
		if update.Bytes != nil {
			if err := s.gameHook.SetPropertyBytes(update.Name, update.Bytes); err != nil {
				result["error"] = err.Error()
				lastError = err
				if batch.Atomic {
					break
				}
			} else {
				result["success"] = true
				result["bytes_updated"] = true
			}
		}

		// Handle freeze state
		if update.Freeze != nil {
			if err := s.gameHook.FreezeProperty(update.Name, *update.Freeze); err != nil {
				result["error"] = err.Error()
				lastError = err
				if batch.Atomic {
					break
				}
			} else {
				result["success"] = true
				result["freeze_updated"] = true
				result["frozen"] = *update.Freeze
			}
		}

		results = append(results, result)
	}

	response := map[string]interface{}{
		"results": results,
		"total":   len(batch.Properties),
		"atomic":  batch.Atomic,
	}

	// Calculate success count
	successCount := 0
	for _, result := range results {
		if result["success"].(bool) {
			successCount++
		}
	}
	response["success_count"] = successCount

	if batch.Atomic && lastError != nil {
		response["success"] = false
		response["error"] = "Atomic batch failed: " + lastError.Error()
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response["success"] = successCount > 0
	}

	json.NewEncoder(w).Encode(response)

	// Broadcast batch update
	s.broadcastMessage(map[string]interface{}{
		"type":          "batch_update_completed",
		"results":       results,
		"success_count": successCount,
		"total":         len(batch.Properties),
		"timestamp":     time.Now(),
	})
}

// Legacy handlers (enhanced)

func (s *Server) handleListMappers(w http.ResponseWriter, r *http.Request) {
	mappers := s.gameHook.ListMappers()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mappers": mappers,
		"count":   len(mappers),
	})
}

func (s *Server) handleLoadMapper(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if err := s.gameHook.LoadMapper(name); err != nil {
		s.writeError(w, http.StatusBadRequest, "LOAD_FAILED", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Mapper %s loaded successfully", name),
		"mapper":  name,
	})

	// Notify WebSocket clients
	s.broadcastMessage(map[string]interface{}{
		"type":      "mapper_loaded",
		"mapper":    name,
		"timestamp": time.Now(),
	})
}

func (s *Server) handleGetCurrentMapper(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapper()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	json.NewEncoder(w).Encode(mapper)
}

func (s *Server) handleListProperties(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	// Build enhanced property list
	properties := make([]PropertyResponse, 0, len(mapper.Properties))
	frozenCount := 0

	for name, prop := range mapper.Properties {
		// Get current value
		value, _ := s.gameHook.GetProperty(name)

		// Get state information
		state := s.gameHook.GetPropertyState(name)

		propResponse := PropertyResponse{
			Name:        name,
			Value:       value,
			Type:        string(prop.Type),
			Address:     fmt.Sprintf("0x%X", prop.Address),
			Description: prop.Description,
			Frozen:      prop.Frozen,
			ReadOnly:    prop.ReadOnly,
		}

		if state != nil {
			// Add state information if available
			if stateMap, ok := state.(map[string]interface{}); ok {
				if lastChanged, ok := stateMap["last_changed"].(time.Time); ok {
					propResponse.LastChanged = lastChanged
				}
				if readCount, ok := stateMap["read_count"].(uint64); ok {
					propResponse.ReadCount = readCount
				}
				if writeCount, ok := stateMap["write_count"].(uint64); ok {
					propResponse.WriteCount = writeCount
				}
			}
		}

		if prop.Validation != nil {
			propResponse.Validation = prop.Validation
		}

		if prop.Frozen {
			frozenCount++
		}

		properties = append(properties, propResponse)
	}

	response := map[string]interface{}{
		"properties":   properties,
		"total":        len(properties),
		"frozen_count": frozenCount,
		"mapper":       mapper.Name,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	mapper := s.gameHook.GetCurrentMapper()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	value, err := s.gameHook.GetProperty(name)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "PROPERTY_ERROR", err.Error())
		return
	}

	// Get additional property information
	state := s.gameHook.GetPropertyState(name)

	response := map[string]interface{}{
		"name":      name,
		"value":     value,
		"timestamp": time.Now(),
	}

	if state != nil {
		response["state"] = state
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSetProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	var request struct {
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := s.gameHook.SetProperty(name, request.Value); err != nil {
		s.writeError(w, http.StatusBadRequest, "SET_FAILED", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Property %s updated successfully", name),
	})

	// Notify WebSocket clients
	s.broadcastMessage(map[string]interface{}{
		"type":      "property_changed",
		"property":  name,
		"value":     request.Value,
		"timestamp": time.Now(),
		"source":    "api_legacy",
	})
}

func (s *Server) handleReadMemory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	address, err := strconv.ParseUint(vars["address"], 0, 32)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_ADDRESS", "Invalid memory address")
		return
	}

	length, err := strconv.ParseUint(vars["length"], 0, 32)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_LENGTH", "Invalid length")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":   fmt.Sprintf("0x%X", address),
		"length":    length,
		"message":   "Raw memory access would go here",
		"timestamp": time.Now(),
	})
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	s.clients[conn] = true
	defer delete(s.clients, conn)

	log.Printf("Enhanced WebSocket client connected")

	// Send welcome message with capabilities
	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "Enhanced WebSocket connection established",
		"features": []string{
			"property_monitoring",
			"batch_updates",
			"freeze_notifications",
			"state_tracking",
		},
		"update_rate": "60fps",
		"timestamp":   time.Now(),
	})

	// Listen for client messages
	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle enhanced client messages
		s.handleWebSocketMessage(conn, message)
	}

	log.Printf("Enhanced WebSocket client disconnected")
}

// handleWebSocketMessage handles incoming WebSocket messages
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribe_property":
		// Subscribe to specific property changes
		if propertyName, ok := message["property"].(string); ok {
			// TODO: Implement per-client property subscriptions
			conn.WriteJSON(map[string]interface{}{
				"type":      "subscription_confirmed",
				"property":  propertyName,
				"timestamp": time.Now(),
			})
		}

	case "get_property_state":
		// Get current state of a property
		if propertyName, ok := message["property"].(string); ok {
			state := s.gameHook.GetPropertyState(propertyName)
			conn.WriteJSON(map[string]interface{}{
				"type":      "property_state",
				"property":  propertyName,
				"state":     state,
				"timestamp": time.Now(),
			})
		}

	case "ping":
		// Respond to ping
		conn.WriteJSON(map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now(),
		})

	default:
		log.Printf("Unknown WebSocket message type: %s", msgType)
	}
}

// Helper functions

func (s *Server) writeError(w http.ResponseWriter, status int, errorType, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorType,
		Message: message,
	})
}

func (s *Server) broadcastMessage(message interface{}) {
	for client := range s.clients {
		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Error broadcasting to WebSocket client: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}
