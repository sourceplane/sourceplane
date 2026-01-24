package thinci

import (
	"fmt"
	"sort"
	"time"

	"github.com/sourceplane/sourceplane/internal/models"
)

// Planner generates CI execution plans
type Planner struct {
	providerRegistry *ProviderRegistry
}

// NewPlanner creates a new planner
func NewPlanner(registry *ProviderRegistry) *Planner {
	return &Planner{
		providerRegistry: registry,
	}
}

// GeneratePlan creates a complete CI execution plan from a request
func (p *Planner) GeneratePlan(req PlanRequest, intents []*models.Repository) (*Plan, error) {
	// Step 1: Detect changes
	detector := NewChangeDetector(req.RepositoryPath, intents)
	changes, err := detector.DetectChanges(req.ChangedFiles)
	if err != nil {
		return nil, fmt.Errorf("change detection failed: %w", err)
	}

	// If changedOnly flag is set and no changes detected, return empty plan
	if req.ChangedOnly && len(changes) == 0 {
		return p.createEmptyPlan(req), nil
	}

	// Step 2: Expand components into dependency nodes
	nodes, err := p.expandComponents(changes, intents, req)
	if err != nil {
		return nil, fmt.Errorf("component expansion failed: %w", err)
	}

	// Step 3: Build dependency graph and topological sort
	sortedNodes, err := p.buildDependencyGraph(nodes, intents)
	if err != nil {
		return nil, fmt.Errorf("dependency graph construction failed: %w", err)
	}

	// Step 4: Generate jobs from sorted nodes
	jobs := p.generateJobs(sortedNodes, req)

	// Step 5: Construct final plan
	plan := &Plan{
		Target: req.Target,
		Mode:   req.Mode,
		Metadata: PlanMetadata{
			Repository:   req.RepositoryPath,
			BaseRef:      req.BaseRef,
			HeadRef:      req.HeadRef,
			ChangedFiles: req.ChangedFiles,
			Timestamp:    time.Now().Format(time.RFC3339),
			Environment:  req.Environment,
		},
		Jobs: jobs,
	}

	return plan, nil
}

