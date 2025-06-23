// config.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration.
type Config struct {
	MyGithubUser string   `toml:"my_github_user"`
	TeamUsers    []string `toml:"team_users"`
	Organization string   `toml:"organization"`
}

// LoadConfig loads the configuration from a TOML file.
// It searches for the config file in the XDG config directory.
func LoadConfig() (*Config, error) {
	// Get the XDG config home directory
	configHome, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user config directory: %w", err)
	}

	// Construct the full path to the config file
	configPath := filepath.Join(configHome, "gh-dash", "config.toml")

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("error decoding config file %s: %w", configPath, err)
	}

	return &cfg, nil
}
