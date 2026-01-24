package cmd

import (
	"fmt"

	"github.com/sourceplane/cli/internal/parser"
	"github.com/sourceplane/cli/internal/provider"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validate the intent.yaml file",
	Long:  "Check for errors and validate component definitions in intent.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, err := parser.FindIntentYaml()
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		fmt.Printf("Linting: %s\n\n", repoPath)

		repo, err := parser.LoadRepository(repoPath)
		if err != nil {
			return fmt.Errorf("❌ Failed to parse: %v", err)
		}

		// Basic validation
		errors := []string{}
		warnings := []string{}

		// Check required fields
		if repo.APIVersion == "" {
			errors = append(errors, "apiVersion is required")
		}

		if repo.Kind == "" {
			errors = append(errors, "kind is required")
		} else if repo.Kind != "Repository" && repo.Kind != "Intent" {
			warnings = append(warnings, fmt.Sprintf("kind should typically be 'Repository' or 'Intent', found '%s'", repo.Kind))
		}

		if repo.Metadata.Name == "" {
			errors = append(errors, "metadata.name is required")
		}

		// Check components
		if len(repo.Components) == 0 {
			warnings = append(warnings, "no components defined")
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

		// List available providers for helpful error messages
		availableProviders, _ := provider.ListAvailableProviders()
		if len(errors) > 0 && len(availableProviders) > 0 {
			// This will be shown after errors
		}

		// Display results
		if len(errors) > 0 {
			fmt.Println("❌ Errors:")
			for _, err := range errors {
				fmt.Printf("  • %s\n", err)
			}
			fmt.Println()
			
			// Show available providers if there were provider errors
			availableProviders, err := provider.ListAvailableProviders()
			if err == nil && len(availableProviders) > 0 {
				fmt.Println("Available providers:")
				for _, p := range availableProviders {
					fmt.Printf("  • %s\n", p)
				}
				fmt.Println()
			}
		}

		if len(warnings) > 0 {
			fmt.Println("⚠️  Warnings:")
			for _, warn := range warnings {
				fmt.Printf("  • %s\n", warn)
			}
			fmt.Println()
		}

		if len(errors) == 0 && len(warnings) == 0 {
			fmt.Println("✅ No issues found")
			fmt.Printf("\nRepository: %s\n", repo.Metadata.Name)
			fmt.Printf("Components: %d\n", len(repo.Components))
			
			// Show provider summary
			providerTypes := make(map[string]int)
			for _, comp := range repo.Components {
				providerName := provider.GetProviderNameFromType(comp.Type)
				providerTypes[providerName]++
			}
			
			if len(providerTypes) > 0 {
				fmt.Println("\nProviders in use:")
				for prov, count := range providerTypes {
					fmt.Printf("  • %s: %d component(s)\n", prov, count)
				}
			}
			
			if repo.Provider != "" {
				fmt.Printf("\nLegacy provider: %s\n", repo.Provider)
			}
		}

		if len(errors) > 0 {
			return fmt.Errorf("validation failed with %d error(s)", len(errors))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
