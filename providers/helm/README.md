# Helm Provider Documentation

## Overview

The Helm provider enables declarative Kubernetes deployments using Helm charts within Sourceplane's intent-driven model.

**Provider Name:** `helm`  
**Supported Types:** `helm.service`

---

## 1. Provider Metadata

### Basic Information
- **Name:** helm
- **Version:** 0.1.0
- **Type:** CLI wrapper around Helm
- **Execution Mode:** Declarative, plan/apply pattern

### Versioning Strategy
- **Format:** Semantic versioning (major.minor.patch)
- **Major:** Breaking schema changes
- **Minor:** New features, backward compatible
- **Patch:** Bug fixes only
- **Compatibility:** Backward compatible within minor versions

### Supported Kinds
| Kind | Type | Description |
|------|------|-------------|
| service | `helm.service` | Deploy a Helm chart as a Kubernetes service |

---

## 2. Type Definition: `helm.service`

### Purpose
Deploy and manage Kubernetes applications using Helm charts.

### Required Fields
- `chart` - Chart specification (repo or path)

### Optional Fields
- `release` - Helm release configuration
- `values` - Chart value overrides
- `lifecycle` - Deployment behavior controls
- `hooks` - CI/CD integration hooks
- `dependsOn` - Component dependencies

### Inferred / Convention-Based Fields

**Automatically inferred:**
- `release.name` → component name (kebab-cased)
- `release.namespace` → provider default (usually `default`)
- `values.image.tag` → from CI context (git SHA, version tag)
- `values.replicas` → provider default (usually `2`)

**Convention-based behaviors:**
- Chart paths are relative to `intent.yaml`
- Remote charts are cached locally
- Release names follow kebab-case convention

### Sensible Defaults

```yaml
release:
  namespace: default
  createNamespace: false

lifecycle:
  atomic: true
  wait: true
  timeout: "10m"
  cleanupOnFail: true

values:
  replicas: 2
  resources:
    cpu: "250m"
    memory: "256Mi"
```

---

## 3. Spec Schema

See [schema.yaml](./schema.yaml) for the complete JSON Schema definition.

### Chart Definition

**Option 1: Remote Repository**
```yaml
chart:
  repo: https://charts.bitnami.com/bitnami
  name: nginx
  version: "15.0.0"
```

**Option 2: Local Path**
```yaml
chart:
  path: ./charts/my-service
```

### Values Override
```yaml
values:
  image:
    repository: myorg/myapp
    tag: "1.0.0"
  replicas: 3
  resources:
    limits:
      cpu: "1"
      memory: "1Gi"
  env:
    DATABASE_URL: postgres://...
```

### Lifecycle Controls
```yaml
lifecycle:
  atomic: true        # Rollback on failure
  wait: true          # Wait for readiness
  timeout: "15m"      # Max wait time
  force: false        # Force updates via recreate
  recreatePods: false # Restart pods
```

### Hooks
```yaml
hooks:
  preDeploy:
    - kubectl apply -f migrations/
  postDeploy:
    - ./scripts/smoke-test.sh
  lint: true
```

---

## 4. Provider-Level Defaults

### Configuration

Define provider-wide defaults in `intent.yaml`:

```yaml
providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"
    
    defaults:
      service:
        namespace: production
        values:
          replicas: 2
          resources:
            cpu: "250m"
            memory: "256Mi"
          image:
            pullPolicy: IfNotPresent
```

### Merge Order

1. **Provider defaults** (from `providers.helm.defaults`)
2. **Component spec** (individual component definition)
3. **Environment overrides** (from CLI flags or CI variables)

### Override Rules

- **Scalars:** Completely replaced
- **Objects/Maps:** Deep-merged (recursive)
- **Arrays:** Completely replaced (not merged)

**Example:**

```yaml
# Provider defaults:
defaults:
  service:
    values:
      replicas: 2
      resources:
        cpu: "250m"
        memory: "256Mi"

# Component spec:
components:
  - name: api
    type: helm.service
    spec:
      values:
        replicas: 5          # Overrides default
        resources:
          memory: "512Mi"    # Merges with cpu from default
        ports:
          - 8080             # New field, added
```

**Result:**
```yaml
replicas: 5
resources:
  cpu: "250m"      # From default
  memory: "512Mi"  # From component
ports:
  - 8080           # From component
```

### Validation Behavior

- Schema validation occurs **after** merge
- Invalid overrides fail at plan time
- Missing required fields show clear error messages

