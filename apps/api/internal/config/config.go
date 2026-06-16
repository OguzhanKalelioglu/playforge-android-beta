package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv      string
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string

	IyzicoAPIKey     string
	IyzicoSecretKey  string
	IyzicoBaseURL    string

	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	OrchestratorAPIToken string
	ActivityIngestToken  string
	Domain               string
	LogLevel             string
	SentryDSN            string

	// Task types (Asynq'da kullanılır)
	TaskTypeTestStart        string
	TaskTypeDailyEngagement  string
	TaskTypeWriteReview      string
	TaskTypeLoginGoogle      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("IYZICO_BASE_URL", "https://api.iyzico.com")
	viper.SetDefault("SMTP_PORT", 587)

	cfg := &Config{
		AppEnv:               viper.GetString("APP_ENV"),
		Port:                 viper.GetString("PORT"),
		DatabaseURL:          viper.GetString("DATABASE_URL"),
		RedisURL:             viper.GetString("REDIS_URL"),
		JWTSecret:            viper.GetString("JWT_SECRET"),
		IyzicoAPIKey:         viper.GetString("IYZICO_API_KEY"),
		IyzicoSecretKey:      viper.GetString("IYZICO_SECRET_KEY"),
		IyzicoBaseURL:        viper.GetString("IYZICO_BASE_URL"),
		SMTPHost:             viper.GetString("SMTP_HOST"),
		SMTPPort:             viper.GetInt("SMTP_PORT"),
		SMTPUser:             viper.GetString("SMTP_USER"),
		SMTPPassword:         viper.GetString("SMTP_PASSWORD"),
		OrchestratorAPIToken: viper.GetString("ORCHESTRATOR_API_TOKEN"),
		ActivityIngestToken:  viper.GetString("ACTIVITY_INGEST_TOKEN"),
		Domain:               viper.GetString("DOMAIN"),
		LogLevel:             viper.GetString("LOG_LEVEL"),
		SentryDSN:            viper.GetString("SENTRY_DSN"),
		TaskTypeTestStart:        "test_start",
		TaskTypeDailyEngagement:  "daily_engagement",
		TaskTypeWriteReview:      "write_review",
		TaskTypeLoginGoogle:      "login_google",
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
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.JWTSecret == "" || len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}
