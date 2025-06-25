package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
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
	Version    = "dev"
)

// GameHook is the main application struct
type GameHook struct {
	config        *config.Config
	driver        drivers.Driver
	memory        *memory.Manager
	mappers       *mappers.Loader
	currentMapper *mappers.Mapper
	server        *server.Server
	ctx           context.Context
	cancel        context.CancelFunc
}

// Config represents the simplified configuration for compatibility
type Config struct {
	Port           int
	RetroArchHost  string
	RetroArchPort  int
	UpdateInterval time.Duration
	RequestTimeout time.Duration
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "gamehook",
		Short: "GameHook - Retro game memory manipulation server",
		Long: `GameHook is a modern retro game memory manipulation tool that connects to
emulators like RetroArch to read and modify game memory in real-time.

By default, this starts the web server. Use subcommands for testing and utilities.`,
		Version: Version,
		RunE:    runGameHook,
	}

	// Server flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.Flags().String("host", "0.0.0.0", "server host")
	rootCmd.Flags().Int("port", 8080, "server port")
	rootCmd.Flags().String("retroarch-host", "127.0.0.1", "RetroArch host")
	rootCmd.Flags().Int("retroarch-port", 55355, "RetroArch port")
	rootCmd.Flags().Duration("update-interval", 5*time.Millisecond, "memory update interval")
	rootCmd.Flags().Duration("request-timeout", 64*time.Millisecond, "request timeout")
	rootCmd.Flags().String("mappers-dir", "./mappers", "mappers directory")
	rootCmd.Flags().String("uis-dir", "./uis", "UIs directory")

	// Add utility subcommands for testing and debugging
	rootCmd.AddCommand(createTestCommands()...)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// createTestCommands creates utility commands for testing and debugging
func createTestCommands() []*cobra.Command {
	// Test command
	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Test RetroArch connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _ := cmd.Root().Flags().GetString("retroarch-host")
			port, _ := cmd.Root().Flags().GetInt("retroarch-port")
			timeout, _ := cmd.Root().Flags().GetDuration("request-timeout")

			driver := drivers.NewRetroArchDriver(host, port, timeout)

			fmt.Printf("Testing connection to RetroArch at %s:%d...\n", host, port)

			if err := driver.Connect(); err != nil {
				return fmt.Errorf("connection failed: %w", err)
			}
			defer driver.Close()

			fmt.Println("✓ Connection successful")

			// Test reading a small amount of memory
			fmt.Println("Testing memory read...")
			data, err := driver.ReadMemory(0x0000, 16)
			if err != nil {
				return fmt.Errorf("memory read failed: %w", err)
			}

			fmt.Printf("✓ Read 16 bytes from 0x0000: %x\n", data)
			return nil
		},
	}

	// Mappers command
	mappersCmd := &cobra.Command{
		Use:   "mappers",
		Short: "Mapper utilities",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available mappers",
		RunE: func(cmd *cobra.Command, args []string) error {
			mappersDir, _ := cmd.Root().Flags().GetString("mappers-dir")
			loader := mappers.NewLoader(mappersDir)
			mapperList := loader.List()

			fmt.Printf("Available mappers in %s:\n", mappersDir)
			if len(mapperList) == 0 {
				fmt.Println("  (none found)")
			} else {
				for _, name := range mapperList {
					fmt.Printf("  %s\n", name)
				}
			}
			return nil
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate [mapper-name]",
		Short: "Validate mapper CUE files",
		RunE: func(cmd *cobra.Command, args []string) error {
			mappersDir, _ := cmd.Root().Flags().GetString("mappers-dir")
			loader := mappers.NewLoader(mappersDir)

			if len(args) == 0 {
				// Validate all mappers
				mapperList := loader.List()
				if len(mapperList) == 0 {
					fmt.Printf("No mappers found in %s\n", mappersDir)
					return nil
				}

				fmt.Printf("Validating %d mappers...\n", len(mapperList))
				valid := 0
				invalid := 0

				for _, name := range mapperList {
					_, err := loader.Load(name)
					if err != nil {
						fmt.Printf("  ✗ %s: %v\n", name, err)
						invalid++
					} else {
						fmt.Printf("  ✓ %s\n", name)
						valid++
					}
				}

				fmt.Printf("\nResults: %d valid, %d invalid\n", valid, invalid)
				if invalid > 0 {
					return fmt.Errorf("%d mappers failed validation", invalid)
				}
			} else {
				// Validate specific mapper
				mapperName := args[0]
				_, err := loader.Load(mapperName)
				if err != nil {
					return fmt.Errorf("validation failed: %w", err)
				}
				fmt.Printf("✓ Mapper %s is valid\n", mapperName)
			}

			return nil
		},
	}

	mappersCmd.AddCommand(listCmd, validateCmd)

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GameHook v%s\n", Version)
			fmt.Printf("Go version: %s\n", strings.TrimPrefix(runtime.Version(), "go"))
			fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}

	return []*cobra.Command{testCmd, mappersCmd, versionCmd}
}

