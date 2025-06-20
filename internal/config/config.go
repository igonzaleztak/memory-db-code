package config

import (
	"fmt"
	"memorydb/internal/enums"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Common configuration
	Verbose enums.VerboseLevel `mapstructure:"VERBOSE" validate:"required"`

	// API configuration
	ApiVersion string `mapstructure:"API_VERSION" validate:"required"`
	Port       *int   `mapstructure:"PORT" validate:"required"`
	HealthPort *int   `mapstructure:"HEALTH_PORT" validate:"required"`

	// Database configuration
	DefaultTTL             time.Duration `mapstructure:"DEFAULT_TTL" validate:"required"`
	DefaultCleanupInterval time.Duration `mapstructure:"DEFAULT_CLEANUP_INTERVAL" validate:"required"`
	PersistenceEnabled     bool          `mapstructure:"PERSISTENCE_ENABLED"`
	DBPath                 string        `mapstructure:"DB_PATH"` // Optional field that indicates the path where the database is stored
}

func (c *Config) SetDefaults() {
	viper.SetDefault("VERBOSE", enums.VerboseLevelInfo.String())
	viper.SetDefault("API_VERSION", "v1")
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("HEALTH_PORT", 8081)
	viper.SetDefault("DEFAULT_TTL", 5*time.Minute)
	viper.SetDefault("DEFAULT_CLEANUP_INTERVAL", 10*time.Minute)
	viper.SetDefault("PERSISTENCE_ENABLED", false)
	viper.SetDefault("DB_PATH", "/tmp/memorydb.db") // Default path for the database file
}

// LoadConfig loads the configuration from environment variables and sets defaults.
func LoadConfig() (*Config, error) {
	cfg := new(Config)
	cfg.SetDefaults()

	viper.AutomaticEnv() // Automatically read environment variables
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err // Return error if unmarshalling fails
	}

	// Validate the configuration
	if !cfg.Verbose.IsValid() {
		return nil, fmt.Errorf("invalid verbose level: %s", cfg.Verbose)
	}

	if cfg.PersistenceEnabled && cfg.DBPath == "" {
		return nil, fmt.Errorf("DB_PATH must be set when persistence is enabled")
	}

	return cfg, nil
}
