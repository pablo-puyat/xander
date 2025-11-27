// Package config provides configuration management for the comic parser application.
// It supports loading configuration from JSON files and environment variables.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// Default API configuration
	defaultAnthropicModel      = "claude-3-5-sonnet-20240620"
	defaultAnthropicMaxTokens  = 1024
	defaultAnthropicAPIBaseURL = "https://api.anthropic.com/v1"
	defaultComicVineAPIBaseURL = "https://comicvine.gamespot.com/api"

	// Default processing settings
	defaultWorkerCount       = 3
	defaultRateLimitPerMin   = 30
	defaultRetryAttempts     = 3
	defaultRetryDelaySeconds = 2

	// Default cache settings
	defaultCacheDir = ".cache"

	// Default output settings
	defaultOutputFile   = "results.json"
	defaultOutputFormat = "json"

	// Environment variable names
	envAnthropicAPIKey = "ANTHROPIC_API_KEY"
	envComicVineAPIKey = "COMICVINE_API_KEY"
)

// Config holds all configuration for the application
type Config struct {
	// API Keys
	AnthropicAPIKey string `json:"anthropic_api_key"`
	ComicVineAPIKey string `json:"comicvine_api_key"`

	// Anthropic settings
	AnthropicModel      string `json:"anthropic_model"`
	AnthropicMaxTokens  int    `json:"anthropic_max_tokens"`
	AnthropicAPIBaseURL string `json:"anthropic_api_base_url"`

	// ComicVine settings
	ComicVineAPIBaseURL string `json:"comicvine_api_base_url"`

	// Processing settings
	WorkerCount       int    `json:"worker_count"`
	RateLimitPerMin   int    `json:"rate_limit_per_min"`
	RetryAttempts     int    `json:"retry_attempts"`
	RetryDelaySeconds int    `json:"retry_delay_seconds"`
	CacheEnabled      bool   `json:"cache_enabled"`
	CacheDir          string `json:"cache_dir"`

	// Output settings
	OutputFile   string `json:"output_file"`
	OutputFormat string `json:"output_format"` // json, csv
	Verbose      bool   `json:"verbose"`
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		AnthropicModel:      defaultAnthropicModel,
		AnthropicMaxTokens:  defaultAnthropicMaxTokens,
		AnthropicAPIBaseURL: defaultAnthropicAPIBaseURL,
		ComicVineAPIBaseURL: defaultComicVineAPIBaseURL,
		WorkerCount:         defaultWorkerCount,
		RateLimitPerMin:     defaultRateLimitPerMin,
		RetryAttempts:       defaultRetryAttempts,
		RetryDelaySeconds:   defaultRetryDelaySeconds,
		CacheEnabled:        true,
		CacheDir:            defaultCacheDir,
		OutputFile:          defaultOutputFile,
		OutputFormat:        defaultOutputFormat,
		Verbose:             false,
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Use defaults if file doesn't exist
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

// LoadFromEnv loads API keys from environment variables.
func (c *Config) LoadFromEnv() {
	if key := os.Getenv(envAnthropicAPIKey); key != "" {
		c.AnthropicAPIKey = key
	}
	if key := os.Getenv(envComicVineAPIKey); key != "" {
		c.ComicVineAPIKey = key
	}
}

// Validate checks that required configuration is present.
func (c *Config) Validate() error {
	if c.AnthropicAPIKey == "" {
		return fmt.Errorf("anthropic API key is required (set %s env var or in config)", envAnthropicAPIKey)
	}
	if c.ComicVineAPIKey == "" {
		return fmt.Errorf("comicvine API key is required (set %s env var or in config)", envComicVineAPIKey)
	}
	return nil
}

// SaveConfig saves the configuration to a JSON file
func (c *Config) SaveConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
