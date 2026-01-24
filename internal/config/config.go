package config

import (
	"os"
	"path/filepath"
)

// Config holds the CLI configuration
type Config struct {
	// ProvidersPath is the path to the providers directory
	ProvidersPath string

	// CachePath is the path to the cache directory
	CachePath string

	// WorkingDir is the current working directory
	WorkingDir string
}

// Default returns a default configuration
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	return &Config{
		ProvidersPath: filepath.Join(cwd, "providers"),
		CachePath:     filepath.Join(homeDir, ".sourceplane"),
		WorkingDir:    cwd,
	}
}

// Load loads configuration from environment or returns default
func Load() (*Config, error) {
	cfg := Default()

	// Override with environment variables if set
	if providersPath := os.Getenv("SOURCEPLANE_PROVIDERS_PATH"); providersPath != "" {
		cfg.ProvidersPath = providersPath
	}

	if cachePath := os.Getenv("SOURCEPLANE_CACHE_PATH"); cachePath != "" {
		cfg.CachePath = cachePath
	}

	return cfg, nil
}

// EnsureCacheDir creates the cache directory if it doesn't exist
func (c *Config) EnsureCacheDir() error {
	return os.MkdirAll(c.CachePath, 0755)
}
