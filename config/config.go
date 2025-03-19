// config/config.go
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	// Update settings
	UpdateEnabled   bool          `json:"update_enabled"`
	UpdateInterval  time.Duration `json:"update_interval"`
	GithubRepo      string        `json:"github_repo"`
	GithubToken     string        `json:"github_token"`
	CheckPrerelease bool          `json:"check_prerelease"`

	// Application settings
	LogLevel string `json:"log_level"`
	DataDir  string `json:"data_dir"`
}

// LoadConfig loads the configuration from the specified file
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	config := &Config{
		UpdateEnabled:   true,
		UpdateInterval:  1 * time.Minute,
		GithubRepo:      "noamstrauss/ota-updater",
		CheckPrerelease: false,
		LogLevel:        "info",
		DataDir:         "./data",
	}

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(configPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Save the default config
		if err := config.SaveConfig(configPath); err != nil {
			return nil, err
		}

		return config, nil
	}

	// Read the config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Parse the config file
	if err := json.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}

	// Check for environment variables
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.GithubToken = token
	}

	return config, nil
}

// SaveConfig saves the configuration to the specified file
func (c *Config) SaveConfig(configPath string) error {
	// Create the file
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the config
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
