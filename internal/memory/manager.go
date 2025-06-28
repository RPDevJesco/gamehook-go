package memory

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// ===== ENHANCED VALIDATION SYSTEM =====

// ValidationError represents a property validation error with enhanced details
type ValidationError struct {
	Property  string      `json:"property"`
	Rule      string      `json:"rule"`
	Message   string      `json:"message"`
	Value     interface{} `json:"value"`
	Severity  string      `json:"severity"` // "error", "warning", "info"
	Timestamp time.Time   `json:"timestamp"`
	Context   interface{} `json:"context,omitempty"`
}

// PerformanceMetrics represents detailed performance tracking for properties
type PerformanceMetrics struct {
	ReadCount       uint64        `json:"read_count"`
	WriteCount      uint64        `json:"write_count"`
	ErrorCount      uint64        `json:"error_count"`
	AvgReadTime     time.Duration `json:"avg_read_time"`
	AvgWriteTime    time.Duration `json:"avg_write_time"`
	MaxReadTime     time.Duration `json:"max_read_time"`
	MaxWriteTime    time.Duration `json:"max_write_time"`
	LastReadTime    time.Duration `json:"last_read_time"`
	LastWriteTime   time.Duration `json:"last_write_time"`
	CacheHits       uint64        `json:"cache_hits"`
	CacheMisses     uint64        `json:"cache_misses"`
	CacheHitRatio   float64       `json:"cache_hit_ratio"`
	LastCacheTime   time.Time     `json:"last_cache_time"`
	TotalReadBytes  uint64        `json:"total_read_bytes"`
	TotalWriteBytes uint64        `json:"total_write_bytes"`
	FirstAccess     time.Time     `json:"first_access"`
	LastAccess      time.Time     `json:"last_access"`
}

// PropertyEvent represents an event that occurred for a property
type PropertyEvent struct {
	Type      string                 `json:"type"` // "read", "write", "freeze", "unfreeze", "validate", "transform"
	Timestamp time.Time              `json:"timestamp"`
	Value     interface{}            `json:"value,omitempty"`
	OldValue  interface{}            `json:"old_value,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Source    string                 `json:"source,omitempty"` // "api", "websocket", "batch", "computed"
	Context   map[string]interface{} `json:"context,omitempty"`
}

// PropertyChangePattern represents detected patterns in property changes
type PropertyChangePattern struct {
	PropertyName    string        `json:"property_name"`
	PatternType     string        `json:"pattern_type"` // "periodic", "threshold", "sequence", "correlation"
	Confidence      float64       `json:"confidence"`   // 0.0 to 1.0
	Period          time.Duration `json:"period,omitempty"`
	Threshold       interface{}   `json:"threshold,omitempty"`
	RelatedProperty string        `json:"related_property,omitempty"`
	LastDetected    time.Time     `json:"last_detected"`
	DetectionCount  uint64        `json:"detection_count"`
	Description     string        `json:"description"`
}

// FrozenProperty represents enhanced frozen property with expiration and conditions
type FrozenProperty struct {
	Address          uint32                 `json:"address"`
	Data             []byte                 `json:"data"`
	LastWritten      time.Time              `json:"last_written"`
	FreezeTime       time.Time              `json:"freeze_time"`
	ExpiryTime       *time.Time             `json:"expiry_time,omitempty"`
	Condition        string                 `json:"condition,omitempty"` // CUE expression for conditional freezing
	WriteAttempts    uint64                 `json:"write_attempts"`
	LastWriteAttempt time.Time              `json:"last_write_attempt"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Source           string                 `json:"source"` // "manual", "automatic", "computed"
}

// PropertyState tracks enhanced state and history of a property
type PropertyState struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Bytes       []byte      `json:"bytes"`
	Address     uint32      `json:"address"`
	Type        string      `json:"type"`
	Frozen      bool        `json:"frozen"`
	LastChanged time.Time   `json:"last_changed"`
	LastRead    time.Time   `json:"last_read"`
	LastWrite   time.Time   `json:"last_write"`

	// Enhanced tracking
	Performance      *PerformanceMetrics     `json:"performance"`
	ValidationErrors []ValidationError       `json:"validation_errors"`
	Events           []PropertyEvent         `json:"events"`
	Patterns         []PropertyChangePattern `json:"patterns"`

	// Value history
	ValueHistory   []ValueHistoryEntry `json:"value_history"`
	MaxHistorySize uint                `json:"max_history_size"`

	// Dependencies and relationships
	Dependencies []string `json:"dependencies"`
	Dependents   []string `json:"dependents"`

	// UI and display hints
	UIHints         map[string]interface{} `json:"ui_hints,omitempty"`
	DisplayPriority uint                   `json:"display_priority"`

	// Caching
	CachedValue interface{} `json:"cached_value,omitempty"`
	CacheExpiry *time.Time  `json:"cache_expiry,omitempty"`
	CacheValid  bool        `json:"cache_valid"`

	// Monitoring
	WatchEnabled   bool   `json:"watch_enabled"`
	WatchCondition string `json:"watch_condition,omitempty"`

	// Statistics
	Statistics *PropertyStatistics `json:"statistics,omitempty"`
}

// ValueHistoryEntry represents a historical value entry
type ValueHistoryEntry struct {
	Value     interface{}   `json:"value"`
	Timestamp time.Time     `json:"timestamp"`
	Source    string        `json:"source"`
	Duration  time.Duration `json:"duration,omitempty"`
}

// PropertyStatistics represents statistical analysis of property values
type PropertyStatistics struct {
	Min                 interface{}   `json:"min,omitempty"`
	Max                 interface{}   `json:"max,omitempty"`
	Mean                float64       `json:"mean,omitempty"`
	Median              float64       `json:"median,omitempty"`
	Mode                interface{}   `json:"mode,omitempty"`
	StandardDev         float64       `json:"standard_deviation,omitempty"`
	Variance            float64       `json:"variance,omitempty"`
	SampleCount         uint64        `json:"sample_count"`
	UniqueValues        uint64        `json:"unique_values"`
	LastCalculated      time.Time     `json:"last_calculated"`
	CalculationDuration time.Duration `json:"calculation_duration"`
}

