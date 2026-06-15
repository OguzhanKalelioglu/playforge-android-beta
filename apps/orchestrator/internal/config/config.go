package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Port                string
	DatabaseURL         string
	AppEnv              string
	LogLevel            string
	EmulatorCount       int
	AdbHost             string
	AdbPort             int
	APIToken            string
	ComposePath         string
	ComposeProject      string
	ServicePrefix       string
	AutoStartEmul       bool
	CheckInterval       int

	// Task runner
	AppiumURL           string
	ActivityAPIURL      string
	ActivityAPIToken    string
	ActivityFallbackDir string
	TaskWatchdogMin     int
	ScreenshotDir       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault("PORT", "9000")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("EMULATOR_MAX_INSTANCES", 25)
	viper.SetDefault("ADB_SERVER_PORT", 5037)
	viper.SetDefault("ADB_HOST", "127.0.0.1")
	viper.SetDefault("COMPOSE_PROJECT", "testers-minipc")
	viper.SetDefault("COMPOSE_PATH", "../../infra/minipc/docker-compose.yml")
	viper.SetDefault("SERVICE_PREFIX", "emulator")
	viper.SetDefault("AUTO_START_EMULATORS", true)
	viper.SetDefault("EMULATOR_CHECK_INTERVAL", 60)

	// Task runner defaults
	viper.SetDefault("APPIUM_URL", "http://127.0.0.1:4723")
	viper.SetDefault("ACTIVITY_API_URL", "")
	viper.SetDefault("ACTIVITY_API_TOKEN", "")
	viper.SetDefault("ACTIVITY_FALLBACK_DIR", "./activity-fallback")
	viper.SetDefault("TASK_WATCHDOG_MIN", 10)
	viper.SetDefault("SCREENSHOT_DIR", "./screenshots")
	viper.AutomaticEnv()

	cfg := &Config{
		Port:               viper.GetString("PORT"),
		DatabaseURL:        viper.GetString("DATABASE_URL"),
		AppEnv:             viper.GetString("APP_ENV"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		EmulatorCount:      viper.GetInt("EMULATOR_MAX_INSTANCES"),
		AdbHost:            viper.GetString("ADB_HOST"),
		AdbPort:            viper.GetInt("ADB_SERVER_PORT"),
		APIToken:           viper.GetString("ORCHESTRATOR_API_TOKEN"),
		ComposePath:        viper.GetString("COMPOSE_PATH"),
		ComposeProject:     viper.GetString("COMPOSE_PROJECT"),
		ServicePrefix:      viper.GetString("SERVICE_PREFIX"),
		AutoStartEmul:      viper.GetBool("AUTO_START_EMULATORS"),
		CheckInterval:      viper.GetInt("EMULATOR_CHECK_INTERVAL"),
		AppiumURL:          viper.GetString("APPIUM_URL"),
		ActivityAPIURL:     viper.GetString("ACTIVITY_API_URL"),
		ActivityAPIToken:   viper.GetString("ACTIVITY_API_TOKEN"),
		ActivityFallbackDir: viper.GetString("ACTIVITY_FALLBACK_DIR"),
		TaskWatchdogMin:    viper.GetInt("TASK_WATCHDOG_MIN"),
		ScreenshotDir:      viper.GetString("SCREENSHOT_DIR"),
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
