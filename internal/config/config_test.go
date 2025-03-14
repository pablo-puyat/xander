package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	
	// Check default values
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "", cfg.ComicVineAPIKey)
	assert.Equal(t, "", cfg.TVDBAPIKey)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("XANDER_API_KEY", "test-api-key")
	t.Setenv("XANDER_ENV", "development")
	t.Setenv("XANDER_COMICVINE_API_KEY", "test-comicvine-key")
	t.Setenv("XANDER_TVDB_API_KEY", "test-tvdb-key")
	
	// Load config and check values
	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	
	assert.Equal(t, "test-api-key", cfg.APIKey)
	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "test-comicvine-key", cfg.ComicVineAPIKey)
	assert.Equal(t, "test-tvdb-key", cfg.TVDBAPIKey)
}

func TestLoadConfigPartialEnv(t *testing.T) {
	// Clear all environment variables first
	os.Unsetenv("XANDER_API_KEY")
	os.Unsetenv("XANDER_ENV")
	os.Unsetenv("XANDER_COMICVINE_API_KEY")
	os.Unsetenv("XANDER_TVDB_API_KEY")
	
	// Set only one environment variable
	t.Setenv("XANDER_COMICVINE_API_KEY", "test-comicvine-key")
	
	// Load config and check values
	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	
	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "test-comicvine-key", cfg.ComicVineAPIKey)
	assert.Equal(t, "", cfg.TVDBAPIKey)
}

func TestLoadConfigFromFile(t *testing.T) {
	// Clear all environment variables first
	os.Unsetenv("XANDER_API_KEY")
	os.Unsetenv("XANDER_ENV")
	os.Unsetenv("XANDER_COMICVINE_API_KEY")
	os.Unsetenv("XANDER_TVDB_API_KEY")
	
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "xander-test-config")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Create a config YAML file
	testConfig := &Config{
		APIKey:          "file-api-key",
		Environment:     "staging",
		ComicVineAPIKey: "file-comicvine-key",
		TVDBAPIKey:      "file-tvdb-key",
	}
	
	configDir := filepath.Join(tempDir, ".config", "xander")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	configPath := filepath.Join(configDir, "config.yaml")
	configData, err := yaml.Marshal(testConfig)
	require.NoError(t, err)
	
	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err)
	
	// Load config and check values
	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	
	assert.Equal(t, "file-api-key", cfg.APIKey)
	assert.Equal(t, "staging", cfg.Environment)
	assert.Equal(t, "file-comicvine-key", cfg.ComicVineAPIKey)
	assert.Equal(t, "file-tvdb-key", cfg.TVDBAPIKey)
}

func TestLoadConfigNoFile(t *testing.T) {
	// Clear all environment variables first
	os.Unsetenv("XANDER_API_KEY")
	os.Unsetenv("XANDER_ENV")
	os.Unsetenv("XANDER_COMICVINE_API_KEY")
	os.Unsetenv("XANDER_TVDB_API_KEY")
	
	// Create a temporary home directory with no config file
	tempDir, err := os.MkdirTemp("", "xander-test-empty")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Load config and check defaults
	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	
	assert.Equal(t, "", cfg.APIKey)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "", cfg.ComicVineAPIKey)
	assert.Equal(t, "", cfg.TVDBAPIKey)
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Clear all environment variables first
	os.Unsetenv("XANDER_API_KEY")
	os.Unsetenv("XANDER_ENV")
	os.Unsetenv("XANDER_COMICVINE_API_KEY")
	os.Unsetenv("XANDER_TVDB_API_KEY")
	
	// Create a temporary config file with invalid YAML
	tempDir, err := os.MkdirTemp("", "xander-test-invalid")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Create invalid config file
	configDir := filepath.Join(tempDir, ".config", "xander")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	configPath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configPath, []byte("this is not valid yaml: :::"), 0600)
	require.NoError(t, err)
	
	// Load config should return error
	cfg, err := LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary home directory
	tempDir, err := os.MkdirTemp("", "xander-test-save")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Create a config to save
	cfg := &Config{
		APIKey:          "save-api-key",
		Environment:     "test",
		ComicVineAPIKey: "save-comicvine-key",
		TVDBAPIKey:      "save-tvdb-key",
	}
	
	// Save the config
	err = cfg.Save()
	require.NoError(t, err)
	
	// Check file exists
	configPath, err := getConfigPath()
	require.NoError(t, err)
	
	fileInfo, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
	
	// Load the config back to verify
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	
	var loadedCfg Config
	err = yaml.Unmarshal(data, &loadedCfg)
	require.NoError(t, err)
	
	assert.Equal(t, "save-api-key", loadedCfg.APIKey)
	assert.Equal(t, "test", loadedCfg.Environment)
	assert.Equal(t, "save-comicvine-key", loadedCfg.ComicVineAPIKey)
	assert.Equal(t, "save-tvdb-key", loadedCfg.TVDBAPIKey)
}

