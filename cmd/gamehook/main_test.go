package main

import (
	"testing"
	"time"

	"gamehook/internal/drivers"
	"gamehook/internal/mappers"
	"gamehook/internal/memory"
)

// MockDriver for testing
type MockDriver struct {
	memoryData map[uint32][]byte
	writeLog   []WriteOperation
}

type WriteOperation struct {
	Address uint32
	Data    []byte
}

func NewMockDriver() *MockDriver {
	return &MockDriver{
		memoryData: make(map[uint32][]byte),
		writeLog:   make([]WriteOperation, 0),
	}
}

func (d *MockDriver) Connect() error {
	return nil
}

func (d *MockDriver) ReadMemoryBlocks(blocks []drivers.MemoryBlock) (map[uint32][]byte, error) {
	result := make(map[uint32][]byte)

	for _, block := range blocks {
		// Generate mock data for each block
		blockSize := block.End - block.Start + 1
		data := make([]byte, blockSize)

		// Fill with pattern data for testing
		for i := uint32(0); i < blockSize; i++ {
			data[i] = byte((block.Start + i) & 0xFF)
		}

		result[block.Start] = data
	}

	return result, nil
}

func (d *MockDriver) WriteBytes(address uint32, data []byte) error {
	d.writeLog = append(d.writeLog, WriteOperation{
		Address: address,
		Data:    make([]byte, len(data)),
	})
	copy(d.writeLog[len(d.writeLog)-1].Data, data)
	return nil
}

func (d *MockDriver) Close() error {
	return nil
}

func (d *MockDriver) SetMemoryData(address uint32, data []byte) {
	d.memoryData[address] = data
}

