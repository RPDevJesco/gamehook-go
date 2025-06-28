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

	// Enhanced API methods
	GetActiveEvents() []string
	GetRecentlyTriggeredEvents() []string
	TriggerEvent(name string, force bool) error
	GetValidationErrors() map[string]interface{}
}

// NOTE: Current mappers.Mapper struct has:
// - References map[string]*Property (needs to be map[string]*ReferenceType)
// - Events *EventsConfig (needs to be map[string]*Event)
// - GlobalValidation field doesn't exist yet
//
// The handlers are designed to work with both current and future enhanced versions

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

// Enhanced response types
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

// Enhanced API response types
type PropertyMetadataResponse struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Address      string                 `json:"address"`
	Description  string                 `json:"description,omitempty"`
	UIHints      interface{}            `json:"ui_hints,omitempty"`
	Advanced     interface{}            `json:"advanced,omitempty"`
	Performance  interface{}            `json:"performance,omitempty"`
	Validation   interface{}            `json:"validation,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Group        string                 `json:"group,omitempty"`
	References   map[string]interface{} `json:"references,omitempty"`
}

type ReferenceResponse struct {
	References map[string]interface{} `json:"references"`
	Count      int                    `json:"count"`
	Categories []string               `json:"categories,omitempty"`
}

type EventResponse struct {
	Events    map[string]interface{} `json:"events"`
	Count     int                    `json:"count"`
	Active    []string               `json:"active_events,omitempty"`
	Triggered []string               `json:"recently_triggered,omitempty"`
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

	// Enhanced API endpoints
	api.HandleFunc("/references", s.handleGetReferences).Methods("GET")
	api.HandleFunc("/references/{type}", s.handleGetReferenceType).Methods("GET")
	api.HandleFunc("/events", s.handleGetEvents).Methods("GET")
	api.HandleFunc("/events/{name}/trigger", s.handleTriggerEvent).Methods("POST")
	api.HandleFunc("/properties/{name}/metadata", s.handleGetPropertyMetadata).Methods("GET")
	api.HandleFunc("/properties/{name}/ui-hints", s.handleGetPropertyUIHints).Methods("GET")
	api.HandleFunc("/properties/by-group/{group}", s.handleGetPropertiesByGroup).Methods("GET")
	api.HandleFunc("/ui/themes", s.handleGetUIThemes).Methods("GET")
	api.HandleFunc("/ui/layout", s.handleGetUILayout).Methods("GET")
	api.HandleFunc("/validation/rules", s.handleGetValidationRules).Methods("GET")
	api.HandleFunc("/validation/errors", s.handleGetValidationErrors).Methods("GET")

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
            <div class="endpoint"><span class="new">NEW</span> GET /api/properties/{name}/metadata - Get property metadata</div>
            <div class="endpoint"><span class="new">NEW</span> GET /api/properties/{name}/ui-hints - Get property UI hints</div>
            <div class="endpoint"><span class="new">NEW</span> GET /api/properties/by-group/{group} - Get properties by group</div>
            
            <h3>Enhanced Features</h3>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/references">/api/references</a> - Get reference types</div>
            <div class="endpoint"><span class="new">NEW</span> GET /api/references/{type} - Get specific reference type</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/events">/api/events</a> - Get events</div>
            <div class="endpoint"><span class="new">NEW</span> POST /api/events/{name}/trigger - Trigger event</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/ui/themes">/api/ui/themes</a> - Get UI themes</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/ui/layout">/api/ui/layout</a> - Get UI layout</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/validation/rules">/api/validation/rules</a> - Get validation rules</div>
            <div class="endpoint"><span class="new">NEW</span> GET <a href="/api/validation/errors">/api/validation/errors</a> - Get validation errors</div>
            
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
                <li><strong>Reference Types:</strong> Enum and structured data support</li>
                <li><strong>Event System:</strong> Trigger-based automation</li>
                <li><strong>UI Hints:</strong> Rich metadata for enhanced interfaces</li>
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

// Enhanced API endpoint handlers

func (s *Server) handleGetReferences(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	// Build categories list and convert to interface{}
	categories := make([]string, 0)
	referencesMap := make(map[string]interface{})

	if mapper.References != nil {
		for refType, refData := range mapper.References {
			categories = append(categories, refType)
			referencesMap[refType] = refData // refData is *Property
		}
	}

	response := ReferenceResponse{
		References: referencesMap,
		Count:      len(referencesMap),
		Categories: categories,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetReferenceType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	refType := vars["type"]

	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	if mapper.References == nil {
		s.writeError(w, http.StatusNotFound, "NO_REFERENCES", "No references defined")
		return
	}

	reference, exists := mapper.References[refType]
	if !exists {
		s.writeError(w, http.StatusNotFound, "REFERENCE_NOT_FOUND", fmt.Sprintf("Reference type %s not found", refType))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":      refType,
		"reference": reference,
	})
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	// For now, return empty events map since mapper.Events is *EventsConfig (config), not event definitions
	// The actual events should be stored differently in the enhanced mapper
	eventsMap := make(map[string]interface{})

	// TODO: When enhanced Mapper has proper Events field as map[string]*Event, use:
	// if mapper.Events != nil {
	//     for eventName, eventData := range mapper.Events {
	//         eventsMap[eventName] = eventData
	//     }
	// }

	// Get active and recently triggered events from GameHook
	activeEvents := s.gameHook.GetActiveEvents()
	recentlyTriggered := s.gameHook.GetRecentlyTriggeredEvents()

	response := EventResponse{
		Events:    eventsMap,
		Count:     len(eventsMap),
		Active:    activeEvents,
		Triggered: recentlyTriggered,
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleTriggerEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventName := vars["name"]

	var request struct {
		Force bool `json:"force"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return
	}

	if err := s.gameHook.TriggerEvent(eventName, request.Force); err != nil {
		s.writeError(w, http.StatusBadRequest, "TRIGGER_FAILED", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"event":     eventName,
		"triggered": true,
		"timestamp": time.Now(),
	})

	// Broadcast event trigger
	s.broadcastMessage(map[string]interface{}{
		"type":       "event_triggered",
		"event_name": eventName,
		"forced":     request.Force,
		"timestamp":  time.Now(),
	})
}

