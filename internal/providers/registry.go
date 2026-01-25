package providers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProviderSource represents where to fetch a provider from
type ProviderSource struct {
	Type    string // "github", "local", "registry"
	URL     string
	Version string
}

// ProviderCache manages local provider caching
type ProviderCache struct {
	baseDir string
}

// NewProviderCache creates a new provider cache
func NewProviderCache() (*ProviderCache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".sourceplane", "providers")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ProviderCache{baseDir: cacheDir}, nil
}

// ParseProviderSource parses a provider source string
func ParseProviderSource(source string) (*ProviderSource, error) {
	if source == "" {
		return nil, fmt.Errorf("empty provider source")
	}

	// Check if it's a local path
	if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "file://") {
		return &ProviderSource{
			Type: "local",
			URL:  strings.TrimPrefix(source, "file://"),
		}, nil
	}

	// Check if it's a GitHub source
	if strings.HasPrefix(source, "github.com/") {
		return &ProviderSource{
			Type: "github",
			URL:  source,
		}, nil
	}

	// Default to registry
	return &ProviderSource{
		Type: "registry",
		URL:  source,
	}, nil
}

// GetProviderPath returns the cached path for a provider or downloads it
func (c *ProviderCache) GetProviderPath(source, version string) (string, error) {
	ps, err := ParseProviderSource(source)
	if err != nil {
		return "", err
	}

	switch ps.Type {
	case "local":
		return ps.URL, nil
	case "github":
		return c.getGitHubProvider(ps.URL, version)
	default:
		return "", fmt.Errorf("unsupported provider source type: %s", ps.Type)
	}
}

// getGitHubProvider downloads and caches a provider from GitHub
func (c *ProviderCache) getGitHubProvider(source, version string) (string, error) {
	// Parse GitHub source: github.com/owner/repo
	parts := strings.Split(strings.TrimPrefix(source, "github.com/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid GitHub source format: %s", source)
	}

	owner := parts[0]
	repo := parts[1]

	// Clean version string (remove >= or other operators)
	cleanVersion := strings.TrimPrefix(version, ">=")
	cleanVersion = strings.TrimPrefix(cleanVersion, "~>")
	cleanVersion = strings.TrimSpace(cleanVersion)

	// Create provider cache directory
	providerDir := filepath.Join(c.baseDir, owner, repo, cleanVersion)
	lockFile := filepath.Join(providerDir, ".lock")

	// Check if already cached
	if _, err := os.Stat(lockFile); err == nil {
		return providerDir, nil
	}

	// Download provider
	if err := c.downloadGitHubProvider(owner, repo, cleanVersion, providerDir); err != nil {
		return "", err
	}

	// Create lock file
	if err := os.WriteFile(lockFile, []byte(cleanVersion), 0644); err != nil {
		return "", fmt.Errorf("failed to create lock file: %w", err)
	}

	return providerDir, nil
}

// downloadGitHubProvider downloads a provider release from GitHub
func (c *ProviderCache) downloadGitHubProvider(owner, repo, version, destDir string) error {
	// If version doesn't start with 'v', add it
	tagVersion := version
	if !strings.HasPrefix(tagVersion, "v") {
		tagVersion = "v" + tagVersion
	}

	// Construct GitHub release URL
	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", owner, repo, tagVersion)

	fmt.Fprintf(os.Stderr, "Downloading provider from %s...\n", url)

	// Download tarball
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download provider: HTTP %d", resp.StatusCode)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract tarball
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Skip the root directory in the archive
		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		relativePath := parts[1]

		target := filepath.Join(destDir, relativePath)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			f.Close()
		}
	}

	fmt.Printf("Provider cached at %s\n", destDir)
	return nil
}

// LoadProviderFromCache loads a provider definition from the cache
func (c *ProviderCache) LoadProviderFromCache(source, version string) (*Provider, error) {
	providerPath, err := c.GetProviderPath(source, version)
	if err != nil {
		return nil, err
	}

	// Load provider.yaml
	providerFile := filepath.Join(providerPath, "provider.yaml")
	data, err := os.ReadFile(providerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read provider.yaml: %w", err)
	}

	var provider Provider
	if err := yaml.Unmarshal(data, &provider); err != nil {
		return nil, fmt.Errorf("failed to parse provider.yaml: %w", err)
	}

	return &provider, nil
}

// ClearCache removes all cached providers
func (c *ProviderCache) ClearCache() error {
	return os.RemoveAll(c.baseDir)
}

// ListCachedProviders returns a list of all cached providers
func (c *ProviderCache) ListCachedProviders() ([]CachedProvider, error) {
	var cached []CachedProvider

	err := filepath.Walk(c.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".lock" {
			relPath, _ := filepath.Rel(c.baseDir, filepath.Dir(path))
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 3 {
				cached = append(cached, CachedProvider{
					Owner:   parts[0],
					Repo:    parts[1],
					Version: parts[2],
					Path:    filepath.Dir(path),
				})
			}
		}

		return nil
	})

	return cached, err
}

// CachedProvider represents a cached provider
type CachedProvider struct {
	Owner   string
	Repo    string
	Version string
	Path    string
}

// SaveCacheManifest saves a manifest of installed providers
func (c *ProviderCache) SaveCacheManifest(providers map[string]string) error {
	manifestPath := filepath.Join(c.baseDir, "manifest.json")
	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

// LoadCacheManifest loads the manifest of installed providers
func (c *ProviderCache) LoadCacheManifest() (map[string]string, error) {
	manifestPath := filepath.Join(c.baseDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}

	var manifest map[string]string
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	return manifest, nil
}
