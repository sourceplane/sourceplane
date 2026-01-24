# Sourceplane CLI Examples

This directory contains examples demonstrating how to use the Sourceplane CLI.

## Example 1: Single Repository

See [intent.yaml](../intent.yaml) in the root directory for a comprehensive example showing:
- Multiple providers (Helm and Terraform)
- Different component types
- Provider-level defaults
- Component relationships

### Commands to try:

```bash
# List components
sp component list

# View component tree
sp component tree

# Describe a specific component
sp component describe auth-service

# Validate the intent definition
sp lint
```

## Example 2: Helm Provider

See [providers/helm/examples/](../providers/helm/examples/) for Helm-specific examples:
- Minimal service deployment
- Remote chart usage
- Production configurations

## Example 3: Blueprint Workflow

1. **Initialize a blueprint:**
   ```bash
   sp init blueprint
   ```

2. **Edit blueprint.yaml** to define your repositories

3. **Preview what will be created:**
   ```bash
   sp plan
   ```

4. **Create the repositories:**
   ```bash
   sp apply
   ```

## Example 4: Organization Analysis

If you have multiple repositories with intent.yaml files:

```bash
# View org-wide component tree
sp org tree

# Generate architectural graph
sp org graph
```

## Example intent.yaml Structure

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: my-service
  owner: team-backend

providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"
    defaults:
      service:
        namespace: default
        values:
          replicas: 2

components:
  - name: api
    type: helm.service
    spec:
      chart:
        path: ./charts/api
      values:
        image:
          repository: myorg/api
          tag: "1.0.0"
```

## Example blueprint.yaml

```yaml
kind: Blueprint
apiVersion: sourceplane.io/v1

provider: github-terraform@v1

repos:
  - name: payments-api
    components:
      - type: service.api
        name: payments
        inputs:
          language: node
          port: 3000
          
  - name: users-api
    components:
      - type: service.api
        name: users
        inputs:
          language: python
          port: 8000
```
