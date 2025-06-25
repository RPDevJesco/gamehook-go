package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Server ServerConfig `mapstructure:"server"`

	// RetroArch driver configuration
	RetroArch RetroArchConfig `mapstructure:"retroarch"`

	// BizHawk driver configuration
	BizHawk BizHawkConfig `mapstructure:"bizhawk"`

	// File paths
	Paths PathsConfig `mapstructure:"paths"`

	// Performance settings
	Performance PerformanceConfig `mapstructure:"performance"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`

	// Features toggle
	Features FeaturesConfig `mapstructure:"features"`

	// Security settings
	Security SecurityConfig `mapstructure:"security"`
}

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
}

type SecurityConfig struct {
	EnableCORS     bool            `mapstructure:"enable_cors"`
	AllowedOrigins []string        `mapstructure:"allowed_origins"`
	APIKeys        []string        `mapstructure:"api_keys"`
	RateLimit      RateLimitConfig `mapstructure:"rate_limit"`
}

type RateLimitConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	RequestsPerSecond int           `mapstructure:"requests_per_second"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
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
			Metrics:           true,
			Profiling:         false,
			AutoMapperReload:  true,
			CacheProperties:   true,
			BackgroundSave:    false,
			MemoryCompression: false,
		},
		Security: SecurityConfig{
			EnableCORS:     true,
			AllowedOrigins: []string{"*"},
			APIKeys:        []string{},
			RateLimit: RateLimitConfig{
				Enabled:           false,
				RequestsPerSecond: 100,
				BurstSize:         20,
				CleanupInterval:   1 * time.Minute,
			},
		},
	}
}

// LoadConfig loads configuration from files and environment variables
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

// setDefaults sets all default values in viper
func setDefaults(v *viper.Viper, config *Config) {
	// Server
	v.SetDefault("server.host", config.Server.Host)
	v.SetDefault("server.port", config.Server.Port)
	v.SetDefault("server.read_timeout", config.Server.ReadTimeout)
	v.SetDefault("server.write_timeout", config.Server.WriteTimeout)
	v.SetDefault("server.idle_timeout", config.Server.IdleTimeout)
	v.SetDefault("server.tls.enabled", config.Server.TLS.Enabled)

	// RetroArch
	v.SetDefault("retroarch.host", config.RetroArch.Host)
	v.SetDefault("retroarch.port", config.RetroArch.Port)
	v.SetDefault("retroarch.request_timeout", config.RetroArch.RequestTimeout)
	v.SetDefault("retroarch.max_retries", config.RetroArch.MaxRetries)
	v.SetDefault("retroarch.retry_delay", config.RetroArch.RetryDelay)

	// BizHawk
	v.SetDefault("bizhawk.memory_map_name", config.BizHawk.MemoryMapName)
	v.SetDefault("bizhawk.data_map_name", config.BizHawk.DataMapName)
	v.SetDefault("bizhawk.timeout", config.BizHawk.Timeout)

	// Paths
	v.SetDefault("paths.mappers_dir", config.Paths.MappersDir)
	v.SetDefault("paths.uis_dir", config.Paths.UIsDir)
	v.SetDefault("paths.data_dir", config.Paths.DataDir)
	v.SetDefault("paths.log_dir", config.Paths.LogDir)
	v.SetDefault("paths.cache_dir", config.Paths.CacheDir)

	// Performance
	v.SetDefault("performance.update_interval", config.Performance.UpdateInterval)
	v.SetDefault("performance.max_clients", config.Performance.MaxClients)
	v.SetDefault("performance.memory_buffer_size", config.Performance.MemoryBufferSize)
	v.SetDefault("performance.websocket_buffer", config.Performance.WebSocketBuffer)
	v.SetDefault("performance.gc_target_percent", config.Performance.GCTargetPercent)
	v.SetDefault("performance.max_memory_usage_mb", config.Performance.MaxMemoryUsageMB)

	// Logging
	v.SetDefault("logging.level", config.Logging.Level)
	v.SetDefault("logging.format", config.Logging.Format)
	v.SetDefault("logging.output", config.Logging.Output)
	v.SetDefault("logging.max_size", config.Logging.MaxSize)
	v.SetDefault("logging.max_age", config.Logging.MaxAge)
	v.SetDefault("logging.max_backups", config.Logging.MaxBackups)
	v.SetDefault("logging.compress", config.Logging.Compress)

	// Features
	v.SetDefault("features.metrics", config.Features.Metrics)
	v.SetDefault("features.profiling", config.Features.Profiling)
	v.SetDefault("features.auto_mapper_reload", config.Features.AutoMapperReload)
	v.SetDefault("features.cache_properties", config.Features.CacheProperties)
	v.SetDefault("features.background_save", config.Features.BackgroundSave)
	v.SetDefault("features.memory_compression", config.Features.MemoryCompression)

	// Security
	v.SetDefault("security.enable_cors", config.Security.EnableCORS)
	v.SetDefault("security.allowed_origins", config.Security.AllowedOrigins)
	v.SetDefault("security.api_keys", config.Security.APIKeys)
	v.SetDefault("security.rate_limit.enabled", config.Security.RateLimit.Enabled)
	v.SetDefault("security.rate_limit.requests_per_second", config.Security.RateLimit.RequestsPerSecond)
	v.SetDefault("security.rate_limit.burst_size", config.Security.RateLimit.BurstSize)
	v.SetDefault("security.rate_limit.cleanup_interval", config.Security.RateLimit.CleanupInterval)
}

// normalizeConfig validates and normalizes configuration values
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

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
