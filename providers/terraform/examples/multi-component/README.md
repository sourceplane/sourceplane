# Multi-Component Infrastructure Example

This example demonstrates a complete infrastructure stack with proper dependency management.

## Architecture

```
┌─────────────────────────────────────────────┐
│              vpc-network                     │
│         (VPC, Subnets, NAT)                  │
└─────────┬──────────────────┬─────────────────┘
          │                   │
          │                   │
┌─────────▼──────────┐  ┌────▼──────────────┐
│   eks-cluster      │  │   rds-database    │
│   (EKS + nodes)    │  │   (PostgreSQL)    │
└────────────────────┘  └───────────────────┘

┌────────────────────┐
│   app-storage      │
│   (S3 bucket)      │
│   (independent)    │
└────────────────────┘
```

## Expected Thin-CI Plan

When all components are changed:

```json
{
  "jobs": [
    // Layer 1: Foundation (parallel where possible)
    {"id": "vpc-network-validate", "dependsOn": []},
    {"id": "vpc-network-plan", "dependsOn": ["vpc-network-validate"]},
    
    {"id": "app-storage-validate", "dependsOn": []},
    {"id": "app-storage-plan", "dependsOn": ["app-storage-validate"]},
    
    // Layer 2: Platform (after vpc-network)
    {"id": "eks-cluster-validate", "dependsOn": []},
    {"id": "eks-cluster-plan", "dependsOn": ["eks-cluster-validate", "vpc-network-plan"]},
    
    {"id": "rds-database-validate", "dependsOn": []},
    {"id": "rds-database-plan", "dependsOn": ["rds-database-validate", "vpc-network-plan"]}
  ]
}
```

## Execution Order

1. **Parallel Execution** (no dependencies):
   - `vpc-network-validate`
   - `app-storage-validate`
   - `eks-cluster-validate`
   - `rds-database-validate`

2. **Sequential** (after validation):
   - `vpc-network-plan` → runs after vpc-network-validate

3. **Parallel after VPC** (dependent on vpc-network):
   - `eks-cluster-plan` → waits for vpc-network-plan
   - `rds-database-plan` → waits for vpc-network-plan

4. **Independent**:
   - `app-storage-plan` → only waits for app-storage-validate

## Testing Change Detection

### Scenario 1: Only VPC changed

```bash
# Simulate VPC change
touch terraform/vpc/main.tf

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: Only vpc-network jobs
```

### Scenario 2: VPC + Database changed

```bash
# Simulate changes
touch terraform/vpc/main.tf
touch terraform/rds/main.tf

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: vpc-network jobs + rds-database jobs
# rds-database jobs will depend on vpc-network-plan
```

### Scenario 3: All components changed

```bash
# Simulate intent change (affects all)
touch intent.yaml

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: All component jobs with proper dependencies
```

## Running the Example

```bash
cd providers/terraform/examples/multi-component

# Generate plan (mocked git diff)
sourceplane thin-ci plan --github --mode=plan

# View plan in YAML
sourceplane thin-ci plan --github --mode=plan --output=yaml

# Generate apply plan
sourceplane thin-ci plan --github --mode=apply --env=prod
```

## Key Learnings

1. **Dependency Resolution**: Components with `depends_on` relationships ensure proper ordering
2. **Parallel Execution**: Independent components can run in parallel
3. **Action Chaining**: Each component's actions (validate → plan → apply) are chained
4. **Change Isolation**: Only changed components generate jobs when `--changed-only` is used
