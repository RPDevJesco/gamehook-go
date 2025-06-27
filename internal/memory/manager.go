package memory

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// FrozenProperty represents a property that should maintain a constant value
type FrozenProperty struct {
	Address     uint32
	Data        []byte
	LastWritten time.Time
	FreezeTime  time.Time
}

// PropertyState tracks the state and history of a property
type PropertyState struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value"`
	Bytes       []byte      `json:"bytes"`
	Address     uint32      `json:"address"`
	Frozen      bool        `json:"frozen"`
	LastChanged time.Time   `json:"last_changed"`
	LastRead    time.Time   `json:"last_read"`
	ReadCount   uint64      `json:"read_count"`
	WriteCount  uint64      `json:"write_count"`
}

// MemoryFragment represents a contiguous piece of memory with metadata
type MemoryFragment struct {
	StartAddress uint32
	Data         []byte
	LastUpdated  time.Time
	AccessCount  uint64
}

// MemoryNamespace groups related memory fragments
type MemoryNamespace struct {
	Name      string
	Fragments map[uint32]*MemoryFragment
}

// Manager handles memory storage and access with enhanced features
type Manager struct {
	mu              sync.RWMutex
	blocks          map[uint32][]byte           // startAddress -> data
	frozenProps     map[uint32]*FrozenProperty  // address -> frozen property
	propertyStates  map[string]*PropertyState   // property name -> state
	namespaces      map[string]*MemoryNamespace // namespace -> fragments
	changeListeners []func(address uint32, oldData, newData []byte)
	debugMode       bool // Add debug mode flag
	// Add separate mutex for property states to prevent recursive locking
	stateMu sync.RWMutex
}

// NewManager creates a new enhanced memory manager
func NewManager() *Manager {
	return &Manager{
		blocks:          make(map[uint32][]byte),
		frozenProps:     make(map[uint32]*FrozenProperty),
		propertyStates:  make(map[string]*PropertyState),
		namespaces:      make(map[string]*MemoryNamespace),
		changeListeners: make([]func(address uint32, oldData, newData []byte), 0),
		debugMode:       false, // Disable debug by default
	}
}

// SetDebugMode enables or disables debug logging
func (m *Manager) SetDebugMode(enabled bool) {
	m.debugMode = enabled
}

// debugLog logs only if debug mode is enabled
func (m *Manager) debugLog(format string, args ...interface{}) {
	if m.debugMode {
		log.Printf(format, args...)
	}
}

// AddChangeListener adds a callback for memory changes
func (m *Manager) AddChangeListener(listener func(address uint32, oldData, newData []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.changeListeners = append(m.changeListeners, listener)
}

// Update updates memory blocks with new data and manages frozen properties
func (m *Manager) Update(memoryData map[uint32][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for address, data := range memoryData {
		oldData := make([]byte, len(data))
		if existingData, exists := m.blocks[address]; exists {
			copy(oldData, existingData)
		}

		m.blocks[address] = make([]byte, len(data))
		copy(m.blocks[address], data)

		m.updateFragment("default", address, data)
		m.applyFrozenProperties(address, data)

		// Notify listeners without holding lock
		for _, listener := range m.changeListeners {
			go listener(address, oldData, data)
		}
	}
}

// applyFrozenProperties writes back frozen values to memory
func (m *Manager) applyFrozenProperties(blockAddress uint32, blockData []byte) {
	for frozenAddr, frozen := range m.frozenProps {
		// Check if this frozen property is within the current block
		blockEnd := blockAddress + uint32(len(blockData)) - 1
		frozenEnd := frozenAddr + uint32(len(frozen.Data)) - 1

		if frozenAddr >= blockAddress && frozenEnd <= blockEnd {
			// Calculate offset within the block
			offset := frozenAddr - blockAddress

			// Check if the data has changed from frozen value
			changed := false
			for i, b := range frozen.Data {
				if blockData[offset+uint32(i)] != b {
					changed = true
					break
				}
			}

			// If changed, restore frozen value
			if changed {
				copy(blockData[offset:offset+uint32(len(frozen.Data))], frozen.Data)
				frozen.LastWritten = time.Now()
			}
		}
	}
}

// updateFragment updates a memory fragment in a namespace
func (m *Manager) updateFragment(namespace string, address uint32, data []byte) {
	if m.namespaces[namespace] == nil {
		m.namespaces[namespace] = &MemoryNamespace{
			Name:      namespace,
			Fragments: make(map[uint32]*MemoryFragment),
		}
	}

	fragment := m.namespaces[namespace].Fragments[address]
	if fragment == nil {
		fragment = &MemoryFragment{
			StartAddress: address,
			Data:         make([]byte, len(data)),
		}
		m.namespaces[namespace].Fragments[address] = fragment
	}

	copy(fragment.Data, data)
	fragment.LastUpdated = time.Now()
	fragment.AccessCount++
}

// FreezeProperty freezes a property at a specific address with given data
func (m *Manager) FreezeProperty(address uint32, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create frozen property
	frozen := &FrozenProperty{
		Address:     address,
		Data:        make([]byte, len(data)),
		FreezeTime:  time.Now(),
		LastWritten: time.Now(),
	}
	copy(frozen.Data, data)

	m.frozenProps[address] = frozen

	// Immediately apply the frozen value to current memory
	m.writeBytesInternal(address, data)

	return nil
}

// UnfreezeProperty removes the freeze on a property at the given address
func (m *Manager) UnfreezeProperty(address uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.frozenProps, address)
	return nil
}

// IsFrozen checks if a property at the given address is frozen
func (m *Manager) IsFrozen(address uint32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.frozenProps[address]
	return exists
}

// GetFrozenProperties returns all currently frozen properties
func (m *Manager) GetFrozenProperties() map[uint32]*FrozenProperty {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uint32]*FrozenProperty)
	for addr, prop := range m.frozenProps {
		// Make a copy
		result[addr] = &FrozenProperty{
			Address:     prop.Address,
			Data:        make([]byte, len(prop.Data)),
			FreezeTime:  prop.FreezeTime,
			LastWritten: prop.LastWritten,
		}
		copy(result[addr].Data, prop.Data)
	}
	return result
}

