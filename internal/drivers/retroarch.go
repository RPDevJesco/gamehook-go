package drivers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// RetroArchDriver implements the Driver interface for RetroArch
type RetroArchDriver struct {
	host           string
	port           int
	requestTimeout time.Duration
	conn           *net.UDPConn
}

// NewRetroArchDriver creates a new RetroArch driver
func NewRetroArchDriver(host string, port int, requestTimeout time.Duration) *RetroArchDriver {
	return &RetroArchDriver{
		host:           host,
		port:           port,
		requestTimeout: requestTimeout,
	}
}

// Connect establishes connection to RetroArch
func (d *RetroArchDriver) Connect() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", d.host, d.port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to RetroArch: %w", err)
	}

	// Set larger buffer sizes to handle large memory blocks
	// Game Boy WRAM can be up to 8KB, plus overhead
	const bufferSize = 64 * 1024 // 64KB should be more than enough

	if err := conn.SetReadBuffer(bufferSize); err != nil {
		conn.Close()
		return fmt.Errorf("failed to set read buffer size: %w", err)
	}

	if err := conn.SetWriteBuffer(bufferSize); err != nil {
		conn.Close()
		return fmt.Errorf("failed to write buffer size: %w", err)
	}

	d.conn = conn

	// Test connection with a simple command
	_, err = d.sendCommand("VERSION")
	if err != nil {
		d.conn.Close()
		d.conn = nil
		return fmt.Errorf("failed to communicate with RetroArch: %w", err)
	}

	return nil
}

// ReadMemoryBlocks reads multiple memory blocks from RetroArch
func (d *RetroArchDriver) ReadMemoryBlocks(blocks []MemoryBlock) (map[uint32][]byte, error) {
	if d.conn == nil {
		if err := d.Connect(); err != nil {
			return nil, err
		}
	}

	result := make(map[uint32][]byte)

	for _, block := range blocks {
		data, err := d.ReadMemory(block.Start, block.End-block.Start+1)
		if err != nil {
			return nil, fmt.Errorf("failed to read block %s: %w", block.Name, err)
		}
		result[block.Start] = data
	}

	return result, nil
}

// ReadMemory reads memory from a specific address
func (d *RetroArchDriver) ReadMemory(address uint32, length uint32) ([]byte, error) {
	if d.conn == nil {
		return nil, fmt.Errorf("not connected to RetroArch")
	}

	// Convert address to hex string (RetroArch expects lowercase hex without 0x prefix)
	addrStr := fmt.Sprintf("%x", address)

	// Build command: READ_CORE_MEMORY <address> <length>
	command := fmt.Sprintf("READ_CORE_MEMORY %s %d", addrStr, length)

	response, err := d.sendCommand(command)
	if err != nil {
		return nil, err
	}

	// Parse response: "READ_CORE_MEMORY <address> <byte1> <byte2> ..."
	parts := strings.Fields(response)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid response format: %s", response)
	}

	// Check for error response
	if parts[2] == "-1" {
		return nil, fmt.Errorf("RetroArch returned error for address %s", addrStr)
	}

	// Skip command name and address, parse the bytes
	byteStrings := parts[2:]
	if len(byteStrings) != int(length) {
		return nil, fmt.Errorf("expected %d bytes, got %d", length, len(byteStrings))
	}

	data := make([]byte, length)
	for i, byteStr := range byteStrings {
		b, err := strconv.ParseUint(byteStr, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %s: %w", byteStr, err)
		}
		data[i] = byte(b)
	}

	return data, nil
}

// WriteBytes writes bytes to RetroArch
func (d *RetroArchDriver) WriteBytes(address uint32, data []byte) error {
	if d.conn == nil {
		return fmt.Errorf("not connected to RetroArch")
	}

	// Convert bytes to hex strings
	hexBytes := make([]string, len(data))
	for i, b := range data {
		hexBytes[i] = fmt.Sprintf("%02x", b)
	}

	// Build command: WRITE_CORE_MEMORY <address> <byte1> <byte2> ...
	addrStr := fmt.Sprintf("%x", address)
	command := fmt.Sprintf("WRITE_CORE_MEMORY %s %s", addrStr, strings.Join(hexBytes, " "))

	_, err := d.sendCommand(command)
	return err
}

// Close closes the connection
func (d *RetroArchDriver) Close() error {
	if d.conn != nil {
		err := d.conn.Close()
		d.conn = nil
		return err
	}
	return nil
}

// sendCommand sends a command to RetroArch and returns the response
func (d *RetroArchDriver) sendCommand(command string) (string, error) {
	if d.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// Set timeout for this operation
	if err := d.conn.SetDeadline(time.Now().Add(d.requestTimeout)); err != nil {
		return "", fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send command
	_, err := d.conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Read response with larger buffer
	buffer := make([]byte, 65536) // 64KB buffer for large responses
	n, err := d.conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	response := strings.TrimSpace(string(buffer[:n]))
	return response, nil
}
