package drivers

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Driver interface for emulator communication
type Driver interface {
	Connect() error
	ReadMemoryBlocks(blocks []MemoryBlock) (map[uint32][]byte, error)
	WriteBytes(address uint32, data []byte) error
	Close() error
}

// MemoryBlock represents a contiguous block of memory to read
type MemoryBlock struct {
	Name  string
	Start uint32
	End   uint32
}

// RetroArchDriver handles UDP communication with RetroArch
type RetroArchDriver struct {
	host           string
	port           int
	requestTimeout time.Duration

	conn        net.Conn
	mu          sync.Mutex
	responseMap map[string]chan string
}

// NewRetroArchDriver creates a new RetroArch driver
func NewRetroArchDriver(host string, port int, timeout time.Duration) *RetroArchDriver {
	return &RetroArchDriver{
		host:           host,
		port:           port,
		requestTimeout: timeout,
		responseMap:    make(map[string]chan string),
	}
}

// Connect establishes connection to RetroArch
func (d *RetroArchDriver) Connect() error {
	address := fmt.Sprintf("%s:%d", d.host, d.port)

	conn, err := net.Dial("udp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to RetroArch at %s: %w", address, err)
	}

	d.conn = conn

	// Start response reader goroutine
	go d.readResponses()

	return nil
}

// Close closes the connection
func (d *RetroArchDriver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// ReadMemoryBlocks reads multiple memory blocks
func (d *RetroArchDriver) ReadMemoryBlocks(blocks []MemoryBlock) (map[uint32][]byte, error) {
	result := make(map[uint32][]byte)

	for _, block := range blocks {
		length := block.End - block.Start + 1
		data, err := d.ReadMemory(block.Start, length)
		if err != nil {
			return nil, fmt.Errorf("failed to read block %s: %w", block.Name, err)
		}
		result[block.Start] = data
	}

	return result, nil
}

// ReadMemory reads memory from RetroArch at specified address
func (d *RetroArchDriver) ReadMemory(address uint32, length uint32) ([]byte, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("not connected to RetroArch")
	}

	// Format command: "READ_CORE_MEMORY address length"
	addressHex := d.formatAddress(address)
	command := fmt.Sprintf("READ_CORE_MEMORY %s %d", addressHex, length)

	response, err := d.sendCommand(command)
	if err != nil {
		return nil, err
	}

	return d.parseMemoryResponse(response)
}

// WriteBytes writes data to RetroArch memory
func (d *RetroArchDriver) WriteBytes(address uint32, data []byte) error {
	if d.conn == nil {
		return fmt.Errorf("not connected to RetroArch")
	}

	// Format command: "WRITE_CORE_MEMORY address data..."
	addressHex := d.formatAddress(address)

	// Convert bytes to hex strings
	hexStrings := make([]string, len(data))
	for i, b := range data {
		hexStrings[i] = fmt.Sprintf("%02x", b)
	}

	command := fmt.Sprintf("WRITE_CORE_MEMORY %s %s", addressHex, strings.Join(hexStrings, " "))

	_, err := d.sendCommand(command)
	return err
}

// sendCommand sends a command and waits for response
func (d *RetroArchDriver) sendCommand(command string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create response channel
	responseChan := make(chan string, 1)
	d.responseMap[command] = responseChan
	defer delete(d.responseMap, command)

	// Send command
	_, err := d.conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for response with timeout
	ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout)
	defer cancel()

	select {
	case response := <-responseChan:
		return response, nil
	case <-ctx.Done():
		return "", fmt.Errorf("command timeout: %s", command)
	}
}

// readResponses continuously reads responses from RetroArch
func (d *RetroArchDriver) readResponses() {
	scanner := bufio.NewScanner(d.conn)

	for scanner.Scan() {
		response := strings.TrimSpace(scanner.Text())
		if response == "" {
			continue
		}

		// Find matching command and send response
		d.mu.Lock()
		for command, responseChan := range d.responseMap {
			if d.matchesCommand(command, response) {
				select {
				case responseChan <- response:
				default:
					// Channel full, skip
				}
				break
			}
		}
		d.mu.Unlock()
	}
}

// matchesCommand checks if a response matches a command
func (d *RetroArchDriver) matchesCommand(command, response string) bool {
	// Extract command type and address from both
	cmdParts := strings.Split(command, " ")
	respParts := strings.Split(response, " ")

	if len(cmdParts) < 3 || len(respParts) < 2 {
		return false
	}

	// Check if command type and address match
	return cmdParts[0] == respParts[0] && cmdParts[1] == respParts[1]
}

// parseMemoryResponse converts hex response to bytes
func (d *RetroArchDriver) parseMemoryResponse(response string) ([]byte, error) {
	parts := strings.Split(response, " ")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid response format: %s", response)
	}

	// Skip command type and address, parse hex bytes
	hexBytes := parts[2:]
	result := make([]byte, len(hexBytes))

	for i, hexByte := range hexBytes {
		if hexByte == "-1" {
			return nil, fmt.Errorf("RetroArch error: %s", response)
		}

		value, err := strconv.ParseUint(hexByte, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid hex byte %s: %w", hexByte, err)
		}
		result[i] = byte(value)
	}

	return result, nil
}

// formatAddress formats address for RetroArch (handles single digits vs hex)
func (d *RetroArchDriver) formatAddress(address uint32) string {
	if address <= 9 {
		return strconv.FormatUint(uint64(address), 10)
	}
	return fmt.Sprintf("%x", address)
}
