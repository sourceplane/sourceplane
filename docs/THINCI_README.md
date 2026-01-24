# Sourceplane Thin-CI Planning Engine

A deterministic CI/CD planning system that generates execution plans without executing them.

## What is Thin-CI?

Thin-CI is Sourceplane's approach to CI/CD planning that separates **planning** from **execution**. Instead of directly generating workflow YAML files, thin-ci creates a pure data structure (execution plan) that describes what jobs should run, in what order, and with what inputs.

### Key Concepts

- **Intent-Driven**: Components and relationships defined in `intent.yaml`
- **Provider-Agnostic**: Works with Terraform, Helm, and custom providers
- **Change-Aware**: Only processes components affected by git changes
- **Dependency-Respecting**: Builds execution DAG from component relationships
- **Platform-Neutral**: Plans can render to GitHub Actions, GitLab CI, etc.

## Quick Start

```bash
# Generate a plan for GitHub Actions
sourceplane thin-ci plan --github --mode=plan
# Or using the alias:
sourceplane thinci plan --github --mode=plan

# Generate an apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=production

# Only include changed components
sourceplane thin-ci plan --github --mode=plan --changed-only
```

## Example

Given this intent:

```yaml
components:
  - name: vpc-network
    type: terraform.network
    
  - name: eks-cluster
    type: terraform.cluster
    relationships:
      - target: vpc-network
        type: depends_on
```

And these changed files:
- `terraform/vpc/main.tf`
- `terraform/eks/main.tf`

Thin-CI generates:

```json
{
  "target": "github",
  "mode": "plan",
  "jobs": [
    {
      "id": "vpc-network-validate",
      "component": "vpc-network",
      "provider": "terraform",
      "action": "validate",
      "dependsOn": []
    },
    {
      "id": "vpc-network-plan",
      "component": "vpc-network",
      "action": "plan",
      "dependsOn": ["vpc-network-validate"]
    },
    {
      "id": "eks-cluster-validate",
      "component": "eks-cluster",
      "provider": "terraform",
      "action": "validate",
      "dependsOn": []
    },
    {
      "id": "eks-cluster-plan",
      "component": "eks-cluster",
      "action": "plan",
      "dependsOn": ["eks-cluster-validate", "vpc-network-plan"]
    }
  ]
}
```

## How It Works

```
┌─────────────┐
│ Git Changes │──┐
└─────────────┘  │
                 │
┌─────────────┐  │    ┌──────────────────┐
│ Intent YAML │──┼───▶│ Change Detection │
└─────────────┘  │    └────────┬─────────┘
                 │             │
┌─────────────┐  │             ▼
│  Providers  │──┘    ┌──────────────────┐
└─────────────┘       │    Expansion     │
                      └────────┬─────────┘
                               │
                               ▼
                      ┌──────────────────┐
                      │  Dependency DAG  │
                      └────────┬─────────┘
                               │
                               ▼
                      ┌──────────────────┐
                      │ Job Generation   │
                      └────────┬─────────┘
                               │
                               ▼
                      ┌──────────────────┐
                      │ Execution Plan   │
                      └──────────────────┘
```

### Planning Pipeline

1. **Change Detection**
   - Maps changed files → affected components
   - Respects provider-specific paths
   - Detects shared module changes

2. **Component Expansion**
   - Loads provider metadata
   - Determines required actions (validate/plan/apply)
   - Extracts dependencies

