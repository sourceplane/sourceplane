# Thin-CI Implementation Guide

This document provides a comprehensive guide to understanding and extending the Sourceplane thin-ci planning engine.

## Quick Start

### Basic Usage

```bash
# Generate a plan for GitHub Actions
sourceplane thin-ci plan --github --mode=plan

# Generate an apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=production

# Only include changed components
sourceplane thin-ci plan --github --mode=plan --changed-only

# Output in YAML format
sourceplane thin-ci plan --github --mode=plan --output=yaml
```

### Command Line Flags

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `--github` | Generate plan for GitHub Actions | - | `--github` |
| `--gitlab` | Generate plan for GitLab CI | - | `--gitlab` |
| `--mode` | CI mode: plan, apply, or destroy | `plan` | `--mode=apply` |
| `--base` | Base git ref for comparison | `main` | `--base=develop` |
| `--head` | Head git ref for comparison | `HEAD` | `--head=feature-branch` |
| `--changed-only` | Only include changed components | `true` | `--changed-only=false` |
| `--env` | Target environment | - | `--env=prod` |
| `--output` | Output format: json or yaml | `json` | `--output=yaml` |

## How It Works

### Planning Pipeline

```
┌────────────────┐
│  Git Changes   │───┐
└────────────────┘   │
                     │
┌────────────────┐   │      ┌─────────────────────┐
│ Intent Files   │───┼─────▶│ Change Detection    │
└────────────────┘   │      └──────────┬──────────┘
                     │                 │
┌────────────────┐   │                 │
│ Provider Defs  │───┘                 │
└────────────────┘                     │
                                       ▼
                              ┌─────────────────────┐
                              │ Component Expansion │
                              └──────────┬──────────┘
                                         │
                                         ▼
                              ┌─────────────────────┐
                              │  Dependency Graph   │
                              └──────────┬──────────┘
                                         │
                                         ▼
                              ┌─────────────────────┐
                              │  Job Generation     │
                              └──────────┬──────────┘
                                         │
                                         ▼
                              ┌─────────────────────┐
                              │   Execution Plan    │
                              └─────────────────────┘
```

### 1. Change Detection

The change detector maps file changes to affected components:

**Detection Rules**:
- **Intent changes**: `intent.yaml` → all components
- **Component files**: Files in component path → that component
- **Provider config**: `providers/<name>/provider.yaml` → all components of that provider
- **Shared modules**: Module files → all components using that module

**Example**:
```
Changed: terraform/vpc/main.tf
Result:  vpc-network component affected
```

### 2. Component Expansion

For each affected component:
1. Load provider metadata
2. Determine required actions based on mode
3. Extract dependencies from relationships

**Mode → Actions Mapping**:
- `plan`: validate → plan
- `apply`: validate → plan → apply
- `destroy`: destroy

### 3. Dependency Graph