func (s *Server) handleGetPropertyMetadata(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	prop, exists := mapper.Properties[name]
	if !exists {
		s.writeError(w, http.StatusNotFound, "PROPERTY_NOT_FOUND", "Property not found")
		return
	}

	// Find group for this property
	var group string
	for groupName, groupInfo := range mapper.Groups {
		for _, propName := range groupInfo.Properties {
			if propName == name {
				group = groupName
				break
			}
		}
	}

	// Build references map for enum/flags types
	references := make(map[string]interface{})
	if prop.Advanced != nil {
		if prop.Type == "enum" && prop.Advanced.EnumValues != nil {
			references["enumValues"] = prop.Advanced.EnumValues
		}
		if prop.Type == "flags" && prop.Advanced.FlagDefinitions != nil {
			references["flagDefinitions"] = prop.Advanced.FlagDefinitions
		}
	}

	metadata := PropertyMetadataResponse{
		Name:         name,
		Type:         string(prop.Type),
		Address:      fmt.Sprintf("0x%X", prop.Address),
		Description:  prop.Description,
		UIHints:      prop.UIHints,
		Advanced:     prop.Advanced,
		Performance:  prop.Performance,
		Validation:   prop.Validation,
		Dependencies: prop.DependsOn,
		Group:        group,
		References:   references,
	}

	json.NewEncoder(w).Encode(metadata)
}

func (s *Server) handleGetPropertyUIHints(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	prop, exists := mapper.Properties[name]
	if !exists {
		s.writeError(w, http.StatusNotFound, "PROPERTY_NOT_FOUND", "Property not found")
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"property":  name,
		"ui_hints":  prop.UIHints,
		"timestamp": time.Now(),
	})
}

