package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the enhanced application configuration
type Config struct {
	// Core configuration (existing)
	Server      ServerConfig      `mapstructure:"server"`
	RetroArch   RetroArchConfig   `mapstructure:"retroarch"`
	BizHawk     BizHawkConfig     `mapstructure:"bizhawk"`
	Paths       PathsConfig       `mapstructure:"paths"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Features    FeaturesConfig    `mapstructure:"features"`
	Security    SecurityConfig    `mapstructure:"security"`

	// Enhanced configuration sections
	UI                 UIConfig                 `mapstructure:"ui"`
	PropertyMonitoring PropertyMonitoringConfig `mapstructure:"property_monitoring"`
	BatchOperations    BatchOperationsConfig    `mapstructure:"batch_operations"`
	Validation         ValidationConfig         `mapstructure:"validation"`
	Events             EventsConfig             `mapstructure:"events"`
	Memory             AdvancedMemoryConfig     `mapstructure:"memory"`
	Metrics            MetricsConfig            `mapstructure:"metrics"`
	Database           DatabaseConfig           `mapstructure:"database"`
}

// Core configuration types (existing)
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type RetroArchConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
	MaxRetries     int           `mapstructure:"max_retries"`
	RetryDelay     time.Duration `mapstructure:"retry_delay"`
}

type BizHawkConfig struct {
	MemoryMapName string        `mapstructure:"memory_map_name"`
	DataMapName   string        `mapstructure:"data_map_name"`
	Timeout       time.Duration `mapstructure:"timeout"`
}

type PathsConfig struct {
	MappersDir string `mapstructure:"mappers_dir"`
	UIsDir     string `mapstructure:"uis_dir"`
	DataDir    string `mapstructure:"data_dir"`
	LogDir     string `mapstructure:"log_dir"`
	CacheDir   string `mapstructure:"cache_dir"`
}

type PerformanceConfig struct {
	UpdateInterval   time.Duration `mapstructure:"update_interval"`
	MaxClients       int           `mapstructure:"max_clients"`
	MemoryBufferSize int           `mapstructure:"memory_buffer_size"`
	WebSocketBuffer  int           `mapstructure:"websocket_buffer"`
	GCTargetPercent  int           `mapstructure:"gc_target_percent"`
	MaxMemoryUsageMB int           `mapstructure:"max_memory_usage_mb"`
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type FeaturesConfig struct {
	Metrics           bool `mapstructure:"metrics"`
	Profiling         bool `mapstructure:"profiling"`
	AutoMapperReload  bool `mapstructure:"auto_mapper_reload"`
	CacheProperties   bool `mapstructure:"cache_properties"`
	BackgroundSave    bool `mapstructure:"background_save"`
	MemoryCompression bool `mapstructure:"memory_compression"`

	// Enhanced features
	AdvancedPropertyTypes bool `mapstructure:"advanced_property_types"`
	ComputedProperties    bool `mapstructure:"computed_properties"`
	PropertyFreezing      bool `mapstructure:"property_freezing"`
	RealTimeValidation    bool `mapstructure:"real_time_validation"`
	WebSocketStreaming    bool `mapstructure:"websocket_streaming"`
}

type SecurityConfig struct {
	EnableCORS              bool            `mapstructure:"enable_cors"`
	AllowedOrigins          []string        `mapstructure:"allowed_origins"`
	APIKeys                 []string        `mapstructure:"api_keys"`
	RateLimit               RateLimitConfig `mapstructure:"rate_limit"`
	MaxWebSocketConnections int             `mapstructure:"max_websocket_connections"`
	EnableRequestLogging    bool            `mapstructure:"enable_request_logging"`
}

type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerSecond int           `mapstructure:"requests_per_second"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// Enhanced configuration types
type UIConfig struct {
	EnableAdminPanel     bool   `mapstructure:"enable_admin_panel"`
	EnableLegacyUI       bool   `mapstructure:"enable_legacy_ui"`
	DefaultTheme         string `mapstructure:"default_theme"`
	EnablePropertyEditor bool   `mapstructure:"enable_property_editor"`
	EnableMemoryViewer   bool   `mapstructure:"enable_memory_viewer"`
	AutoRefreshInterval  string `mapstructure:"auto_refresh_interval"`
	MaxPropertyHistory   int    `mapstructure:"max_property_history"`
}

type PropertyMonitoringConfig struct {
	UpdateInterval        time.Duration `mapstructure:"update_interval"`
	EnableStatistics      bool          `mapstructure:"enable_statistics"`
	HistorySize           int           `mapstructure:"history_size"`
	ChangeThreshold       float64       `mapstructure:"change_threshold"`
	EnableChangeDetection bool          `mapstructure:"enable_change_detection"`
	BatchChangeEvents     bool          `mapstructure:"batch_change_events"`
	MaxEventsPerBatch     int           `mapstructure:"max_events_per_batch"`
}

type BatchOperationsConfig struct {
	MaxBatchSize      int           `mapstructure:"max_batch_size"`
	Timeout           time.Duration `mapstructure:"timeout"`
	EnableAtomic      bool          `mapstructure:"enable_atomic"`
	ParallelExecution bool          `mapstructure:"parallel_execution"`
	ValidationMode    string        `mapstructure:"validation_mode"`
}

type ValidationConfig struct {
	EnableStrict            bool          `mapstructure:"enable_strict"`
	LogValidation           bool          `mapstructure:"log_validation"`
	FailOnError             bool          `mapstructure:"fail_on_error"`
	CacheValidationResults  bool          `mapstructure:"cache_validation_results"`
	ValidationTimeout       time.Duration `mapstructure:"validation_timeout"`
	CustomValidatorsEnabled bool          `mapstructure:"custom_validators_enabled"`
}

type EventsConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	MaxEventHistory    int           `mapstructure:"max_event_history"`
	EventBatchSize     int           `mapstructure:"event_batch_size"`
	ProcessingTimeout  time.Duration `mapstructure:"processing_timeout"`
	EnableCustomEvents bool          `mapstructure:"enable_custom_events"`
	LogEventTriggers   bool          `mapstructure:"log_event_triggers"`
}

