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

type Config struct {
	UpdateInterval time.Duration `json:"update_interval"`
	GithubRepo     string        `json:"github_repo"`
	GithubToken    string        `json:"github_token,omitempty"`
	LogLevel       string        `json:"log_level"`
}

// DefaultConfig returns a Config struct with default values
func DefaultConfig() *Config {
	return &Config{
		UpdateInterval: 1 * time.Minute,
		GithubRepo:     "noamstrauss/ota-updater",
		LogLevel:       "info",
	}
}

// LoadConfig loads the config from the specified file or creates a default one if missing
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := saveDefaultConfig(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	overrideWithEnv(config)

	return config, nil
}

// SaveConfig saves the config to the specified file
func (c *Config) SaveConfig(configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config JSON: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// saveDefaultConfig creates a default config file
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

	if updateInterval := os.Getenv("UPDATE_INTERVAL"); updateInterval != "" {
		parsed, err := strconv.Atoi(updateInterval)
		if err == nil {
			config.UpdateInterval = time.Duration(parsed) * time.Minute
		}
	}
}