// MemoryFragment represents enhanced memory fragment with analytics
type MemoryFragment struct {
	StartAddress     uint32                 `json:"start_address"`
	Data             []byte                 `json:"data"`
	LastUpdated      time.Time              `json:"last_updated"`
	AccessCount      uint64                 `json:"access_count"`
	AccessPattern    string                 `json:"access_pattern"` // "sequential", "random", "sparse"
	CompressionRatio float64                `json:"compression_ratio"`
	Checksum         uint32                 `json:"checksum"`
	Dirty            bool                   `json:"dirty"`
	Protected        bool                   `json:"protected"`
	Cached           bool                   `json:"cached"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryNamespace groups related memory fragments with enhanced organization
type MemoryNamespace struct {
	Name              string                     `json:"name"`
	Description       string                     `json:"description"`
	Fragments         map[uint32]*MemoryFragment `json:"fragments"`
	AccessPolicy      string                     `json:"access_policy"` // "read_only", "write_only", "read_write"
	CompressionLevel  uint                       `json:"compression_level"`
	EncryptionEnabled bool                       `json:"encryption_enabled"`
	Created           time.Time                  `json:"created"`
	LastAccessed      time.Time                  `json:"last_accessed"`
	TotalSize         uint64                     `json:"total_size"`
	UsedSize          uint64                     `json:"used_size"`
	Metadata          map[string]interface{}     `json:"metadata,omitempty"`
}

// PropertyCache represents a cached property value with smart invalidation
type PropertyCache struct {
	Value         interface{}            `json:"value"`
	CachedAt      time.Time              `json:"cached_at"`
	ExpiresAt     time.Time              `json:"expires_at"`
	HitCount      uint64                 `json:"hit_count"`
	Dependencies  []string               `json:"dependencies"`
	Invalidated   bool                   `json:"invalidated"`
	InvalidReason string                 `json:"invalid_reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ===== ENHANCED MEMORY MANAGER =====

// Manager handles enhanced memory storage and access with advanced features
type Manager struct {
	mu             sync.RWMutex
	blocks         map[uint32][]byte           // startAddress -> data
	frozenProps    map[uint32]*FrozenProperty  // address -> frozen property
	propertyStates map[string]*PropertyState   // property name -> state
	namespaces     map[string]*MemoryNamespace // namespace -> fragments
	propertyCache  map[string]*PropertyCache   // property name -> cache

	// Enhanced tracking
	changeListeners   []func(address uint32, oldData, newData []byte)
	propertyListeners []func(name string, event *PropertyEvent)
	validationEnabled bool
	debugMode         bool

	// Performance optimization
	batchingEnabled    bool
	batchOperations    chan BatchOperation
	compressionEnabled bool

	// Statistics and monitoring
	globalStats     *GlobalStatistics
	alertThresholds map[string]interface{}

	// Thread safety for different operations
	stateMu  sync.RWMutex // Separate mutex for property states
	cacheMu  sync.RWMutex // Separate mutex for cache operations
	frozenMu sync.RWMutex // Separate mutex for frozen properties
}

// BatchOperation represents a batch memory operation
type BatchOperation struct {
	Type       string                 `json:"type"` // "read", "write", "freeze", "validate"
	Operations []Operation            `json:"operations"`
	Atomic     bool                   `json:"atomic"`
	Priority   uint                   `json:"priority"`
	Timeout    time.Duration          `json:"timeout"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Response   chan BatchResult       `json:"-"`
}

// Operation represents a single operation in a batch
type Operation struct {
	Address  uint32                 `json:"address"`
	Data     []byte                 `json:"data,omitempty"`
	Property string                 `json:"property,omitempty"`
	Value    interface{}            `json:"value,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Success  bool                   `json:"success"`
	Results  []OperationResult      `json:"results"`
	Duration time.Duration          `json:"duration"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// OperationResult represents the result of a single operation
type OperationResult struct {
	Success  bool          `json:"success"`
	Data     []byte        `json:"data,omitempty"`
	Value    interface{}   `json:"value,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// GlobalStatistics represents global memory manager statistics
type GlobalStatistics struct {
	TotalReads            uint64        `json:"total_reads"`
	TotalWrites           uint64        `json:"total_writes"`
	TotalErrors           uint64        `json:"total_errors"`
	AvgOperationTime      time.Duration `json:"avg_operation_time"`
	PeakMemoryUsage       uint64        `json:"peak_memory_usage"`
	CurrentMemoryUsage    uint64        `json:"current_memory_usage"`
	CacheEfficiency       float64       `json:"cache_efficiency"`
	ValidationSuccessRate float64       `json:"validation_success_rate"`
	UptimeStart           time.Time     `json:"uptime_start"`
	LastReset             time.Time     `json:"last_reset"`
}

// NewManager creates a new enhanced memory manager
func NewManager() *Manager {
	manager := &Manager{
		blocks:             make(map[uint32][]byte),
		frozenProps:        make(map[uint32]*FrozenProperty),
		propertyStates:     make(map[string]*PropertyState),
		namespaces:         make(map[string]*MemoryNamespace),
		propertyCache:      make(map[string]*PropertyCache),
		changeListeners:    make([]func(address uint32, oldData, newData []byte), 0),
		propertyListeners:  make([]func(name string, event *PropertyEvent), 0),
		validationEnabled:  true,
		debugMode:          false,
		batchingEnabled:    true,
		batchOperations:    make(chan BatchOperation, 1000),
		compressionEnabled: false,
		alertThresholds:    make(map[string]interface{}),
		globalStats: &GlobalStatistics{
			UptimeStart: time.Now(),
			LastReset:   time.Now(),
		},
	}

	// Start batch processor
	go manager.processBatchOperations()

	return manager
}

// ===== CONFIGURATION METHODS =====

// SetDebugMode enables or disables debug logging
func (m *Manager) SetDebugMode(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugMode = enabled
}

// SetValidationEnabled enables or disables property validation
func (m *Manager) SetValidationEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validationEnabled = enabled
}

// SetCompressionEnabled enables or disables memory compression
func (m *Manager) SetCompressionEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.compressionEnabled = enabled
}

