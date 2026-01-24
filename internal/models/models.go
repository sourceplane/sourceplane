package models

// Component represents a component in a repository
type Component struct {
	Name string                 `yaml:"name"`
	Type string                 `yaml:"type"`
	Spec map[string]interface{} `yaml:"spec,omitempty"`
	// Deprecated: use Spec instead
	Inputs map[string]interface{} `yaml:"inputs,omitempty"`
}

// Provider configuration with defaults
type Provider struct {
	Source   string                 `yaml:"source"`
	Version  string                 `yaml:"version"`
	Defaults map[string]interface{} `yaml:"defaults,omitempty"`
}

// Repository represents an intent.yaml file (new format) or legacy sourceplane.yaml
type Repository struct {
	APIVersion    string              `yaml:"apiVersion"`
	Kind          string              `yaml:"kind"`
	Metadata      RepositoryMetadata  `yaml:"metadata"`
	Providers     map[string]Provider `yaml:"providers,omitempty"`
	Provider      string              `yaml:"provider,omitempty"` // Legacy support
	Components    []Component         `yaml:"components"`
	Relationships []Relationship      `yaml:"relationships,omitempty"`
}

// Relationship between components
type Relationship struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Type string `yaml:"type"`
}

// RepositoryMetadata contains metadata about a repository
type RepositoryMetadata struct {
	Name        string `yaml:"name"`
	Owner       string `yaml:"owner,omitempty"`
	Domain      string `yaml:"domain,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Blueprint represents a blueprint.yaml file
type Blueprint struct {
	Kind       string          `yaml:"kind"`
	APIVersion string          `yaml:"apiVersion"`
	Provider   string          `yaml:"provider"`
	Repos      []BlueprintRepo `yaml:"repos"`
}

// BlueprintRepo represents a repository definition in a blueprint
type BlueprintRepo struct {
	Name       string      `yaml:"name"`
	Components []Component `yaml:"components"`
}
