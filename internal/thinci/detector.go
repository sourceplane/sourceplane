package thinci

import (
	"path/filepath"
	"strings"

	"github.com/sourceplane/sourceplane/internal/models"
)

// ChangeDetector identifies which components are affected by file changes
type ChangeDetector struct {
	repositoryPath string
	intents        []*models.Repository
}

// NewChangeDetector creates a new change detector
func NewChangeDetector(repositoryPath string, intents []*models.Repository) *ChangeDetector {
	return &ChangeDetector{
		repositoryPath: repositoryPath,
		intents:        intents,
	}
}

// DetectChanges analyzes changed files and returns affected components
func (cd *ChangeDetector) DetectChanges(changedFiles []string) ([]ComponentChange, error) {
	changes := make(map[string]*ComponentChange)

	for _, intent := range cd.intents {
		for _, component := range intent.Components {
			change := cd.checkComponentAffected(component, intent, changedFiles)
			if change != nil {
				// Use component name as key to deduplicate
				if existing, ok := changes[component.Name]; ok {
					// Merge affected paths
					existing.AffectedPaths = append(existing.AffectedPaths, change.AffectedPaths...)
				} else {
					changes[component.Name] = change
				}
			}
		}
	}

	// Convert map to slice
	result := make([]ComponentChange, 0, len(changes))
	for _, change := range changes {
		result = append(result, *change)
	}

	return result, nil
}

// checkComponentAffected checks if a component is affected by the changed files
func (cd *ChangeDetector) checkComponentAffected(
	component models.Component,
	intent *models.Repository,
	changedFiles []string,
) *ComponentChange {
	var affectedPaths []string
	var reason string

	// Extract provider name from component type (e.g., "terraform.database" -> "terraform")
	provider := extractProvider(component.Type)

	// Check if intent.yaml itself changed
	for _, file := range changedFiles {
		if strings.HasSuffix(file, "intent.yaml") || strings.HasSuffix(file, "sourceplane.yaml") {
			affectedPaths = append(affectedPaths, file)
			reason = "Intent definition changed"
			break
		}
	}

	// Check component-specific paths
	componentPaths := cd.getComponentPaths(component, provider)
	for _, file := range changedFiles {
		for _, compPath := range componentPaths {
			if cd.pathMatches(file, compPath) {
				affectedPaths = append(affectedPaths, file)
				if reason == "" {
					reason = "Component files changed"
				}
			}
		}
	}

	// Check provider-level changes
	providerPaths := cd.getProviderPaths(provider)
	for _, file := range changedFiles {
		for _, provPath := range providerPaths {
			if cd.pathMatches(file, provPath) {
				affectedPaths = append(affectedPaths, file)
				if reason == "" {
					reason = "Provider configuration changed"
				}
			}
		}
	}

	// Check shared module dependencies
	sharedModulePaths := cd.getSharedModulePaths(component, provider)
	for _, file := range changedFiles {
		for _, modPath := range sharedModulePaths {
			if cd.pathMatches(file, modPath) {
				affectedPaths = append(affectedPaths, file)
				if reason == "" {
					reason = "Shared module changed"
				}
			}
		}
	}

	if len(affectedPaths) == 0 {
		return nil
	}

	return &ComponentChange{
		ComponentName: component.Name,
		Provider:      provider,
		ComponentType: component.Type,
		Reason:        reason,
		AffectedPaths: affectedPaths,
	}
}

// getComponentPaths returns paths that are specific to a component
func (cd *ChangeDetector) getComponentPaths(component models.Component, provider string) []string {
	paths := []string{}

	// Extract path from component spec
	if spec := component.Spec; spec != nil {
		switch provider {
		case "terraform":
			// Check for module source path or inline path
			if module, ok := spec["module"].(map[string]interface{}); ok {
				if source, ok := module["source"].(string); ok && !strings.HasPrefix(source, "terraform-") {
					// Local module path
					paths = append(paths, source)
				}
			}
			if path, ok := spec["path"].(string); ok {
				paths = append(paths, path)
			}

		case "helm":
			// Check for chart path
			if chart, ok := spec["chart"].(map[string]interface{}); ok {
				if path, ok := chart["path"].(string); ok {
					paths = append(paths, path)
				}
			}
			if chartPath, ok := spec["chartPath"].(string); ok {
				paths = append(paths, chartPath)
			}
			if valuesPath, ok := spec["valuesPath"].(string); ok {
				paths = append(paths, valuesPath)
			}
		}
	}

	// Fallback: convention-based paths
	if len(paths) == 0 {
		conventionPath := cd.getConventionBasedPath(component.Name, provider)
		if conventionPath != "" {
			paths = append(paths, conventionPath)
		}
	}

	return paths
}

// getConventionBasedPath returns conventional paths based on provider
func (cd *ChangeDetector) getConventionBasedPath(componentName, provider string) string {
	switch provider {
	case "terraform":
		return filepath.Join("terraform", componentName)
	case "helm":
		return filepath.Join("helm", componentName)
	default:
		return filepath.Join(provider, componentName)
	}
}

// getProviderPaths returns paths that affect all components of a provider
func (cd *ChangeDetector) getProviderPaths(provider string) []string {
	return []string{
		filepath.Join("providers", provider, "provider.yaml"),
		filepath.Join("providers", provider, "schema.yaml"),
		filepath.Join(".sourceplane", "providers", provider),
	}
}

// getSharedModulePaths returns paths to shared modules that a component depends on
func (cd *ChangeDetector) getSharedModulePaths(component models.Component, provider string) []string {
	paths := []string{}

	switch provider {
	case "terraform":
		// Check for shared modules in spec
		if spec := component.Spec; spec != nil {
			if module, ok := spec["module"].(map[string]interface{}); ok {
				if source, ok := module["source"].(string); ok {
					if strings.HasPrefix(source, "./") || strings.HasPrefix(source, "../") {
						paths = append(paths, source)
					}
				}
			}
		}
		// Common shared modules paths
		paths = append(paths, "terraform/modules")
	case "helm":
		paths = append(paths, "helm/charts")
	}

	return paths
}

// pathMatches checks if a file path matches a pattern
func (cd *ChangeDetector) pathMatches(file, pattern string) bool {
	// Direct match
	if file == pattern {
		return true
	}

	// Prefix match (file is under pattern directory)
	if strings.HasPrefix(file, pattern+"/") {
		return true
	}

	// Pattern match
	matched, _ := filepath.Match(pattern, file)
	return matched
}

// extractProvider extracts provider name from component type
// e.g., "terraform.database" -> "terraform"
func extractProvider(componentType string) string {
	parts := strings.Split(componentType, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
