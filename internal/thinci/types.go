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

// Job represents a single CI job to be executed
type Job struct {
	ID        string         `json:"id"`
	Component string         `json:"component"`
	Provider  string         `json:"provider"`
	Action    string         `json:"action"` // plan, apply, destroy, validate
	Inputs    map[string]any `json:"inputs"`
	DependsOn []string       `json:"dependsOn"`
	Metadata  JobMetadata    `json:"metadata"`
	Condition *JobCondition  `json:"condition,omitempty"`
}

// JobMetadata contains platform-specific job configuration
type JobMetadata struct {
	RunsOn          string            `json:"runsOn,omitempty"`      // e.g., "ubuntu-latest"
	Permissions     []string          `json:"permissions,omitempty"` // e.g., ["id-token", "contents"]
	Environment     map[string]string `json:"env,omitempty"`
	Timeout         int               `json:"timeout,omitempty"` // in minutes
	ContinueOnError bool              `json:"continueOnError,omitempty"`
}

// JobCondition defines when a job should run
type JobCondition struct {
	ChangedPaths []string `json:"changedPaths,omitempty"` // Run only if these paths changed
	Always       bool     `json:"always,omitempty"`       // Always run regardless of changes
}

// ProviderAction describes what a provider can do in CI
type ProviderAction struct {
	Name        string         `json:"name"` // plan, apply, destroy, validate
	Description string         `json:"description"`
	Order       int            `json:"order"` // Execution order relative to other actions
	PreSteps    []ActionStep   `json:"preSteps,omitempty"`
	PostSteps   []ActionStep   `json:"postSteps,omitempty"`
	Inputs      map[string]any `json:"inputs,omitempty"`
	Outputs     []string       `json:"outputs,omitempty"`
}

// ActionStep represents a single step within an action
type ActionStep struct {
	Name    string         `json:"name"`
	Command string         `json:"command"`
	Inputs  map[string]any `json:"inputs,omitempty"`
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
