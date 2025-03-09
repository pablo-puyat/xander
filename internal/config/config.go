package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey          string `yaml:"api_key"`
	Environment     string `yaml:"environment"`
	ComicVineAPIKey string `yaml:"comicvine_api_key"`
	TVDBAPIKey      string `yaml:"tvdb_api_key"`
}

func NewConfig() *Config {
	return &Config{
		Environment: "production", // Set defaults
	}
}

func LoadConfig() (*Config, error) {
	// First check environment variables
	cfg := NewConfig()
	
	// Load from environment variables if available
	if apiKey := os.Getenv("XANDER_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	
	if env := os.Getenv("XANDER_ENV"); env != "" {
		cfg.Environment = env
	}
	
	if comicVineAPIKey := os.Getenv("XANDER_COMICVINE_API_KEY"); comicVineAPIKey != "" {
		cfg.ComicVineAPIKey = comicVineAPIKey
	}
	
	if tvdbAPIKey := os.Getenv("XANDER_TVDB_API_KEY"); tvdbAPIKey != "" {
		cfg.TVDBAPIKey = tvdbAPIKey
	}
	
	// If any key is set via environment, return the config
	if cfg.APIKey != "" || cfg.ComicVineAPIKey != "" || cfg.TVDBAPIKey != "" {
		return cfg, nil
	}

	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "xander")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.yaml"), nil
}
