package memory

import (
	"encoding/binary"
	"fmt"
	"sync"
)

// Manager handles memory storage and access
type Manager struct {
	mu     sync.RWMutex
	blocks map[uint32][]byte // startAddress -> data
}

// NewManager creates a new memory manager
func NewManager() *Manager {
	return &Manager{
		blocks: make(map[uint32][]byte),
	}
}

// Update updates memory blocks with new data
func (m *Manager) Update(memoryData map[uint32][]byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for address, data := range memoryData {
		// Make a copy to avoid external modifications
		m.blocks[address] = make([]byte, len(data))
		copy(m.blocks[address], data)
	}
}

// ReadBytes reads bytes from a specific address
func (m *Manager) ReadBytes(address uint32, length uint32) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Find the block containing this address
	for blockStart, blockData := range m.blocks {
		blockEnd := blockStart + uint32(len(blockData)) - 1

		if address >= blockStart && address <= blockEnd {
			offset := address - blockStart

			// Check if we have enough data
			if offset+length > uint32(len(blockData)) {
				return nil, fmt.Errorf("not enough data at address 0x%X (need %d bytes, have %d)",
					address, length, uint32(len(blockData))-offset)
			}

			// Return copy of requested bytes
			result := make([]byte, length)
			copy(result, blockData[offset:offset+length])
			return result, nil
		}
	}

	return nil, fmt.Errorf("address 0x%X not found in loaded memory", address)
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
		if b == 0 {
			break // null terminator
		}

		if char, exists := charMap[b]; exists {
			result += char
		} else {
			result += fmt.Sprintf("\\x%02X", b) // show unknown bytes
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
