// config/config.go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Update settings
	UpdateEnabled  bool          `json:"update_enabled"`
	UpdateInterval time.Duration `json:"update_interval"`
	GithubRepo     string        `json:"github_repo"`
	GithubToken    string        `json:"github_token,omitempty"`

	// Application settings
	LogLevel string `json:"log_level"`
}

// DefaultConfig returns a Config struct with default values
func DefaultConfig() *Config {
	return &Config{
		UpdateEnabled:  true,
		UpdateInterval: 1 * time.Minute,
		GithubRepo:     "noamstrauss/ota-updater",
		LogLevel:       "info",
	}
}

// LoadConfig loads the configuration from the specified file or creates a default one if missing
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := saveDefaultConfig(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Read the config file
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config file
	if err := json.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Apply environment variable overrides
	overrideWithEnv(config)

	return config, nil
}

// SaveConfig saves the configuration to the specified file
func (c *Config) SaveConfig(configPath string) error {
	// Create the config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// saveDefaultConfig creates a default configuration file
func saveDefaultConfig(configPath string, config *Config) error {
	return config.SaveConfig(configPath)
}

// overrideWithEnv updates config values with environment variables if set
func overrideWithEnv(config *Config) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.GithubToken = token
	}

	if repo := os.Getenv("GITHUB_REPO"); repo != "" {
		config.GithubRepo = repo
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if updateEnabled := os.Getenv("UPDATE_ENABLED"); updateEnabled != "" {
		parsed, err := strconv.ParseBool(updateEnabled)
		if err == nil {
			config.UpdateEnabled = parsed
		}
	}

	if updateInterval := os.Getenv("UPDATE_INTERVAL"); updateInterval != "" {
		parsed, err := strconv.Atoi(updateInterval)
		if err == nil {
			config.UpdateInterval = time.Duration(parsed) * time.Minute
		}
	}
}
