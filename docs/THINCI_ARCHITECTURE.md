# Thin-CI Planning Engine - Architecture

## Overview

The Sourceplane Thin-CI Planning Engine is a deterministic CI/CD planning system that separates **intent** (what to deploy) from **execution** (how to deploy). It generates execution plans without executing them, enabling CI systems like GitHub Actions or GitLab CI to become "render targets" rather than hardcoded implementations.

## Design Philosophy

### Core Principles

1. **Sourceplane owns intent** - Component definitions, relationships, and desired state
2. **Providers own behavior** - How to validate, plan, apply, and destroy resources
3. **Thin-CI owns execution planning** - Dependency resolution, change detection, job generation
4. **CI systems are render targets** - GitHub Actions, GitLab CI, etc. are output formats

### Separation of Concerns

```
┌─────────────────────────────────────────────────────────┐
│                    Intent (intent.yaml)                  │
│  - Components                                            │
│  - Relationships                                         │
│  - Provider configurations                               │
└────────────────────────┬────────────────────────────────┘
                         │
                         │ Input to
                         ▼
┌─────────────────────────────────────────────────────────┐
│              Thin-CI Planning Engine                     │
│  1. Change Detection                                     │
│  2. Component Expansion                                  │
│  3. Dependency Resolution                                │
│  4. Job Generation                                       │
└────────────────────────┬────────────────────────────────┘
                         │
                         │ Produces
                         ▼
┌─────────────────────────────────────────────────────────┐
│                 Execution Plan (JSON)                    │
│  - Jobs with dependencies                                │
│  - Actions and inputs                                    │
│  - Platform-specific metadata                            │
└────────────────────────┬────────────────────────────────┘
                         │
                         │ Rendered to
                         ▼
┌─────────────────────────────────────────────────────────┐
│             CI System (GitHub Actions, etc.)             │
│  - Workflow YAML                                         │
│  - Actual execution                                      │
└─────────────────────────────────────────────────────────┘
```

## Architecture Components

### 1. Data Structures (`internal/thinci/types.go`)

#### Plan
The complete execution plan output:
```go
type Plan struct {
    Target   string         // "github", "gitlab"
    Mode     string         // "plan", "apply", "destroy"
    Metadata PlanMetadata
    Jobs     []Job
}
```

#### Job
A single unit of work in the plan:
```go
type Job struct {
    ID          string            // Unique identifier
    Component   string            // Component name
    Provider    string            // Provider (terraform, helm)
    Action      string            // Action (validate, plan, apply)
    Inputs      map[string]any    // Action inputs
    DependsOn   []string          // Job dependencies
    Metadata    JobMetadata       // Platform-specific metadata
}
```

#### Provider Action
Defines what a provider can do:
```go
type ProviderAction struct {
    Name        string
    Description string
    Order       int
    PreSteps    []ActionStep
    PostSteps   []ActionStep
    Inputs      map[string]any
    Outputs     []string
}
```

### 2. Change Detection (`internal/thinci/detector.go`)

**Purpose**: Identify which components are affected by file changes

**Algorithm**:
```
For each component in intent:
    1. Check if intent.yaml changed → affects all components
    2. Check if component-specific paths changed
        - Terraform: *.tf files in component path
        - Helm: Chart.yaml, values.yaml, templates/
    3. Check if provider configuration changed
        - providers/<provider>/provider.yaml
        - providers/<provider>/schema.yaml
    4. Check if shared modules changed
        - terraform/modules/*
        - helm/charts/*
    
    If any match → component is affected
```

**Key Features**:
- Convention-based path detection
- Provider-specific file patterns
- Shared module detection
- Intent schema change propagation

### 3. Planning Engine (`internal/thinci/planner.go`)

**Purpose**: Generate deterministic execution plans from detected changes

**Pipeline**:
```
Input: PlanRequest + Intent Files
  ↓
Step 1: Change Detection
  → ComponentChange[]
  ↓
Step 2: Component Expansion
  → DependencyNode[]
  ↓
Step 3: Dependency Graph + Topological Sort
  → Sorted DependencyNode[]
  ↓
Step 4: Job Generation
  → Job[]
  ↓
Output: Plan
```

#### Step 1: Change Detection
```go
detector := NewChangeDetector(repositoryPath, intents)
changes := detector.DetectChanges(changedFiles)
```

#### Step 2: Component Expansion
```go
// For each changed component:
// 1. Load provider metadata
// 2. Determine required actions based on mode
// 3. Extract component dependencies
// 4. Create dependency node

nodes := expandComponents(changes, intents, request)
```

#### Step 3: Dependency Resolution
```go
// Build directed acyclic graph (DAG)
// Use Kahn's algorithm for topological sort
// Detect cycles
// Return sorted nodes

sortedNodes := buildDependencyGraph(nodes, intents)
```