3. **Dependency Graph**
   - Builds DAG from relationships
   - Topological sort (Kahn's algorithm)
   - Detects circular dependencies

4. **Job Generation**
   - Creates jobs for each action
   - Chains actions per component
   - Respects cross-component dependencies

## Supported Providers

### Terraform

Actions: `validate` → `plan` → `apply`

```yaml
components:
  - name: vpc-network
    type: terraform.network
    spec:
      path: terraform/vpc
      module:
        source: "terraform-aws-modules/vpc/aws"
        version: "5.0.0"
```

See [Terraform examples](providers/terraform/examples/)

### Helm

Actions: `validate` → `plan` → `apply`

```yaml
components:
  - name: api-service
    type: helm.service
    spec:
      chart:
        path: helm/api-service
      namespace: production
```

See [Helm examples](providers/helm/examples/)

## Command Reference

### `sourceplane thin-ci plan` (or `sourceplane thinci plan`)

Generate a CI execution plan.

**Note:** Both `thin-ci` and `thinci` work interchangeably.

**Flags:**

| Flag | Description | Default |
|------|-------------|---------|
| `--github` | Generate plan for GitHub Actions | - |
| `--gitlab` | Generate plan for GitLab CI | - |
| `--mode` | CI mode: plan, apply, destroy | `plan` |
| `--base` | Base git ref | `main` |
| `--head` | Head git ref | `HEAD` |
| `--changed-only` | Only changed components | `true` |
| `--env` | Target environment | - |
| `--output` | Output format: json, yaml | `json` |

**Examples:**

```bash
# Basic plan
sourceplane thin-ci plan --github

# Apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=prod

# Include all components (not just changed)
sourceplane thin-ci plan --github --changed-only=false

# YAML output
sourceplane thin-ci plan --github --output=yaml
```

## Plan Structure

### Plan Object

```json
{
  "target": "github",
  "mode": "plan",
  "metadata": {
    "repository": "/path/to/repo",
    "baseRef": "main",
    "headRef": "HEAD",
    "changedFiles": ["..."],
    "timestamp": "2026-01-24T10:00:00Z",
    "environment": "prod"
  },
  "jobs": [...]
}
```

### Job Object

```json
{
  "id": "component-action",
  "component": "component-name",
  "provider": "terraform",
  "action": "plan",
  "inputs": {
    "path": "terraform/component",
    "workspace": "default"
  },
  "dependsOn": ["other-job-id"],
  "metadata": {
    "runsOn": "ubuntu-latest",
    "permissions": ["id-token", "contents"],
    "env": {
      "SP_COMPONENT": "component-name",
      "SP_PROVIDER": "terraform",
      "SP_ACTION": "plan"
    },
    "timeout": 30
  }
}
```

## Design Principles

### 1. Sourceplane Owns Intent

Components and relationships are defined in `intent.yaml`, not scattered across CI configs.

### 2. Providers Own Behavior

Each provider defines its own CI actions, steps, and requirements in `provider.yaml`.

### 3. Thin-CI Owns Planning

The planner resolves dependencies, detects changes, and generates jobs without hardcoded provider logic.

### 4. CI Systems Are Render Targets

GitHub Actions, GitLab CI, etc. are output formats, not the source of truth.

## Use Cases

### 1. Pull Request CI

```bash
sourceplane thin-ci plan --github --mode=plan --changed-only
```

Only validates and plans changed components, providing fast feedback.

### 2. Production Deployment

```bash
sourceplane thin-ci plan --github --mode=apply --env=production
```

Generates apply plan with production-specific configuration.

### 3. Infrastructure Teardown

```bash
sourceplane thin-ci plan --github --mode=destroy --env=staging
```

Creates destroy plan for staging environment cleanup.

### 4. Full Validation

```bash
sourceplane thin-ci plan --github --mode=plan --changed-only=false
```

Validates all components, not just changed ones.

## Advanced Features

### Change Detection

Thin-CI detects changes through multiple mechanisms:

- **Direct file changes**: Files in component directories
- **Intent changes**: Modifications to `intent.yaml`
- **Provider changes**: Updates to provider configurations
- **Shared modules**: Changes to reusable modules

### Dependency Resolution

Supports multiple dependency types:

- **Explicit**: `depends_on` in relationships
- **Implicit**: Cross-component references in specs
- **Provider-level**: Provider-defined ordering

### Parallel Optimization

The planner maximizes parallelization:

- Independent components run in parallel
- Validation phase runs before plan phase
- Only enforces necessary dependencies

## Examples

### Multi-Tier Infrastructure

See [terraform/examples/multi-component](providers/terraform/examples/multi-component/)

Shows VPC → EKS/RDS dependency chain with parallel execution where possible.

### Microservices Application

See [helm/examples/microservices](providers/helm/examples/microservices/)

Demonstrates database → services → gateway deployment flow.

## Documentation

- **[Architecture](docs/THINCI_ARCHITECTURE.md)** - Deep dive into design and algorithms
- **[Implementation Guide](docs/THINCI_IMPLEMENTATION.md)** - How to extend and customize
- **[Terraform Examples](providers/terraform/examples/)** - Infrastructure examples
- **[Helm Examples](providers/helm/examples/)** - Application deployment examples

## Roadmap

- [ ] **Workflow Rendering**: Convert plans to actual GitHub Actions YAML
- [ ] **Caching**: Skip unchanged components based on content hash
- [ ] **Plan Diff**: Compare two plans
- [ ] **Visualization**: Graphical dependency graph
- [ ] **Cost Estimation**: Predict CI costs from plan
- [ ] **Matrix Jobs**: Multi-environment execution
- [ ] **Dry Run**: Simulate execution without running
- [ ] **More Providers**: Pulumi, CDK, Ansible, etc.
- [ ] **More Targets**: CircleCI, Jenkins, Azure DevOps

## Contributing

We welcome contributions! Areas where you can help:

1. **New Providers**: Add support for new IaC tools
2. **New Targets**: Support additional CI platforms
3. **Examples**: Share real-world use cases
4. **Documentation**: Improve guides and tutorials
5. **Testing**: Add test cases and fixtures

## FAQ

### Q: Why not just generate workflow YAML directly?

**A:** Separation of concerns. The plan is a pure data structure that can be:
- Tested independently
- Transformed/filtered
- Rendered to multiple CI platforms
- Analyzed for cost/complexity

### Q: How does change detection work?

**A:** The planner maps file paths to components using:
- Component spec paths
- Provider conventions (e.g., `terraform/<component-name>`)
- Shared module references
- Provider configuration files

### Q: What if I have circular dependencies?

**A:** The planner detects cycles and returns an error. Component dependencies must form a DAG (directed acyclic graph).

### Q: Can I run only specific components?

**A:** Yes, use `--changed-only` with appropriate git refs, or manually edit the plan JSON before execution.

### Q: How do I add a custom provider?

**A:** Create a `provider.yaml` with thin-ci configuration. See [Implementation Guide](docs/THINCI_IMPLEMENTATION.md#adding-a-new-provider).

### Q: Does this work with monorepos?

**A:** Yes! Thin-CI is designed for monorepos with multiple components and providers.

## License

MIT License - see [LICENSE](LICENSE) for details
