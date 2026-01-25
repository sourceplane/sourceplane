# Provider-Driven Job Structure

## Overview

The job data structure is now fully provider-driven, giving each provider complete flexibility to define their own job schema and logic through the `jobTemplate` field in their provider.yaml.

## Architecture

### 1. Flexible Job Type

The `Job` type is now a flexible `map[string]any` instead of a rigid struct:

```go
// Job represents a single CI job with flexible provider-defined structure
type Job map[string]any
```

This allows providers to include any fields they need without being constrained by a fixed schema.

### 2. Core Fields

All jobs must include these minimal core fields:
- `id`: Unique job identifier
- `component`: Component name
- `provider`: Provider name
- `action`: Action type (validate, plan, apply, destroy)
- `dependsOn`: Array of job IDs this job depends on

### 3. Provider Job Templates

Providers define their job structure in the `jobTemplate` section of each action in provider.yaml:

```yaml
actions:
  - name: validate
    order: 1
    jobTemplate:
      # Commands to execute
      commands:
        - helm lint {{.chartPath}}
        - helm template {{.releaseName}} {{.chartPath}} --validate
      
      # Metadata specific to this job type
      metadata:
        runsOn: ubuntu-latest
        timeout: 10
        continueOnError: false
      
      # Artifacts produced
      artifacts:
        - name: lint-report
          path: lint-report.txt
      
      # Any other provider-specific fields
      customField: customValue
```

## Example: Helm Provider

### Provider Definition (provider.yaml)

```yaml
thinCI:
  actions:
    - name: validate
      jobTemplate:
        commands:
          - helm lint {{.chartPath}}
        metadata:
          runsOn: ubuntu-latest
          timeout: 10
        artifacts:
          - name: lint-report
            path: lint-report.txt
    
    - name: plan
      jobTemplate:
        commands:
          - helm template {{.releaseName}} {{.chartPath}}
          - helm diff upgrade {{.releaseName}} {{.chartPath}}
        metadata:
          runsOn: ubuntu-latest
          timeout: 15
          permissions:
            - contents: read
            - id-token: write
        cache:
          enabled: true
          key: helm-{{.releaseName}}
          paths:
            - ~/.cache/helm
        artifacts:
          - name: manifests
            path: manifests/
    
    - name: apply
      jobTemplate:
        commands:
          - helm upgrade --install {{.releaseName}} {{.chartPath}}
        metadata:
          runsOn: ubuntu-latest
          timeout: 30
          environment:
            name: "{{.environment}}"
            url: "https://{{.namespace}}.example.com"
        requiresApproval: true
        retryPolicy:
          maxAttempts: 3
          backoff: exponential
        notifications:
          onSuccess:
            - slack: "#deployments"
          onFailure:
            - slack: "#alerts"
```

### Generated Job Output

When the planner generates a job, it merges the provider's template with runtime values:

```json
{
  "id": "hello-app-apply",
  "component": "hello-app",
  "provider": "helm",
  "action": "apply",
  "dependsOn": ["hello-app-plan"],
  
  "commands": [
    "helm upgrade --install hello-app ./charts/hello-app"
  ],
  
  "metadata": {
    "runsOn": "ubuntu-latest",
    "timeout": 30,
    "environment": {
      "name": "production",
      "url": "https://default.example.com"
    }
  },
  
  "requiresApproval": true,
  
  "retryPolicy": {
    "maxAttempts": 3,
    "backoff": "exponential"
  },
  
  "notifications": {
    "onSuccess": [
      {"slack": "#deployments"}
    ],
    "onFailure": [
      {"slack": "#alerts"}
    ]
  },
  
  "inputs": {
    "chartPath": "./charts/hello-app",
    "namespace": "default",
    "releaseName": "hello-app"
  },
  
  "preSteps": [
    {
      "name": "verify-cluster",
      "command": "kubectl cluster-info"
    }
  ],
  
  "postSteps": [
    {
      "name": "verify-deployment",
      "command": "kubectl rollout status"
    }
  ]
}
```

## Benefits

### 1. **Provider Flexibility**
Providers can define any job structure they need:
- Custom metadata fields
- Provider-specific configuration
- Advanced features (caching, artifacts, notifications)
- Retry policies, approval gates, etc.

### 2. **No Code Changes Required**
Adding new provider capabilities doesn't require modifying the core planner code. Providers simply update their YAML configuration.

### 3. **Backward Compatibility**
The system still ensures core fields are present while allowing unlimited extension.

### 4. **Type Safety for Core Fields**
Helper methods provide type-safe access to required fields:
```go
job.GetID()
job.GetComponent()
job.GetProvider()
job.GetAction()
job.GetDependsOn()
```

## Creating a Custom Provider

To create a provider with custom job logic:

1. **Define your jobTemplate** in provider.yaml:
```yaml
thinCI:
  actions:
    - name: custom-action
      jobTemplate:
        # Your custom fields
        executionMode: parallel
        sharding:
          enabled: true
          count: 5
        healthChecks:
          - endpoint: /health
            interval: 30s
        # Standard fields
        commands:
          - ./my-custom-command
        metadata:
          runsOn: custom-runner
```

2. **The planner automatically merges your template** with core fields and runtime values

3. **Your CI renderer** can access all custom fields:
```go
if sharding, ok := job["sharding"].(map[string]any); ok {
    if enabled, ok := sharding["enabled"].(bool); enabled {
        // Use sharding configuration
    }
}
```

## Template Variables

Job templates support Go template syntax for dynamic values:

- `{{.chartPath}}` - From inputs
- `{{.releaseName}}` - From component name or inputs
- `{{.namespace}}` - From inputs or defaults
- `{{.environment}}` - From CLI flags
- `{{.checksum}}` - Computed values

These are interpolated at runtime when the job is executed.

## Migration Guide

If you have existing providers, update them to use jobTemplate:

**Before:**
```yaml
actions:
  - name: deploy
    commands:
      - kubectl apply -f manifest.yaml
```

**After:**
```yaml
actions:
  - name: deploy
    jobTemplate:
      commands:
        - kubectl apply -f manifest.yaml
      metadata:
        runsOn: ubuntu-latest
      # Add any custom fields you need
      kubectl:
        version: "1.28"
        context: production
```

The planner will automatically use the jobTemplate to construct the job with all your custom fields included in the final output.