func TestMemoryManager(t *testing.T) {
	manager := memory.NewManager()

	// Test data
	testData := map[uint32][]byte{
		0x1000: {0x01, 0x02, 0x03, 0x04},
		0x2000: {0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
	}

	// Update memory
	manager.Update(testData)

	// Test reading bytes
	data, err := manager.ReadBytes(0x1000, 4)
	if err != nil {
		t.Fatalf("Failed to read bytes: %v", err)
	}

	expected := []byte{0x01, 0x02, 0x03, 0x04}
	for i, b := range data {
		if b != expected[i] {
			t.Errorf("Expected byte %d to be 0x%02X, got 0x%02X", i, expected[i], b)
		}
	}

	// Test reading uint8
	val8, err := manager.ReadUint8(0x1000)
	if err != nil {
		t.Fatalf("Failed to read uint8: %v", err)
	}
	if val8 != 0x01 {
		t.Errorf("Expected uint8 to be 0x01, got 0x%02X", val8)
	}

	// Test reading uint16 little endian
	val16, err := manager.ReadUint16(0x1000, true)
	if err != nil {
		t.Fatalf("Failed to read uint16: %v", err)
	}
	expected16 := uint16(0x0201) // Little endian: 0x01, 0x02 -> 0x0201
	if val16 != expected16 {
		t.Errorf("Expected uint16 to be 0x%04X, got 0x%04X", expected16, val16)
	}

	// Test reading beyond available data
	_, err = manager.ReadBytes(0x1000, 10)
	if err == nil {
		t.Error("Expected error when reading beyond available data")
	}

	// Test reading from non-existent address
	_, err = manager.ReadBytes(0x9999, 1)
	if err == nil {
		t.Error("Expected error when reading from non-existent address")
	}
}

func TestMockDriver(t *testing.T) {
	driver := NewMockDriver()

	// Test connection
	err := driver.Connect()
	if err != nil {
		t.Fatalf("Failed to connect mock driver: %v", err)
	}

	// Test reading memory blocks
	blocks := []drivers.MemoryBlock{
		{Name: "Test Block", Start: 0x1000, End: 0x1003},
	}

	data, err := driver.ReadMemoryBlocks(blocks)
	if err != nil {
		t.Fatalf("Failed to read memory blocks: %v", err)
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 memory block, got %d", len(data))
	}

	blockData, exists := data[0x1000]
	if !exists {
		t.Error("Expected memory block at 0x1000")
	}

	if len(blockData) != 4 {
		t.Errorf("Expected block size 4, got %d", len(blockData))
	}

	// Test writing bytes
	testWrite := []byte{0xAA, 0xBB, 0xCC}
	err = driver.WriteBytes(0x2000, testWrite)
	if err != nil {
		t.Fatalf("Failed to write bytes: %v", err)
	}

	if len(driver.writeLog) != 1 {
		t.Errorf("Expected 1 write operation, got %d", len(driver.writeLog))
	}

	writeOp := driver.writeLog[0]
	if writeOp.Address != 0x2000 {
		t.Errorf("Expected write address 0x2000, got 0x%04X", writeOp.Address)
	}

	for i, b := range writeOp.Data {
		if b != testWrite[i] {
			t.Errorf("Expected write data[%d] to be 0x%02X, got 0x%02X", i, testWrite[i], b)
		}
	}
}

func TestGameHookIntegration(t *testing.T) {
	// Create mock components
	driver := NewMockDriver()
	memManager := memory.NewManager()
	mapperLoader := mappers.NewLoader("../../mappers")

	// Create minimal GameHook instance
	gameHook := &GameHook{
		driver:  driver,
		memory:  memManager,
		mappers: mapperLoader,
		config: Config{
			UpdateInterval: 10 * time.Millisecond,
		},
	}

	// Test that components are properly initialized
	if gameHook.driver == nil {
		t.Error("Driver not initialized")
	}

	if gameHook.memory == nil {
		t.Error("Memory manager not initialized")
	}

	if gameHook.mappers == nil {
		t.Error("Mapper loader not initialized")
	}

	// Test initial state
	if gameHook.currentMapper != nil {
		t.Error("Expected no mapper to be loaded initially")
	}

	// Test getting property without mapper
	_, err := gameHook.GetProperty("test")
	if err == nil {
		t.Error("Expected error when getting property without mapper")
	}

	// Test setting property without mapper
	err = gameHook.SetProperty("test", 123)
	if err == nil {
		t.Error("Expected error when setting property without mapper")
	}
}

func TestMemoryManagerWriteOperations(t *testing.T) {
	manager := memory.NewManager()

	// Initialize with some data
	testData := map[uint32][]byte{
		0x1000: {0x00, 0x00, 0x00, 0x00},
	}
	manager.Update(testData)

	// Test writing uint8
	data := manager.WriteUint8(0x1000, 0xAA)
	if len(data) != 1 {
		t.Errorf("Expected 1 byte, got %d", len(data))
	}
	if data[0] != 0xAA {
		t.Errorf("Expected 0xAA, got 0x%02X", data[0])
	}

	// Test writing uint16 little endian
	data = manager.WriteUint16(0x1000, 0x1234, true)
	if len(data) != 2 {
		t.Errorf("Expected 2 bytes, got %d", len(data))
	}
	// Little endian: 0x1234 -> [0x34, 0x12]
	if data[0] != 0x34 || data[1] != 0x12 {
		t.Errorf("Expected [0x34, 0x12], got [0x%02X, 0x%02X]", data[0], data[1])
	}

	// Test writing uint16 big endian
	data = manager.WriteUint16(0x1000, 0x1234, false)
	if len(data) != 2 {
		t.Errorf("Expected 2 bytes, got %d", len(data))
	}
	// Big endian: 0x1234 -> [0x12, 0x34]
	if data[0] != 0x12 || data[1] != 0x34 {
		t.Errorf("Expected [0x12, 0x34], got [0x%02X, 0x%02X]", data[0], data[1])
	}

	// Test writing bool
	data = manager.WriteBool(0x1000, true)
	if len(data) != 1 {
		t.Errorf("Expected 1 byte, got %d", len(data))
	}
	if data[0] != 0x01 {
		t.Errorf("Expected 0x01 for true, got 0x%02X", data[0])
	}

	data = manager.WriteBool(0x1000, false)
	if data[0] != 0x00 {
		t.Errorf("Expected 0x00 for false, got 0x%02X", data[0])
	}
}

func TestMapperLoader(t *testing.T) {
	loader := mappers.NewLoader("../../mappers")

	// Test listing mappers (this will be empty in test environment)
	mapperList := loader.List()
	if mapperList == nil {
		t.Error("Mapper list should not be nil")
	}

	// Test loading non-existent mapper
	_, err := loader.Load("non_existent_mapper")
	if err == nil {
		t.Error("Expected error when loading non-existent mapper")
	}
}

func BenchmarkMemoryManagerRead(b *testing.B) {
	manager := memory.NewManager()

	// Initialize with test data
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i & 0xFF)
	}

	testData := map[uint32][]byte{
		0x1000: data,
	}
	manager.Update(testData)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := manager.ReadUint8(0x1000 + uint32(i%1024))
		if err != nil {
			b.Fatalf("Read error: %v", err)
		}
	}
}

func BenchmarkMemoryManagerUpdate(b *testing.B) {
	manager := memory.NewManager()

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	testData := map[uint32][]byte{
		0x1000: data,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Update(testData)
	}
}

func TestConfigParsing(t *testing.T) {
	// Test default config values
	config := Config{
		Port:           8080,
		RetroArchHost:  "127.0.0.1",
		RetroArchPort:  55355,
		UpdateInterval: 5 * time.Millisecond,
		RequestTimeout: 64 * time.Millisecond,
	}

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}

	if config.RetroArchHost != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.RetroArchHost)
	}

	if config.UpdateInterval != 5*time.Millisecond {
		t.Errorf("Expected update interval 5ms, got %v", config.UpdateInterval)
	}
}
