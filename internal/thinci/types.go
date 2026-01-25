package thinci

// Plan represents a complete CI execution plan
type Plan struct {
	Target   string       `json:"target"` // e.g., "github", "gitlab"
	Mode     string       `json:"mode"`   // "plan" or "apply"
	Metadata PlanMetadata `json:"metadata"`
	Jobs     []Job        `json:"jobs"`
}

// PlanMetadata contains contextual information about the plan
type PlanMetadata struct {
	Repository   string   `json:"repository"`
	BaseRef      string   `json:"baseRef"`
	HeadRef      string   `json:"headRef"`
	ChangedFiles []string `json:"changedFiles"`
	Timestamp    string   `json:"timestamp"`
	Environment  string   `json:"environment,omitempty"`
}

// Job represents a single CI job with flexible provider-defined structure
type Job map[string]any

// JobCore contains the minimal required fields that all jobs must have
// Providers can extend beyond these core fields
type JobCore struct {
	ID        string   `json:"id"`
	Component string   `json:"component"`
	Provider  string   `json:"provider"`
	Action    string   `json:"action"`
	DependsOn []string `json:"dependsOn"`
}

// Helper methods for Job to access core fields with type safety
func (j Job) GetID() string {
	if id, ok := j["id"].(string); ok {
		return id
	}
	return ""
}

func (j Job) GetComponent() string {
	if comp, ok := j["component"].(string); ok {
		return comp
	}
	return ""
}

func (j Job) GetProvider() string {
	if prov, ok := j["provider"].(string); ok {
		return prov
	}
	return ""
}

func (j Job) GetAction() string {
	if action, ok := j["action"].(string); ok {
		return action
	}
	return ""
}

func (j Job) GetDependsOn() []string {
	if deps, ok := j["dependsOn"].([]string); ok {
		return deps
	}
	// Handle []interface{} conversion
	if depsInterface, ok := j["dependsOn"].([]interface{}); ok {
		deps := make([]string, 0, len(depsInterface))
		for _, dep := range depsInterface {
			if depStr, ok := dep.(string); ok {
				deps = append(deps, depStr)
			}
		}
		return deps
	}
	return []string{}
}

// ProviderAction describes what a provider can do in CI
type ProviderAction struct {
	Name        string         `json:"name" yaml:"name"` // plan, apply, destroy, validate
	Description string         `json:"description" yaml:"description"`
	Order       int            `json:"order" yaml:"order"` // Execution order relative to other actions
	JobTemplate map[string]any `json:"jobTemplate,omitempty" yaml:"jobTemplate,omitempty"` // Provider-defined job structure template
	Commands    []string       `json:"commands,omitempty" yaml:"commands,omitempty"`
	PreSteps    []ActionStep   `json:"preSteps,omitempty" yaml:"preSteps,omitempty"`
	PostSteps   []ActionStep   `json:"postSteps,omitempty" yaml:"postSteps,omitempty"`
	Inputs      map[string]any `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs     []string       `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// ActionStep represents a single step within an action
type ActionStep struct {
	Name    string         `json:"name" yaml:"name"`
	Command string         `json:"command" yaml:"command"`
	Inputs  map[string]any `json:"inputs,omitempty" yaml:"inputs,omitempty"`
}

// ComponentChange tracks which component is affected by file changes
type ComponentChange struct {
	ComponentName string
	Provider      string
	ComponentType string
	Reason        string   // Why this component is affected
	AffectedPaths []string // Which paths triggered the change
}

// DependencyNode represents a node in the dependency graph
type DependencyNode struct {
	ComponentName string
	Provider      string
	Actions       []string // Which actions this component needs
	Dependencies  []string // Component names this depends on
}

// PlanRequest contains all inputs needed to generate a plan
type PlanRequest struct {
	// Git context
	BaseRef      string
	HeadRef      string
	ChangedFiles []string

	// Repository state
	RepositoryPath string
	IntentFiles    []string // Paths to intent.yaml files

	// CLI flags
	Target      string // github, gitlab, etc.
	Mode        string // plan, apply
	ChangedOnly bool
	Environment string

	// Optional overrides
	ProviderOverrides map[string]map[string]any
}