// SetBatchingEnabled enables or disables batch operations
func (m *Manager) SetBatchingEnabled(enabled bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.batchingEnabled = enabled
}

// debugLog logs only if debug mode is enabled
func (m *Manager) debugLog(format string, args ...interface{}) {
	if m.debugMode {
		log.Printf("[MemoryManager] "+format, args...)
	}
}

// ===== ENHANCED LISTENER SYSTEM =====

// AddChangeListener adds a callback for memory changes
func (m *Manager) AddChangeListener(listener func(address uint32, oldData, newData []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.changeListeners = append(m.changeListeners, listener)
}

// AddPropertyListener adds a callback for property events
func (m *Manager) AddPropertyListener(listener func(name string, event *PropertyEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.propertyListeners = append(m.propertyListeners, listener)
}

// ===== ENHANCED MEMORY OPERATIONS =====

// Update updates memory blocks with enhanced processing and frozen properties management
func (m *Manager) Update(memoryData map[uint32][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	updateStart := time.Now()
	m.globalStats.TotalReads++

	for address, data := range memoryData {
		oldData := make([]byte, len(data))
		if existingData, exists := m.blocks[address]; exists {
			copy(oldData, existingData)
		}

		// Create new data copy
		newData := make([]byte, len(data))
		copy(newData, data)

		// Update memory block
		m.blocks[address] = newData

		// Update namespace fragments
		m.updateFragment("default", address, newData)

		// Apply frozen properties
		m.applyFrozenProperties(address, newData)

		// Detect changes and notify listeners
		if !m.bytesEqual(oldData, newData) {
			// Notify change listeners without holding lock
			for _, listener := range m.changeListeners {
				go listener(address, oldData, newData)
			}
		}

		// Update global statistics
		m.globalStats.CurrentMemoryUsage = m.calculateCurrentMemoryUsage()
		if m.globalStats.CurrentMemoryUsage > m.globalStats.PeakMemoryUsage {
			m.globalStats.PeakMemoryUsage = m.globalStats.CurrentMemoryUsage
		}
	}

	// Update average operation time
	duration := time.Since(updateStart)
	m.updateAverageOperationTime(duration)

	m.debugLog("Updated %d memory blocks in %v", len(memoryData), duration)
}

// applyFrozenProperties writes back frozen values to memory with enhanced tracking
func (m *Manager) applyFrozenProperties(blockAddress uint32, blockData []byte) {
	m.frozenMu.RLock()
	frozenProps := make(map[uint32]*FrozenProperty)
	for addr, prop := range m.frozenProps {
		frozenProps[addr] = prop
	}
	m.frozenMu.RUnlock()

	for frozenAddr, frozen := range frozenProps {
		// Check if this frozen property is within the current block
		blockEnd := blockAddress + uint32(len(blockData)) - 1
		frozenEnd := frozenAddr + uint32(len(frozen.Data)) - 1

		if frozenAddr >= blockAddress && frozenEnd <= blockEnd {
			// Calculate offset within the block
			offset := frozenAddr - blockAddress

			// Check if the data has changed from frozen value
			changed := false
			for i, b := range frozen.Data {
				if offset+uint32(i) < uint32(len(blockData)) && blockData[offset+uint32(i)] != b {
					changed = true
					break
				}
			}

			// If changed, restore frozen value and track attempt
			if changed {
				m.frozenMu.Lock()
				if frozen, exists := m.frozenProps[frozenAddr]; exists {
					copy(blockData[offset:offset+uint32(len(frozen.Data))], frozen.Data)
					frozen.LastWritten = time.Now()
					frozen.WriteAttempts++
					frozen.LastWriteAttempt = time.Now()
				}
				m.frozenMu.Unlock()

				m.debugLog("Restored frozen property at 0x%X", frozenAddr)
			}
		}
	}
}

// updateFragment updates a memory fragment in a namespace with enhanced metadata
func (m *Manager) updateFragment(namespace string, address uint32, data []byte) {
	if m.namespaces[namespace] == nil {
		m.namespaces[namespace] = &MemoryNamespace{
			Name:         namespace,
			Description:  "Default namespace",
			Fragments:    make(map[uint32]*MemoryFragment),
			AccessPolicy: "read_write",
			Created:      time.Now(),
			Metadata:     make(map[string]interface{}),
		}
	}

	ns := m.namespaces[namespace]
	fragment := ns.Fragments[address]
	if fragment == nil {
		fragment = &MemoryFragment{
			StartAddress: address,
			Data:         make([]byte, len(data)),
			Cached:       true,
			Metadata:     make(map[string]interface{}),
		}
		ns.Fragments[address] = fragment
	}

	// Update fragment data and metadata
	copy(fragment.Data, data)
	fragment.LastUpdated = time.Now()
	fragment.AccessCount++
	fragment.Checksum = m.calculateChecksum(data)
	fragment.Dirty = true

	// Update namespace statistics
	ns.LastAccessed = time.Now()
	ns.TotalSize = m.calculateNamespaceSize(ns)
	ns.UsedSize = ns.TotalSize // Simplified calculation
}

// ===== ENHANCED FROZEN PROPERTIES =====

// FreezeProperty freezes a property with enhanced options
func (m *Manager) FreezeProperty(address uint32, data []byte) error {
	return m.FreezePropertyWithOptions(address, data, nil)
}

// FreezePropertyWithOptions freezes a property with advanced options
func (m *Manager) FreezePropertyWithOptions(address uint32, data []byte, options map[string]interface{}) error {
	m.frozenMu.Lock()
	defer m.frozenMu.Unlock()

	// Create enhanced frozen property
	frozen := &FrozenProperty{
		Address:       address,
		Data:          make([]byte, len(data)),
		FreezeTime:    time.Now(),
		LastWritten:   time.Now(),
		WriteAttempts: 0,
		Source:        "manual",
		Metadata:      make(map[string]interface{}),
	}
	copy(frozen.Data, data)

	// Apply options
	if options != nil {
		if expiry, exists := options["expiry"]; exists {
			if expiryTime, ok := expiry.(time.Time); ok {
				frozen.ExpiryTime = &expiryTime
			}
		}
		if condition, exists := options["condition"]; exists {
			if condStr, ok := condition.(string); ok {
				frozen.Condition = condStr
			}
		}
		if source, exists := options["source"]; exists {
			if srcStr, ok := source.(string); ok {
				frozen.Source = srcStr
			}
		}
		for key, value := range options {
			if key != "expiry" && key != "condition" && key != "source" {
				frozen.Metadata[key] = value
			}
		}
	}

	m.frozenProps[address] = frozen

	// Immediately apply the frozen value to current memory
	m.writeBytesInternal(address, data)

	m.debugLog("Frozen property at 0x%X with %d bytes", address, len(data))
	return nil
}

// UnfreezeProperty removes the freeze on a property at the given address
func (m *Manager) UnfreezeProperty(address uint32) error {
	m.frozenMu.Lock()
	defer m.frozenMu.Unlock()

	if _, exists := m.frozenProps[address]; exists {
		delete(m.frozenProps, address)
		m.debugLog("Unfrozen property at 0x%X", address)
		return nil
	}

	return fmt.Errorf("property at address 0x%X is not frozen", address)
}

// IsFrozen checks if a property at the given address is frozen
func (m *Manager) IsFrozen(address uint32) bool {
	m.frozenMu.RLock()
	defer m.frozenMu.RUnlock()
	_, exists := m.frozenProps[address]
	return exists
}

// GetFrozenProperties returns all currently frozen properties with enhanced metadata
func (m *Manager) GetFrozenProperties() map[uint32]*FrozenProperty {
	m.frozenMu.RLock()
	defer m.frozenMu.RUnlock()

	result := make(map[uint32]*FrozenProperty)
	for addr, prop := range m.frozenProps {
		// Create a deep copy
		result[addr] = &FrozenProperty{
			Address:          prop.Address,
			Data:             append([]byte(nil), prop.Data...),
			FreezeTime:       prop.FreezeTime,
			LastWritten:      prop.LastWritten,
			ExpiryTime:       prop.ExpiryTime,
			Condition:        prop.Condition,
			WriteAttempts:    prop.WriteAttempts,
			LastWriteAttempt: prop.LastWriteAttempt,
			Source:           prop.Source,
			Metadata:         make(map[string]interface{}),
		}
		for k, v := range prop.Metadata {
			result[addr].Metadata[k] = v
		}
	}
	return result
}

// CleanupExpiredFrozenProperties removes expired frozen properties
func (m *Manager) CleanupExpiredFrozenProperties() int {
	m.frozenMu.Lock()
	defer m.frozenMu.Unlock()

	now := time.Now()
	cleaned := 0

	for address, frozen := range m.frozenProps {
		if frozen.ExpiryTime != nil && now.After(*frozen.ExpiryTime) {
			delete(m.frozenProps, address)
			cleaned++
			m.debugLog("Cleaned up expired frozen property at 0x%X", address)
		}
	}

	return cleaned
}

// ===== ENHANCED PROPERTY STATE MANAGEMENT =====

// GetPropertyState returns enhanced property state with separate mutex
func (m *Manager) GetPropertyState(name string) *PropertyState {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()

	if state, exists := m.propertyStates[name]; exists {
		// Return a deep copy to prevent race conditions
		return m.copyPropertyState(state)
	}
	return nil
}

// GetAllPropertyStates returns all property states with separate mutex
func (m *Manager) GetAllPropertyStates() map[string]*PropertyState {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()

	result := make(map[string]*PropertyState)
	for name, state := range m.propertyStates {
		result[name] = m.copyPropertyState(state)
	}
	return result
}

// UpdatePropertyState updates property state with enhanced tracking using separate mutex
func (m *Manager) UpdatePropertyState(name string, value interface{}, bytes []byte, address uint32) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	updateStart := time.Now()

	state := m.propertyStates[name]
	if state == nil {
		state = &PropertyState{
			Name:           name,
			Address:        address,
			Performance:    &PerformanceMetrics{FirstAccess: time.Now()},
			Events:         make([]PropertyEvent, 0),
			ValueHistory:   make([]ValueHistoryEntry, 0),
			MaxHistorySize: 100, // Default history size
			Dependencies:   make([]string, 0),
			Dependents:     make([]string, 0),
			UIHints:        make(map[string]interface{}),
			Statistics:     &PropertyStatistics{},
		}
		m.propertyStates[name] = state
	}

	// Check if value actually changed
	changed := false
	oldValue := state.Value
	if !m.valuesEqual(state.Value, value) {
		changed = true
		state.Value = value
		state.LastChanged = time.Now()

		// Add to value history
		m.addValueToHistory(state, value, "update")

		// Update statistics
		m.updatePropertyStatistics(state, value)

		// Create change event
		event := &PropertyEvent{
			Type:      "read",
			Timestamp: time.Now(),
			Value:     value,
			OldValue:  oldValue,
			Duration:  time.Since(updateStart),
			Source:    "memory_update",
		}
		m.addPropertyEvent(state, event)

		// Notify property listeners
		for _, listener := range m.propertyListeners {
			go listener(name, event)
		}
	}

	// Update basic state
	state.Bytes = append([]byte(nil), bytes...)
	state.LastRead = time.Now()

	// Update performance metrics
	state.Performance.ReadCount++
	state.Performance.LastAccess = time.Now()
	state.Performance.TotalReadBytes += uint64(len(bytes))

	// Update read time statistics
	duration := time.Since(updateStart)
	state.Performance.LastReadTime = duration
	if duration > state.Performance.MaxReadTime {
		state.Performance.MaxReadTime = duration
	}

	// Update average read time
	totalReadTime := time.Duration(state.Performance.ReadCount) * state.Performance.AvgReadTime
	state.Performance.AvgReadTime = (totalReadTime + duration) / time.Duration(state.Performance.ReadCount)

	// Check frozen status without holding main mutex
	state.Frozen = m.isFrozenNoLock(address)

	if changed {
		state.Performance.WriteCount++
	}

	m.debugLog("Updated property state for %s: value=%v, changed=%t", name, value, changed)
}

// addValueToHistory adds a value to property history with size management
func (m *Manager) addValueToHistory(state *PropertyState, value interface{}, source string) {
	entry := ValueHistoryEntry{
		Value:     value,
		Timestamp: time.Now(),
		Source:    source,
	}

	state.ValueHistory = append(state.ValueHistory, entry)

	// Maintain history size limit
	if len(state.ValueHistory) > int(state.MaxHistorySize) {
		state.ValueHistory = state.ValueHistory[1:]
	}
}

// addPropertyEvent adds an event to property event history
func (m *Manager) addPropertyEvent(state *PropertyState, event *PropertyEvent) {
	state.Events = append(state.Events, *event)

	// Keep only last 50 events to prevent memory bloat
	if len(state.Events) > 50 {
		state.Events = state.Events[1:]
	}
}

// updatePropertyStatistics updates statistical analysis of property values
func (m *Manager) updatePropertyStatistics(state *PropertyState, value interface{}) {
	if state.Statistics == nil {
		state.Statistics = &PropertyStatistics{}
	}

	stats := state.Statistics
	stats.SampleCount++
	stats.LastCalculated = time.Now()

	// Convert value to float64 for statistical calculations
	if numValue, ok := m.convertToFloat64(value); ok {
		// Update min/max
		if stats.SampleCount == 1 {
			stats.Min = numValue
			stats.Max = numValue
			stats.Mean = numValue
		} else {
			if minFloat, ok := m.convertToFloat64(stats.Min); ok && numValue < minFloat {
				stats.Min = numValue
			}
			if maxFloat, ok := m.convertToFloat64(stats.Max); ok && numValue > maxFloat {
				stats.Max = numValue
			}

			// Update running mean
			stats.Mean = (stats.Mean*float64(stats.SampleCount-1) + numValue) / float64(stats.SampleCount)
		}

		// Calculate variance and standard deviation (simplified online algorithm)
		if stats.SampleCount > 1 {
			delta := numValue - stats.Mean
			stats.Variance += delta * delta / float64(stats.SampleCount)
			stats.StandardDev = math.Sqrt(stats.Variance)
		}
	}
}

// ===== ENHANCED CACHING SYSTEM =====

// GetCachedProperty gets a cached property value
func (m *Manager) GetCachedProperty(name string) (interface{}, bool) {
	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()

	if cache, exists := m.propertyCache[name]; exists {
		if !cache.Invalidated && time.Now().Before(cache.ExpiresAt) {
			cache.HitCount++
			m.debugLog("Cache hit for property %s", name)
			return cache.Value, true
		} else {
			// Cache expired or invalidated
			delete(m.propertyCache, name)
			m.debugLog("Cache miss for property %s (expired/invalidated)", name)
		}
	}

	return nil, false
}

// SetCachedProperty sets a cached property value
func (m *Manager) SetCachedProperty(name string, value interface{}, duration time.Duration, dependencies []string) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	cache := &PropertyCache{
		Value:        value,
		CachedAt:     time.Now(),
		ExpiresAt:    time.Now().Add(duration),
		HitCount:     0,
		Dependencies: append([]string(nil), dependencies...),
		Invalidated:  false,
		Metadata:     make(map[string]interface{}),
	}

	m.propertyCache[name] = cache
	m.debugLog("Cached property %s for %v", name, duration)
}

// InvalidateCache invalidates cached property values
func (m *Manager) InvalidateCache(pattern string) int {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	invalidated := 0
	for name, cache := range m.propertyCache {
		if m.matchesPattern(name, pattern) {
			cache.Invalidated = true
			cache.InvalidReason = "manual_invalidation"
			invalidated++
		}
	}

	m.debugLog("Invalidated %d cached properties matching pattern %s", invalidated, pattern)
	return invalidated
}

// ===== ENHANCED READ OPERATIONS =====

// ReadBytes reads bytes from a specific address with enhanced error handling and performance tracking
func (m *Manager) ReadBytes(address uint32, length uint32) ([]byte, error) {
	readStart := time.Now()
	m.debugLog("📖 ReadBytes: address=0x%X, length=%d", address, length)

	// Update global statistics
	m.mu.Lock()
	m.globalStats.TotalReads++
	m.mu.Unlock()

	// Create a local copy to work with (prevents holding lock too long)
	var targetBlock []byte
	var blockStart uint32
	var found bool

	// Very fast critical section - just find and copy the block
	m.mu.RLock()
	for bs, bd := range m.blocks {
		blockEnd := bs + uint32(len(bd)) - 1
		if address >= bs && address <= blockEnd {
			blockStart = bs
			targetBlock = make([]byte, len(bd))
			copy(targetBlock, bd)
			found = true
			break
		}
	}
	m.mu.RUnlock()

	if !found {
		m.mu.Lock()
		m.globalStats.TotalErrors++
		m.mu.Unlock()
		return nil, fmt.Errorf("address 0x%X not found in loaded memory", address)
	}

	// Work with copied data (no locks needed)
	offset := address - blockStart
	if offset+length > uint32(len(targetBlock)) {
		m.mu.Lock()
		m.globalStats.TotalErrors++
		m.mu.Unlock()
		return nil, fmt.Errorf("not enough data at address 0x%X (requested %d bytes, available %d)",
			address, length, uint32(len(targetBlock))-offset)
	}

	result := make([]byte, length)
	copy(result, targetBlock[offset:offset+length])

	// Update performance statistics
	duration := time.Since(readStart)
	m.updateAverageOperationTime(duration)

	m.debugLog("✅ Read %d bytes from 0x%X in %v", length, address, duration)
	return result, nil
}

// ===== BATCH OPERATIONS =====

// SubmitBatchOperation submits a batch operation for processing
func (m *Manager) SubmitBatchOperation(operation BatchOperation) <-chan BatchResult {
	if !m.batchingEnabled {
		// Process immediately if batching is disabled
		result := m.processBatchOperationSync(operation)
		resultChan := make(chan BatchResult, 1)
		resultChan <- result
		close(resultChan)
		return resultChan
	}

	// Submit to batch processor
	select {
	case m.batchOperations <- operation:
		return operation.Response
	default:
		// Channel full, process synchronously
		result := m.processBatchOperationSync(operation)
		resultChan := make(chan BatchResult, 1)
		resultChan <- result
		close(resultChan)
		return resultChan
	}
}

// processBatchOperations processes batch operations asynchronously
func (m *Manager) processBatchOperations() {
	for operation := range m.batchOperations {
		result := m.processBatchOperationSync(operation)

		select {
		case operation.Response <- result:
		case <-time.After(operation.Timeout):
			// Timeout sending result
			m.debugLog("Timeout sending batch operation result")
		}
	}
}

// processBatchOperationSync processes a batch operation synchronously
func (m *Manager) processBatchOperationSync(operation BatchOperation) BatchResult {
	start := time.Now()
	results := make([]OperationResult, len(operation.Operations))

	for i, op := range operation.Operations {
		opStart := time.Now()
		var opResult OperationResult

		switch operation.Type {
		case "read":
			data, err := m.ReadBytes(op.Address, uint32(len(op.Data)))
			if err != nil {
				opResult = OperationResult{
					Success:  false,
					Error:    err.Error(),
					Duration: time.Since(opStart),
				}
			} else {
				opResult = OperationResult{
					Success:  true,
					Data:     data,
					Duration: time.Since(opStart),
				}
			}
		case "write":
			data := m.WriteBytes(op.Address, op.Data)
			opResult = OperationResult{
				Success:  true,
				Data:     data,
				Duration: time.Since(opStart),
			}
		default:
			opResult = OperationResult{
				Success:  false,
				Error:    fmt.Sprintf("unsupported operation type: %s", operation.Type),
				Duration: time.Since(opStart),
			}
		}

		results[i] = opResult

		// If atomic and any operation fails, abort
		if operation.Atomic && !opResult.Success {
			break
		}
	}

	// Determine overall success
	success := true
	for _, result := range results {
		if !result.Success {
			success = false
			break
		}
	}

	return BatchResult{
		Success:  success,
		Results:  results,
		Duration: time.Since(start),
		Metadata: operation.Metadata,
	}
}

// ===== ENHANCED UTILITY METHODS =====

// isFrozenNoLock checks frozen status with minimal locking
func (m *Manager) isFrozenNoLock(address uint32) bool {
	m.frozenMu.RLock()
	defer m.frozenMu.RUnlock()
	_, exists := m.frozenProps[address]
	return exists
}

// copyPropertyState creates a deep copy of property state
func (m *Manager) copyPropertyState(state *PropertyState) *PropertyState {
	if state == nil {
		return nil
	}

	copy := &PropertyState{
		Name:            state.Name,
		Value:           state.Value,
		Bytes:           append([]byte(nil), state.Bytes...),
		Address:         state.Address,
		Type:            state.Type,
		Frozen:          state.Frozen,
		LastChanged:     state.LastChanged,
		LastRead:        state.LastRead,
		LastWrite:       state.LastWrite,
		MaxHistorySize:  state.MaxHistorySize,
		Dependencies:    append([]string(nil), state.Dependencies...),
		Dependents:      append([]string(nil), state.Dependents...),
		DisplayPriority: state.DisplayPriority,
		CachedValue:     state.CachedValue,
		CacheExpiry:     state.CacheExpiry,
		CacheValid:      state.CacheValid,
		WatchEnabled:    state.WatchEnabled,
		WatchCondition:  state.WatchCondition,
	}

	// Copy performance metrics
	if state.Performance != nil {
		copy.Performance = &PerformanceMetrics{
			ReadCount:       state.Performance.ReadCount,
			WriteCount:      state.Performance.WriteCount,
			ErrorCount:      state.Performance.ErrorCount,
			AvgReadTime:     state.Performance.AvgReadTime,
			AvgWriteTime:    state.Performance.AvgWriteTime,
			MaxReadTime:     state.Performance.MaxReadTime,
			MaxWriteTime:    state.Performance.MaxWriteTime,
			LastReadTime:    state.Performance.LastReadTime,
			LastWriteTime:   state.Performance.LastWriteTime,
			CacheHits:       state.Performance.CacheHits,
			CacheMisses:     state.Performance.CacheMisses,
			CacheHitRatio:   state.Performance.CacheHitRatio,
			LastCacheTime:   state.Performance.LastCacheTime,
			TotalReadBytes:  state.Performance.TotalReadBytes,
			TotalWriteBytes: state.Performance.TotalWriteBytes,
			FirstAccess:     state.Performance.FirstAccess,
			LastAccess:      state.Performance.LastAccess,
		}
	}

	// Copy validation errors
	copy.ValidationErrors = append([]ValidationError(nil), state.ValidationErrors...)

	// Copy events
	copy.Events = append([]PropertyEvent(nil), state.Events...)

	// Copy patterns
	copy.Patterns = append([]PropertyChangePattern(nil), state.Patterns...)

	// Copy value history
	copy.ValueHistory = append([]ValueHistoryEntry(nil), state.ValueHistory...)

	// Copy UI hints
	copy.UIHints = make(map[string]interface{})
	for k, v := range state.UIHints {
		copy.UIHints[k] = v
	}

	// Copy statistics
	if state.Statistics != nil {
		copy.Statistics = &PropertyStatistics{
			Min:                 state.Statistics.Min,
			Max:                 state.Statistics.Max,
			Mean:                state.Statistics.Mean,
			Median:              state.Statistics.Median,
			Mode:                state.Statistics.Mode,
			StandardDev:         state.Statistics.StandardDev,
			Variance:            state.Statistics.Variance,
			SampleCount:         state.Statistics.SampleCount,
			UniqueValues:        state.Statistics.UniqueValues,
			LastCalculated:      state.Statistics.LastCalculated,
			CalculationDuration: state.Statistics.CalculationDuration,
		}
	}

	return copy
}

// bytesEqual compares two byte slices for equality
func (m *Manager) bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// valuesEqual compares two values for equality
func (m *Manager) valuesEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// convertToFloat64 converts various numeric types to float64
func (m *Manager) convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

// matchesPattern checks if a string matches a simple pattern (supports *)
func (m *Manager) matchesPattern(str, pattern string) bool {
	// Simple pattern matching - support for * wildcard
	if pattern == "*" {
		return true
	}
	if pattern == str {
		return true
	}
	// TODO: Implement more sophisticated pattern matching if needed
	return false
}

// calculateChecksum calculates a simple checksum for data
func (m *Manager) calculateChecksum(data []byte) uint32 {
	var sum uint32
	for _, b := range data {
		sum += uint32(b)
	}
	return sum
}

// calculateCurrentMemoryUsage calculates current memory usage
func (m *Manager) calculateCurrentMemoryUsage() uint64 {
	var total uint64
	for _, data := range m.blocks {
		total += uint64(len(data))
	}
	return total
}

// calculateNamespaceSize calculates the total size of a namespace
func (m *Manager) calculateNamespaceSize(ns *MemoryNamespace) uint64 {
	var total uint64
	for _, fragment := range ns.Fragments {
		total += uint64(len(fragment.Data))
	}
	return total
}

// updateAverageOperationTime updates the global average operation time
func (m *Manager) updateAverageOperationTime(duration time.Duration) {
	totalOps := m.globalStats.TotalReads + m.globalStats.TotalWrites
	if totalOps > 0 {
		totalTime := time.Duration(totalOps-1) * m.globalStats.AvgOperationTime
		m.globalStats.AvgOperationTime = (totalTime + duration) / time.Duration(totalOps)
	} else {
		m.globalStats.AvgOperationTime = duration
	}
}

// ===== ORIGINAL COMPATIBILITY METHODS =====

// ReadUint8 reads a single byte as uint8
func (m *Manager) ReadUint8(address uint32) (uint8, error) {
	data, err := m.ReadBytes(address, 1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

// ReadUint16 reads 2 bytes as uint16 with specified endianness
func (m *Manager) ReadUint16(address uint32, littleEndian bool) (uint16, error) {
	data, err := m.ReadBytes(address, 2)
	if err != nil {
		return 0, err
	}

	if littleEndian {
		return binary.LittleEndian.Uint16(data), nil
	}
	return binary.BigEndian.Uint16(data), nil
}

// ReadUint32 reads 4 bytes as uint32 with specified endianness
func (m *Manager) ReadUint32(address uint32, littleEndian bool) (uint32, error) {
	data, err := m.ReadBytes(address, 4)
	if err != nil {
		return 0, err
	}

	if littleEndian {
		return binary.LittleEndian.Uint32(data), nil
	}
	return binary.BigEndian.Uint32(data), nil
}

// ReadInt8 reads a single byte as int8
func (m *Manager) ReadInt8(address uint32) (int8, error) {
	val, err := m.ReadUint8(address)
	return int8(val), err
}

// ReadInt16 reads 2 bytes as int16 with specified endianness
func (m *Manager) ReadInt16(address uint32, littleEndian bool) (int16, error) {
	val, err := m.ReadUint16(address, littleEndian)
	return int16(val), err
}

// ReadInt32 reads 4 bytes as int32 with specified endianness
func (m *Manager) ReadInt32(address uint32, littleEndian bool) (int32, error) {
	val, err := m.ReadUint32(address, littleEndian)
	return int32(val), err
}

// ReadFloat32 reads 4 bytes as float32 with specified endianness
func (m *Manager) ReadFloat32(address uint32, littleEndian bool) (float32, error) {
	data, err := m.ReadBytes(address, 4)
	if err != nil {
		return 0, err
	}
	if littleEndian {
		return math.Float32frombits(binary.LittleEndian.Uint32(data)), nil
	}
	return math.Float32frombits(binary.BigEndian.Uint32(data)), nil
}

// ReadFloat64 reads 8 bytes as float64 with specified endianness
func (m *Manager) ReadFloat64(address uint32, littleEndian bool) (float64, error) {
	data, err := m.ReadBytes(address, 8)
	if err != nil {
		return 0, err
	}
	if littleEndian {
		return math.Float64frombits(binary.LittleEndian.Uint64(data)), nil
	}
	return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
}

// ReadBool reads a single byte as boolean (0 = false, anything else = true)
func (m *Manager) ReadBool(address uint32) (bool, error) {
	val, err := m.ReadUint8(address)
	return val != 0, err
}

// ReadBCD reads Binary Coded Decimal values (used in Pokemon games)
func (m *Manager) ReadBCD(address uint32, length uint32) (uint32, error) {
	data, err := m.ReadBytes(address, length)
	if err != nil {
		return 0, err
	}

	result := uint32(0)
	for _, bcd := range data {
		result *= 100
		result += uint32(10*(bcd>>4) + (bcd & 0x0F))
	}
	return result, nil
}

// ReadString reads a null-terminated string with character mapping
func (m *Manager) ReadString(address uint32, maxLength uint32, charMap map[uint8]string) (string, error) {
	data, err := m.ReadBytes(address, maxLength)
	if err != nil {
		return "", err
	}

	var result string
	for _, b := range data {
		if b == 0 || b == 0xFF { // null terminator or Pokemon string terminator
			break
		}

		if charMap != nil {
			if char, exists := charMap[b]; exists {
				result += char
			} else {
				// For Pokemon strings, skip unknown characters rather than showing hex
				if b >= 0x80 && b <= 0xF6 {
					result += "?"
				}
			}
		} else {
			// No character map, use raw bytes for printable ASCII
			if b >= 0x20 && b <= 0x7E {
				result += string(b)
			}
		}
	}

	return result, nil
}

// ReadBitfield reads bytes and returns individual bit values
func (m *Manager) ReadBitfield(address uint32, length uint32) ([]bool, error) {
	data, err := m.ReadBytes(address, length)
	if err != nil {
		return nil, err
	}

	bits := make([]bool, length*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b & (1 << j)) != 0
		}
	}

	return bits, nil
}

// WriteBytes writes bytes to memory and returns the data for driver to write
func (m *Manager) WriteBytes(address uint32, data []byte) []byte {
	return m.writeBytesInternal(address, data)
}

// writeBytesInternal is the internal implementation of WriteBytes
func (m *Manager) writeBytesInternal(address uint32, data []byte) []byte {
	writeStart := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update global statistics
	m.globalStats.TotalWrites++

	// Find the block containing this address and update it
	for blockStart, blockData := range m.blocks {
		blockEnd := blockStart + uint32(len(blockData)) - 1

		if address >= blockStart && address <= blockEnd {
			offset := address - blockStart

			// Update our internal copy
			if offset+uint32(len(data)) <= uint32(len(blockData)) {
				copy(blockData[offset:], data)
			}
			break
		}
	}

	// Update performance statistics
	duration := time.Since(writeStart)
	m.updateAverageOperationTime(duration)

	// Return copy for driver to write
	result := make([]byte, len(data))
	copy(result, data)

	m.debugLog("✍️ Wrote %d bytes to 0x%X in %v", len(data), address, duration)
	return result
}

// WriteUint8 writes a uint8 value
func (m *Manager) WriteUint8(address uint32, value uint8) []byte {
	return m.WriteBytes(address, []byte{value})
}

// WriteUint16 writes a uint16 value with specified endianness
func (m *Manager) WriteUint16(address uint32, value uint16, littleEndian bool) []byte {
	data := make([]byte, 2)
	if littleEndian {
		binary.LittleEndian.PutUint16(data, value)
	} else {
		binary.BigEndian.PutUint16(data, value)
	}
	return m.WriteBytes(address, data)
}

// WriteUint32 writes a uint32 value with specified endianness
func (m *Manager) WriteUint32(address uint32, value uint32, littleEndian bool) []byte {
	data := make([]byte, 4)
	if littleEndian {
		binary.LittleEndian.PutUint32(data, value)
	} else {
		binary.BigEndian.PutUint32(data, value)
	}
	return m.WriteBytes(address, data)
}

// WriteBool writes a boolean as a byte (false = 0, true = 1)
func (m *Manager) WriteBool(address uint32, value bool) []byte {
	var b byte
	if value {
		b = 1
	}
	return m.WriteUint8(address, b)
}

// GetLoadedBlocks returns information about currently loaded memory blocks
func (m *Manager) GetLoadedBlocks() map[uint32]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uint32]int)
	for address, data := range m.blocks {
		result[address] = len(data)
	}
	return result
}

