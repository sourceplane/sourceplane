package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/sourceplane/sourceplane/internal/models"
	"github.com/sourceplane/sourceplane/internal/parser"
	provider "github.com/sourceplane/sourceplane/internal/providers"
	"github.com/sourceplane/sourceplane/internal/thinci"
)

var (
	thinCITarget      string
	thinCIMode        string
	thinCIBaseRef     string
	thinCIHeadRef     string
	thinCIChangedOnly bool
	thinCIEnvironment string
	thinCIOutput      string
	intentPath        string
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
	thinCIPlanCmd.Flags().StringVarP(&intentPath, "intent", "i", "", "Path to intent.yaml file (default: ./intent.yaml)")

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

	// Determine intent file path
	if intentPath == "" {
		intentPath = filepath.Join(cwd, "intent.yaml")
	} else if !filepath.IsAbs(intentPath) {
		intentPath = filepath.Join(cwd, intentPath)
	}

	// Check if intent.yaml exists
	if _, err := os.Stat(intentPath); os.IsNotExist(err) {
		return fmt.Errorf("could not find intent.yaml at %s", intentPath)
	}

	// Load the intent file
	intentFiles := []string{intentPath}
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

	// Find intent.yaml in repo
	intentPath := filepath.Join(repoPath, "intent.yaml")
	if _, err := os.Stat(intentPath); os.IsNotExist(err) {
		// Try legacy approach - load from providers directory
		return loadProvidersLegacy(repoPath)
	}

	// Load providers from intent.yaml using new loader
	providers, err := provider.LoadProvidersFromIntent(intentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers from intent: %w", err)
	}

	// Convert to thin-ci provider metadata and register
	for providerName, prov := range providers {
		metadata, err := convertToProviderMetadata(providerName, prov)
		if err != nil {
			return nil, fmt.Errorf("failed to convert provider %s: %w", providerName, err)
		}
		registry.RegisterProvider(metadata)
	}

	return registry, nil
}

// loadProvidersLegacy loads providers from local providers directory (backward compatibility)
func loadProvidersLegacy(repoPath string) (*thinci.ProviderRegistry, error) {
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
		providerMeta, err := loadProviderMetadata(providerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load provider %s: %w", providerName, err)
		}

		registry.RegisterProvider(providerMeta)
	}

	return registry, nil
}

// convertToProviderMetadata converts a provider.Provider to thinci.ProviderMetadata
func convertToProviderMetadata(name string, prov *provider.Provider) (*thinci.ProviderMetadata, error) {
	// Extract thin-ci configuration from provider
	thinCIConfig, ok := prov.Extensions["thinCI"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("provider %s missing thinCI configuration", name)
	}

	// Convert actions
	actionsRaw, ok := thinCIConfig["actions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("provider %s missing thinCI actions", name)
	}

	var actions []thinci.ProviderAction
	for _, actionRaw := range actionsRaw {
		actionData, err := yaml.Marshal(actionRaw)
		if err != nil {
			return nil, err
		}

		var action thinci.ProviderAction
		if err := yaml.Unmarshal(actionData, &action); err != nil {
			return nil, err
		}

		actions = append(actions, action)
	}

	// Extract defaults
	defaults := make(map[string]any)
	if defaultsRaw, ok := thinCIConfig["defaults"].(map[string]interface{}); ok {
		defaults = defaultsRaw
	}

	// Extract ordering
	var ordering []string
	if orderingRaw, ok := thinCIConfig["ordering"].([]interface{}); ok {
		for _, o := range orderingRaw {
			if s, ok := o.(string); ok {
				ordering = append(ordering, s)
			}
		}
	}

	return &thinci.ProviderMetadata{
		Name:    name,
		Version: prov.Version,
		ThinCI: thinci.ThinCIConfig{
			Actions:  actions,
			Defaults: defaults,
			Ordering: ordering,
		},
	}, nil
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