func TestGetConfigPath(t *testing.T) {
	// Create a temporary home directory
	tempDir, err := os.MkdirTemp("", "xander-test-path")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Get config path
	configPath, err := getConfigPath()
	require.NoError(t, err)
	
	// Verify path is correct
	expected := filepath.Join(tempDir, ".config", "xander", "config.yaml")
	assert.Equal(t, expected, configPath)
	
	// Verify directory was created
	configDir := filepath.Join(tempDir, ".config", "xander")
	fileInfo, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, fileInfo.IsDir())
}

func TestEnvironmentPrecedence(t *testing.T) {
	// This test verifies that environment variables take precedence over file config
	
	// Create a temporary config file
	tempDir, err := os.MkdirTemp("", "xander-test-priority")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Setup the test environment
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tempDir)
	defer t.Setenv("HOME", oldHome)
	
	// Create a config YAML file
	fileConfig := &Config{
		APIKey:          "file-api-key",
		Environment:     "file-env",
		ComicVineAPIKey: "file-comicvine-key",
		TVDBAPIKey:      "file-tvdb-key",
	}
	
	configDir := filepath.Join(tempDir, ".config", "xander")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	configPath := filepath.Join(configDir, "config.yaml")
	configData, err := yaml.Marshal(fileConfig)
	require.NoError(t, err)
	
	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err)
	
	// Test that ALL environment variables take precedence
	t.Run("All environment variables", func(t *testing.T) {
		// Set environment variables
		t.Setenv("XANDER_API_KEY", "env-api-key")
		t.Setenv("XANDER_ENV", "env-env")
		t.Setenv("XANDER_COMICVINE_API_KEY", "env-comicvine-key")
		t.Setenv("XANDER_TVDB_API_KEY", "env-tvdb-key")
		
		// Load config and check values
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// All values should come from environment
		assert.Equal(t, "env-api-key", cfg.APIKey)
		assert.Equal(t, "env-env", cfg.Environment)
		assert.Equal(t, "env-comicvine-key", cfg.ComicVineAPIKey)
		assert.Equal(t, "env-tvdb-key", cfg.TVDBAPIKey)
	})
	
	// Test that if ANY environment variable is set, file is not loaded
	t.Run("Partial environment variables", func(t *testing.T) {
		// Clear all variables
		os.Unsetenv("XANDER_API_KEY")
		os.Unsetenv("XANDER_ENV")
		os.Unsetenv("XANDER_COMICVINE_API_KEY")
		os.Unsetenv("XANDER_TVDB_API_KEY")
		
		// Set just one environment variable
		t.Setenv("XANDER_API_KEY", "env-api-key")
		
		// Load config and check values
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// The API key should come from environment
		assert.Equal(t, "env-api-key", cfg.APIKey)
		// Default value from NewConfig(), not from file
		assert.Equal(t, "production", cfg.Environment)
		// Not set in env or config
		assert.Equal(t, "", cfg.ComicVineAPIKey)
		assert.Equal(t, "", cfg.TVDBAPIKey)
	})
	
	// Test that with no environment variables, file is loaded
	t.Run("No environment variables", func(t *testing.T) {
		// Clear all variables
		os.Unsetenv("XANDER_API_KEY")
		os.Unsetenv("XANDER_ENV")
		os.Unsetenv("XANDER_COMICVINE_API_KEY")
		os.Unsetenv("XANDER_TVDB_API_KEY")
		
		// Load config and check values
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		
		// All values should come from file
		assert.Equal(t, "file-api-key", cfg.APIKey)
		assert.Equal(t, "file-env", cfg.Environment)
		assert.Equal(t, "file-comicvine-key", cfg.ComicVineAPIKey)
		assert.Equal(t, "file-tvdb-key", cfg.TVDBAPIKey)
	})
}