Build a directed acyclic graph (DAG):
1. Create nodes for each component
2. Add edges from dependencies
3. Topological sort (Kahn's algorithm)
4. Detect cycles

**Example**:
```
eks-cluster depends on vpc-network

Graph:
vpc-network ──▶ eks-cluster

Sorted:
[vpc-network, eks-cluster]
```

### 4. Job Generation

Generate CI jobs from sorted nodes:
1. For each component, create a job per action
2. Chain actions: validate → plan → apply
3. Respect cross-component dependencies
4. Add platform-specific metadata

**Dependency Chaining**:
```
Component A: validate → plan
Component B: validate → plan (depends on A)

Jobs:
- A-validate (no deps)
- A-plan (depends on A-validate)
- B-validate (no deps)
- B-plan (depends on B-validate AND A-plan)
```

## Adding a New Provider

### 1. Create Provider Definition

Create `providers/my-provider/provider.yaml`:

```yaml
name: my-provider
version: 0.1.0
apiVersion: sourceplane.io/v1
kind: Provider

kinds:
  - name: service
    fullType: my-provider.service
    description: Deploy a service

thinCI:
  actions:
    - name: validate
      description: Validate configuration
      order: 1
      preSteps: []
      postSteps: []
      inputs:
        path: "."
      outputs: []
      
    - name: plan
      description: Generate deployment plan
      order: 2
      preSteps:
        - name: init
          command: my-provider init
      postSteps:
        - name: save-plan
          command: my-provider show
      inputs:
        path: "."
        config: ""
      outputs:
        - plan.json
        
    - name: apply
      description: Deploy service
      order: 3
      preSteps:
        - name: verify
          command: my-provider verify
      postSteps:
        - name: health-check
          command: my-provider status
      inputs:
        path: "."
        planFile: ""
      outputs:
        - deployment.json
        
  defaults:
    timeout: 900
    version: "1.0.0"
    
  ordering:
    - validate
    - plan
    - apply
```

### 2. Create Schema

Create `providers/my-provider/schema.yaml` with JSON schema for component specs.

### 3. Add Examples

Create examples in `providers/my-provider/examples/`:
- `minimal/` - Simplest possible usage
- `complete/` - Full-featured example
- `README.md` - Documentation

### 4. Register Provider

The provider is automatically discovered from the `providers/` directory.

## Extending the Planner

### Adding a New Target Platform

To add support for a new CI platform (e.g., CircleCI):

1. **Update Job Metadata**

Edit `internal/thinci/planner.go`:

```go
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
    case "circleci":
        metadata.RunsOn = "docker"
        metadata.Image = "cimg/base:stable"
        metadata.Timeout = 30
    }

    return metadata
}
```

2. **Add CLI Flag**

Edit `cmd/thinci.go`:

```go
thinCIPlanCmd.Flags().StringVar(&thinCITarget, "circleci", "", "Generate plan for CircleCI")
```

3. **Update Validation**

```go
func runThinCIPlan(cmd *cobra.Command, args []string) error {
    target := ""
    if cmd.Flags().Changed("github") {
        target = "github"
    } else if cmd.Flags().Changed("gitlab") {
        target = "gitlab"
    } else if cmd.Flags().Changed("circleci") {
        target = "circleci"
    }
    // ...
}
```

### Adding Custom Change Detection Rules

To customize how changes are detected:

Edit `internal/thinci/detector.go`:

```go
func (cd *ChangeDetector) getComponentPaths(component models.Component, provider string) []string {
    paths := []string{}

    switch provider {
    case "terraform":
        // Existing logic
    case "helm":
        // Existing logic
    case "my-provider":
        // Custom logic
        if source, ok := component.Spec["source"].(string); ok {
            paths = append(paths, source)
        }
    }

    return paths
}
```

### Adding Provider-Specific Logic

To add provider-specific behavior:

Edit `internal/thinci/planner.go`:

```go
func (p *Planner) determineActions(mode string, provider *ProviderMetadata) []string {
    actions := []string{}

    // Get provider's supported actions
    supportedActions := provider.ThinCI.Actions

    switch mode {
    case "plan":
        // Standard logic
    case "apply":
        // Standard logic
    case "custom-mode":
        // Custom logic for special modes
        if p.hasAction(supportedActions, "custom-action") {
            actions = append(actions, "custom-action")
        }
    }

    return actions
}
```

## Testing

### Unit Testing

Test individual components:

```go
func TestChangeDetection(t *testing.T) {
    detector := thinci.NewChangeDetector("/repo", intents)
    
    changes, err := detector.DetectChanges([]string{
        "terraform/vpc/main.tf",
    })
    
    assert.NoError(t, err)
    assert.Len(t, changes, 1)
    assert.Equal(t, "vpc-network", changes[0].ComponentName)
}
```

### Integration Testing

Test the full pipeline:

```bash
# Setup test repository
cd test-fixtures/sample-repo

# Generate plan
sourceplane thin-ci plan --github --mode=plan > plan.json

# Verify plan structure
jq '.jobs | length' plan.json  # Should return number of jobs
jq '.jobs[0].dependsOn' plan.json  # Check dependencies
```

### Example Test Cases

1. **Single component change**
   - Change one file
   - Verify only that component in plan

2. **Dependency chain**
   - Change base component
   - Verify dependent components included

3. **Parallel execution**
   - Change independent components
   - Verify no artificial dependencies

4. **Circular dependency detection**
   - Create circular dependency
   - Verify error is raised

## Troubleshooting

### No jobs generated

**Symptoms**: Empty plan with no jobs

**Debugging**:
```bash
# Check if files are detected as changed
sourceplane thin-ci plan --github --mode=plan --changed-only=false

# Verify intent files are found
ls -la intent.yaml

# Check provider registration
ls -la providers/
```

**Solutions**:
- Ensure changed files match component paths
- Disable `--changed-only` flag
- Verify intent.yaml syntax
- Check provider.yaml exists

### Circular dependency error

**Symptoms**: Error: "circular dependency detected in component graph"

**Debugging**:
```yaml
# Check relationships in intent.yaml
relationships:
  - from: A
    to: B
  - from: B
    to: A  # ❌ Circular!
```

**Solutions**:
- Review component relationships
- Ensure DAG structure
- Remove circular references

### Wrong execution order

**Symptoms**: Jobs running in incorrect order

**Debugging**:
```bash
# Generate plan and check dependencies
sourceplane thin-ci plan --github --output=yaml | grep -A5 "dependsOn"
```

**Solutions**:
- Verify relationships in intent.yaml
- Check provider action ordering
- Ensure topological sort is correct

### Provider not found

**Symptoms**: Error: "provider 'X' not found"

**Debugging**:
```bash
# Check providers directory
ls -la providers/

# Verify provider.yaml exists
cat providers/X/provider.yaml
```

**Solutions**:
- Create provider.yaml if missing
- Fix YAML syntax errors
- Ensure provider name matches directory name

## Best Practices

### 1. Component Organization

```
terraform/
  vpc-network/
    main.tf
    variables.tf
  eks-cluster/
    main.tf
    variables.tf
    
helm/
  api-service/
    Chart.yaml
    values.yaml
    templates/
```

### 2. Explicit Dependencies

Always declare dependencies explicitly:

```yaml
components:
  - name: eks-cluster
    type: terraform.cluster
    relationships:
      - target: vpc-network
        type: depends_on
```

### 3. Provider Defaults

Set common values in provider configuration:

```yaml
thinCI:
  defaults:
    timeout: 1800
    terraformVersion: "1.5.0"
```

### 4. Environment-Specific Configuration

Use `--env` flag and conditional logic:

```yaml
spec:
  workspace: ${ENV}
  variables:
    instance_count: ${ENV == "prod" ? 5 : 2}
```

### 5. Change Detection Optimization

Structure files to minimize false positives:
- Keep component files isolated
- Use separate directories for modules
- Avoid shared configuration files

## Performance Considerations

### Scaling to Large Repositories

The planner is designed to handle large-scale repositories:

- **Time Complexity**: O(V + E) for graph operations (V=components, E=dependencies)
- **Space Complexity**: O(V) for storing nodes
- **Parallel Processing**: Independent validations run in parallel

### Optimization Tips

1. **Use --changed-only**: Reduces components to process
2. **Minimize shared modules**: Reduces cascading changes
3. **Cache provider metadata**: Load providers once
4. **Lazy loading**: Only load intent files when needed

## Next Steps

1. **Learn by Example**: Explore [examples](../providers/)
2. **Read Architecture**: See [THINCI_ARCHITECTURE.md](./THINCI_ARCHITECTURE.md)
3. **Create Provider**: Follow provider creation guide above
4. **Contribute**: Submit PRs for new features

## Resources

- [Architecture Documentation](./THINCI_ARCHITECTURE.md)
- [Terraform Provider Examples](../providers/terraform/examples/)
- [Helm Provider Examples](../providers/helm/examples/)
- [GitHub Issues](https://github.com/sourceplane/cli/issues)