// GetMemoryNamespaces returns all memory namespaces with enhanced metadata
func (m *Manager) GetMemoryNamespaces() map[string]*MemoryNamespace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*MemoryNamespace)
	for name, namespace := range m.namespaces {
		result[name] = &MemoryNamespace{
			Name:              namespace.Name,
			Description:       namespace.Description,
			AccessPolicy:      namespace.AccessPolicy,
			CompressionLevel:  namespace.CompressionLevel,
			EncryptionEnabled: namespace.EncryptionEnabled,
			Created:           namespace.Created,
			LastAccessed:      namespace.LastAccessed,
			TotalSize:         namespace.TotalSize,
			UsedSize:          namespace.UsedSize,
			Fragments:         make(map[uint32]*MemoryFragment),
			Metadata:          make(map[string]interface{}),
		}
		for addr, fragment := range namespace.Fragments {
			result[name].Fragments[addr] = &MemoryFragment{
				StartAddress:     fragment.StartAddress,
				Data:             append([]byte(nil), fragment.Data...),
				LastUpdated:      fragment.LastUpdated,
				AccessCount:      fragment.AccessCount,
				AccessPattern:    fragment.AccessPattern,
				CompressionRatio: fragment.CompressionRatio,
				Checksum:         fragment.Checksum,
				Dirty:            fragment.Dirty,
				Protected:        fragment.Protected,
				Cached:           fragment.Cached,
				Metadata:         make(map[string]interface{}),
			}
			for k, v := range fragment.Metadata {
				result[name].Fragments[addr].Metadata[k] = v
			}
		}
		for k, v := range namespace.Metadata {
			result[name].Metadata[k] = v
		}
	}
	return result
}