type AdvancedMemoryConfig struct {
	EnableCompression    bool   `mapstructure:"enable_compression"`
	CompressionAlgorithm string `mapstructure:"compression_algorithm"`
	CacheSizeMB          int    `mapstructure:"cache_size_mb"`
	EnableMemoryMapping  bool   `mapstructure:"enable_memory_mapping"`
	PrefetchEnabled      bool   `mapstructure:"prefetch_enabled"`
	MemoryAlignment      int    `mapstructure:"memory_alignment"`
}

type MetricsConfig struct {
	Enabled              bool          `mapstructure:"enabled"`
	Endpoint             string        `mapstructure:"endpoint"`
	IncludeSystemMetrics bool          `mapstructure:"include_system_metrics"`
	IncludeGoMetrics     bool          `mapstructure:"include_go_metrics"`
	IncludeCustomMetrics bool          `mapstructure:"include_custom_metrics"`
	ExportInterval       time.Duration `mapstructure:"export_interval"`
}

type DatabaseConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	Driver           string `mapstructure:"driver"`
	ConnectionString string `mapstructure:"connection_string"`
	MaxConnections   int    `mapstructure:"max_connections"`
	MigrationEnabled bool   `mapstructure:"migration_enabled"`
}

// DefaultConfig returns a configuration with enhanced sensible defaults
func DefaultConfig() *Config {
	return &Config{
		// Core configuration defaults
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		RetroArch: RetroArchConfig{
			Host:           "127.0.0.1",
			Port:           55355,
			RequestTimeout: 64 * time.Millisecond,
			MaxRetries:     3,
			RetryDelay:     100 * time.Millisecond,
		},
		BizHawk: BizHawkConfig{
			MemoryMapName: "GAMEHOOK_BIZHAWK.bin",
			DataMapName:   "GAMEHOOK_BIZHAWK_DATA.bin",
			Timeout:       1 * time.Second,
		},
		Paths: PathsConfig{
			MappersDir: "./mappers",
			UIsDir:     "./uis",
			DataDir:    "./data",
			LogDir:     "./logs",
			CacheDir:   "./cache",
		},
		Performance: PerformanceConfig{
			UpdateInterval:   5 * time.Millisecond,
			MaxClients:       100,
			MemoryBufferSize: 1024 * 1024, // 1MB
			WebSocketBuffer:  256,
			GCTargetPercent:  100,
			MaxMemoryUsageMB: 512,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100, // MB
			MaxAge:     7,   // days
			MaxBackups: 3,
			Compress:   true,
		},
		Features: FeaturesConfig{
			Metrics:               true,
			Profiling:             false,
			AutoMapperReload:      true,
			CacheProperties:       true,
			BackgroundSave:        false,
			MemoryCompression:     false,
			AdvancedPropertyTypes: true,
			ComputedProperties:    true,
			PropertyFreezing:      true,
			RealTimeValidation:    true,
			WebSocketStreaming:    true,
		},
		Security: SecurityConfig{
			EnableCORS:              true,
			AllowedOrigins:          []string{"*"},
			APIKeys:                 []string{},
			MaxWebSocketConnections: 50,
			EnableRequestLogging:    false,
			RateLimit: RateLimitConfig{
				Enabled:           false,
				RequestsPerSecond: 100,
				BurstSize:         20,
				CleanupInterval:   1 * time.Minute,
			},
		},

		// Enhanced configuration defaults
		UI: UIConfig{
			EnableAdminPanel:     true,
			EnableLegacyUI:       true,
			DefaultTheme:         "dark",
			EnablePropertyEditor: true,
			EnableMemoryViewer:   true,
			AutoRefreshInterval:  "1s",
			MaxPropertyHistory:   100,
		},

		PropertyMonitoring: PropertyMonitoringConfig{
			UpdateInterval:        16 * time.Millisecond, // 60 FPS
			EnableStatistics:      true,
			HistorySize:           100000,
			ChangeThreshold:       0.01,
			EnableChangeDetection: true,
			BatchChangeEvents:     true,
			MaxEventsPerBatch:     50,
		},

		BatchOperations: BatchOperationsConfig{
			MaxBatchSize:      50,
			Timeout:           5 * time.Second,
			EnableAtomic:      true,
			ParallelExecution: false,
			ValidationMode:    "strict", // "strict", "warn", "ignore"
		},

		Validation: ValidationConfig{
			EnableStrict:            true,
			LogValidation:           true,
			FailOnError:             false,
			CacheValidationResults:  true,
			ValidationTimeout:       1 * time.Second,
			CustomValidatorsEnabled: true,
		},

		Events: EventsConfig{
			Enabled:            true,
			MaxEventHistory:    1000,
			EventBatchSize:     10,
			ProcessingTimeout:  5 * time.Second,
			EnableCustomEvents: true,
			LogEventTriggers:   true,
		},

		Memory: AdvancedMemoryConfig{
			EnableCompression:    false,
			CompressionAlgorithm: "gzip", // gzip, lz4, snappy
			CacheSizeMB:          64,
			EnableMemoryMapping:  false,
			PrefetchEnabled:      true,
			MemoryAlignment:      4,
		},

		Metrics: MetricsConfig{
			Enabled:              true,
			Endpoint:             "/metrics",
			IncludeSystemMetrics: true,
			IncludeGoMetrics:     true,
			IncludeCustomMetrics: true,
			ExportInterval:       30 * time.Second,
		},

		Database: DatabaseConfig{
			Enabled:          false,
			Driver:           "sqlite",
			ConnectionString: "./data/gamehook.db",
			MaxConnections:   10,
			MigrationEnabled: true,
		},
	}
}

