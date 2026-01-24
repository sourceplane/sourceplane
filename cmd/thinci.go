package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/sourceplane/cli/internal/models"
	"github.com/sourceplane/cli/internal/parser"
	"github.com/sourceplane/cli/internal/thinci"
)

var (
	thinCITarget      string
	thinCIMode        string
	thinCIBaseRef     string
	thinCIHeadRef     string
	thinCIChangedOnly bool
	thinCIEnvironment string
	thinCIOutput      string
)

var thinCICmd = &cobra.Command{
	Use:     "thin-ci",
	Aliases: []string{"thinci"},
	Short:   "Thin CI planning engine for Sourceplane",
	Long: `Thin CI generates deterministic execution plans for CI systems.
It does not execute CI, it only creates plans that can be rendered into workflows.`,
}

var thinCIPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate a CI execution plan",
	Long: `Generate a deterministic execution plan based on repository state and git diff.
The plan includes affected components, required provider actions, dependencies, and CI job metadata.`,
	RunE: runThinCIPlan,
}

func init() {
	// Flags for plan command
	thinCIPlanCmd.Flags().StringVar(&thinCITarget, "github", "", "Generate plan for GitHub Actions (use --github)")
	thinCIPlanCmd.Flags().StringVar(&thinCITarget, "gitlab", "", "Generate plan for GitLab CI (use --gitlab)")
	thinCIPlanCmd.Flags().StringVarP(&thinCIMode, "mode", "m", "plan", "CI mode: plan, apply, or destroy")
	thinCIPlanCmd.Flags().StringVar(&thinCIBaseRef, "base", "main", "Base git ref for comparison")
	thinCIPlanCmd.Flags().StringVar(&thinCIHeadRef, "head", "HEAD", "Head git ref for comparison")
	thinCIPlanCmd.Flags().BoolVar(&thinCIChangedOnly, "changed-only", true, "Only include changed components")
	thinCIPlanCmd.Flags().StringVarP(&thinCIEnvironment, "env", "e", "", "Target environment (prod, staging, etc.)")
	thinCIPlanCmd.Flags().StringVarP(&thinCIOutput, "output", "o", "json", "Output format: json or yaml")

	// Mark target as required (at least one)
	thinCIPlanCmd.MarkFlagsOneRequired("github", "gitlab")
	
	// Add plan command to thin-ci command (for use as subcommand of sp)
	thinCICmd.AddCommand(thinCIPlanCmd)
}