---

## 5. Execution Model

### Architecture

The Helm provider is a **CLI wrapper** around the Helm binary:

```
Sourceplane CLI → Helm Provider → helm CLI → Kubernetes API
```

### Execution Modes

| Command | Action | Helm Equivalent |
|---------|--------|-----------------|
| `sp plan` | Dry-run preview | `helm template --dry-run` |
| `sp apply` | Deploy/upgrade | `helm upgrade --install` |
| `sp diff` | Show changes | `helm diff upgrade` |
| `sp destroy` | Uninstall | `helm uninstall` |
| `sp validate` | Lint chart | `helm lint` |

### How Execution Works

1. **Parse** `intent.yaml` and load component definitions
2. **Resolve** chart (download if remote, validate if local)
3. **Merge** values (defaults + spec + overrides)
4. **Generate** values file (temp YAML)
5. **Execute** Helm command with generated values
6. **Capture** outputs (release info, manifests, errors)

### Manifest Rendering

- Provider generates a temporary `values.yaml`
- Helm templates the chart with merged values
- Rendered manifests are stored in `.sourceplane/rendered/`
- Manifests can be committed for GitOps workflows

### Plan/Apply/Diff Support

✅ **Plan (--dry-run)**
```bash
sp plan component api
# → helm template api ./charts/api --values /tmp/values.yaml --dry-run
```

✅ **Apply**
```bash
sp apply component api
# → helm upgrade --install api ./charts/api --values /tmp/values.yaml
```

✅ **Diff**
```bash
sp diff component api
# → helm diff upgrade api ./charts/api --values /tmp/values.yaml
```

---

## 6. Validation & Errors

### Schema Validation

**When:** Before any execution (plan/apply)

**What:**
- Required fields present
- Field types correct
- Values match schema constraints

**Example Error:**
```
Error: validation failed for component 'api'
  - spec.chart is required
  - spec.values.replicas must be a number
```

### Chart Resolution Errors

**Local Chart:**
```
Error: chart not found at path './charts/api'
  → Check that the path is relative to intent.yaml
  → Ensure the chart directory contains Chart.yaml
```

**Remote Chart:**
```
Error: failed to fetch chart 'nginx' from 'https://charts.bitnami.com'
  → Repository may be unreachable
  → Chart name or version may not exist
  → Run: helm repo add bitnami https://charts.bitnami.com
```

### Dry-Run Behavior

**Before apply:**
1. Validate schema
2. Resolve chart
3. Merge values
4. Render templates
5. Show preview (no kubectl apply)

**Failures surface early:**
- Template syntax errors
- Missing required values
- Invalid Kubernetes manifests

### Failure Surfaces

```yaml
validation:
  - Schema validation errors
  - Chart resolution failures
  
execution:
  - Helm template errors
  - Kubernetes API errors
  - Timeout errors
  
post-execution:
  - Pod crashloops (if wait: true)
  - Resource creation failures
  - Hook execution failures
```

---

## 7. Outputs & Artifacts

### Outputs

After successful apply, the provider outputs:

```yaml
outputs:
  releaseName: api
  namespace: production
  chartVersion: "1.2.3"
  revision: 5
  status: deployed
  lastDeployed: "2026-01-23T10:30:00Z"
  notes: |
    Your application is available at:
    http://api.example.com
```

### Artifacts

**Rendered Manifests:**
```
.sourceplane/
  rendered/
    api/
      deployment.yaml
      service.yaml
      ingress.yaml
      values.yaml
```

**Purpose:**
- GitOps workflows (commit manifests)
- Audit trail
- Debugging
- Manual review

### References for Graph Modeling

The provider exposes relationships for dependency graphs:

```yaml
component: api
type: helm.service
provides:
  - service: api.production.svc.cluster.local
  - http: https://api.example.com
depends:
  - component: database
    type: datastore
  - component: redis
    type: cache
```

**Graph output:**
```
api (helm.service)
  ├─→ database (postgres)
  └─→ redis (cache)
```

---

## 8. Example Files

### Example 1: Minimal Intent

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: simple-app

providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"

components:
  - name: api
    type: helm.service
    spec:
      chart:
        path: ./charts/api
```

### Example 2: Remote Chart

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: postgres-stack

providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"
    defaults:
      service:
        namespace: databases

components:
  - name: postgres
    type: helm.service
    spec:
      chart:
        repo: https://charts.bitnami.com/bitnami
        name: postgresql
        version: "15.0.0"
      values:
        auth:
          username: app
          database: myapp
        primary:
          persistence:
            size: 20Gi
```

