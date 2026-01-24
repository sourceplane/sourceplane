# Microservices Application Example

This example demonstrates a complete microservices application with proper dependency management using Helm and thin-ci.

## Architecture

```
┌────────────────────────────┐
│      api-gateway           │
│   (Frontend/Gateway)       │
└──────┬────────────┬────────┘
       │            │
       │            │
┌──────▼──────┐  ┌─▼───────────┐
│user-service │  │order-service│
│ (Backend)   │  │  (Backend)  │
└──────┬──────┘  └─┬───────────┘
       │            │
       └────┬───────┘
            │
     ┌──────▼──────┐
     │ postgres-db │
     │   (Data)    │
     └─────────────┘
```

## Components

### Layer 1: Data (postgres-db)
- PostgreSQL database using Bitnami Helm chart
- Persistent storage enabled
- Namespace: `data`

### Layer 2: Backend Services
- **user-service**: User management API
- **order-service**: Order processing API
- Both depend on postgres-db
- Namespace: `backend`

### Layer 3: Frontend (api-gateway)
- API Gateway routing requests to backend services
- Depends on both user-service and order-service
- Ingress enabled for external access
- Namespace: `frontend`

## Expected Thin-CI Plan

When all components are changed:

```json
{
  "jobs": [
    // Layer 1: Database (foundation)
    {"id": "postgres-db-validate", "dependsOn": []},
    {"id": "postgres-db-plan", "dependsOn": ["postgres-db-validate"]},
    
    // Layer 2: Backend Services (after database)
    {"id": "user-service-validate", "dependsOn": []},
    {"id": "user-service-plan", "dependsOn": ["user-service-validate", "postgres-db-plan"]},
    
    {"id": "order-service-validate", "dependsOn": []},
    {"id": "order-service-plan", "dependsOn": ["order-service-validate", "postgres-db-plan"]},
    
    // Layer 3: API Gateway (after backend services)
    {"id": "api-gateway-validate", "dependsOn": []},
    {"id": "api-gateway-plan", "dependsOn": ["api-gateway-validate", "user-service-plan", "order-service-plan"]}
  ]
}
```

## Execution Order

1. **Parallel Execution** (validation phase):
   - All validate jobs run in parallel
   - No dependencies between validation jobs

2. **Layer 1** (data):
   - `postgres-db-plan` runs after postgres-db-validate

3. **Layer 2** (backend, parallel after database):
   - `user-service-plan` waits for postgres-db-plan
   - `order-service-plan` waits for postgres-db-plan
   - Both can run in parallel

4. **Layer 3** (frontend, after backend):
   - `api-gateway-plan` waits for both user-service-plan and order-service-plan

## Testing Change Detection

### Scenario 1: Only database changed

```bash
# Simulate database values change
echo "# changed" >> helm/postgres/values.yaml

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: Only postgres-db jobs
```

### Scenario 2: Backend service changed

```bash
# Simulate user-service change
touch helm/user-service/templates/deployment.yaml

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: user-service jobs (database not included since not changed)
```

### Scenario 3: Database + Gateway changed

```bash
# Simulate changes
touch helm/postgres/values.yaml
touch helm/api-gateway/values.yaml

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected:
# - postgres-db jobs
# - api-gateway jobs
# - api-gateway depends on postgres-db completion
# - Backend services NOT included (not changed, even though gateway depends on them)
```

### Scenario 4: All components changed

```bash
# Simulate intent change (affects all)
touch intent.yaml

# Generate plan
sourceplane thin-ci plan --github --changed-only

# Expected: All component jobs with proper dependency chain
```

## Running the Example

```bash
cd providers/helm/examples/microservices

# Generate plan (mocked git diff)
sourceplane thin-ci plan --github --mode=plan

# View plan in YAML
sourceplane thin-ci plan --github --mode=plan --output=yaml

# Generate apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=production
```

## GitHub Actions Workflow Example

Here's how the generated plan translates to GitHub Actions:

```yaml
name: Microservices CI/CD
on:
  pull_request:
    paths:
      - 'helm/**'
      - 'intent.yaml'

jobs:
  # Layer 1: Database
  postgres-db-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Helm
        uses: azure/setup-helm@v3
        with:
          version: '3.12.0'
      - name: Add Bitnami Repo
        run: helm repo add bitnami https://charts.bitnami.com/bitnami
      - name: Validate Chart
        run: helm template app-postgres bitnami/postgresql --version 12.5.8

  postgres-db-plan:
    runs-on: ubuntu-latest
    needs: [postgres-db-validate]
    steps:
      - uses: actions/checkout@v3
      - name: Setup Helm
        uses: azure/setup-helm@v3
      - name: Generate Manifests
        run: |
          helm template app-postgres bitnami/postgresql \
            --version 12.5.8 \
            --namespace data \
            --values values.yaml > manifests.yaml

  # Layer 2: Backend Services (parallel)
  user-service-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Lint Chart
        run: helm lint helm/user-service

  user-service-plan:
    runs-on: ubuntu-latest
    needs: [user-service-validate, postgres-db-plan]
    steps:
      - uses: actions/checkout@v3
      - name: Generate Manifests
        run: helm template user-service helm/user-service

  order-service-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Lint Chart
        run: helm lint helm/order-service

  order-service-plan:
    runs-on: ubuntu-latest
    needs: [order-service-validate, postgres-db-plan]
    steps:
      - uses: actions/checkout@v3
      - name: Generate Manifests
        run: helm template order-service helm/order-service

  # Layer 3: API Gateway
  api-gateway-validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Lint Chart
        run: helm lint helm/api-gateway

  api-gateway-plan:
    runs-on: ubuntu-latest
    needs: [api-gateway-validate, user-service-plan, order-service-plan]
    steps:
      - uses: actions/checkout@v3
      - name: Generate Manifests
        run: helm template api-gateway helm/api-gateway
```

## Key Learnings

1. **Multi-Layer Dependencies**: The system properly handles 3 layers of dependencies
2. **Parallel Optimization**: Independent components at the same layer run in parallel
3. **Validation First**: All validations can run in parallel before plans
4. **Change Isolation**: Only changed components generate jobs with `--changed-only`
5. **Transitive Dependencies**: Gateway depends on services, services depend on database

## Best Practices Demonstrated

1. **Namespace Isolation**: Different namespaces for data, backend, and frontend layers
2. **Resource Limits**: All components have resource requests and limits
3. **External Charts**: Database uses maintained external chart (Bitnami)
4. **Local Charts**: Services use local charts for custom applications
5. **Service Discovery**: DNS-based service discovery between components
6. **Explicit Dependencies**: Clear dependency declarations in relationships
