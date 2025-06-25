package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

// GameHookAPI interface for the server to interact with the main application
type GameHookAPI interface {
	LoadMapper(name string) error
	GetCurrentMapper() interface{}
	GetProperty(name string) (interface{}, error)
	SetProperty(name string, value interface{}) error
	ListMappers() []string
}

// Server handles HTTP requests and WebSocket connections
type Server struct {
	gameHook GameHookAPI
	uisDir   string
	port     int
	router   *mux.Router
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

// PropertyResponse represents a property API response
type PropertyResponse struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Address     string      `json:"address"`
	Description string      `json:"description,omitempty"`
}

// MapperResponse represents a mapper API response
type MapperResponse struct {
	Name       string                      `json:"name"`
	Game       string                      `json:"game"`
	Platform   string                      `json:"platform"`
	Properties map[string]PropertyResponse `json:"properties"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewServer creates a new HTTP server
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
		clients: make(map[*websocket.Conn]bool),
	}

	server.setupRoutes()
	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Server listening on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
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

	// Property access
	api.HandleFunc("/properties", s.handleListProperties).Methods("GET")
	api.HandleFunc("/properties/{name}", s.handleGetProperty).Methods("GET")
	api.HandleFunc("/properties/{name}", s.handleSetProperty).Methods("PUT")

	// Raw memory access
	api.HandleFunc("/memory/{address}/{length}", s.handleReadMemory).Methods("GET")

	// WebSocket for real-time updates
	api.HandleFunc("/stream", s.handleWebSocket)

	// UI serving
	s.router.PathPrefix("/ui/").Handler(http.StripPrefix("/ui/", http.FileServer(http.Dir(s.uisDir))))

	// Default route
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
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

// API Handlers

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>GameHook Go</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #1a1a1a; color: #fff; }
        .container { max-width: 800px; margin: 0 auto; }
        h1 { color: #4CAF50; }
        .section { margin: 20px 0; padding: 20px; background: #2a2a2a; border-radius: 8px; }
        .endpoint { background: #333; padding: 10px; margin: 5px 0; border-radius: 4px; font-family: monospace; }
        a { color: #4CAF50; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>GameHook Go</h1>
        <p>A modern retro game memory manipulation tool built with Go and CUE.</p>
        
        <div class="section">
            <h2>API Endpoints</h2>
            <div class="endpoint">GET <a href="/api/mappers">/api/mappers</a> - List available mappers</div>
            <div class="endpoint">POST /api/mappers/{name}/load - Load a mapper</div>
            <div class="endpoint">GET <a href="/api/mapper">/api/mapper</a> - Get current mapper info</div>
            <div class="endpoint">GET <a href="/api/properties">/api/properties</a> - List all properties</div>
            <div class="endpoint">GET /api/properties/{name} - Get specific property</div>
            <div class="endpoint">PUT /api/properties/{name} - Set property value</div>
            <div class="endpoint">WS /api/stream - WebSocket for real-time updates</div>
        </div>
        
        <div class="section">
            <h2>User Interfaces</h2>
            <p>Place your custom UIs in the <code>uis/</code> directory and access them at:</p>
            <div class="endpoint">GET <a href="/ui/">/ui/{folder-name}/</a></div>
        </div>
    </div>
</body>
</html>
	`)
}

func (s *Server) handleListMappers(w http.ResponseWriter, r *http.Request) {
	mappers := s.gameHook.ListMappers()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"mappers": mappers,
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
	})

	// Notify WebSocket clients
	s.broadcastMessage(map[string]interface{}{
		"type":   "mapper_loaded",
		"mapper": name,
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
	mapper := s.gameHook.GetCurrentMapper()
	if mapper == nil {
		s.writeError(w, http.StatusNotFound, "NO_MAPPER", "No mapper currently loaded")
		return
	}

	// This would need to be implemented based on the actual mapper structure
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Properties list would go here",
	})
}

func (s *Server) handleGetProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	value, err := s.gameHook.GetProperty(name)
	if err != nil {
		s.writeError(w, http.StatusNotFound, "PROPERTY_NOT_FOUND", err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":  name,
		"value": value,
	})
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
		"type":     "property_changed",
		"property": name,
		"value":    request.Value,
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

	// This would need to access the memory manager directly
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": fmt.Sprintf("0x%X", address),
		"length":  length,
		"message": "Raw memory access would go here",
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

	log.Printf("WebSocket client connected")

	// Send welcome message
	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"message": "WebSocket connection established",
	})

	// Listen for client messages (optional)
	for {
		var message map[string]interface{}
		err := conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages here if needed
		log.Printf("Received WebSocket message: %v", message)
	}

	log.Printf("WebSocket client disconnected")
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
