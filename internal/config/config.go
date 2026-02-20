package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AccessToken  string `json:"access_token"`
	BaseURL      string `json:"base_url"`
}

// Load reads a configuration from the specified file path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to the specified file path
// It creates the directory if it doesn't exist with 0700 permissions
// and writes the file with 0600 permissions
func (c *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file with 0600 permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks that all required configuration fields are present
func (c *Config) Validate() error {
	if c.ClientID == "" {
		return fmt.Errorf("ClientID is required")
	}
	if c.ClientSecret == "" {
		return fmt.Errorf("ClientSecret is required")
	}
	if c.AccessToken == "" {
		return fmt.Errorf("AccessToken is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("BaseURL is required")
	}
	return nil
}

// DefaultConfigPath returns the default configuration file path
// Returns ~/.dida365/config.json or empty string if home directory cannot be determined
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".dida365", "config.json")
}
