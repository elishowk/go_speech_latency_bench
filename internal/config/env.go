package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Try to load from .env file in current directory
	if err := godotenv.Load(); err != nil {
		// If not found in current directory, try to find it in parent directories
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Look for .env file in parent directories up to 3 levels
		for i := 0; i < 3; i++ {
			envPath := filepath.Join(dir, ".env")
			if err := godotenv.Load(envPath); err == nil {
				return nil
			}
			dir = filepath.Dir(dir)
		}

		// If no .env file found, that's okay - we'll use system environment variables
		return nil
	}
	return nil
}

// GetProviderAPIKey gets the API key for the specified provider
func GetProviderAPIKey(providerName string) (string, error) {
	key := os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(providerName)))
	if key == "" {
		return "", fmt.Errorf("%s_API_KEY environment variable is required", strings.ToUpper(providerName))
	}
	return key, nil
}

// GetEnvWithDefault gets an environment variable with a default value
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 