func runThinCIPlan(cmd *cobra.Command, args []string) error {
	// Determine target from flags
	target := ""
	if cmd.Flags().Changed("github") {
		target = "github"
	} else if cmd.Flags().Changed("gitlab") {
		target = "gitlab"
	}

	if target == "" {
		return fmt.Errorf("target platform required: use --github or --gitlab")
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find all intent.yaml files
	intentFiles, err := findIntentFiles(cwd)
	if err != nil {
		return fmt.Errorf("failed to find intent files: %w", err)
	}

	if len(intentFiles) == 0 {
		return fmt.Errorf("no intent.yaml files found in repository")
	}

	// Load all intent files
	intents, err := loadIntentFiles(intentFiles)
	if err != nil {
		return fmt.Errorf("failed to load intent files: %w", err)
	}

	// Get changed files from git
	changedFiles, err := getChangedFiles(cwd, thinCIBaseRef, thinCIHeadRef)
	if err != nil {
		return fmt.Errorf("failed to get changed files: %w", err)
	}

	// Load provider registry
	registry, err := loadProviderRegistry(cwd)
	if err != nil {
		return fmt.Errorf("failed to load providers: %w", err)
	}

	// Create plan request
	planReq := thinci.PlanRequest{
		BaseRef:        thinCIBaseRef,
		HeadRef:        thinCIHeadRef,
		ChangedFiles:   changedFiles,
		RepositoryPath: cwd,
		IntentFiles:    intentFiles,
		Target:         target,
		Mode:           thinCIMode,
		ChangedOnly:    thinCIChangedOnly,
		Environment:    thinCIEnvironment,
	}

	// Generate plan
	planner := thinci.NewPlanner(registry)
	plan, err := planner.GeneratePlan(planReq, intents)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Output plan
	return outputPlan(plan, thinCIOutput)
}

// findIntentFiles recursively finds all intent.yaml files
func findIntentFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common ignore patterns
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == ".terraform" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file is intent.yaml or sourceplane.yaml
		if info.Name() == "intent.yaml" || info.Name() == "sourceplane.yaml" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// loadIntentFiles loads and parses intent files
func loadIntentFiles(paths []string) ([]*models.Repository, error) {
	intents := make([]*models.Repository, 0, len(paths))

	for _, path := range paths {
		intent, err := parser.LoadRepository(path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		intents = append(intents, intent)
	}

	return intents, nil
}

// getChangedFiles gets list of changed files from git
func getChangedFiles(repoPath, baseRef, headRef string) ([]string, error) {
	// This is a placeholder - in production, use git command or library
	// For now, return a mock list for testing
	
	// TODO: Implement actual git diff
	// Example: git diff --name-only base..head
	
	return []string{
		"intent.yaml",
		"terraform/vpc-network/main.tf",
		"helm/api-service/values.yaml",
	}, nil
}

// loadProviderRegistry loads all providers and creates a registry
func loadProviderRegistry(repoPath string) (*thinci.ProviderRegistry, error) {
	registry := thinci.NewProviderRegistry()

	// Find providers directory
	providersDir := filepath.Join(repoPath, "providers")
	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("providers directory not found at %s", providersDir)
	}

	// Read all provider directories
	entries, err := os.ReadDir(providersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read providers directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		providerName := entry.Name()
		providerPath := filepath.Join(providersDir, providerName, "provider.yaml")

		// Check if provider.yaml exists
		if _, err := os.Stat(providerPath); os.IsNotExist(err) {
			continue
		}

		// Load provider
		provider, err := loadProviderMetadata(providerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load provider %s: %w", providerName, err)
		}

		registry.RegisterProvider(provider)
	}

	return registry, nil
}

// loadProviderMetadata loads provider metadata including thin-ci config
func loadProviderMetadata(path string) (*thinci.ProviderMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	name, _ := raw["name"].(string)
	version, _ := raw["version"].(string)

	// Parse thin-ci configuration
	thinCIConfig := thinci.ThinCIConfig{
		Actions:  []thinci.ProviderAction{},
		Defaults: make(map[string]any),
	}

	if thinCIRaw, ok := raw["thinCI"].(map[string]interface{}); ok {
		// Parse actions
		if actionsRaw, ok := thinCIRaw["actions"].([]interface{}); ok {
			for _, actionRaw := range actionsRaw {
				if actionMap, ok := actionRaw.(map[string]interface{}); ok {
					action := thinci.ProviderAction{
						Name:        getString(actionMap, "name"),
						Description: getString(actionMap, "description"),
						Order:       getInt(actionMap, "order"),
					}
					thinCIConfig.Actions = append(thinCIConfig.Actions, action)
				}
			}
		}

		// Parse defaults
		if defaultsRaw, ok := thinCIRaw["defaults"].(map[string]interface{}); ok {
			thinCIConfig.Defaults = defaultsRaw
		}

		// Parse ordering
		if orderingRaw, ok := thinCIRaw["ordering"].([]interface{}); ok {
			ordering := make([]string, 0, len(orderingRaw))
			for _, o := range orderingRaw {
				if s, ok := o.(string); ok {
					ordering = append(ordering, s)
				}
			}
			thinCIConfig.Ordering = ordering
		}
	}

	return &thinci.ProviderMetadata{
		Name:    name,
		Version: version,
		ThinCI:  thinCIConfig,
	}, nil
}

// outputPlan outputs the plan in the specified format
func outputPlan(plan *thinci.Plan, format string) error {
	var output []byte
	var err error

	switch format {
	case "json":
		output, err = json.MarshalIndent(plan, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(plan)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}
