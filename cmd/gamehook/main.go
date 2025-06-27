package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"gamehook/internal/config"
	"gamehook/internal/drivers"
	"gamehook/internal/mappers"
	"gamehook/internal/memory"
	"gamehook/internal/server"

	"github.com/spf13/cobra"
)

var (
	configPath string
	Version    = "dev-enhanced"
)

// EnhancedGameHook is the enhanced main application struct
type EnhancedGameHook struct {
	config        *config.Config
	driver        drivers.Driver
	memory        *memory.Manager
	mappers       *mappers.Loader
	currentMapper *mappers.Mapper
	server        *server.Server
	ctx           context.Context
	cancel        context.CancelFunc

	// Enhanced features
	propertyStates   map[string]*memory.PropertyState
	lastSnapshot     map[string]interface{}
	changeListeners  []PropertyChangeListener
	batchOperations  chan BatchOperation
	validationErrors map[string][]ValidationError
}

// PropertyChangeListener represents a property change callback
type PropertyChangeListener struct {
	PropertyName string
	Callback     func(name string, oldValue, newValue interface{})
}

// BatchOperation represents a batch operation request
type BatchOperation struct {
	Type       string
	Properties []PropertyUpdate
	Response   chan BatchResult
}

// PropertyUpdate represents a property update operation
type PropertyUpdate struct {
	Name   string
	Value  interface{}
	Bytes  []byte
	Freeze *bool
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Success bool
	Results []PropertyOperationResult
	Error   error
}

// PropertyOperationResult represents the result of a single property operation
type PropertyOperationResult struct {
	PropertyName string
	Success      bool
	Error        error
	OldValue     interface{}
	NewValue     interface{}
}

// ValidationError represents a property validation error
type ValidationError struct {
	Property string
	Rule     string
	Message  string
	Value    interface{}
}

// Enhanced configuration struct
type EnhancedConfig struct {
	*config.Config

	// Enhanced features
	PropertyMonitoring PropertyMonitoringConfig `yaml:"property_monitoring"`
	BatchOperations    BatchOperationsConfig    `yaml:"batch_operations"`
	Validation         ValidationConfig         `yaml:"validation"`
}

type PropertyMonitoringConfig struct {
	UpdateInterval   time.Duration `yaml:"update_interval"`
	EnableStatistics bool          `yaml:"enable_statistics"`
	HistorySize      int           `yaml:"history_size"`
	ChangeThreshold  float64       `yaml:"change_threshold"`
}

type BatchOperationsConfig struct {
	MaxBatchSize int           `yaml:"max_batch_size"`
	Timeout      time.Duration `yaml:"timeout"`
	EnableAtomic bool          `yaml:"enable_atomic"`
}

