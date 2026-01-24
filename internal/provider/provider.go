package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderMetadata represents a provider.yaml file
type ProviderMetadata struct {
	Name       string         `yaml:"name"`
	Version    string         `yaml:"version"`
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Kinds      []ProviderKind `yaml:"kinds"`
}

// ProviderKind represents a supported component kind
type ProviderKind struct {
	Name        string `yaml:"name"`
	FullType    string `yaml:"fullType"`
	Description string `yaml:"description"`
	Category    string `yaml:"category"`
}

// LoadProvider loads a provider definition from the providers directory
func LoadProvider(providerName string) (*ProviderMetadata, error) {
	// Get the directory where the CLI is running or look for providers/ directory
	providersDir := findProvidersDirectory()
	if providersDir == "" {
		return nil, fmt.Errorf("providers directory not found")
	}

	providerPath := filepath.Join(providersDir, providerName, "provider.yaml")
	
	// Check if provider.yaml exists
	if _, err := os.Stat(providerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("provider '%s' not found (expected at %s)", providerName, providerPath)
	}

	// Read and parse provider.yaml
	data, err := os.ReadFile(providerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider.yaml for '%s': %w", providerName, err)
	}

	var metadata ProviderMetadata
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse provider.yaml for '%s': %w", providerName, err)
	}

	return &metadata, nil
}

// findProvidersDirectory searches for the providers/ directory
func findProvidersDirectory() string {
	// Try current directory first
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Check current directory
	providersPath := filepath.Join(cwd, "providers")
	if _, err := os.Stat(providersPath); err == nil {
		return providersPath
	}

	// Walk up the directory tree
	dir := cwd
	for {
		providersPath := filepath.Join(dir, "providers")
		if _, err := os.Stat(providersPath); err == nil {
			return providersPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// ValidateComponentType checks if a component type is supported by a provider
func (p *ProviderMetadata) ValidateComponentType(componentType string) error {
	// Check if the component type matches provider.kind format
	expectedPrefix := p.Name + "."
	if !strings.HasPrefix(componentType, expectedPrefix) {
		return fmt.Errorf("component type '%s' does not match provider '%s' (expected format: %s<kind>)", 
			componentType, p.Name, expectedPrefix)
	}

	// Extract the kind from the type (e.g., "helm.service" -> "service")
	kind := strings.TrimPrefix(componentType, expectedPrefix)

	// Check if this kind is supported by the provider
	for _, supportedKind := range p.Kinds {
		if supportedKind.Name == kind || supportedKind.FullType == componentType {
			return nil
		}
	}

	// Build list of supported types
	supportedTypes := make([]string, len(p.Kinds))
	for i, k := range p.Kinds {
		supportedTypes[i] = k.FullType
	}

	return fmt.Errorf("component type '%s' is not supported by provider '%s' (supported types: %s)", 
		componentType, p.Name, strings.Join(supportedTypes, ", "))
}

// GetProviderNameFromType extracts the provider name from a component type
// e.g., "helm.service" -> "helm"
func GetProviderNameFromType(componentType string) string {
	parts := strings.SplitN(componentType, ".", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// ListAvailableProviders returns a list of all available providers in the providers directory
func ListAvailableProviders() ([]string, error) {
	providersDir := findProvidersDirectory()
	if providersDir == "" {
		return nil, fmt.Errorf("providers directory not found")
	}

	entries, err := os.ReadDir(providersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read providers directory: %w", err)
	}

	var providers []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Check if this directory contains a provider.yaml
			providerYaml := filepath.Join(providersDir, entry.Name(), "provider.yaml")
			if _, err := os.Stat(providerYaml); err == nil {
				providers = append(providers, entry.Name())
			}
		}
	}

	return providers, nil
}