// isFrozenNoLock checks frozen status with minimal locking
func (m *Manager) isFrozenNoLock(address uint32) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.frozenProps[address]
	return exists
}

// GetPropertyState uses separate mutex
func (m *Manager) GetPropertyState(name string) *PropertyState {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()

	if state, exists := m.propertyStates[name]; exists {
		return &PropertyState{
			Name:        state.Name,
			Value:       state.Value,
			Bytes:       append([]byte(nil), state.Bytes...),
			Address:     state.Address,
			Frozen:      state.Frozen,
			LastChanged: state.LastChanged,
			LastRead:    state.LastRead,
			ReadCount:   state.ReadCount,
			WriteCount:  state.WriteCount,
		}
	}
	return nil
}

// GetAllPropertyStates uses separate mutex
func (m *Manager) GetAllPropertyStates() map[string]*PropertyState {
	m.stateMu.RLock()
	defer m.stateMu.RUnlock()

	result := make(map[string]*PropertyState)
	for name, state := range m.propertyStates {
		result[name] = &PropertyState{
			Name:        state.Name,
			Value:       state.Value,
			Bytes:       append([]byte(nil), state.Bytes...),
			Address:     state.Address,
			Frozen:      state.Frozen,
			LastChanged: state.LastChanged,
			LastRead:    state.LastRead,
			ReadCount:   state.ReadCount,
			WriteCount:  state.WriteCount,
		}
	}
	return result
}

// ReadBytes reads bytes from a specific address with optimized locking
func (m *Manager) ReadBytes(address uint32, length uint32) ([]byte, error) {
	m.debugLog("🔍 ReadBytes: address=0x%X, length=%d", address, length)

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
		return nil, fmt.Errorf("address 0x%X not found in loaded memory", address)
	}

	// Work with copied data (no locks needed)
	offset := address - blockStart
	if offset+length > uint32(len(targetBlock)) {
		return nil, fmt.Errorf("not enough data at address 0x%X", address)
	}

	result := make([]byte, length)
	copy(result, targetBlock[offset:offset+length])
	return result, nil
}

// UpdatePropertyState uses separate mutex to prevent deadlock
func (m *Manager) UpdatePropertyState(name string, value interface{}, bytes []byte, address uint32) {
	m.stateMu.Lock()
	defer m.stateMu.Unlock()

	state := m.propertyStates[name]
	if state == nil {
		state = &PropertyState{
			Name:    name,
			Address: address,
		}
		m.propertyStates[name] = state
	}

	// Check if value actually changed
	changed := false
	if state.Value != value {
		changed = true
		state.Value = value
		state.LastChanged = time.Now()
	}

	state.Bytes = make([]byte, len(bytes))
	copy(state.Bytes, bytes)
	state.LastRead = time.Now()
	state.ReadCount++

	// Check frozen status without holding main mutex
	state.Frozen = m.isFrozenNoLock(address)

	if changed {
		state.WriteCount++
	}
}

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

		if char, exists := charMap[b]; exists {
			result += char
		} else {
			// For Pokemon strings, skip unknown characters rather than showing hex
			if b >= 0x80 && b <= 0xF6 {
				result += "?"
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
	m.mu.Lock()
	defer m.mu.Unlock()

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

	// Return copy for driver to write
	result := make([]byte, len(data))
	copy(result, data)
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

// GetMemoryNamespaces returns all memory namespaces
func (m *Manager) GetMemoryNamespaces() map[string]*MemoryNamespace {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*MemoryNamespace)
	for name, namespace := range m.namespaces {
		result[name] = &MemoryNamespace{
			Name:      namespace.Name,
			Fragments: make(map[uint32]*MemoryFragment),
		}
		for addr, fragment := range namespace.Fragments {
			result[name].Fragments[addr] = &MemoryFragment{
				StartAddress: fragment.StartAddress,
				Data:         append([]byte(nil), fragment.Data...),
				LastUpdated:  fragment.LastUpdated,
				AccessCount:  fragment.AccessCount,
			}
		}
	}
	return result
}
