package cmd

import (
	"fmt"

	"github.com/sourceplane/cli/internal/parser"
	"github.com/sourceplane/cli/internal/validator"
	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage and inspect components",
	Long:  "Commands for listing, describing, and creating components in repositories",
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all components in the current repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, err := parser.FindIntentYaml()
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		repo, err := parser.LoadRepository(repoPath)
		if err != nil {
			return err
		}

		// Validate before proceeding
		if err := validator.ValidateRepository(repo); err != nil {
			return err
		}

		if len(repo.Components) == 0 {
			fmt.Println("No components found in this repository")
			return nil
		}

		fmt.Printf("Components in %s:\n\n", repo.Metadata.Name)
		for _, comp := range repo.Components {
			fmt.Printf("  • %s (%s)\n", comp.Name, comp.Type)
		}
		fmt.Printf("\nTotal: %d components\n", len(repo.Components))

		return nil
	},
}

var componentTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display component tree with dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath, err := parser.FindIntentYaml()
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		repo, err := parser.LoadRepository(repoPath)
		if err != nil {
			return err
		}

		// Validate before proceeding
		if err := validator.ValidateRepository(repo); err != nil {
			return err
		}

		if len(repo.Components) == 0 {
			fmt.Println("No components found in this repository")
			return nil
		}

		fmt.Printf("Component Tree for %s:\n\n", repo.Metadata.Name)
		fmt.Printf("Repository: %s\n", repo.Metadata.Name)
		for i, comp := range repo.Components {
			isLast := i == len(repo.Components)-1
			prefix := "├──"
			if isLast {
				prefix = "└──"
			}
			fmt.Printf("%s %s [%s]\n", prefix, comp.Name, comp.Type)

			// Show spec/inputs if present
			inputs := comp.Spec
			if len(inputs) == 0 {
				inputs = comp.Inputs // fallback to legacy
			}
			if len(inputs) > 0 {
				indent := "│   "
				if isLast {
					indent = "    "
				}
				for key, value := range inputs {
					fmt.Printf("%s  ▸ %s: %v\n", indent, key, value)
				}
			}
		}

		return nil
	},
}

var componentDescribeCmd = &cobra.Command{
	Use:   "describe [component-name]",
	Short: "Describe a specific component",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentName := args[0]

		repoPath, err := parser.FindIntentYaml()
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		repo, err := parser.LoadRepository(repoPath)
		if err != nil {
			return err
		}

		// Validate before proceeding
		if err := validator.ValidateRepository(repo); err != nil {
			return err
		}

		var foundComponent *struct {
			Name   string
			Type   string
			Spec   map[string]interface{}
			Inputs map[string]interface{}
		}

		for _, comp := range repo.Components {
			if comp.Name == componentName {
				foundComponent = &struct {
					Name   string
					Type   string
					Spec   map[string]interface{}
					Inputs map[string]interface{}
				}{
					Name:   comp.Name,
					Type:   comp.Type,
					Spec:   comp.Spec,
					Inputs: comp.Inputs,
				}
				break
			}
		}

		if foundComponent == nil {
			return fmt.Errorf("component '%s' not found", componentName)
		}

		fmt.Printf("Component: %s\n", foundComponent.Name)
		fmt.Printf("Type: %s\n", foundComponent.Type)
		fmt.Printf("Repository: %s\n\n", repo.Metadata.Name)

		// Prefer Spec over legacy Inputs
		data := foundComponent.Spec
		if len(data) == 0 {
			data = foundComponent.Inputs
		}

		if len(data) > 0 {
			fmt.Println("Spec:")
			for key, value := range data {
				fmt.Printf("  %s: %v\n", key, value)
			}
		} else {
			fmt.Println("Spec: (none)")
		}

		if repo.Provider != "" {
			fmt.Printf("\nProvider: %s\n", repo.Provider)
		}

		return nil
	},
}

var componentCreateCmd = &cobra.Command{
	Use:   "create [component-name]",
	Short: "Create a new component (requires provider)",
	Long:  "Bootstrap a new component in the current repository using a provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentName := args[0]
		componentType, _ := cmd.Flags().GetString("type")
		provider, _ := cmd.Flags().GetString("provider")

		if componentType == "" {
			return fmt.Errorf("--type flag is required")
		}

		if provider == "" {
			return fmt.Errorf("--provider flag is required")
		}

		fmt.Printf("Creating component '%s' with type '%s' using provider '%s'\n",
			componentName, componentType, provider)
		fmt.Println("\n⚠️  Provider functionality not yet implemented")
		fmt.Println("This will bootstrap the component with implementation files")

		return nil
	},
}

func init() {
	componentCreateCmd.Flags().String("type", "", "Component type (e.g., service.api)")
	componentCreateCmd.Flags().String("provider", "", "Provider to use (e.g., my-provider@v1)")

	componentCmd.AddCommand(componentListCmd)
	componentCmd.AddCommand(componentTreeCmd)
	componentCmd.AddCommand(componentDescribeCmd)
	componentCmd.AddCommand(componentCreateCmd)
	rootCmd.AddCommand(componentCmd)
}