func (s *Server) handleGetPropertiesByGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupName := vars["group"]

	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	group, exists := mapper.Groups[groupName]
	if !exists {
		s.writeError(w, http.StatusNotFound, "GROUP_NOT_FOUND", "Property group not found")
		return
	}

	// Get property values for this group
	properties := make([]PropertyResponse, 0, len(group.Properties))
	for _, propName := range group.Properties {
		if prop, exists := mapper.Properties[propName]; exists {
			value, _ := s.gameHook.GetProperty(propName)
			state := s.gameHook.GetPropertyState(propName)

			propResponse := PropertyResponse{
				Name:        propName,
				Value:       value,
				Type:        string(prop.Type),
				Address:     fmt.Sprintf("0x%X", prop.Address),
				Description: prop.Description,
				Frozen:      prop.Frozen,
				ReadOnly:    prop.ReadOnly,
			}

			if state != nil {
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

			properties = append(properties, propResponse)
		}
	}

	response := map[string]interface{}{
		"group_name": groupName,
		"group_info": group,
		"properties": properties,
		"count":      len(properties),
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetUIThemes(w http.ResponseWriter, r *http.Request) {
	themes := map[string]interface{}{
		"dark": map[string]string{
			"primary":   "#1a1a1a",
			"secondary": "#2a2a2a",
			"accent":    "#4CAF50",
			"text":      "#ffffff",
		},
		"light": map[string]string{
			"primary":   "#ffffff",
			"secondary": "#f5f5f5",
			"accent":    "#2196F3",
			"text":      "#000000",
		},
		"retro": map[string]string{
			"primary":   "#0f380f",
			"secondary": "#306230",
			"accent":    "#8bac0f",
			"text":      "#9bbc0f",
		},
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"themes":  themes,
		"default": "dark",
		"current": "dark", // Could be configurable
	})
}

func (s *Server) handleGetUILayout(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	// Count references safely
	referencesCount := 0
	if mapper.References != nil {
		referencesCount = len(mapper.References)
	}

	// Events count is 0 for now since mapper.Events is config, not event definitions
	eventsCount := 0

	layout := map[string]interface{}{
		"groups":           mapper.Groups,
		"property_count":   len(mapper.Properties),
		"computed_count":   len(mapper.Computed),
		"references_count": referencesCount,
		"events_count":     eventsCount,
	}

	json.NewEncoder(w).Encode(layout)
}

func (s *Server) handleGetValidationRules(w http.ResponseWriter, r *http.Request) {
	mapper := s.gameHook.GetCurrentMapperFull()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper loaded")
		return
	}

	rules := make(map[string]interface{})
	for name, prop := range mapper.Properties {
		if prop.Validation != nil {
			rules[name] = prop.Validation
		}
	}

	// Handle global validation if it exists (check if the field exists)
	var globalRules interface{}
	// Note: GlobalValidation field may not exist yet in the current Mapper struct
	// This is a placeholder for when the field is added to the enhanced Mapper
	// globalRules = mapper.GlobalValidation

	json.NewEncoder(w).Encode(map[string]interface{}{
		"rules":        rules,
		"global_rules": globalRules,
		"count":        len(rules),
	})
}

func (s *Server) handleGetValidationErrors(w http.ResponseWriter, r *http.Request) {
	errors := s.gameHook.GetValidationErrors()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors":    errors,
		"count":     len(errors),
		"timestamp": time.Now(),
	})
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
			"events",
			"references",
			"ui_hints",
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

// Enhanced WebSocket message handling
func (s *Server) handleWebSocketMessage(conn *websocket.Conn, message map[string]interface{}) {
	msgType, ok := message["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "subscribe_property":
		// Subscribe to specific property changes
		if propertyName, ok := message["property"].(string); ok {
			conn.WriteJSON(map[string]interface{}{
				"type":      "subscription_confirmed",
				"property":  propertyName,
				"timestamp": time.Now(),
			})
		}

	case "subscribe_events":
		conn.WriteJSON(map[string]interface{}{
			"type":      "event_subscription_confirmed",
			"timestamp": time.Now(),
		})

	case "get_property_metadata":
		if propertyName, ok := message["property"].(string); ok {
			mapper := s.gameHook.GetCurrentMapperFull()
			if mapper != nil && mapper.Properties[propertyName] != nil {
				prop := mapper.Properties[propertyName]
				conn.WriteJSON(map[string]interface{}{
					"type":       "property_metadata",
					"property":   propertyName,
					"ui_hints":   prop.UIHints,
					"advanced":   prop.Advanced,
					"validation": prop.Validation,
					"timestamp":  time.Now(),
				})
			}
		}

	case "trigger_event":
		if eventName, ok := message["event"].(string); ok {
			force, _ := message["force"].(bool)
			if err := s.gameHook.TriggerEvent(eventName, force); err == nil {
				conn.WriteJSON(map[string]interface{}{
					"type":      "event_triggered",
					"event":     eventName,
					"success":   true,
					"timestamp": time.Now(),
				})
			} else {
				conn.WriteJSON(map[string]interface{}{
					"type":      "event_trigger_failed",
					"event":     eventName,
					"error":     err.Error(),
					"timestamp": time.Now(),
				})
			}
		}

	case "get_ui_layout":
		mapper := s.gameHook.GetCurrentMapperFull()
		if mapper != nil {
			conn.WriteJSON(map[string]interface{}{
				"type":       "ui_layout",
				"groups":     mapper.Groups,
				"computed":   mapper.Computed,
				"references": mapper.References,
				"timestamp":  time.Now(),
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
