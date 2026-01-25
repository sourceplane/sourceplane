package thinci

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProviderFetcher handles fetching remote providers
type ProviderFetcher struct {
	cacheDir string
}

// NewProviderFetcher creates a new provider fetcher
func NewProviderFetcher() (*ProviderFetcher, error) {
	// Default cache location: ~/.sourceplane/providers
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	cacheDir := filepath.Join(homeDir, ".sourceplane", "providers")
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return &ProviderFetcher{
		cacheDir: cacheDir,
	}, nil
}

// FetchProvider downloads a provider from a remote source if needed
// Returns the local path to the provider
func (f *ProviderFetcher) FetchProvider(source, version string) (string, error) {
	// Parse the source to determine provider name and repo
	providerName, repoURL := f.parseSource(source)
	
	// Check if provider is already cached
	providerPath := filepath.Join(f.cacheDir, providerName)
	
	// Check if provider exists and is a git repo
	gitDir := filepath.Join(providerPath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		// Provider exists, try to update it
		fmt.Fprintf(os.Stderr, "Updating provider %s from %s...\n", providerName, source)
		if err := f.updateProvider(providerPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update provider: %v\n", err)
			// Continue with existing version
		}
	} else {
		// Provider doesn't exist, clone it
		fmt.Fprintf(os.Stderr, "Fetching provider %s from %s...\n", providerName, source)
		if err := f.cloneProvider(repoURL, providerPath); err != nil {
			return "", fmt.Errorf("failed to fetch provider: %w", err)
		}
	}
	
	// Verify provider.yaml exists
	providerYamlPath := filepath.Join(providerPath, "provider.yaml")
	if _, err := os.Stat(providerYamlPath); os.IsNotExist(err) {
		return "", fmt.Errorf("provider.yaml not found in %s", providerPath)
	}
	
	return providerPath, nil
}

// parseSource extracts provider name and repository URL from source string
// Examples:
//   - github.com/sourceplane/providers/helm -> (helm, https://github.com/sourceplane/providers)
//   - github.com/org/provider-name -> (provider-name, https://github.com/org/provider-name)
func (f *ProviderFetcher) parseSource(source string) (string, string) {
	// Remove protocol if present
	source = strings.TrimPrefix(source, "https://")
	source = strings.TrimPrefix(source, "http://")
	
	parts := strings.Split(source, "/")
	
	if len(parts) < 3 {
		// Invalid source, return as-is
		return source, "https://" + source
	}
	
	// Extract provider name from last part
	providerName := parts[len(parts)-1]
	
	// Build repo URL
	repoURL := "https://" + source
	
	return providerName, repoURL
}

// cloneProvider clones a git repository
func (f *ProviderFetcher) cloneProvider(repoURL, destPath string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, destPath)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	
	return nil
}

// updateProvider pulls latest changes for a provider
func (f *ProviderFetcher) updateProvider(providerPath string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = providerPath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	
	return nil
}

// IsRemoteSource checks if a source is remote (vs local path)
func IsRemoteSource(source string) bool {
	// Remote sources typically start with domain names
	return strings.Contains(source, "github.com") ||
		strings.Contains(source, "gitlab.com") ||
		strings.Contains(source, "bitbucket.org") ||
		strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "http://") ||
		strings.HasPrefix(source, "git@")
}
