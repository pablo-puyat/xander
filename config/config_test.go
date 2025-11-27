package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorkerCount != defaultWorkerCount {
		t.Errorf("DefaultConfig().WorkerCount = %d; want %d", cfg.WorkerCount, defaultWorkerCount)
	}
	if cfg.AnthropicModel != defaultAnthropicModel {
		t.Errorf("DefaultConfig().AnthropicModel = %s; want %s", cfg.AnthropicModel, defaultAnthropicModel)
	}
	if cfg.CacheEnabled != true {
		t.Error("DefaultConfig().CacheEnabled = false; want true")
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save current env vars
	oldAnthropic := os.Getenv(envAnthropicAPIKey)
	oldCV := os.Getenv(envComicVineAPIKey)
	defer func() {
		os.Setenv(envAnthropicAPIKey, oldAnthropic)
		os.Setenv(envComicVineAPIKey, oldCV)
	}()

	// Set test env vars
	testAnthropicKey := "test-anthropic-key"
	testCVKey := "test-cv-key"
	os.Setenv(envAnthropicAPIKey, testAnthropicKey)
	os.Setenv(envComicVineAPIKey, testCVKey)

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	if cfg.AnthropicAPIKey != testAnthropicKey {
		t.Errorf("LoadFromEnv() AnthropicAPIKey = %s; want %s", cfg.AnthropicAPIKey, testAnthropicKey)
	}
	if cfg.ComicVineAPIKey != testCVKey {
		t.Errorf("LoadFromEnv() ComicVineAPIKey = %s; want %s", cfg.ComicVineAPIKey, testCVKey)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid Config",
			config: &Config{
				AnthropicAPIKey: "key1",
				ComicVineAPIKey: "key2",
			},
			wantErr: false,
		},
		{
			name: "Missing Anthropic Key",
			config: &Config{
				AnthropicAPIKey: "",
				ComicVineAPIKey: "key2",
			},
			wantErr: true,
		},
		{
			name: "Missing ComicVine Key",
			config: &Config{
				AnthropicAPIKey: "key1",
				ComicVineAPIKey: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
