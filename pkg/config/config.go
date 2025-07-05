package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// WhatsApp Configuration
	TargetPhoneNumber string
	SessionFilePath   string

	// Discord Configuration
	DiscordWebhookURL string

	// Google Cloud Configuration (optional)
	GoogleCloudProject string
	GoogleCloudBucket  string

	// Application Configuration
	LogLevel string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		TargetPhoneNumber:  getEnv("TARGET_PHONE_NUMBER", ""),
		SessionFilePath:    getEnv("SESSION_FILE_PATH", "./sessions/"),
		DiscordWebhookURL:  getEnv("DISCORD_WEBHOOK_URL", ""),
		GoogleCloudProject: getEnv("GOOGLE_CLOUD_PROJECT", ""),
		GoogleCloudBucket:  getEnv("GOOGLE_CLOUD_BUCKET", ""),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields
	if config.TargetPhoneNumber == "" {
		return nil, fmt.Errorf("TARGET_PHONE_NUMBER is required")
	}

	if config.DiscordWebhookURL == "" {
		return nil, fmt.Errorf("DISCORD_WEBHOOK_URL is required")
	}

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean with a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
