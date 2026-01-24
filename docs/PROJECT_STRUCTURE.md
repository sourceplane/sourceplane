# Sourceplane CLI - Project Structure

## Overview

A fully functional Go-based CLI tool for managing component-driven software organizations with intent-as-code philosophy.

## Project Structure

```
.
├── main.go                          # Application entry point
├── cmd/                             # CLI commands
│   ├── root.go                      # Root command and CLI setup
│   ├── component.go                 # Component management commands
│   ├── blueprint.go                 # Blueprint commands (init, plan, apply)
│   ├── lint.go                      # Validation commands
│   └── org.go                       # Organization-wide commands
├── internal/
│   ├── models/
│   │   └── models.go                # Data structures for YAML parsing
│   └── parser/
│       └── parser.go                # YAML file parsing utilities
├── providers/
│   └── helm/                        # Helm provider (example)
│       ├── provider.yaml            # Provider metadata
│       ├── schema.yaml              # Component type schemas
│       ├── README.md                # Provider documentation
│       └── examples/                # Usage examples
├── examples/
│   └── README.md                    # Usage examples
├── README.md                        # Main documentation
├── intent.yaml                      # Example intent definition
├── go.mod                           # Go module definition
├── go.sum                           # Go dependencies checksum
├── Makefile                         # Build and development tasks
└── .gitignore                       # Git ignore patterns

```

## Implemented Commands

### Component Commands
- `sp component list` - List all components in repository
- `sp component tree` - Display component tree with inputs
- `sp component describe <name>` - Describe a specific component
- `sp component create <name>` - Create new component (placeholder for provider integration)

### Blueprint Commands
- `sp init blueprint` - Initialize a new blueprint.yaml
- `sp plan` - Preview what will be created from blueprint
- `sp apply` - Create repositories from blueprint

### Validation
- `sp lint` - Validate sourceplane.yaml file

### Organization Commands
- `sp org tree` - Display org-wide component tree
- `sp org graph` - Generate architectural graph with statistics

### CI/CD Commands
- `sp ci render` - Render CI/CD workflows (placeholder for provider integration)

## Key Features

✅ **Complete YAML Parsing** - Fully functional intent.yaml and blueprint.yaml parsing  
✅ **Intent-Driven Design** - Support for Intent kind with provider definitions  
✅ **Repository Introspection** - Analyze components without provider  
✅ **Organization Analysis** - Scan multiple repos for org-wide insights  
✅ **Blueprint Workflow** - Initialize, plan, and apply blueprints  
✅ **Provider Validation** - Compile-time validation against actual provider definitions  
✅ **Type Safety** - Components must match provider schemas  
✅ **Provider System** - Example Helm provider with full documentation  
✅ **Clean CLI Interface** - Cobra-based with help and subcommands  
✅ **Spec-Based Components** - Modern spec field instead of legacy inputs

## Technologies Used

- **Go 1.25+** - Primary language
- **Cobra** - CLI framework
- **YAML v3** - YAML parsing

## Next Steps for Enhancement

1. **Provider Integration** - Implement actual provider fetching and component rendering
2. **Git Integration** - Initialize Git repos, create commits
3. **CI/CD Templates** - Generate actual GitHub Actions/GitLab CI files
4. **Component Dependencies** - Track and visualize dependencies between components
5. **Drift Detection** - Compare intent vs actual files
6. **Remote Providers** - Fetch providers from Git repositories
7. **Tests** - Add comprehensive unit and integration tests

## Building & Running

```bash
# Build
make build

# Or manually
go build -o sp

# Run
./sp --help
```

## Example Workflows

### Analyze a Repository
```bash
sp component list
sp component tree
sp lint
```

### Create from Blueprint
```bash
sp init blueprint
# Edit blueprint.yaml
sp plan
sp apply
```

### Organization Analysis
```bash
sp org tree
sp org graph
```

## File Formats

### intent.yaml (New Format)
```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: my-service
  owner: team-name
  description: Service description

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
  - name: component-name
    type: helm.service
    spec:
      chart:
        path: ./charts/component
      values:
        key: value

relationships:
  - from: component-a
    to: component-b
    type: depends-on
```

### Legacy Format (Still Supported)
```yaml
apiVersion: sourceplane.io/v1
kind: Repository
metadata:
  name: my-service
provider: provider@v1
components:
  - name: component-name
    type: component.type
    inputs:  # Legacy field
      key: value
```

### blueprint.yaml
```yaml
kind: Blueprint
apiVersion: sourceplane.io/v1
provider: provider@v1
repos:
  - name: repo-name
    components:
      - name: component-name
        type: component.type
        inputs:
          key: value
```
