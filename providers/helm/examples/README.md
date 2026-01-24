# Helm Provider Examples

This directory contains working examples of the Helm provider for Sourceplane, including thin-ci planning examples.

## Examples

### 1. [minimal/](./minimal/) - Minimal Service
The simplest possible Helm service deployment using local chart.

### 2. [remote-chart/](./remote-chart/) - Remote Chart
Deploying from a Helm repository (e.g., Bitnami).

### 3. [production/](./production/) - Production Setup
Full production configuration with multiple services, dependencies, and custom hooks.

### 4. [microservices/](./microservices/) - Microservices with Thin-CI
Complete microservices application showing thin-ci planning with dependencies.

## Quick Start

```bash
# Try the minimal example
cd minimal
sp plan
sp apply

# Try with a remote chart
cd ../remote-chart
sp plan
sp apply

# Generate thin-ci plan for microservices
cd ../microservices
sourceplane thin-ci plan --github --mode=plan
```

## Thin-CI Planning

The Helm provider supports Sourceplane's thin-ci planning engine for CI/CD workflows.

### Generate a Plan

```bash
# Generate plan for GitHub Actions
sourceplane thin-ci plan --github --mode=plan

# Generate apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=production

# Only include changed components
sourceplane thin-ci plan --github --mode=plan --changed-only
```

### Provider Actions

The Helm provider supports these thin-ci actions:

1. **validate** - Lint Helm chart and validate syntax
2. **plan** - Generate Kubernetes manifests with helm template
3. **apply** - Deploy Helm chart to Kubernetes cluster
4. **destroy** - Uninstall Helm release

### Example Output

```json
{
  "target": "github",
  "mode": "plan",
  "jobs": [
    {
      "id": "api-gateway-validate",
      "component": "api-gateway",
      "provider": "helm",
      "action": "validate",
      "dependsOn": []
    },
    {
      "id": "api-gateway-plan",
      "component": "api-gateway",
      "provider": "helm",
      "action": "plan",
      "dependsOn": ["api-gateway-validate"]
    }
  ]
}
```

## File Structure

Each example contains:
- `intent.yaml` - The Sourceplane intent definition
- `charts/` - Local Helm charts (where applicable)
- `README.md` - Example-specific documentation
