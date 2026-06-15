package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	DatabaseURL   string
	AppEnv        string
	LogLevel      string
	EmulatorCount int
	AdbHost       string
	AdbPort       int
	APIToken      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.SetDefault("PORT", "9000")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("EMULATOR_MAX_INSTANCES", 25)
	viper.SetDefault("ADB_SERVER_PORT", 5037)
	viper.SetDefault("ADB_HOST", "127.0.0.1")
	viper.AutomaticEnv()

	cfg := &Config{
		Port:          viper.GetString("PORT"),
		DatabaseURL:   viper.GetString("DATABASE_URL"),
		AppEnv:        viper.GetString("APP_ENV"),
		LogLevel:      viper.GetString("LOG_LEVEL"),
		EmulatorCount: viper.GetInt("EMULATOR_MAX_INSTANCES"),
		AdbHost:       viper.GetString("ADB_HOST"),
		AdbPort:       viper.GetInt("ADB_SERVER_PORT"),
		APIToken:      viper.GetString("ORCHESTRATOR_API_TOKEN"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.EmulatorCount < 1 || c.EmulatorCount > 50 {
		return fmt.Errorf("EMULATOR_MAX_INSTANCES must be 1-50, got %d", c.EmulatorCount)
	}
	return nil
}
