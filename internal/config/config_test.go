package config_test

import (
	"memorydb/internal/config"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func (suite *ConfigSuite) TestLoadConfig() {

	suite.Run("Defaut", func() {
		// Test loading configuration with default values
		cfg, err := config.LoadConfig()
		suite.NoError(err)
		suite.NotNil(cfg)

		// Check if defaults are set correctly
		suite.Equal("v1", cfg.ApiVersion)
	})
	suite.Run("Custom env", func() {
		viper.Set("VERBOSE", "info")
		viper.Set("API_VERSION", "v2")
		viper.Set("PORT", 9090)
		viper.Set("HEALTH_PORT", 9091)
		viper.Set("DEFAULT_TTL", "10m")
		viper.Set("DEFAULT_CLEANUP_INTERVAL", "15m")

		cfg, err := config.LoadConfig()
		suite.NoError(err)
		suite.NotNil(cfg)

		suite.Equal("v2", cfg.ApiVersion)
	})

	suite.Run("Invalid env", func() {
		viper.Set("VERBOSE", "invalid_level") // Set an invalid verbose level
		_, err := config.LoadConfig()
		suite.Error(err, "Expected error when loading config with invalid verbose level")
		suite.Contains(err.Error(), "invalid verbose level")
	})

}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
