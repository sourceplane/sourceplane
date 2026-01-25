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
	
	// Run command flags
	runPlanFile   string
	runJobID      string
	runVerbose    bool
	runDryRun     bool
	runGitHub     bool
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

var thinCIRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a job from a plan locally",
	Long: `Execute a specific job from a generated plan file.
Runs pre-steps, main commands, and post-steps with verbose output.
Useful for testing CI jobs locally before pushing to CI/CD platform.`,
	RunE: runThinCIRun,
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

	// Flags for run command
	thinCIRunCmd.Flags().StringVarP(&runPlanFile, "plan", "p", "plan.json", "Path to plan file")
	thinCIRunCmd.Flags().StringVar(&runJobID, "job-id", "", "Job ID to execute (required)")
	thinCIRunCmd.Flags().BoolVarP(&runVerbose, "verbose", "v", true, "Verbose output")
	thinCIRunCmd.Flags().BoolVar(&runDryRun, "dry-run", false, "Dry run mode (don't execute commands)")
	thinCIRunCmd.Flags().BoolVar(&runGitHub, "github", false, "Running in GitHub Actions context")
	
	// Mark required flags
	thinCIRunCmd.MarkFlagRequired("job-id")

	// Add plan command to thin-ci command (for use as subcommand of sp)
	thinCICmd.AddCommand(thinCIPlanCmd)
	thinCICmd.AddCommand(thinCIRunCmd)
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

	// Load provider registry (from intent file and local/remote sources)
	registry, err := loadProviderRegistry(cwd, intents)
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
// loadProviderRegistry loads all providers from intent files and local/remote sources
func loadProviderRegistry(repoPath string, intents []*models.Repository) (*thinci.ProviderRegistry, error) {
	registry := thinci.NewProviderRegistry()
	fetcher, err := thinci.NewProviderFetcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create provider fetcher: %w", err)
	}

	// Collect all providers from all intent files
	providerSources := make(map[string]models.Provider)
	for _, intent := range intents {
		for name, provider := range intent.Providers {
			if _, exists := providerSources[name]; !exists {
				providerSources[name] = provider
			}
		}
	}

	// Load each provider
	for name, providerConfig := range providerSources {
		var providerPath string
		
		// Check if source is remote or local
		if providerConfig.Source != "" && thinci.IsRemoteSource(providerConfig.Source) {
			// Fetch remote provider
			path, err := fetcher.FetchProvider(providerConfig.Source, providerConfig.Version)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch provider %s: %w", name, err)
			}
			providerPath = filepath.Join(path, "provider.yaml")
		} else {
			// Try local providers directory first
			providerPath = filepath.Join(repoPath, "providers", name, "provider.yaml")
			
			// If not found locally, search up the directory tree
			if _, err := os.Stat(providerPath); os.IsNotExist(err) {
				searchPath := repoPath
				found := false
				for i := 0; i < 5; i++ {
					testPath := filepath.Join(searchPath, "providers", name, "provider.yaml")
					if _, err := os.Stat(testPath); err == nil {
						providerPath = testPath
						found = true
						break
					}
					searchPath = filepath.Dir(searchPath)
				}
				
				if !found {
					return nil, fmt.Errorf("provider '%s' not found locally and no remote source specified", name)
				}
			}
		}

		// Verify provider.yaml exists
		if _, err := os.Stat(providerPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("provider.yaml not found for provider %s at %s", name, providerPath)
		}

		fmt.Fprintf(os.Stderr, "Loading provider: %s from %s\n", name, providerPath)

		// Load provider metadata
		providerMeta, err := loadProviderMetadata(providerPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load provider %s: %w", name, err)
		}

		registry.RegisterProvider(providerMeta)
	}

	// If no providers were loaded from intents, fall back to loading from local providers directory
	if len(providerSources) == 0 {
		fmt.Fprintln(os.Stderr, "No providers defined in intent files, searching local providers directory...")
		return loadLocalProviders(repoPath)
	}

	return registry, nil
}

// loadLocalProviders loads providers from local providers directory (fallback/legacy)
func loadLocalProviders(repoPath string) (*thinci.ProviderRegistry, error) {
	registry := thinci.NewProviderRegistry()

	// For thin-ci, always load providers from providers directory
	// because we need the full thinCI configuration from provider.yaml
	providersDir := filepath.Join(repoPath, "providers")
	
	// Check if we're in a subdirectory - walk up to find providers dir
	searchPath := repoPath
	for i := 0; i < 5; i++ { // Max 5 levels up
		testPath := filepath.Join(searchPath, "providers")
		if _, err := os.Stat(testPath); err == nil {
			providersDir = testPath
			break
		}
		searchPath = filepath.Dir(searchPath)
	}

	if _, err := os.Stat(providersDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("providers directory not found (searched from %s)", repoPath)
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

		fmt.Fprintf(os.Stderr, "Loading provider: %s\n", providerName)

		// Load provider metadata directly from provider.yaml
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
				// Marshal back to YAML and unmarshal into ProviderAction struct
				// This properly handles all fields including jobTemplate
				actionData, err := yaml.Marshal(actionRaw)
				if err != nil {
					continue
				}

				var action thinci.ProviderAction
				if err := yaml.Unmarshal(actionData, &action); err != nil {
					continue
				}

				thinCIConfig.Actions = append(thinCIConfig.Actions, action)
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

// runThinCIRun executes a specific job from a plan file
func runThinCIRun(cmd *cobra.Command, args []string) error {
	// Load the plan file
	planData, err := os.ReadFile(runPlanFile)
	if err != nil {
		return fmt.Errorf("failed to read plan file: %w", err)
	}
	
	// Parse the plan
	var plan thinci.Plan
	if err := json.Unmarshal(planData, &plan); err != nil {
		return fmt.Errorf("failed to parse plan file: %w", err)
	}
	
	// Find the job with the specified ID
	var targetJob *thinci.Job
	for i := range plan.Jobs {
		if plan.Jobs[i].GetID() == runJobID {
			targetJob = &plan.Jobs[i]
			break
		}
	}
	
	if targetJob == nil {
		return fmt.Errorf("job '%s' not found in plan", runJobID)
	}
	
	// Create executor
	executor := thinci.NewExecutor(runVerbose, runDryRun)
	
	// Execute the job
	fmt.Printf("Sourceplane Thin-CI Job Executor\n")
	fmt.Printf("Plan: %s\n", runPlanFile)
	
	if runDryRun {
		fmt.Println("\n⚠️  DRY RUN MODE - Commands will not be executed")
	}
	
	if err := executor.ExecuteJob(*targetJob); err != nil {
		return fmt.Errorf("job execution failed: %w", err)
	}
	
	return nil
}
