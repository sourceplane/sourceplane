# Thin-CI Examples for Terraform Provider

This directory contains examples demonstrating how the Terraform provider works with Sourceplane's thin-ci planning engine.

## Overview

The Terraform provider enables infrastructure as code components to participate in CI planning. When you run `sourceplane thin-ci plan --github`, the planner:

1. Detects which Terraform components changed
2. Determines required actions (validate → plan → apply)
3. Builds dependency graph respecting component relationships
4. Generates GitHub Actions job definitions

## Examples

### 1. Multi-Component Infrastructure

**Location**: `examples/multi-component/`

Shows a complete infrastructure stack with dependencies:
- VPC network (foundation)
- EKS cluster (depends on network)
- RDS database (depends on network)
- S3 bucket (independent)

**What it demonstrates**:
- Component dependencies
- Parallel job execution where possible
- Sequential execution where required
- Change detection across multiple components

### 2. Changed Component Only

**Location**: `examples/changed-only/`

Demonstrates the `--changed-only` flag behavior.

**What it demonstrates**:
- Only affected components generate jobs
- Unchanged components are skipped
- Dependency checking for changed components

### 3. Environment-Specific Deployment

**Location**: `examples/environment-specific/`

Shows how to use `--env` flag for environment-specific plans.

**What it demonstrates**:
- Environment-specific configuration
- Workspace selection
- Variable file selection

## Running Examples

### Generate a Plan

```bash
# From the terraform provider examples directory
cd providers/terraform/examples/multi-component

# Generate plan for GitHub Actions
sourceplane thin-ci plan --github --mode=plan --base=main --head=HEAD

# Generate plan for specific environment
sourceplane thin-ci plan --github --mode=apply --env=prod

# Only include changed components
sourceplane thin-ci plan --github --mode=plan --changed-only
```

### Expected Output

```json
{
  "target": "github",
  "mode": "plan",
  "metadata": {
    "repository": "/path/to/repo",
    "baseRef": "main",
    "headRef": "HEAD",
    "changedFiles": ["terraform/vpc/main.tf"],
    "timestamp": "2026-01-24T10:30:00Z"
  },
  "jobs": [
    {
      "id": "vpc-network-validate",
      "component": "vpc-network",
      "provider": "terraform",
      "action": "validate",
      "inputs": {
        "path": "terraform/vpc",
        "terraformVersion": "1.5.0"
      },
      "dependsOn": [],
      "metadata": {
        "runsOn": "ubuntu-latest",
        "permissions": ["id-token", "contents"],
        "env": {
          "SP_COMPONENT": "vpc-network",
          "SP_PROVIDER": "terraform",
          "SP_ACTION": "validate"
        }
      }
    },
    {
      "id": "vpc-network-plan",
      "component": "vpc-network",
      "provider": "terraform",
      "action": "plan",
      "inputs": {
        "path": "terraform/vpc",
        "workspace": "default",
        "terraformVersion": "1.5.0"
      },
      "dependsOn": ["vpc-network-validate"],
      "metadata": {
        "runsOn": "ubuntu-latest",
        "permissions": ["id-token", "contents"],
        "env": {
          "SP_COMPONENT": "vpc-network",
          "SP_PROVIDER": "terraform",
          "SP_ACTION": "plan"
        }
      }
    }
  ]
}
```

## Integration with GitHub Actions

The generated plan can be converted into GitHub Actions workflow:

```yaml
name: Terraform CI
on: [pull_request]

jobs:
  vpc-network-validate:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: read
    env:
      SP_COMPONENT: vpc-network
      SP_PROVIDER: terraform
      SP_ACTION: validate
    steps:
      - uses: actions/checkout@v3
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
      - name: Terraform Format Check
        run: terraform fmt -check
        working-directory: terraform/vpc
      - name: Terraform Validate
        run: terraform validate
        working-directory: terraform/vpc

  vpc-network-plan:
    runs-on: ubuntu-latest
    needs: [vpc-network-validate]
    permissions:
      id-token: write
      contents: read
    env:
      SP_COMPONENT: vpc-network
      SP_PROVIDER: terraform
      SP_ACTION: plan
    steps:
      - uses: actions/checkout@v3
      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2
        with:
          terraform_version: 1.5.0
      - name: Terraform Init
        run: terraform init
        working-directory: terraform/vpc
      - name: Terraform Plan
        run: terraform plan -out=plan.tfplan
        working-directory: terraform/vpc
      - name: Save Plan
        run: terraform show -json plan.tfplan > plan.json
        working-directory: terraform/vpc
      - name: Upload Plan
        uses: actions/upload-artifact@v3
        with:
          name: vpc-network-plan
          path: terraform/vpc/plan.json
```

## Provider Actions

The Terraform provider supports these thin-ci actions:

### 1. validate
- **Purpose**: Check syntax and configuration
- **Pre-steps**: `terraform fmt -check`
- **Command**: `terraform validate`
- **Outputs**: validationReport

### 2. plan
- **Purpose**: Generate execution plan
- **Pre-steps**: `terraform init`
- **Command**: `terraform plan -out=plan.tfplan`
- **Post-steps**: `terraform show -json`
- **Outputs**: plan.tfplan, plan.json, planSummary

### 3. apply
- **Purpose**: Apply infrastructure changes
- **Command**: `terraform apply plan.tfplan`
- **Post-steps**: `terraform output -json`
- **Outputs**: terraform.tfstate, outputs.json

### 4. destroy
- **Purpose**: Destroy infrastructure
- **Command**: `terraform destroy`
- **Post-steps**: State cleanup

## Change Detection

The planner detects Terraform component changes through:

1. **Direct file changes**
   - `*.tf` files in component directory
   - `*.tfvars` files
   - `terraform.tfstate` (though typically ignored)

2. **Module changes**
   - Local modules referenced in component
   - Shared modules in `terraform/modules`

3. **Provider configuration changes**
   - `providers/terraform/provider.yaml`
   - `providers/terraform/schema.yaml`

4. **Intent definition changes**
   - `intent.yaml` modifications

## Best Practices

1. **Use explicit dependencies**
   ```yaml
   - name: eks-cluster
     type: terraform.cluster
     relationships:
       - target: vpc-network
         type: depends_on
   ```

2. **Structure components by lifecycle**
   - Foundation (VPC, networks)
   - Platform (clusters, databases)
   - Applications (services)

3. **Use workspaces for environments**
   ```yaml
   spec:
     workspace: prod
     backend:
       config:
         bucket: my-terraform-state-prod
   ```

4. **Leverage provider defaults**
   - Set common values in provider.yaml thinCI.defaults
   - Override per-component in spec

## Troubleshooting

### No jobs generated

Check:
- Are there changed files?
- Is `--changed-only` filtering out your components?
- Are intent files properly formatted?

### Circular dependency error

Review component relationships - you cannot have cycles:
```yaml
# ❌ Invalid
eks → rds → eks

# ✅ Valid
vpc → eks
vpc → rds
```

### Missing provider

Ensure provider.yaml exists and is properly formatted:
```bash
providers/terraform/provider.yaml
```
