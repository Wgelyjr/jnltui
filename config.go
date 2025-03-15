package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	EntriesDir string `yaml:"entries_dir"`
	DevMode    bool   `yaml:"dev_mode"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	// Default to production mode with entries in ~/.journal-tui/entries
	homeDir, err := os.UserHomeDir()
	entriesDir := "entries" // Fallback
	
	if err == nil {
		entriesDir = filepath.Join(homeDir, ".journal-tui", "entries")
	}
	
	return &Config{
		EntriesDir: entriesDir,
		DevMode:    false,
	}
}

// LoadConfig loads the configuration from files in the following order:
// 1. Default configuration
// 2. System-wide configuration (/etc/journal-tui/config.yaml)
// 3. User-specific configuration (~/.journal-tui/config.yaml)
// 4. Current directory configuration (./journal-tui.yaml)
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Try to load system-wide configuration
	if err := loadConfigFile("/etc/journal-tui/config.yaml", config); err != nil {
		// It's okay if the file doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading system config: %w", err)
		}
	}

	// Try to load user-specific configuration
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userConfigPath := filepath.Join(homeDir, ".journal-tui", "config.yaml")
		if err := loadConfigFile(userConfigPath, config); err != nil {
			// It's okay if the file doesn't exist
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("error loading user config: %w", err)
			}
		}
	}

	// Try to load current directory configuration (for development)
	if err := loadConfigFile("journal-tui.yaml", config); err != nil {
		// It's okay if the file doesn't exist
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading local config: %w", err)
		}
	}

	return config, nil
}

// loadConfigFile loads configuration from a YAML file
func loadConfigFile(path string, config *Config) error {
	// Check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Parse the YAML
	return yaml.Unmarshal(data, config)
}

// SaveConfig saves the configuration to the user's config file
func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}

	// Ensure the config directory exists
	configDir := filepath.Join(homeDir, ".journal-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Marshal the config to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write the file
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}

// EnsureEntriesDir ensures the entries directory exists
func (c *Config) EnsureEntriesDir() error {
	if _, err := os.Stat(c.EntriesDir); os.IsNotExist(err) {
		return os.MkdirAll(c.EntriesDir, 0755)
	}
	return nil
}

// CreateDefaultConfigFile creates a default configuration file if it doesn't exist
func CreateDefaultConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}

	// Ensure the config directory exists
	configDir := filepath.Join(homeDir, ".journal-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	// Check if the config file already exists
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		// File already exists, don't overwrite it
		return nil
	}

	// Create a default config
	config := DefaultConfig()

	// Save it
	return SaveConfig(config)
}