**Topological Sort Example**:
```
Input Components:
  vpc-network (no dependencies)
  eks-cluster (depends on vpc-network)
  rds-database (depends on vpc-network)
  app-service (depends on eks-cluster)

Dependency Graph:
  vpc-network → eks-cluster → app-service
  vpc-network → rds-database

Topological Order:
  1. vpc-network
  2. eks-cluster, rds-database (parallel)
  3. app-service
```

#### Step 4: Job Generation
```go
// For each sorted node:
// - Generate job for each action (validate, plan, apply)
// - Chain actions within same component
// - Add dependencies from dependency graph
// - Inject platform-specific metadata

jobs := generateJobs(sortedNodes, request)
```

**Job Dependency Rules**:
1. **Action Chaining**: Within a component, actions run sequentially
   ```
   component-validate → component-plan → component-apply
   ```

2. **Component Dependencies**: Cross-component dependencies respect graph
   ```
   vpc-plan → eks-plan (eks depends on vpc)
   ```

3. **Parallel Execution**: Independent components run in parallel
   ```
   eks-plan ┐
            ├→ (both can run concurrently)
   rds-plan ┘
   ```

### 4. Provider Registry

**Purpose**: Manage provider metadata and capabilities

```go
type ProviderRegistry struct {
    providers map[string]*ProviderMetadata
}

func (r *ProviderRegistry) GetProvider(name string) (*ProviderMetadata, error)
func (r *ProviderRegistry) RegisterProvider(provider *ProviderMetadata)
```

**Provider Metadata Structure**:
```yaml
thinCI:
  actions:
    - name: validate
      order: 1
      preSteps: [...]
      postSteps: [...]
      inputs: {...}
      outputs: [...]
  
  defaults:
    timeout: 600
    version: "1.5.0"
  
  ordering:
    - validate
    - plan
    - apply
```

### 5. CLI Integration (`cmd/thinci.go`)

**Command Structure**:
```
sourceplane thin-ci plan --github --mode=plan
```

**Workflow**:
```
1. Parse CLI flags
2. Find all intent.yaml files
3. Load intent files
4. Get changed files from git
5. Load provider registry
6. Create plan request
7. Generate plan
8. Output plan (JSON/YAML)
```

## Algorithms

### Change Detection Algorithm

```
function detectChanges(changedFiles, intents):
    changes = []
    
    for each component in intents:
        affectedPaths = []
        
        // Check intent changes
        if "intent.yaml" in changedFiles:
            affectedPaths.add("intent.yaml")
        
        // Check component-specific paths
        componentPaths = getComponentPaths(component)
        for file in changedFiles:
            if file matches any componentPath:
                affectedPaths.add(file)
        
        // Check provider changes
        providerPaths = getProviderPaths(component.provider)
        for file in changedFiles:
            if file matches any providerPath:
                affectedPaths.add(file)
        
        // Check shared module changes
        sharedPaths = getSharedModulePaths(component)
        for file in changedFiles:
            if file matches any sharedPath:
                affectedPaths.add(file)
        
        if affectedPaths is not empty:
            changes.add(ComponentChange{
                component: component.name,
                provider: component.provider,
                affectedPaths: affectedPaths
            })
    
    return changes
```

### Dependency Resolution Algorithm (Kahn's)

```
function topologicalSort(nodes):
    // Build graph
    graph = {}
    inDegree = {}
    
    for node in nodes:
        graph[node.name] = []
        inDegree[node.name] = 0
    
    // Add edges
    for node in nodes:
        for dependency in node.dependencies:
            if dependency in graph:
                graph[dependency].append(node.name)
                inDegree[node.name]++
    
    // Find nodes with no dependencies
    queue = []
    for name, degree in inDegree:
        if degree == 0:
            queue.append(name)
    
    // Process queue
    sorted = []
    while queue is not empty:
        current = queue.dequeue()
        sorted.append(current)
        
        for neighbor in graph[current]:
            inDegree[neighbor]--
            if inDegree[neighbor] == 0:
                queue.append(neighbor)
    
    // Check for cycles
    if len(sorted) != len(nodes):
        throw "Circular dependency detected"
    
    return sorted
```

### Job Generation Algorithm

```
function generateJobs(sortedNodes, request):
    jobs = []
    
    for node in sortedNodes:
        provider = getProvider(node.provider)
        
        for action in node.actions:
            jobID = node.component + "-" + action
            
            // Determine dependencies
            deps = []
            
            // If not first action, depend on previous action
            if action is not first in node.actions:
                prevAction = previous action in node.actions
                deps.add(node.component + "-" + prevAction)
            else:
                // First action depends on last actions of dependencies
                for depComponent in node.dependencies:
                    if depComponent in sortedNodes:
                        lastAction = last action of depComponent
                        deps.add(depComponent + "-" + lastAction)
            
            // Get provider action metadata
            providerAction = provider.getAction(action)
            
            // Build inputs
            inputs = mergeInputs(
                provider.defaults,
                providerAction.inputs,
                request.overrides
            )
            
            // Create job
            job = Job{
                id: jobID,
                component: node.component,
                provider: node.provider,
                action: action,
                inputs: inputs,
                dependsOn: deps,
                metadata: createMetadata(request.target, node, action)
            }
            
            jobs.add(job)
    
    return jobs
```