### Example 3: Full Production Setup

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: payments-platform
  owner: platform-team

providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"
    defaults:
      service:
        namespace: production
        values:
          replicas: 3
          resources:
            cpu: "500m"
            memory: "512Mi"

components:
  - name: payments-api
    type: helm.service
    spec:
      chart:
        path: ./charts/payments-api
      release:
        namespace: production
        createNamespace: true
      values:
        image:
          repository: myorg/payments-api
          tag: "2.3.1"
        replicas: 5  # Override default
        ingress:
          enabled: true
          host: payments.example.com
      lifecycle:
        atomic: true
        wait: true
        timeout: "15m"
      hooks:
        preDeploy:
          - kubectl apply -f migrations/
        postDeploy:
          - ./scripts/smoke-test.sh
      dependsOn:
        - payments-db
```

### Example 4: Execution Plan

**Command:**
```bash
sp plan
```

**Output:**
```
📋 Execution Plan

Provider: helm@0.1.0

Components to deploy:

  ✓ payments-api (helm.service)
    Chart: ./charts/payments-api
    Release: payments-api
    Namespace: production
    
    Changes:
    ├─ Deployment/payments-api
    │  └─ replicas: 3 → 5
    ├─ Service/payments-api
    │  └─ (no changes)
    └─ Ingress/payments-api
       └─ host: payments.example.com (new)
    
    Dependencies:
    └─→ payments-db (must be deployed first)

Actions:
  - Validate chart
  - Render templates
  - Helm upgrade --install

Run 'sp apply' to execute this plan.
```

---

## 9. Non-Goals

### What This Provider Does NOT Handle

❌ **Cluster Provisioning**
- Creating Kubernetes clusters
- Managing cloud infrastructure (EKS, GKE, AKS)
- → Use separate `terraform` or `pulumi` providers

❌ **Secret Management**
- Creating secrets
- Rotating credentials
- Vault integration
- → Use `secrets` provider or external secret operators

❌ **Policy Enforcement**
- OPA/Gatekeeper policies
- RBAC configuration
- Network policies
- → Use `policy` provider or separate tooling

❌ **Certificate Management**
- TLS certificate generation
- cert-manager setup
- → Use separate cert operators

❌ **Monitoring Setup**
- Prometheus/Grafana installation
- Alert configuration
- → Use separate monitoring stack

❌ **GitOps Orchestration**
- ArgoCD/Flux setup
- Sync wave management
- → Use GitOps providers

### Composability

The Helm provider is **intentionally narrow**:
- Focuses only on Helm chart deployment
- Composes with other providers for full stack
- Avoids environment-specific logic

**Example: Full Stack Composition**
```yaml
providers:
  terraform:
    # Provisions infrastructure
  helm:
    # Deploys applications
  secrets:
    # Manages secrets
  policy:
    # Enforces policies

components:
  - name: eks-cluster
    type: terraform.eks
    
  - name: api
    type: helm.service
    dependsOn: [eks-cluster]
    
  - name: api-secrets
    type: secrets.vault
    
  - name: network-policy
    type: policy.networkpolicy
```

---

## Constraints & Design Principles

### Keep It Composable
- Single responsibility (Helm deployment only)
- Works alongside other providers
- Clear boundaries

### Avoid Environment Logic
- No hardcoded environments (dev/staging/prod)
- Environments managed by Sourceplane orchestration
- Provider stays environment-agnostic

### Convention Over Configuration
- Sensible defaults for 90% of cases
- Short intent YAMLs are the goal
- Explicit when needed, implicit when safe

### Optimize for Short Intents
```yaml
# This should be enough for most services:
- name: api
  type: helm.service
  spec:
    chart:
      path: ./charts/api
```

---

## Getting Started

### 1. Install Provider

```bash
sp provider install helm
```

### 2. Create Chart

```bash
helm create charts/my-service
```

### 3. Define Intent

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: my-app

providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0"

components:
  - name: my-service
    type: helm.service
    spec:
      chart:
        path: ./charts/my-service
```

### 4. Deploy

```bash
sp plan      # Preview
sp apply     # Deploy
```

---

## Additional Resources

- [Helm Documentation](https://helm.sh/docs/)
- [JSON Schema Reference](./schema.yaml)
- [Provider Source Code](https://github.com/sourceplane/providers/helm)
- [Example Projects](./examples/)
