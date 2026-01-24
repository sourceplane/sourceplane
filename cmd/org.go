package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourceplane/cli/internal/models"
	"github.com/sourceplane/cli/internal/parser"
	"github.com/sourceplane/cli/internal/validator"
	"github.com/spf13/cobra"
)

var orgCmd = &cobra.Command{
	Use:   "org",
	Short: "Organization-wide operations",
	Long:  "Analyze and visualize your entire organization's architecture",
}

var orgTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display org-wide component tree",
	Long:  "Scan all repositories with intent.yaml and display components",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, _ := cmd.Flags().GetString("root")
		if rootDir == "" {
			var err error
			rootDir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		fmt.Printf("🔍 Scanning organization from: %s\n\n", rootDir)

		repos, err := findAllRepositories(rootDir)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories with sourceplane.yaml found")
			return nil
		}

		fmt.Printf("Found %d repository(ies):\n\n", len(repos))

		for _, repoPath := range repos {
			repo, err := parser.LoadRepository(repoPath)
			if err != nil {
				fmt.Printf("⚠️  Error loading %s: %v\n", repoPath, err)
				continue
			}

			// Validate each repository
			if err := validator.ValidateRepository(repo); err != nil {
				fmt.Printf("⚠️  Validation failed for %s:\n%v\n", repo.Metadata.Name, err)
				continue
			}

			fmt.Printf("📦 %s\n", repo.Metadata.Name)
			if repo.Metadata.Owner != "" {
				fmt.Printf("   Owner: %s\n", repo.Metadata.Owner)
			}
			if repo.Provider != "" {
				fmt.Printf("   Provider: %s\n", repo.Provider)
			}

			if len(repo.Components) > 0 {
				fmt.Println("   Components:")
				for i, comp := range repo.Components {
					isLast := i == len(repo.Components)-1
					prefix := "├──"
					if isLast {
						prefix = "└──"
					}
					fmt.Printf("   %s %s [%s]\n", prefix, comp.Name, comp.Type)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

var orgGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate org-wide architectural graph",
	Long:  "Build a dependency and relationship graph across all repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir, _ := cmd.Flags().GetString("root")
		if rootDir == "" {
			var err error
			rootDir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		fmt.Printf("🔍 Building org graph from: %s\n\n", rootDir)

		repos, err := findAllRepositories(rootDir)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories with sourceplane.yaml found")
			return nil
		}

		// Build graph
		componentCount := 0
		repoMap := make(map[string]int)

		repoList := []*models.Repository{}
		for _, repoPath := range repos {
			repo, err := parser.LoadRepository(repoPath)
			if err != nil {
				continue
			}
			
			// Validate each repository
			if err := validator.ValidateRepository(repo); err != nil {
				fmt.Printf("⚠️  Skipping %s: validation failed\n", repo.Metadata.Name)
				continue
			}
			
			repoList = append(repoList, repo)
			repoMap[repo.Metadata.Name] = len(repo.Components)
			componentCount += len(repo.Components)
		}

		fmt.Println("Organization Graph:")
		fmt.Printf("─────────────────────\n")
		fmt.Printf("Repositories: %d\n", len(repoMap))
		fmt.Printf("Components:   %d\n", componentCount)
		fmt.Println()

		// Component type distribution
		typeCount := make(map[string]int)
		for _, repo := range repoList {
			for _, comp := range repo.Components {
				typeCount[comp.Type]++
			}
		}

		if len(typeCount) > 0 {
			fmt.Println("Component Types:")
			for cType, count := range typeCount {
				fmt.Printf("  • %s: %d\n", cType, count)
			}
			fmt.Println()
		}

		// Owner distribution
		ownerCount := make(map[string]int)
		for _, repo := range repoList {
			if repo.Metadata.Owner != "" {
				ownerCount[repo.Metadata.Owner]++
			}
		}

		if len(ownerCount) > 0 {
			fmt.Println("Repository Owners:")
			for owner, count := range ownerCount {
				fmt.Printf("  • %s: %d repo(s)\n", owner, count)
			}
		}

		return nil
	},
}

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "CI/CD operations",
	Long:  "Generate and manage CI/CD pipelines from component definitions",
}

var ciRenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render CI/CD workflows from intent.yaml",
	Long:  "Generate CI workflows based on repository components and provider",
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

		fmt.Printf("🔄 Rendering CI/CD for: %s\n\n", repo.Metadata.Name)

		if repo.Provider == "" {
			return fmt.Errorf("no provider specified in sourceplane.yaml")
		}

		fmt.Printf("Provider: %s\n", repo.Provider)
		fmt.Printf("Components: %d\n\n", len(repo.Components))

		fmt.Println("⚠️  Provider-based CI rendering not yet implemented")
		fmt.Println("This will generate CI workflows based on component definitions")

		return nil
	},
}

// findAllRepositories recursively searches for intent.yaml files
func findAllRepositories(root string) ([]string, error) {
	var repos []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't access
		}

		// Skip hidden directories
		if info.IsDir() && len(info.Name()) > 0 && info.Name()[0] == '.' {
			return filepath.SkipDir
		}

		// Skip common directories that won't have sourceplane.yaml
		if info.IsDir() && (info.Name() == "node_modules" || info.Name() == "vendor" || info.Name() == "target") {
			return filepath.SkipDir
		}

		if !info.IsDir() && info.Name() == "intent.yaml" {
			repos = append(repos, path)
		}

		return nil
	})

	return repos, err
}

func init() {
	orgTreeCmd.Flags().String("root", "", "Root directory to scan (defaults to current directory)")
	orgGraphCmd.Flags().String("root", "", "Root directory to scan (defaults to current directory)")

	orgCmd.AddCommand(orgTreeCmd)
	orgCmd.AddCommand(orgGraphCmd)
	rootCmd.AddCommand(orgCmd)

	ciCmd.AddCommand(ciRenderCmd)
	rootCmd.AddCommand(ciCmd)
}