func runGameHook(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override with command line flags
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

	// Create and start GameHook server
	gameHook, err := NewGameHook(cfg)
	if err != nil {
		return fmt.Errorf("failed to create GameHook: %w", err)
	}

	log.Printf("🎮 Starting GameHook v%s", Version)
	log.Printf("🌐 Web server: http://localhost:%d", cfg.Server.Port)
	log.Printf("🎯 RetroArch: %s:%d", cfg.RetroArch.Host, cfg.RetroArch.Port)
	log.Printf("📁 Mappers: %s", cfg.Paths.MappersDir)
	log.Printf("🎨 UIs: %s", cfg.Paths.UIsDir)
	log.Printf("")
	log.Printf("🚀 Ready! Open http://localhost:%d in your browser", cfg.Server.Port)
	log.Printf("📚 API docs at http://localhost:%d/api", cfg.Server.Port)
	log.Printf("⚙️  Use 'gamehook test' to verify RetroArch connection")
	log.Printf("")

	return gameHook.Run()
}

// NewGameHook creates a new GameHook instance
func NewGameHook(cfg *config.Config) (*GameHook, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create driver
	driver := drivers.NewRetroArchDriver(
		cfg.RetroArch.Host,
		cfg.RetroArch.Port,
		cfg.RetroArch.RequestTimeout,
	)

	// Create memory manager
	memoryManager := memory.NewManager()

	// Create mappers loader
	mappersLoader := mappers.NewLoader(cfg.Paths.MappersDir)

	gameHook := &GameHook{
		config:  cfg,
		driver:  driver,
		memory:  memoryManager,
		mappers: mappersLoader,
		ctx:     ctx,
		cancel:  cancel,
	}

	// Create server
	gameHook.server = server.New(gameHook, cfg.Paths.UIsDir, cfg.Server.Port)

	return gameHook, nil
}

// Run starts the GameHook application
func (gh *GameHook) Run() error {
	// Start update loop (only when a mapper is loaded)
	go gh.updateLoop()

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
	log.Println("🛑 Shutting down GameHook...")
	gh.cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := gh.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️  Server shutdown error: %v", err)
	}

	if err := gh.driver.Close(); err != nil {
		log.Printf("⚠️  Driver close error: %v", err)
	}

	log.Println("✅ Shutdown complete")
	return nil
}

// updateLoop continuously updates memory from the driver
func (gh *GameHook) updateLoop() {
	ticker := time.NewTicker(gh.config.Performance.UpdateInterval)
	defer ticker.Stop()

	lastErrorLog := time.Time{}
	consecutiveErrors := 0

	for {
		select {
		case <-gh.ctx.Done():
			return
		case <-ticker.C:
			if gh.currentMapper != nil {
				if err := gh.updateMemory(); err != nil {
					consecutiveErrors++

					// Only log errors occasionally to avoid spam, but log first error immediately
					if consecutiveErrors == 1 || time.Since(lastErrorLog) > 30*time.Second {
						log.Printf("⚠️  Memory update error (RetroArch connected? Game loaded?): %v", err)
						lastErrorLog = time.Now()
					}
				} else {
					// Reset error counter on successful read
					if consecutiveErrors > 0 {
						log.Printf("✅ RetroArch connection restored")
						consecutiveErrors = 0
					}
				}
			}
		}
	}
}

// updateMemory reads memory from the driver and updates the memory manager
func (gh *GameHook) updateMemory() error {
	if gh.currentMapper == nil {
		return nil // No mapper loaded, nothing to do
	}

	// The new driver handles connection internally, so we don't need to connect/close each time
	memoryData, err := gh.driver.ReadMemoryBlocks(gh.currentMapper.Platform.MemoryBlocks)
	if err != nil {
		return fmt.Errorf("memory read failed: %w", err)
	}

	gh.memory.Update(memoryData)

	// Optionally log successful reads (remove this in production)
	// log.Printf("Successfully read %d memory blocks", len(memoryData))

	return nil
}

// GameHookAPI implementation

func (gh *GameHook) LoadMapper(name string) error {
	mapper, err := gh.mappers.Load(name)
	if err != nil {
		return err
	}

	gh.currentMapper = mapper
	log.Printf("📍 Loaded mapper: %s (%s)", mapper.Name, mapper.Game)
	log.Printf("🎮 Platform: %s (%s endian)", mapper.Platform.Name, mapper.Platform.Endian)
	log.Printf("📊 Properties: %d defined", len(mapper.Properties))
	return nil
}

// Update the GameHook implementation
func (gh *GameHook) GetCurrentMapperFull() *mappers.Mapper {
	return gh.currentMapper // Return the actual mapper object
}

// Keep the existing method for simple info (backward compatibility)
func (gh *GameHook) GetCurrentMapper() interface{} {
	if gh.currentMapper == nil {
		return nil
	}

	return map[string]interface{}{
		"name":     gh.currentMapper.Name,
		"game":     gh.currentMapper.Game,
		"platform": gh.currentMapper.Platform.Name,
	}
}

func (gh *GameHook) GetProperty(name string) (interface{}, error) {
	if gh.currentMapper == nil {
		return nil, fmt.Errorf("no mapper loaded")
	}

	return gh.currentMapper.GetProperty(name, gh.memory)
}

func (gh *GameHook) SetProperty(name string, value interface{}) error {
	if gh.currentMapper == nil {
		return fmt.Errorf("no mapper loaded")
	}

	return gh.currentMapper.SetProperty(name, value, gh.memory, gh.driver)
}

func (gh *GameHook) ListMappers() []string {
	return gh.mappers.List()
}