// LoadConfig loads enhanced configuration from files and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	v := viper.New()

	// Set default values
	setDefaults(v, config)

	// Configure viper
	v.SetConfigName("gamehook")
	v.SetConfigType("yaml")

	// Add config paths
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("$HOME/.gamehook")
	v.AddConfigPath("/etc/gamehook")

	// Environment variables
	v.SetEnvPrefix("GAMEHOOK")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is not an error, we'll use defaults
	}

	// Unmarshal to struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate and normalize paths
	if err := normalizeConfig(config); err != nil {
		return nil, fmt.Errorf("error normalizing config: %w", err)
	}

	return config, nil
}

// setDefaults sets all default values in viper (enhanced)
func setDefaults(v *viper.Viper, config *Config) {
	// Core configuration defaults
	v.SetDefault("server.host", config.Server.Host)
	v.SetDefault("server.port", config.Server.Port)
	v.SetDefault("server.read_timeout", config.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", config.Server.WriteTimeout)
	v.SetDefault("server.idle_timeout", config.Server.IdleTimeout)
	v.SetDefault("server.tls.enabled", config.Server.TLS.Enabled)

	v.SetDefault("retroarch.host", config.RetroArch.Host)
	v.SetDefault("retroarch.port", config.RetroArch.Port)
	v.SetDefault("retroarch.request_timeout", config.RetroArch.RequestTimeout)
	v.SetDefault("retroarch.max_retries", config.RetroArch.MaxRetries)
	v.SetDefault("retroarch.retry_delay", config.RetroArch.RetryDelay)

	v.SetDefault("bizhawk.memory_map_name", config.BizHawk.MemoryMapName)
	v.SetDefault("bizhawk.data_map_name", config.BizHawk.DataMapName)
	v.SetDefault("bizhawk.timeout", config.BizHawk.Timeout)

	v.SetDefault("paths.mappers_dir", config.Paths.MappersDir)
	v.SetDefault("paths.uis_dir", config.Paths.UIsDir)
	v.SetDefault("paths.data_dir", config.Paths.DataDir)
	v.SetDefault("paths.log_dir", config.Paths.LogDir)
	v.SetDefault("paths.cache_dir", config.Paths.CacheDir)

	v.SetDefault("performance.update_interval", config.Performance.UpdateInterval)
	v.SetDefault("performance.max_clients", config.Performance.MaxClients)
	v.SetDefault("performance.memory_buffer_size", config.Performance.MemoryBufferSize)
	v.SetDefault("performance.websocket_buffer", config.Performance.WebSocketBuffer)
	v.SetDefault("performance.gc_target_percent", config.Performance.GCTargetPercent)
	v.SetDefault("performance.max_memory_usage_mb", config.Performance.MaxMemoryUsageMB)

	v.SetDefault("logging.level", config.Logging.Level)
	v.SetDefault("logging.format", config.Logging.Format)
	v.SetDefault("logging.output", config.Logging.Output)
	v.SetDefault("logging.max_size", config.Logging.MaxSize)
	v.SetDefault("logging.max_age", config.Logging.MaxAge)
	v.SetDefault("logging.max_backups", config.Logging.MaxBackups)
	v.SetDefault("logging.compress", config.Logging.Compress)

	v.SetDefault("features.metrics", config.Features.Metrics)
	v.SetDefault("features.profiling", config.Features.Profiling)
	v.SetDefault("features.auto_mapper_reload", config.Features.AutoMapperReload)
	v.SetDefault("features.cache_properties", config.Features.CacheProperties)
	v.SetDefault("features.background_save", config.Features.BackgroundSave)
	v.SetDefault("features.memory_compression", config.Features.MemoryCompression)
	v.SetDefault("features.advanced_property_types", config.Features.AdvancedPropertyTypes)
	v.SetDefault("features.computed_properties", config.Features.ComputedProperties)
	v.SetDefault("features.property_freezing", config.Features.PropertyFreezing)
	v.SetDefault("features.real_time_validation", config.Features.RealTimeValidation)
	v.SetDefault("features.websocket_streaming", config.Features.WebSocketStreaming)

	v.SetDefault("security.enable_cors", config.Security.EnableCORS)
	v.SetDefault("security.allowed_origins", config.Security.AllowedOrigins)
	v.SetDefault("security.api_keys", config.Security.APIKeys)
	v.SetDefault("security.max_websocket_connections", config.Security.MaxWebSocketConnections)
	v.SetDefault("security.enable_request_logging", config.Security.EnableRequestLogging)
	v.SetDefault("security.rate_limit.enabled", config.Security.RateLimit.Enabled)
	v.SetDefault("security.rate_limit.requests_per_second", config.Security.RateLimit.RequestsPerSecond)
	v.SetDefault("security.rate_limit.burst_size", config.Security.RateLimit.BurstSize)
	v.SetDefault("security.rate_limit.cleanup_interval", config.Security.RateLimit.CleanupInterval)

	// Enhanced configuration defaults
	v.SetDefault("ui.enable_admin_panel", config.UI.EnableAdminPanel)
	v.SetDefault("ui.enable_legacy_ui", config.UI.EnableLegacyUI)
	v.SetDefault("ui.default_theme", config.UI.DefaultTheme)
	v.SetDefault("ui.enable_property_editor", config.UI.EnablePropertyEditor)
	v.SetDefault("ui.enable_memory_viewer", config.UI.EnableMemoryViewer)
	v.SetDefault("ui.auto_refresh_interval", config.UI.AutoRefreshInterval)
	v.SetDefault("ui.max_property_history", config.UI.MaxPropertyHistory)

	v.SetDefault("property_monitoring.update_interval", config.PropertyMonitoring.UpdateInterval)
	v.SetDefault("property_monitoring.enable_statistics", config.PropertyMonitoring.EnableStatistics)
	v.SetDefault("property_monitoring.history_size", config.PropertyMonitoring.HistorySize)
	v.SetDefault("property_monitoring.change_threshold", config.PropertyMonitoring.ChangeThreshold)
	v.SetDefault("property_monitoring.enable_change_detection", config.PropertyMonitoring.EnableChangeDetection)
	v.SetDefault("property_monitoring.batch_change_events", config.PropertyMonitoring.BatchChangeEvents)
	v.SetDefault("property_monitoring.max_events_per_batch", config.PropertyMonitoring.MaxEventsPerBatch)

	v.SetDefault("batch_operations.max_batch_size", config.BatchOperations.MaxBatchSize)
	v.SetDefault("batch_operations.timeout", config.BatchOperations.Timeout)
	v.SetDefault("batch_operations.enable_atomic", config.BatchOperations.EnableAtomic)
	v.SetDefault("batch_operations.parallel_execution", config.BatchOperations.ParallelExecution)
	v.SetDefault("batch_operations.validation_mode", config.BatchOperations.ValidationMode)

	v.SetDefault("validation.enable_strict", config.Validation.EnableStrict)
	v.SetDefault("validation.log_validation", config.Validation.LogValidation)
	v.SetDefault("validation.fail_on_error", config.Validation.FailOnError)
	v.SetDefault("validation.cache_validation_results", config.Validation.CacheValidationResults)
	v.SetDefault("validation.validation_timeout", config.Validation.ValidationTimeout)
	v.SetDefault("validation.custom_validators_enabled", config.Validation.CustomValidatorsEnabled)

	v.SetDefault("events.enabled", config.Events.Enabled)
	v.SetDefault("events.max_event_history", config.Events.MaxEventHistory)
	v.SetDefault("events.event_batch_size", config.Events.EventBatchSize)
	v.SetDefault("events.processing_timeout", config.Events.ProcessingTimeout)
	v.SetDefault("events.enable_custom_events", config.Events.EnableCustomEvents)
	v.SetDefault("events.log_event_triggers", config.Events.LogEventTriggers)

	v.SetDefault("memory.enable_compression", config.Memory.EnableCompression)
	v.SetDefault("memory.compression_algorithm", config.Memory.CompressionAlgorithm)
	v.SetDefault("memory.cache_size_mb", config.Memory.CacheSizeMB)
	v.SetDefault("memory.enable_memory_mapping", config.Memory.EnableMemoryMapping)
	v.SetDefault("memory.prefetch_enabled", config.Memory.PrefetchEnabled)
	v.SetDefault("memory.memory_alignment", config.Memory.MemoryAlignment)

	v.SetDefault("metrics.enabled", config.Metrics.Enabled)
	v.SetDefault("metrics.endpoint", config.Metrics.Endpoint)
	v.SetDefault("metrics.include_system_metrics", config.Metrics.IncludeSystemMetrics)
	v.SetDefault("metrics.include_go_metrics", config.Metrics.IncludeGoMetrics)
	v.SetDefault("metrics.include_custom_metrics", config.Metrics.IncludeCustomMetrics)
	v.SetDefault("metrics.export_interval", config.Metrics.ExportInterval)

	v.SetDefault("database.enabled", config.Database.Enabled)
	v.SetDefault("database.driver", config.Database.Driver)
	v.SetDefault("database.connection_string", config.Database.ConnectionString)
	v.SetDefault("database.max_connections", config.Database.MaxConnections)
	v.SetDefault("database.migration_enabled", config.Database.MigrationEnabled)
}

// normalizeConfig validates and normalizes enhanced configuration values
func normalizeConfig(config *Config) error {
	// Convert relative paths to absolute paths
	var err error

	if config.Paths.MappersDir, err = filepath.Abs(config.Paths.MappersDir); err != nil {
		return fmt.Errorf("invalid mappers directory: %w", err)
	}

	if config.Paths.UIsDir, err = filepath.Abs(config.Paths.UIsDir); err != nil {
		return fmt.Errorf("invalid UIs directory: %w", err)
	}

	if config.Paths.DataDir, err = filepath.Abs(config.Paths.DataDir); err != nil {
		return fmt.Errorf("invalid data directory: %w", err)
	}

	if config.Paths.LogDir, err = filepath.Abs(config.Paths.LogDir); err != nil {
		return fmt.Errorf("invalid log directory: %w", err)
	}

	if config.Paths.CacheDir, err = filepath.Abs(config.Paths.CacheDir); err != nil {
		return fmt.Errorf("invalid cache directory: %w", err)
	}

	// Validate port ranges
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.RetroArch.Port < 1 || config.RetroArch.Port > 65535 {
		return fmt.Errorf("invalid RetroArch port: %d", config.RetroArch.Port)
	}

	// Validate timeouts
	if config.Performance.UpdateInterval < time.Millisecond {
		return fmt.Errorf("update interval too small: %v", config.Performance.UpdateInterval)
	}

	if config.RetroArch.RequestTimeout < time.Millisecond {
		return fmt.Errorf("RetroArch request timeout too small: %v", config.RetroArch.RequestTimeout)
	}

	// Validate enhanced configuration
	if config.PropertyMonitoring.UpdateInterval < time.Millisecond {
		return fmt.Errorf("property monitoring update interval too small: %v", config.PropertyMonitoring.UpdateInterval)
	}

	if config.BatchOperations.MaxBatchSize < 1 {
		return fmt.Errorf("batch operations max batch size must be at least 1: %d", config.BatchOperations.MaxBatchSize)
	}

	if config.Events.MaxEventHistory < 0 {
		return fmt.Errorf("events max event history cannot be negative: %d", config.Events.MaxEventHistory)
	}

	if config.Memory.CacheSizeMB < 0 {
		return fmt.Errorf("memory cache size cannot be negative: %d", config.Memory.CacheSizeMB)
	}

	if config.UI.MaxPropertyHistory < 0 {
		return fmt.Errorf("UI max property history cannot be negative: %d", config.UI.MaxPropertyHistory)
	}

	// Validate validation mode
	validModes := []string{"strict", "warn", "ignore"}
	if !contains(validModes, config.BatchOperations.ValidationMode) {
		return fmt.Errorf("invalid batch operations validation mode: %s", config.BatchOperations.ValidationMode)
	}

	// Validate theme
	validThemes := []string{"dark", "light", "retro"}
	if !contains(validThemes, config.UI.DefaultTheme) {
		return fmt.Errorf("invalid UI default theme: %s", config.UI.DefaultTheme)
	}

	// Validate compression algorithm
	validAlgorithms := []string{"gzip", "lz4", "snappy"}
	if config.Memory.EnableCompression && !contains(validAlgorithms, config.Memory.CompressionAlgorithm) {
		return fmt.Errorf("invalid memory compression algorithm: %s", config.Memory.CompressionAlgorithm)
	}

	// Validate database driver
	if config.Database.Enabled {
		validDrivers := []string{"sqlite", "mysql", "postgres"}
		if !contains(validDrivers, config.Database.Driver) {
			return fmt.Errorf("invalid database driver: %s", config.Database.Driver)
		}
	}

	// Validate logging level
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if !contains(validLevels, config.Logging.Level) {
		return fmt.Errorf("invalid logging level: %s", config.Logging.Level)
	}

	// Create directories if they don't exist
	dirs := []string{
		config.Paths.MappersDir,
		config.Paths.UIsDir,
		config.Paths.DataDir,
		config.Paths.LogDir,
		config.Paths.CacheDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	return nil
}

// Enhanced configuration validation helpers

// ValidateConfig performs comprehensive validation of the configuration
func ValidateConfig(config *Config) error {
	// Core validation
	if err := validateCoreConfig(config); err != nil {
		return fmt.Errorf("core configuration validation failed: %w", err)
	}

	// Enhanced feature validation
	if err := validateEnhancedConfig(config); err != nil {
		return fmt.Errorf("enhanced configuration validation failed: %w", err)
	}

	// Cross-validation between sections
	if err := validateCrossReferences(config); err != nil {
		return fmt.Errorf("cross-reference validation failed: %w", err)
	}

	return nil
}

func validateCoreConfig(config *Config) error {
	// Validate critical core settings
	if config.Server.Port == config.RetroArch.Port {
		return fmt.Errorf("server port and RetroArch port cannot be the same")
	}

	if config.Performance.MaxClients < 1 {
		return fmt.Errorf("max clients must be at least 1")
	}

	return nil
}

func validateEnhancedConfig(config *Config) error {
	// Validate enhanced features are compatible
	if config.Features.PropertyFreezing && !config.Features.AdvancedPropertyTypes {
		return fmt.Errorf("property freezing requires advanced property types to be enabled")
	}

	if config.Events.Enabled && config.Events.ProcessingTimeout < time.Millisecond {
		return fmt.Errorf("event processing timeout must be at least 1ms when events are enabled")
	}

	if config.PropertyMonitoring.BatchChangeEvents && config.PropertyMonitoring.MaxEventsPerBatch < 1 {
		return fmt.Errorf("max events per batch must be at least 1 when batch change events are enabled")
	}

	return nil
}

func validateCrossReferences(config *Config) error {
	// Validate that related settings are compatible
	if config.PropertyMonitoring.UpdateInterval > config.Performance.UpdateInterval {
		return fmt.Errorf("property monitoring update interval cannot be longer than performance update interval")
	}

	if config.Security.MaxWebSocketConnections > config.Performance.MaxClients {
		return fmt.Errorf("max WebSocket connections cannot exceed max clients")
	}

	return nil
}

// GetConfigSummary returns a summary of the current configuration for logging
func GetConfigSummary(config *Config) map[string]interface{} {
	return map[string]interface{}{
		"server_port":         config.Server.Port,
		"retroarch_host":      fmt.Sprintf("%s:%d", config.RetroArch.Host, config.RetroArch.Port),
		"update_interval":     config.Performance.UpdateInterval.String(),
		"property_monitoring": config.PropertyMonitoring.UpdateInterval.String(),
		"features_enabled":    countEnabledFeatures(config),
		"enhanced_features":   config.Features.AdvancedPropertyTypes,
		"event_system":        config.Events.Enabled,
		"validation_enabled":  config.Validation.EnableStrict,
		"ui_theme":            config.UI.DefaultTheme,
		"metrics_enabled":     config.Metrics.Enabled,
		"database_enabled":    config.Database.Enabled,
	}
}

func countEnabledFeatures(config *Config) int {
	count := 0
	if config.Features.Metrics {
		count++
	}
	if config.Features.Profiling {
		count++
	}
	if config.Features.AutoMapperReload {
		count++
	}
	if config.Features.CacheProperties {
		count++
	}
	if config.Features.BackgroundSave {
		count++
	}
	if config.Features.MemoryCompression {
		count++
	}
	if config.Features.AdvancedPropertyTypes {
		count++
	}
	if config.Features.ComputedProperties {
		count++
	}
	if config.Features.PropertyFreezing {
		count++
	}
	if config.Features.RealTimeValidation {
		count++
	}
	if config.Features.WebSocketStreaming {
		count++
	}
	return count
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