## Example Walkthrough

### Input

**Intent** (intent.yaml):
```yaml
components:
  - name: vpc
    type: terraform.network
  
  - name: eks
    type: terraform.cluster
    relationships:
      - target: vpc
        type: depends_on
```

**Changed Files**:
```
- terraform/vpc/main.tf
- terraform/eks/variables.tf
```

**Command**:
```bash
sourceplane thin-ci plan --github --mode=plan
```

### Processing

**Step 1: Change Detection**
```
Detected changes:
- vpc (reason: terraform/vpc/main.tf changed)
- eks (reason: terraform/eks/variables.tf changed)
```

**Step 2: Component Expansion**
```
Nodes:
- DependencyNode{
    component: "vpc",
    provider: "terraform",
    actions: ["validate", "plan"],
    dependencies: []
  }
- DependencyNode{
    component: "eks",
    provider: "terraform",
    actions: ["validate", "plan"],
    dependencies: ["vpc"]
  }
```

**Step 3: Dependency Resolution**
```
Topological order:
1. vpc
2. eks
```

**Step 4: Job Generation**
```
Jobs:
1. vpc-validate (depends: [])
2. vpc-plan (depends: [vpc-validate])
3. eks-validate (depends: [])
4. eks-plan (depends: [eks-validate, vpc-plan])
```

### Output

```json
{
  "target": "github",
  "mode": "plan",
  "jobs": [
    {
      "id": "vpc-validate",
      "component": "vpc",
      "provider": "terraform",
      "action": "validate",
      "dependsOn": []
    },
    {
      "id": "vpc-plan",
      "component": "vpc",
      "provider": "terraform",
      "action": "plan",
      "dependsOn": ["vpc-validate"]
    },
    {
      "id": "eks-validate",
      "component": "eks",
      "provider": "terraform",
      "action": "validate",
      "dependsOn": []
    },
    {
      "id": "eks-plan",
      "component": "eks",
      "provider": "terraform",
      "action": "plan",
      "dependsOn": ["eks-validate", "vpc-plan"]
    }
  ]
}
```

## Provider Integration

Providers participate in thin-ci by declaring their capabilities in `provider.yaml`:

```yaml
thinCI:
  actions:
    - name: validate
      description: Validate configuration
      order: 1
      preSteps:
        - name: setup
          command: terraform init
      inputs:
        path: "."
      outputs:
        - validationReport
    
    - name: plan
      description: Generate execution plan
      order: 2
      preSteps:
        - name: init
          command: terraform init
      inputs:
        path: "."
        workspace: "default"
      outputs:
        - plan.tfplan
        - plan.json
  
  defaults:
    timeout: 1800
    terraformVersion: "1.5.0"
  
  ordering:
    - validate
    - plan
    - apply
```

## Non-Goals

What this system does **NOT** do:

1. ❌ Execute CI/CD pipelines
2. ❌ Render YAML workflows (planned for future)
3. ❌ Hardcode provider-specific logic in core
4. ❌ Assume monorepo structure
5. ❌ Manage git operations beyond change detection
6. ❌ Handle authentication or secrets

## Future Enhancements

1. **Workflow Rendering**: Convert plans to GitHub Actions / GitLab CI YAML
2. **Advanced Change Detection**: Git integration via libgit2
3. **Plan Optimization**: Minimize redundant jobs
4. **Caching**: Cache provider metadata and plans
5. **Drift Detection**: Compare plan with actual state
6. **Cost Estimation**: Estimate resource costs from plans
7. **Policy Enforcement**: Validate plans against policies
8. **Multi-Repository**: Support cross-repo dependencies

## Comparison to Similar Systems

### vs Terraform
- **Terraform**: Provider-specific, state-centric, executes changes
- **Thin-CI**: Provider-agnostic, intent-centric, generates plans only

### vs Nx
- **Nx**: Monorepo task runner, file-based change detection
- **Thin-CI**: Multi-repo support, component-based change detection, CI-focused

### vs Crossplane
- **Crossplane**: Kubernetes operator, runtime reconciliation
- **Thin-CI**: CLI tool, pre-execution planning, CI integration

### vs Bazel
- **Bazel**: Build system, hermetic execution, fine-grained targets
- **Thin-CI**: CI planning, infrastructure focus, coarse-grained components

## Quality Guarantees

The Thin-CI planning engine ensures:

1. **Determinism**: Same input always produces same plan
2. **DAG Enforcement**: Circular dependencies are detected and rejected
3. **Provider Isolation**: Providers cannot interfere with core logic
4. **Type Safety**: Strong typing in Go ensures correctness
5. **Testability**: Pure functions enable comprehensive testing
6. **Extensibility**: New providers can be added without modifying core

---

**Version**: 1.0.0  
**Last Updated**: 2026-01-24  
**Status**: Implemented
