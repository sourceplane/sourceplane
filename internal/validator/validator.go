package validator

import (
	"fmt"

	"github.com/sourceplane/cli/internal/models"
	"github.com/sourceplane/cli/internal/provider"
)

// ValidateRepository validates a repository definition against available providers
func ValidateRepository(repo *models.Repository) error {
	errors := []string{}

	// Basic validation
	if repo.APIVersion == "" {
		errors = append(errors, "apiVersion is required")
	}

	if repo.Kind == "" {
		errors = append(errors, "kind is required")
	}

	if repo.Metadata.Name == "" {
		errors = append(errors, "metadata.name is required")
	}

	// Validate components
	if len(repo.Components) == 0 {
		// Not an error, just no components
		if len(errors) > 0 {
			return fmt.Errorf("validation failed:\n  • %s", joinErrors(errors))
		}
		return nil
	}

	componentNames := make(map[string]bool)
	for i, comp := range repo.Components {
		if comp.Name == "" {
			errors = append(errors, fmt.Sprintf("component[%d]: name is required", i))
		} else {
			if componentNames[comp.Name] {
				errors = append(errors, fmt.Sprintf("duplicate component name: %s", comp.Name))
			}
			componentNames[comp.Name] = true
		}

		if comp.Type == "" {
			errors = append(errors, fmt.Sprintf("component[%d] (%s): type is required", i, comp.Name))
			continue
		}

		// Validate provider for this component
		providerName := provider.GetProviderNameFromType(comp.Type)
		if providerName == "" {
			errors = append(errors, fmt.Sprintf("component '%s': invalid type format '%s' (expected: provider.kind)", comp.Name, comp.Type))
			continue
		}

		// Load and validate against provider definition
		providerMeta, err := provider.LoadProvider(providerName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("component '%s': %v", comp.Name, err))
			continue
		}

		// Validate component type against provider's supported types
		if err := providerMeta.ValidateComponentType(comp.Type); err != nil {
			errors = append(errors, fmt.Sprintf("component '%s': %v", comp.Name, err))
		}
	}

	if len(errors) > 0 {
		// Get available providers for helpful error message
		availableProviders, _ := provider.ListAvailableProviders()
		errorMsg := "validation failed:\n"
		for _, err := range errors {
			errorMsg += fmt.Sprintf("  • %s\n", err)
		}
		if len(availableProviders) > 0 {
			errorMsg += "\nAvailable providers:\n"
			for _, p := range availableProviders {
				errorMsg += fmt.Sprintf("  • %s\n", p)
			}
		}
		return fmt.Errorf(errorMsg)
	}

	return nil
}

func joinErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	result := errors[0]
	for i := 1; i < len(errors); i++ {
		result += "\n  • " + errors[i]
	}
	return result
}
