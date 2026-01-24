package providers

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Provider represents a provider definition loaded from provider.yaml
type Provider struct {
	Name       string                 `yaml:"name"`
	Version    string                 `yaml:"version"`
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Kinds      []ProviderKind         `yaml:"kinds"`
	Extensions map[string]interface{} `yaml:",inline"`
}

// ProviderKind represents a supported component kind
type ProviderKind struct {
	Name        string `yaml:"name"`
	FullType    string `yaml:"fullType"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
}

// IntentProviderConfig represents provider configuration in intent.yaml
type IntentProviderConfig struct {
	Source  string `yaml:"source,omitempty"`
	Version string `yaml:"version"`
}

// LoadProvidersFromIntent loads providers defined in an intent.yaml file
func LoadProvidersFromIntent(intentPath string) (map[string]*Provider, error) {
	// Read intent.yaml
	data, err := os.ReadFile(intentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read intent file: %w", err)
	}

	// Parse intent
	var intent struct {
		Providers map[string]IntentProviderConfig `yaml:"providers"`
	}

	if err := yaml.Unmarshal(data, &intent); err != nil {
		return nil, fmt.Errorf("failed to parse intent file: %w", err)
	}

	// Initialize provider cache
	cache, err := NewProviderCache()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize provider cache: %w", err)
	}

	// Load each provider
	providers := make(map[string]*Provider)
	intentDir := filepath.Dir(intentPath)

	for name, config := range intent.Providers {
		fmt.Printf("Loading provider: %s\n", name)

		var provider *Provider

		if config.Source != "" {
			// Load from remote source
			provider, err = cache.LoadProviderFromCache(config.Source, config.Version)
			if err != nil {
				return nil, fmt.Errorf("failed to load provider %s: %w", name, err)
			}
		} else {
			// Legacy: Load from local providers directory
			providerPath := filepath.Join(intentDir, "providers", name, "provider.yaml")
			providerData, err := os.ReadFile(providerPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read provider %s: %w", name, err)
			}

			provider = &Provider{}
			if err := yaml.Unmarshal(providerData, provider); err != nil {
				return nil, fmt.Errorf("failed to parse provider %s: %w", name, err)
			}
		}

		providers[name] = provider
	}

	return providers, nil
}

// InitProviders downloads all providers specified in intent.yaml
func InitProviders(intentPath string) error {
	// Read intent.yaml
	data, err := os.ReadFile(intentPath)
	if err != nil {
		return fmt.Errorf("failed to read intent file: %w", err)
	}

	var intent struct {
		Providers map[string]IntentProviderConfig `yaml:"providers"`
	}

	if err := yaml.Unmarshal(data, &intent); err != nil {
		return fmt.Errorf("failed to parse intent file: %w", err)
	}

	// Initialize provider cache
	cache, err := NewProviderCache()
	if err != nil {
		return fmt.Errorf("failed to initialize provider cache: %w", err)
	}

	// Download each provider
	manifest := make(map[string]string)

	for name, config := range intent.Providers {
		if config.Source == "" {
			fmt.Printf("Skipping %s (local provider)\n", name)
			continue
		}

		fmt.Printf("Initializing provider: %s@%s from %s\n", name, config.Version, config.Source)

		providerPath, err := cache.GetProviderPath(config.Source, config.Version)
		if err != nil {
			return fmt.Errorf("failed to initialize provider %s: %w", name, err)
		}

		manifest[name] = providerPath
		fmt.Printf("✓ Provider %s ready at %s\n", name, providerPath)
	}

	// Save manifest
	if err := cache.SaveCacheManifest(manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	return nil
}