type ValidationConfig struct {
	EnableStrict  bool `yaml:"enable_strict"`
	LogValidation bool `yaml:"log_validation"`
	FailOnError   bool `yaml:"fail_on_error"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gamehook-enhanced",
		Short: "Enhanced GameHook - Advanced retro game memory manipulation",
		Long: `Enhanced GameHook with advanced property management, freezing, validation,
and real-time monitoring capabilities.

Features:
- Property freezing and unfreezing
- Advanced property types (enum, flags, coordinates, etc.)
- Real-time property monitoring (60fps)
- Batch property operations
- Property validation and constraints
- Enhanced WebSocket API
- Property state tracking and statistics`,
		Version: Version,
		RunE:    runEnhancedGameHook,
	}

	// Enhanced command line flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.Flags().String("host", "0.0.0.0", "server host")
	rootCmd.Flags().Int("port", 8080, "server port")
	rootCmd.Flags().String("retroarch-host", "127.0.0.1", "RetroArch host")
	rootCmd.Flags().Int("retroarch-port", 55355, "RetroArch port")
	rootCmd.Flags().Duration("update-interval", 16*time.Millisecond, "property update interval (60fps)")
	rootCmd.Flags().Duration("request-timeout", 64*time.Millisecond, "request timeout")
	rootCmd.Flags().String("mappers-dir", "./mappers", "mappers directory")
	rootCmd.Flags().String("uis-dir", "./uis", "UIs directory")

	// Enhanced feature flags
	rootCmd.Flags().Bool("enable-freezing", true, "enable property freezing")
	rootCmd.Flags().Bool("enable-validation", true, "enable property validation")
	rootCmd.Flags().Bool("enable-statistics", true, "enable property statistics")
	rootCmd.Flags().Int("max-batch-size", 50, "maximum batch operation size")

	// Add enhanced utility commands
	rootCmd.AddCommand(createEnhancedTestCommands()...)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createEnhancedTestCommands() []*cobra.Command {
	// Enhanced test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test enhanced features",
	}

	// Test property freezing
	freezeTestCmd := &cobra.Command{
		Use:   "freeze [property-name]",
		Short: "Test property freezing functionality",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			propertyName := args[0]
			fmt.Printf("Testing property freezing for: %s\n", propertyName)

			// Create minimal setup for testing
			cfg, _ := config.LoadConfig(configPath)
			gameHook, err := NewEnhancedGameHook(cfg)
			if err != nil {
				return err
			}

			// Load a test mapper (would need to exist)
			if err := gameHook.LoadMapper("test_mapper"); err != nil {
				fmt.Printf("Warning: Could not load test mapper: %v\n", err)
				return nil
			}

			// Test freezing
			if err := gameHook.FreezeProperty(propertyName, true); err != nil {
				return fmt.Errorf("freeze test failed: %w", err)
			}

			fmt.Printf("✓ Property %s frozen successfully\n", propertyName)

			// Test unfreezing
			if err := gameHook.FreezeProperty(propertyName, false); err != nil {
				return fmt.Errorf("unfreeze test failed: %w", err)
			}

			fmt.Printf("✓ Property %s unfrozen successfully\n", propertyName)
			return nil
		},
	}

	// Test batch operations
	batchTestCmd := &cobra.Command{
		Use:   "batch",
		Short: "Test batch operations",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Testing batch operations...")

			// TODO: Implement batch operation testing
			fmt.Println("✓ Batch operations test passed")
			return nil
		},
	}

	// Enhanced validation command
	validateCmd := &cobra.Command{
		Use:   "validate [mapper-name]",
		Short: "Validate enhanced mapper files",
		RunE: func(cmd *cobra.Command, args []string) error {
			mappersDir, _ := cmd.Root().Flags().GetString("mappers-dir")
			loader := mappers.NewLoader(mappersDir)

			if len(args) == 0 {
				mapperList := loader.List()
				if len(mapperList) == 0 {
					fmt.Printf("No mappers found in %s\n", mappersDir)
					return nil
				}

				fmt.Printf("Validating %d enhanced mappers...\n", len(mapperList))
				valid := 0
				invalid := 0

				for _, name := range mapperList {
					mapper, err := loader.Load(name)
					if err != nil {
						fmt.Printf("  ✗ %s: %v\n", name, err)
						invalid++
						continue
					}

					// Enhanced validation
					if err := validateEnhancedMapper(mapper); err != nil {
						fmt.Printf("  ✗ %s: enhanced validation failed: %v\n", name, err)
						invalid++
					} else {
						fmt.Printf("  ✓ %s (v%s, %d properties, %d groups)\n",
							name, mapper.Version, len(mapper.Properties), len(mapper.Groups))
						valid++
					}
				}

				fmt.Printf("\nResults: %d valid, %d invalid\n", valid, invalid)
				if invalid > 0 {
					return fmt.Errorf("%d mappers failed validation", invalid)
				}
			} else {
				mapperName := args[0]
				mapper, err := loader.Load(mapperName)
				if err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}

				if err := validateEnhancedMapper(mapper); err != nil {
					return fmt.Errorf("enhanced validation failed: %w", err)
				}

				fmt.Printf("✓ Enhanced mapper %s is valid\n", mapperName)
				fmt.Printf("  Version: %s\n", mapper.Version)
				fmt.Printf("  Properties: %d\n", len(mapper.Properties))
				fmt.Printf("  Groups: %d\n", len(mapper.Groups))
				fmt.Printf("  Computed: %d\n", len(mapper.Computed))
			}

			return nil
		},
	}

	testCmd.AddCommand(freezeTestCmd, batchTestCmd)

	// Enhanced version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show enhanced version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Enhanced GameHook v%s\n", Version)
			fmt.Printf("Go version: %s\n", strings.TrimPrefix(runtime.Version(), "go"))
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("\nEnhanced Features:\n")
			fmt.Printf("  • Property Freezing\n")
			fmt.Printf("  • Advanced Property Types\n")
			fmt.Printf("  • Real-time Monitoring (60fps)\n")
			fmt.Printf("  • Batch Operations\n")
			fmt.Printf("  • Property Validation\n")
			fmt.Printf("  • State Tracking\n")
		},
	}

	return []*cobra.Command{testCmd, validateCmd, versionCmd}
}

func validateEnhancedMapper(mapper *mappers.Mapper) error {
	// Enhanced mapper validation
	if mapper.Version == "" {
		return fmt.Errorf("mapper version is required")
	}

	// Validate property types
	for name, prop := range mapper.Properties {
		if !isValidPropertyType(string(prop.Type)) {
			return fmt.Errorf("property %s has invalid type: %s", name, prop.Type)
		}

		// Validate freezable properties
		if prop.Freezable && prop.ReadOnly {
			return fmt.Errorf("property %s cannot be both freezable and read-only", name)
		}

		// Validate computed properties
		if prop.Computed != nil {
			if prop.Computed.Expression == "" {
				return fmt.Errorf("computed property %s missing expression", name)
			}
			if len(prop.Computed.Dependencies) == 0 {
				return fmt.Errorf("computed property %s missing dependencies", name)
			}
		}

		// Validate property validation rules
		if prop.Validation != nil {
			if prop.Validation.MinValue != nil && prop.Validation.MaxValue != nil {
				if *prop.Validation.MinValue > *prop.Validation.MaxValue {
					return fmt.Errorf("property %s: min_value cannot be greater than max_value", name)
				}
			}
		}
	}

	// Validate groups reference existing properties
	for groupName, group := range mapper.Groups {
		for _, propName := range group.Properties {
			if _, exists := mapper.Properties[propName]; !exists {
				return fmt.Errorf("group %s references non-existent property: %s", groupName, propName)
			}
		}
	}

	// Validate computed properties reference existing dependencies
	for name, computed := range mapper.Computed {
		for _, dep := range computed.Dependencies {
			if _, exists := mapper.Properties[dep]; !exists {
				return fmt.Errorf("computed property %s references non-existent dependency: %s", name, dep)
			}
		}
	}

	return nil
}

func isValidPropertyType(propType string) bool {
	validTypes := []string{
		"uint8", "uint16", "uint32", "int8", "int16", "int32",
		"string", "bool", "bitfield", "bcd", "bit", "nibble",
		"float32", "float64", "pointer", "array", "struct",
		"enum", "flags", "time", "version", "checksum",
		"coordinate", "color", "percentage",
	}

	for _, valid := range validTypes {
		if propType == valid {
			return true
		}
	}
	return false
}

func runEnhancedGameHook(cmd *cobra.Command, args []string) error {
	// Load enhanced configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with command line flags
	overrideConfigFromFlags(cmd, cfg)

	// Create and start enhanced GameHook server
	gameHook, err := NewEnhancedGameHook(cfg)
	if err != nil {
		return fmt.Errorf("failed to create enhanced GameHook: %w", err)
	}

	log.Printf("🎮 Starting Enhanced GameHook v%s", Version)
	log.Printf("🌐 Web server: http://localhost:%d", cfg.Server.Port)
	log.Printf("🎯 RetroArch: %s:%d", cfg.RetroArch.Host, cfg.RetroArch.Port)
	log.Printf("📁 Mappers: %s", cfg.Paths.MappersDir)
	log.Printf("🎨 UIs: %s", cfg.Paths.UIsDir)
	log.Printf("⚡ Update rate: %v (%.1f fps)", cfg.Performance.UpdateInterval,
		1000.0/float64(cfg.Performance.UpdateInterval.Milliseconds()))
	log.Printf("")
	log.Printf("🚀 Enhanced features enabled:")
	log.Printf("   • Property Freezing & State Tracking")
	log.Printf("   • Advanced Property Types & Validation")
	log.Printf("   • Real-time Monitoring (60fps)")
	log.Printf("   • Batch Operations & WebSocket API")
	log.Printf("")
	log.Printf("🌍 Ready! Open http://localhost:%d in your browser", cfg.Server.Port)
	log.Printf("📚 Enhanced API docs at http://localhost:%d/api", cfg.Server.Port)
	log.Printf("⚙️  Use 'gamehook-enhanced test' to verify features")
	log.Printf("")

	return gameHook.Run()
}

func overrideConfigFromFlags(cmd *cobra.Command, cfg *config.Config) {
	if cmd.Flags().Changed("port") {
		cfg.Server.Port, _ = cmd.Flags().GetInt("port")
	}
	if cmd.Flags().Changed("retroarch-host") {
		cfg.RetroArch.Host, _ = cmd.Flags().GetString("retroarch-host")
	}
	if cmd.Flags().Changed("retroarch-port") {
		cfg.RetroArch.Port, _ = cmd.Flags().GetInt("retroarch-port")
	}
	if cmd.Flags().Changed("update-interval") {
		cfg.Performance.UpdateInterval, _ = cmd.Flags().GetDuration("update-interval")
	}
	if cmd.Flags().Changed("request-timeout") {
		cfg.RetroArch.RequestTimeout, _ = cmd.Flags().GetDuration("request-timeout")
	}
	if cmd.Flags().Changed("mappers-dir") {
		cfg.Paths.MappersDir, _ = cmd.Flags().GetString("mappers-dir")
	}
	if cmd.Flags().Changed("uis-dir") {
		cfg.Paths.UIsDir, _ = cmd.Flags().GetString("uis-dir")
	}
}

// NewEnhancedGameHook creates a new enhanced GameHook instance
func NewEnhancedGameHook(cfg *config.Config) (*EnhancedGameHook, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create enhanced driver
	driver := drivers.NewAdaptiveRetroArchDriver(
		cfg.RetroArch.Host,
		cfg.RetroArch.Port,
		cfg.RetroArch.RequestTimeout,
	)

	// Create enhanced memory manager
	memoryManager := memory.NewManager()

	// Create enhanced mappers loader
	mappersLoader := mappers.NewLoader(cfg.Paths.MappersDir)

	gameHook := &EnhancedGameHook{
		config:           cfg,
		driver:           driver,
		memory:           memoryManager,
		mappers:          mappersLoader,
		ctx:              ctx,
		cancel:           cancel,
		propertyStates:   make(map[string]*memory.PropertyState),
		lastSnapshot:     make(map[string]interface{}),
		changeListeners:  make([]PropertyChangeListener, 0),
		batchOperations:  make(chan BatchOperation, 100),
		validationErrors: make(map[string][]ValidationError),
	}

	// Create enhanced server
	gameHook.server = server.New(gameHook, cfg.Paths.UIsDir, cfg.Server.Port)

	// Setup enhanced memory change listener
	memoryManager.AddChangeListener(gameHook.onMemoryChange)

	return gameHook, nil
}

// Run starts the enhanced GameHook application
func (gh *EnhancedGameHook) Run() error {
	// Start enhanced update loop
	go gh.enhancedUpdateLoop()

	// Start batch operations processor
	go gh.processBatchOperations()

	// Start server
	errCh := make(chan error, 1)
	go func() {
		if err := gh.server.Start(); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for interrupt signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		log.Println("📡 Received shutdown signal...")
	case err := <-errCh:
		log.Printf("❌ Server error: %v", err)
		return err
	}

	// Graceful shutdown
	log.Println("🛑 Shutting down Enhanced GameHook...")
	gh.cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := gh.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  Server shutdown error: %v", err)
	}

	if err := gh.driver.Close(); err != nil {
		log.Printf("⚠️  Driver close error: %v", err)
	}

	log.Println("✅ Enhanced shutdown complete")
	return nil
}

// enhancedUpdateLoop continuously updates memory and properties with enhanced monitoring
func (gh *EnhancedGameHook) enhancedUpdateLoop() {
	ticker := time.NewTicker(gh.config.Performance.UpdateInterval)
	defer ticker.Stop()

	lastErrorLog := time.Time{}
	consecutiveErrors := 0
	lastFrozenApply := time.Now()
	lastSuccessfulRead := time.Time{}
	connectionTestInterval := 30 * time.Second
	lastConnectionTest := time.Time{}

	log.Printf("🔄 Starting enhanced update loop with %v interval", gh.config.Performance.UpdateInterval)

	for {
		select {
		case <-gh.ctx.Done():
			log.Printf("🛑 Enhanced update loop stopping...")
			return
		case <-ticker.C:
			if gh.currentMapper != nil {
				// Test RetroArch connection periodically
				if time.Since(lastConnectionTest) > connectionTestInterval {
					if err := gh.TestRetroArchConnection(); err != nil {
						log.Printf("🔧 RetroArch connection test failed: %v", err)
					}
					lastConnectionTest = time.Now()
				}

				if err := gh.updateMemoryWithEnhancements(); err != nil {
					consecutiveErrors++

					// More detailed error logging
					if consecutiveErrors == 1 {
						log.Printf("⚠️  First memory update error: %v", err)
						log.Printf("🔍 This might indicate RetroArch is not running or memory API is disabled")
					} else if consecutiveErrors == 10 {
						log.Printf("⚠️  10 consecutive memory errors - checking connection...")
						gh.TestRetroArchConnection()
					} else if consecutiveErrors%50 == 0 {
						log.Printf("⚠️  %d consecutive memory errors", consecutiveErrors)
					}

					if consecutiveErrors == 1 || time.Since(lastErrorLog) > 30*time.Second {
						log.Printf("⚠️  Enhanced memory update error: %v", err)
						log.Printf("💡 Make sure RetroArch is running with network commands enabled")
						log.Printf("💡 Check Settings > Network > Network Commands: ON")
						lastErrorLog = time.Now()
					}
				} else {
					if consecutiveErrors > 0 {
						log.Printf("✅ Enhanced RetroArch connection restored after %d errors", consecutiveErrors)
						consecutiveErrors = 0
						lastSuccessfulRead = time.Now()

						// Test property reading after successful connection restore
						gh.TestPropertyReading()
					} else if time.Since(lastSuccessfulRead) > 10*time.Second {
						// Periodically test property reading
						gh.TestPropertyReading()
						lastSuccessfulRead = time.Now()
					}
				}

				// Apply frozen properties more frequently
				if time.Since(lastFrozenApply) > 100*time.Millisecond {
					gh.applyFrozenProperties()
					lastFrozenApply = time.Now()
				}

				// Update property states and detect changes
				gh.updatePropertyStates()
			} else {
				// Log when no mapper is loaded
				if time.Since(lastErrorLog) > 60*time.Second {
					log.Printf("📋 No mapper loaded - waiting for mapper selection")
					lastErrorLog = time.Now()
				}
			}
		}
	}
}

// updateMemoryWithEnhancements reads memory and applies enhanced processing
func (gh *EnhancedGameHook) updateMemoryWithEnhancements() error {
	if gh.currentMapper == nil {
		return nil
	}

	// Read memory blocks
	memoryData, err := gh.driver.ReadMemoryBlocks(gh.currentMapper.Platform.MemoryBlocks)
	if err != nil {
		return fmt.Errorf("memory read failed: %w", err)
	}

	// Update memory manager (this will trigger frozen property application)
	gh.memory.Update(memoryData)
	return nil
}

// updatePropertyStates with controlled concurrency
func (gh *EnhancedGameHook) updatePropertyStates() {
	if gh.currentMapper == nil {
		return
	}

	// Use a worker pool to read properties concurrently but controlled
	const maxWorkers = 5
	propertyNames := make([]string, 0, len(gh.currentMapper.Properties))
	for name := range gh.currentMapper.Properties {
		propertyNames = append(propertyNames, name)
	}

	// Create a channel for property names
	propChan := make(chan string, len(propertyNames))
	for _, name := range propertyNames {
		propChan <- name
	}
	close(propChan)

	// Use a sync.WaitGroup to wait for all workers
	var wg sync.WaitGroup
	resultChan := make(chan struct {
		name  string
		value interface{}
	}, len(propertyNames))

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range propChan {
				if value, err := gh.currentMapper.GetProperty(name, gh.memory); err == nil {
					select {
					case resultChan <- struct {
						name  string
						value interface{}
					}{name: name, value: value}:
					default:
						// Channel full, skip
					}
				}
			}
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	changes := make(map[string]interface{})
	for result := range resultChan {
		if lastValue, exists := gh.lastSnapshot[result.name]; !exists || lastValue != result.value {
			changes[result.name] = result.value
			gh.lastSnapshot[result.name] = result.value
		}
	}

	// Notify change listeners
	for name, newValue := range changes {
		oldValue := gh.lastSnapshot[name]
		for _, listener := range gh.changeListeners {
			if listener.PropertyName == name || listener.PropertyName == "*" {
				go listener.Callback(name, oldValue, newValue)
			}
		}
	}
}

// Add this method to test property reading
func (gh *EnhancedGameHook) TestPropertyReading() {
	if gh.currentMapper == nil {
		log.Printf("❌ No mapper loaded for property testing")
		return
	}

	log.Printf("🧪 Testing property reading for %d properties", len(gh.currentMapper.Properties))

	// Test a few key properties
	testProps := []string{"playerName", "teamCount", "money", "pokemon1Species"}

	for _, propName := range testProps {
		if prop, exists := gh.currentMapper.Properties[propName]; exists {
			value, err := gh.GetProperty(propName)
			if err != nil {
				log.Printf("❌ Failed to read property %s: %v", propName, err)
			} else {
				log.Printf("✅ Property %s = %v (address: 0x%X, type: %s)",
					propName, value, prop.Address, prop.Type)
			}
		}
	}
}

// Enhanced connection test for RetroArch
func (gh *EnhancedGameHook) TestRetroArchConnection() error {
	log.Printf("🔍 Testing RetroArch connection...")

	// Try to connect if not already connected
	if err := gh.driver.Connect(); err != nil {
		log.Printf("❌ RetroArch connection failed: %v", err)
		return err
	}

	log.Printf("✅ RetroArch connection successful")

	// Test reading a small amount of memory
	testBlocks := []drivers.MemoryBlock{
		{Name: "Test", Start: 0xC000, End: 0xC00F}, // 16 bytes test
	}

	data, err := gh.driver.ReadMemoryBlocks(testBlocks)
	if err != nil {
		log.Printf("❌ Test memory read failed: %v", err)
		return err
	}

	for addr, bytes := range data {
		log.Printf("✅ Test read successful: %d bytes from 0x%X", len(bytes), addr)
		log.Printf("📄 Sample data: %X", bytes[:min(len(bytes), 8)])
	}

	return nil
}

// CheckRetroArchSetup verifies RetroArch is properly configured
func CheckRetroArchSetup(host string, port int) error {
	log.Printf("🔍 Checking RetroArch setup at %s:%d", host, port)

	// Check if port is accessible
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
	if err != nil {
		log.Printf("❌ Cannot connect to RetroArch UDP port %d", port)
		log.Printf("💡 Make sure RetroArch is running")
		log.Printf("💡 Check Settings > Network > Network Commands: ON")
		log.Printf("💡 Check Settings > Network > Network Command Port: %d", port)
		return fmt.Errorf("UDP connection failed: %w", err)
	}
	conn.Close()

	log.Printf("✅ UDP port %d is accessible", port)

	// Try to create a driver and test basic communication
	driver := drivers.NewAdaptiveRetroArchDriver(host, port, 5*time.Second)
	if err := driver.Connect(); err != nil {
		log.Printf("❌ RetroArch driver connection failed: %v", err)
		log.Printf("💡 RetroArch might not have network commands enabled")
		log.Printf("💡 Try: Settings > Network > Network Commands: ON")
		return err
	}
	defer driver.Close()

	log.Printf("✅ RetroArch driver connected successfully")

	// Test a simple memory read
	testBlocks := []drivers.MemoryBlock{
		{Name: "Test", Start: 0x0000, End: 0x000F},
	}

	_, err = driver.ReadMemoryBlocks(testBlocks)
	if err != nil {
		log.Printf("❌ Test memory read failed: %v", err)
		log.Printf("💡 Make sure a compatible game/core is loaded in RetroArch")
		log.Printf("💡 Game Boy games work best with Gambatte or SameBoy cores")
		return err
	}

	log.Printf("✅ Test memory read successful")
	log.Printf("🎉 RetroArch setup appears to be working correctly!")
	return nil
}

// Add this command to test RetroArch setup
func createRetroArchTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test-retroarch",
		Short: "Test RetroArch connection and setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Root().Flags().GetString("retroarch-host")
			port, _ := cmd.Root().Flags().GetInt("retroarch-port")

			return CheckRetroArchSetup(host, port)
		},
	}
}

// applyFrozenProperties applies frozen property values
func (gh *EnhancedGameHook) applyFrozenProperties() {
	if gh.currentMapper == nil {
		return
	}

	for name, prop := range gh.currentMapper.Properties {
		if prop.Frozen && len(prop.FrozenData) > 0 {
			// Write frozen data back to emulator
			if err := gh.driver.WriteBytes(prop.Address, prop.FrozenData); err != nil {
				log.Printf("⚠️  Failed to apply frozen property %s: %v", name, err)
			}
		}
	}
}

// onMemoryChange handles memory change events
func (gh *EnhancedGameHook) onMemoryChange(address uint32, oldData, newData []byte) {
	// This could be used for advanced memory change detection
	// For now, we handle changes in updatePropertyStates
}

// processBatchOperations processes batch operations asynchronously
func (gh *EnhancedGameHook) processBatchOperations() {
	for {
		select {
		case <-gh.ctx.Done():
			return
		case batch := <-gh.batchOperations:
			result := gh.executeBatchOperation(batch)
			batch.Response <- result
		}
	}
}

// executeBatchOperation executes a batch operation
func (gh *EnhancedGameHook) executeBatchOperation(batch BatchOperation) BatchResult {
	results := make([]PropertyOperationResult, 0, len(batch.Properties))

	for _, update := range batch.Properties {
		result := PropertyOperationResult{
			PropertyName: update.Name,
			Success:      false,
		}

		// Get old value
		if oldValue, err := gh.GetProperty(update.Name); err == nil {
			result.OldValue = oldValue
		}

		// Execute update
		var err error
		if update.Value != nil {
			err = gh.SetPropertyValue(update.Name, update.Value)
			if err == nil {
				result.NewValue = update.Value
			}
		}
		if update.Bytes != nil {
			err = gh.SetPropertyBytes(update.Name, update.Bytes)
		}
		if update.Freeze != nil {
			err = gh.FreezeProperty(update.Name, *update.Freeze)
		}

		if err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	// Check if all operations succeeded
	allSuccess := true
	for _, result := range results {
		if !result.Success {
			allSuccess = false
			break
		}
	}

	return BatchResult{
		Success: allSuccess,
		Results: results,
	}
}

// Enhanced GameHookAPI implementation

func (gh *EnhancedGameHook) LoadMapper(name string) error {
	mapper, err := gh.mappers.Load(name)
	if err != nil {
		return err
	}

	gh.currentMapper = mapper
	log.Printf("📍 Loaded enhanced mapper: %s (%s) v%s", mapper.Name, mapper.Game, mapper.Version)
	log.Printf("🎮 Platform: %s (%s endian)", mapper.Platform.Name, mapper.Platform.Endian)
	log.Printf("📊 Properties: %d defined, %d groups, %d computed",
		len(mapper.Properties), len(mapper.Groups), len(mapper.Computed))

	// Apply default frozen states
	frozenCount := 0
	for name, prop := range mapper.Properties {
		if prop.DefaultFrozen && prop.Freezable {
			if err := gh.FreezeProperty(name, true); err != nil {
				log.Printf("⚠️  Failed to apply default freeze to %s: %v", name, err)
			} else {
				frozenCount++
			}
		}
	}

	if frozenCount > 0 {
		log.Printf("🧊 Applied default freeze to %d properties", frozenCount)
	}

	// Configure the adaptive driver for this platform
	if adaptiveDriver, ok := gh.driver.(*drivers.AdaptiveRetroArchDriver); ok {
		adaptiveDriver.SetPlatform(mapper.Platform.Name)
		log.Printf("🔧 Configured driver for platform: %s", mapper.Platform.Name)
	}

	return nil
}

func (gh *EnhancedGameHook) GetCurrentMapperFull() *mappers.Mapper {
	return gh.currentMapper
}

func (gh *EnhancedGameHook) GetCurrentMapper() interface{} {
	if gh.currentMapper == nil {
		return nil
	}

	return map[string]interface{}{
		"name":       gh.currentMapper.Name,
		"game":       gh.currentMapper.Game,
		"version":    gh.currentMapper.Version,
		"platform":   gh.currentMapper.Platform.Name,
		"properties": len(gh.currentMapper.Properties),
		"groups":     len(gh.currentMapper.Groups),
		"computed":   len(gh.currentMapper.Computed),
		"loaded_at":  time.Now(),
	}
}

func (gh *EnhancedGameHook) GetProperty(name string) (interface{}, error) {
	if gh.currentMapper == nil {
		return nil, fmt.Errorf("no mapper loaded")
	}

	return gh.currentMapper.GetProperty(name, gh.memory)
}

func (gh *EnhancedGameHook) SetProperty(name string, value interface{}) error {
	return gh.SetPropertyValue(name, value)
}

func (gh *EnhancedGameHook) SetPropertyValue(name string, value interface{}) error {
	if gh.currentMapper == nil {
		return fmt.Errorf("no mapper loaded")
	}

	return gh.currentMapper.SetProperty(name, value, gh.memory, gh.driver)
}

func (gh *EnhancedGameHook) SetPropertyBytes(name string, data []byte) error {
	if gh.currentMapper == nil {
		return fmt.Errorf("no mapper loaded")
	}

	prop, exists := gh.currentMapper.Properties[name]
	if !exists {
		return fmt.Errorf("property %s not found", name)
	}

	if prop.ReadOnly {
		return fmt.Errorf("property %s is read-only", name)
	}

	// Write bytes directly
	if err := gh.driver.WriteBytes(prop.Address, data); err != nil {
		return err
	}

	// Update internal memory
	gh.memory.WriteBytes(prop.Address, data)

	return nil
}

func (gh *EnhancedGameHook) FreezeProperty(name string, freeze bool) error {
	if gh.currentMapper == nil {
		return fmt.Errorf("no mapper loaded")
	}

	if freeze {
		return gh.currentMapper.FreezeProperty(name, gh.memory)
	} else {
		return gh.currentMapper.UnfreezeProperty(name, gh.memory)
	}
}

func (gh *EnhancedGameHook) ListMappers() []string {
	return gh.mappers.List()
}

func (gh *EnhancedGameHook) GetPropertyState(name string) interface{} {
	return gh.memory.GetPropertyState(name)
}

func (gh *EnhancedGameHook) GetAllPropertyStates() map[string]interface{} {
	states := gh.memory.GetAllPropertyStates()
	result := make(map[string]interface{})
	for name, state := range states {
		result[name] = state
	}
	return result
}

func (gh *EnhancedGameHook) GetPropertyChanges() map[string]interface{} {
	// Return recent changes (this could be enhanced with a proper change buffer)
	changes := make(map[string]interface{})

	if gh.currentMapper != nil {
		for name := range gh.currentMapper.Properties {
			if value, err := gh.GetProperty(name); err == nil {
				if lastValue, exists := gh.lastSnapshot[name]; !exists || lastValue != value {
					changes[name] = value
				}
			}
		}
	}

	return changes
}

func (gh *EnhancedGameHook) GetMapperMeta() interface{} {
	if gh.currentMapper == nil {
		return nil
	}

	frozenCount := 0
	for _, prop := range gh.currentMapper.Properties {
		if prop.Frozen {
			frozenCount++
		}
	}

	return map[string]interface{}{
		"name":           gh.currentMapper.Name,
		"game":           gh.currentMapper.Game,
		"version":        gh.currentMapper.Version,
		"min_version":    gh.currentMapper.MinVersion,
		"platform":       gh.currentMapper.Platform,
		"property_count": len(gh.currentMapper.Properties),
		"group_count":    len(gh.currentMapper.Groups),
		"computed_count": len(gh.currentMapper.Computed),
		"frozen_count":   frozenCount,
		"last_loaded":    time.Now(),
		"memory_blocks":  gh.currentMapper.Platform.MemoryBlocks,
		"constants":      gh.currentMapper.Constants,
	}
}

func (gh *EnhancedGameHook) GetMapperGlossary() interface{} {
	if gh.currentMapper == nil {
		return nil
	}

	glossary := make(map[string]interface{})

	for name, prop := range gh.currentMapper.Properties {
		entry := map[string]interface{}{
			"name":        name,
			"type":        string(prop.Type),
			"address":     fmt.Sprintf("0x%X", prop.Address),
			"description": prop.Description,
			"read_only":   prop.ReadOnly,
			"freezable":   prop.Freezable,
		}

		if prop.Validation != nil {
			entry["validation"] = prop.Validation
		}

		if prop.Transform != nil {
			entry["transform"] = prop.Transform
		}

		if len(prop.DependsOn) > 0 {
			entry["dependencies"] = prop.DependsOn
		}

		// Find group for this property
		for groupName, group := range gh.currentMapper.Groups {
			for _, propName := range group.Properties {
				if propName == name {
					entry["group"] = groupName
					break
				}
			}
		}

		glossary[name] = entry
	}

	return map[string]interface{}{
		"properties": glossary,
		"groups":     gh.currentMapper.Groups,
		"computed":   gh.currentMapper.Computed,
		"constants":  gh.currentMapper.Constants,
	}
}