// expandComponents converts component changes into dependency nodes with actions
func (p *Planner) expandComponents(
	changes []ComponentChange,
	intents []*models.Repository,
	req PlanRequest,
) ([]DependencyNode, error) {
	nodes := make([]DependencyNode, 0, len(changes))

	for _, change := range changes {
		// Get provider metadata
		providerMeta, err := p.providerRegistry.GetProvider(change.Provider)
		if err != nil {
			return nil, fmt.Errorf("provider '%s' not found: %w", change.Provider, err)
		}

		// Find component in intent to get relationships
		component := p.findComponent(change.ComponentName, intents)
		if component == nil {
			return nil, fmt.Errorf("component '%s' not found in intent", change.ComponentName)
		}

		// Determine which actions to run based on mode and provider capabilities
		actions := p.determineActions(req.Mode, providerMeta)

		// Build dependency list
		dependencies := p.extractDependencies(component, intents)

		node := DependencyNode{
			ComponentName: change.ComponentName,
			Provider:      change.Provider,
			Actions:       actions,
			Dependencies:  dependencies,
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// determineActions decides which provider actions should run
func (p *Planner) determineActions(mode string, provider *ProviderMetadata) []string {
	actions := []string{}

	// Get provider's supported actions
	supportedActions := provider.ThinCI.Actions

	switch mode {
	case "plan":
		// In plan mode, run validation and plan actions
		if p.hasAction(supportedActions, "validate") {
			actions = append(actions, "validate")
		}
		if p.hasAction(supportedActions, "plan") {
			actions = append(actions, "plan")
		}
	case "apply":
		// In apply mode, run validation, plan, and apply
		if p.hasAction(supportedActions, "validate") {
			actions = append(actions, "validate")
		}
		if p.hasAction(supportedActions, "plan") {
			actions = append(actions, "plan")
		}
		if p.hasAction(supportedActions, "apply") {
			actions = append(actions, "apply")
		}
	case "destroy":
		if p.hasAction(supportedActions, "destroy") {
			actions = append(actions, "destroy")
		}
	}

	return actions
}

// hasAction checks if provider supports an action
func (p *Planner) hasAction(actions []ProviderAction, name string) bool {
	for _, action := range actions {
		if action.Name == name {
			return true
		}
	}
	return false
}

// extractDependencies gets component dependencies from relationships
func (p *Planner) extractDependencies(component *models.Component, intents []*models.Repository) []string {
	dependencies := []string{}

	for _, intent := range intents {
		for _, rel := range intent.Relationships {
			// Check if this component depends on another
			if rel.From == component.Name && (rel.Type == "depends_on" || rel.Type == "uses") {
				dependencies = append(dependencies, rel.To)
			}
		}
	}

	// Also check component-level relationships if they exist
	if compRels, ok := component.Spec["relationships"].([]interface{}); ok {
		for _, relInterface := range compRels {
			if rel, ok := relInterface.(map[string]interface{}); ok {
				if target, ok := rel["target"].(string); ok {
					dependencies = append(dependencies, target)
				}
			}
		}
	}

	return dependencies
}

// buildDependencyGraph performs topological sort on dependency nodes
func (p *Planner) buildDependencyGraph(nodes []DependencyNode, intents []*models.Repository) ([]DependencyNode, error) {
	// Create adjacency list
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	nodeMap := make(map[string]DependencyNode)

	// Initialize graph
	for _, node := range nodes {
		nodeMap[node.ComponentName] = node
		graph[node.ComponentName] = []string{}
		inDegree[node.ComponentName] = 0
	}

	// Build edges
	for _, node := range nodes {
		for _, dep := range node.Dependencies {
			// Only add edge if dependency is also in the changed set
			if _, exists := nodeMap[dep]; exists {
				graph[dep] = append(graph[dep], node.ComponentName)
				inDegree[node.ComponentName]++
			}
		}
	}

	// Kahn's algorithm for topological sort
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	sorted := []DependencyNode{}
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, nodeMap[current])

		// Process neighbors
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(nodes) {
		return nil, fmt.Errorf("circular dependency detected in component graph")
	}

	return sorted, nil
}

// generateJobs creates CI jobs from sorted dependency nodes
func (p *Planner) generateJobs(nodes []DependencyNode, req PlanRequest) []Job {
	jobs := []Job{}

	// Track which jobs depend on which other jobs
	jobDependencies := make(map[string][]string)

	for _, node := range nodes {
		// Get provider metadata for job configuration
		providerMeta, err := p.providerRegistry.GetProvider(node.Provider)
		if err != nil {
			continue // Skip if provider not found
		}

		// Generate a job for each action
		for actionIdx, action := range node.Actions {
			jobID := fmt.Sprintf("%s-%s", node.ComponentName, action)

			// Determine dependencies for this job
			deps := []string{}

			// If not the first action for this component, depend on previous action
			if actionIdx > 0 {
				prevAction := node.Actions[actionIdx-1]
				prevJobID := fmt.Sprintf("%s-%s", node.ComponentName, prevAction)
				deps = append(deps, prevJobID)
			} else {
				// First action depends on last actions of all dependency components
				for _, depComp := range node.Dependencies {
					// Find the last action for the dependency component
					depNode := p.findNode(depComp, nodes)
					if depNode != nil && len(depNode.Actions) > 0 {
						lastAction := depNode.Actions[len(depNode.Actions)-1]
						depJobID := fmt.Sprintf("%s-%s", depComp, lastAction)
						deps = append(deps, depJobID)
					}
				}
			}

			// Get action-specific configuration from provider
			providerAction := p.findProviderAction(providerMeta, action)

			// Build job inputs
			inputs := p.buildJobInputs(node, providerMeta, req)
			if providerAction != nil && providerAction.Inputs != nil {
				for k, v := range providerAction.Inputs {
					if _, exists := inputs[k]; !exists {
						inputs[k] = v
					}
				}
			}

			// Create job metadata based on target platform
			metadata := p.createJobMetadata(req.Target, node, action)

			job := Job{
				ID:        jobID,
				Component: node.ComponentName,
				Provider:  node.Provider,
				Action:    action,
				Inputs:    inputs,
				DependsOn: deps,
				Metadata:  metadata,
			}

			jobs = append(jobs, job)
			jobDependencies[jobID] = deps
		}
	}

	return jobs
}

// buildJobInputs constructs the inputs map for a job
func (p *Planner) buildJobInputs(node DependencyNode, provider *ProviderMetadata, req PlanRequest) map[string]any {
	inputs := make(map[string]any)

	// Add component name
	inputs["component"] = node.ComponentName

	// Add provider-specific defaults
	if provider.ThinCI.Defaults != nil {
		for k, v := range provider.ThinCI.Defaults {
			inputs[k] = v
		}
	}

	// Add CLI overrides
	if req.Environment != "" {
		inputs["environment"] = req.Environment
	}

	// Add provider overrides from request
	if overrides, ok := req.ProviderOverrides[node.Provider]; ok {
		for k, v := range overrides {
			inputs[k] = v
		}
	}

	return inputs
}

// createJobMetadata creates platform-specific job metadata
func (p *Planner) createJobMetadata(target string, node DependencyNode, action string) JobMetadata {
	metadata := JobMetadata{
		Environment: map[string]string{
			"SP_COMPONENT": node.ComponentName,
			"SP_PROVIDER":  node.Provider,
			"SP_ACTION":    action,
		},
	}

	switch target {
	case "github":
		metadata.RunsOn = "ubuntu-latest"
		metadata.Permissions = []string{"id-token", "contents"}
		metadata.Timeout = 30
	case "gitlab":
		metadata.RunsOn = "docker"
		metadata.Timeout = 30
	}

	return metadata
}

// Helper functions

func (p *Planner) findComponent(name string, intents []*models.Repository) *models.Component {
	for _, intent := range intents {
		for _, comp := range intent.Components {
			if comp.Name == name {
				return &comp
			}
		}
	}
	return nil
}

func (p *Planner) findNode(name string, nodes []DependencyNode) *DependencyNode {
	for _, node := range nodes {
		if node.ComponentName == name {
			return &node
		}
	}
	return nil
}

func (p *Planner) findProviderAction(provider *ProviderMetadata, actionName string) *ProviderAction {
	for _, action := range provider.ThinCI.Actions {
		if action.Name == actionName {
			return &action
		}
	}
	return nil
}

func (p *Planner) createEmptyPlan(req PlanRequest) *Plan {
	return &Plan{
		Target: req.Target,
		Mode:   req.Mode,
		Metadata: PlanMetadata{
			Repository:   req.RepositoryPath,
			BaseRef:      req.BaseRef,
			HeadRef:      req.HeadRef,
			ChangedFiles: req.ChangedFiles,
			Timestamp:    time.Now().Format(time.RFC3339),
			Environment:  req.Environment,
		},
		Jobs: []Job{},
	}
}

// ProviderMetadata extends the provider model with thin-ci specific configuration
type ProviderMetadata struct {
	Name    string
	Version string
	ThinCI  ThinCIConfig
}

// ThinCIConfig holds thin-ci specific provider configuration
type ThinCIConfig struct {
	Actions  []ProviderAction `yaml:"actions"`
	Defaults map[string]any   `yaml:"defaults,omitempty"`
	Ordering []string         `yaml:"ordering,omitempty"` // Default action ordering
}

// ProviderRegistry manages loaded providers
type ProviderRegistry struct {
	providers map[string]*ProviderMetadata
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]*ProviderMetadata),
	}
}

// RegisterProvider adds a provider to the registry
func (r *ProviderRegistry) RegisterProvider(provider *ProviderMetadata) {
	r.providers[provider.Name] = provider
}

// GetProvider retrieves a provider by name
func (r *ProviderRegistry) GetProvider(name string) (*ProviderMetadata, error) {
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not registered", name)
	}
	return provider, nil
}

// ListProviders returns all registered providers
func (r *ProviderRegistry) ListProviders() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
