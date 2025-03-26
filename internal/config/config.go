package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/kirsle/configdir"
)

type Config struct {
	DatabasePath        string   `json:"database_path"`
	DefaultProjectPaths []string `json:"default_project_paths"`
	VSCodePath          string   `json:"vscode_path"`
	Theme               string   `json:"theme"`
}

// DefaultConfig provides initial configuration values
func DefaultConfig() *Config {
	appName := "CodeHive"
	configPath := configdir.LocalConfig(appName)

	return &Config{
		DatabasePath: filepath.Join(configPath, "projects.db"),
		DefaultProjectPaths: []string{
			filepath.Join(os.Getenv("HOME"), "Projects"),
			filepath.Join(os.Getenv("USERPROFILE"), "Projects"),
		},
		Theme: "default",
	}
}

// Load reads the configuration file or creates a default one
func Load() (*Config, error) {
	appName := "CodeHive"
	configPath := configdir.LocalConfig(appName)
	configFile := filepath.Join(configPath, "config.json")

	// Ensure config directory exists
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config file
		defaultConfig := DefaultConfig()
		return defaultConfig, defaultConfig.Save()
	}

	// Read existing config
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

// Save writes the configuration to a file
func (c *Config) Save() error {
	appName := "CodeHive"
	configPath := configdir.LocalConfig(appName)
	configFile := filepath.Join(configPath, "config.json")

	configData, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = ioutil.WriteFile(configFile, configData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Update allows updating specific configuration values
func (c *Config) Update(updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "database_path":
			if path, ok := value.(string); ok {
				c.DatabasePath = path
			}
		case "default_project_paths":
			if paths, ok := value.([]string); ok {
				c.DefaultProjectPaths = paths
			}
		case "vscode_path":
			if path, ok := value.(string); ok {
				c.VSCodePath = path
			}
		case "theme":
			if theme, ok := value.(string); ok {
				c.Theme = theme
			}
		default:
			log.Printf("Unknown config key: %s", key)
		}
	}

	return c.Save()
}