// GetGlobalStatistics returns global memory manager statistics
func (m *Manager) GetGlobalStatistics() *GlobalStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate efficiency metrics
	totalOps := m.globalStats.TotalReads + m.globalStats.TotalWrites
	if totalOps > 0 {
		m.globalStats.ValidationSuccessRate = float64(totalOps-m.globalStats.TotalErrors) / float64(totalOps) * 100
	}

	// Calculate cache efficiency
	m.cacheMu.RLock()
	totalCacheOps := uint64(0)
	totalCacheHits := uint64(0)
	for _, cache := range m.propertyCache {
		totalCacheOps += cache.HitCount + 1 // +1 for the miss that created the cache
		totalCacheHits += cache.HitCount
	}
	m.cacheMu.RUnlock()

	if totalCacheOps > 0 {
		m.globalStats.CacheEfficiency = float64(totalCacheHits) / float64(totalCacheOps) * 100
	}

	// Return copy to prevent modification
	return &GlobalStatistics{
		TotalReads:            m.globalStats.TotalReads,
		TotalWrites:           m.globalStats.TotalWrites,
		TotalErrors:           m.globalStats.TotalErrors,
		AvgOperationTime:      m.globalStats.AvgOperationTime,
		PeakMemoryUsage:       m.globalStats.PeakMemoryUsage,
		CurrentMemoryUsage:    m.globalStats.CurrentMemoryUsage,
		CacheEfficiency:       m.globalStats.CacheEfficiency,
		ValidationSuccessRate: m.globalStats.ValidationSuccessRate,
		UptimeStart:           m.globalStats.UptimeStart,
		LastReset:             m.globalStats.LastReset,
	}
}

// ResetStatistics resets global statistics
func (m *Manager) ResetStatistics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.globalStats = &GlobalStatistics{
		UptimeStart: m.globalStats.UptimeStart, // Keep original uptime start
		LastReset:   time.Now(),
	}

	m.debugLog("🔄 Global statistics reset")
}
