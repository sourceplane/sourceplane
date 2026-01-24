# Sourceplane CLI

Sourceplane CLI is a **component-driven tool** for defining, understanding, and managing software repositories and organizations.

It codifies **intent as code** — transforming specifications (defined in `intent.yaml`) into real repositories, components, and CI/CD workflows while enabling organizational introspection and architectural analysis.

---

## Installation

### macOS / Linux

#### Homebrew

```bash
# Install Sourceplane CLI
brew install sourceplane/sourceplane/sp

# Or install Thin-CI standalone
brew install sourceplane/sourceplane/thinci
```

#### Install Script

```bash
# Install Sourceplane CLI
curl -sSfL https://raw.githubusercontent.com/sourceplane/cli/main/install.sh | sh

# Or install Thin-CI standalone
BINARY=thinci curl -sSfL https://raw.githubusercontent.com/sourceplane/cli/main/install.sh | sh
```

#### Manual Download

Download the latest binary from [releases](https://github.com/sourceplane/cli/releases):

```bash
# macOS (Apple Silicon) - sp
curl -LO https://github.com/sourceplane/cli/releases/latest/download/sp_v*_Darwin_arm64.tar.gz
tar -xzf sp_v*_Darwin_arm64.tar.gz
sudo mv sp /usr/local/bin/

# macOS (Intel) - sp
curl -LO https://github.com/sourceplane/cli/releases/latest/download/sp_v*_Darwin_x86_64.tar.gz
tar -xzf sp_v*_Darwin_x86_64.tar.gz
sudo mv sp /usr/local/bin/

# Linux (amd64) - sp
curl -LO https://github.com/sourceplane/cli/releases/latest/download/sp_v*_Linux_x86_64.tar.gz
tar -xzf sp_v*_Linux_x86_64.tar.gz
sudo mv sp /usr/local/bin/

# Linux (arm64) - sp
curl -LO https://github.com/sourceplane/cli/releases/latest/download/sp_v*_Linux_arm64.tar.gz
tar -xzf sp_v*_Linux_arm64.tar.gz
sudo mv sp /usr/local/bin/
```

Replace `sp` with `thinci` in the URLs above to install Thin-CI instead.

### Windows

#### Scoop

```powershell
scoop bucket add sourceplane https://github.com/sourceplane/scoop-bucket
scoop install sp
# or
scoop install thinci
```

#### Manual Download

Download the latest `.zip` from [releases](https://github.com/sourceplane/cli/releases) and add to PATH.

### From Source

```bash
go install github.com/sourceplane/cli/cmd/sp@latest
# or
go install github.com/sourceplane/cli/cmd/thinci@latest
```

Alternatively, clone and build:

```bash
# Clone the repository
git clone https://github.com/sourceplane/cli.git
cd cli

# Build both binaries
make build

# Or build individually
make build-sp      # Sourceplane CLI
make build-thinci  # Thin-CI standalone

# Install
make install-all   # Install both binaries
```

See [BUILD.md](BUILD.md) for detailed build instructions and release processes.

---

## Two Binaries, One Codebase

This repository provides two independent binaries:

### 1. Sourceplane CLI (`sp`)
The full-featured CLI with all commands including component management, linting, organization analysis, and thin-ci.

```bash
sp component list
sp lint
sp thin-ci plan --github  # or: sp thinci plan --github
```

### 2. Thin-CI Standalone (`thinci`)
A lightweight binary containing only the thin-ci planning engine for CI/CD workflows.

```bash
thinci plan --github --mode=apply
```

**Why two binaries?**
- **`sp`**: Use when you need the full Sourceplane feature set
- **`thinci`**: Use in CI/CD environments where you only need plan generation (smaller, faster)

Both share the same core implementation. See [docs/THINCI_README.md](docs/THINCI_README.md) for thin-ci documentation.

---

## Quick Start

### Thin-CI Planning

1. **Initialize providers** (like `terraform init`):

```bash
sp providers init
```

2. **Generate a CI plan**:

```bash
# Using full CLI
sp thin-ci plan --github --mode=plan

# Using standalone
thinci plan --github --mode=plan
```

See [Thin-CI Documentation](docs/THINCI_INDEX.md) for complete guide.

---

## Three Ways to Use Sourceplane

### 1. Repository Analysis & Introspection (Standalone Mode)

Work with individual repositories or your entire organization **without a provider**. This mode reads `intent.yaml` files to understand and analyze your architecture.

**What you can do:**
- List components in a repository
- Create component dependency trees
- Describe individual components
- Lint and validate component definitions
- Analyze org-wide architecture with Git access
- Build organizational component graphs

**Example:**

```bash
# List all components in current repo
sp component list

# Show component tree
sp component tree

# Describe a specific component
sp component describe api

# Lint the repository definition
sp lint

# Org-level: analyze all repos with sourceplane.yaml
sp org tree
sp org graph
```

**Use this when:** You want to understand your architecture, validate component definitions (including provider validation), or explore dependencies without generating any code.

**Note:** The `lint` command validates that all component types match available provider definitions in the `providers/` directory.

---

### 2. Component Bootstrapping (Provider Mode)

When used **with a provider**, Sourceplane becomes more powerful — it can generate and bootstrap components inside repositories.

**What you can do:**
- Bootstrap new components in existing repos
- Generate implementation files based on component types
- Render CI/CD pipelines from component definitions
- Ensure consistent component structure across repos

**Example:**

```bash
# Bootstrap a new API component
sp component create api --type service.api --provider my-provider@v1

# Render CI workflows from component specs
sp ci render
```

A `sourceplane.yaml` file defines the repo's components:

```yaml
apiVersion: sourceplane.io/v1
kind: Repository

metadata:
  name: payments-api
  owner: team-payments

provider: my-provider@v1

components:
  - name: api
    type: service.api
    inputs:
      language: node
      port: 3000
```

**Use this when:** You want to scaffold components with real implementation code and maintain consistency across your repositories.

---

### 3. Blueprint-Driven Organization (Blueprint Mode)

Create entire organizations from a **blueprint.yaml** — a single file that defines multiple repositories and their components.

**What you can do:**
- Define your entire org structure in one file
- Create multiple repositories from a blueprint
- Ensure consistency across repos
- Bootstrap entire projects or microservice ecosystems

**Example blueprint.yaml:**

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
          
  - name: shared-infra
    components:
      - type: terraform.module
        name: vpc
        inputs:
          region: us-east-1
```

**Commands:**

```bash
# Create a blueprint
sp init blueprint

# Preview what will be created
sp plan

# Generate all repositories and components
sp apply
```

**Use this when:** You're setting up a new organization, standardizing multiple repositories, or need to manage infrastructure and services at scale.

---

## Core Concepts

### Components

Everything is a **component** — an API, a Terraform module, a pipeline, or a microservice. Components have:
- A **type** (e.g., `service.api`, `terraform.module`)
- **Inputs** (configuration)
- **Dependencies** on other components

### Providers

Providers define:
- Available component types
- How components render into actual files
- Validation rules and schemas

Providers are versioned and can be shared across teams.

**Built-in Provider Examples:**
- **[Helm Provider](providers/helm/)** - Deploy Kubernetes applications via Helm charts
- Terraform Provider (coming soon)
- Custom providers via schema definitions

### Repository Intent (`intent.yaml`)

The single source of truth for a repository — declares what the repo contains and why it exists.

**Format:**
```yaml
apiVersion: sourceplane.io/v1
kind: Intent  # or Repository for legacy

metadata:
  name: my-service
  owner: team-name

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

### Blueprints

Organization-level specifications that define multiple repositories and their relationships.

---

## Quick Start

### For Repository Analysis:

```bash
# Navigate to a repo with intent.yaml
cd my-repo

# View components
sp component list

# See the tree
sp component tree

# Validate
sp lint
```

### For Bootstrapping with a Provider:

```bash
# Add a new component
sp component create my-service --type service.api --provider my-provider@v1
```

### For Blueprint-Driven Setup:

```bash
# Create a blueprint
sp init blueprint

# Edit blueprint.yaml to define your org

# Preview changes
sp plan

# Create everything
sp apply
```

---

## Why Sourceplane?

| Traditional Approach | Sourceplane |
|---------------------|-------------|
| Implicit architecture | Explicit component definitions |
| Scattered config files | Centralized intent |
| Manual repo setup | Automated from blueprints |
| Guesswork and heuristics | Clear specifications |
| Per-repo inconsistency | Provider-enforced standards |
| Runtime validation | Compile-time provider validation |

## Provider Validation

Sourceplane validates all components against actual provider definitions:

- **Compile-time checks**: Errors are caught before deployment
- **Type safety**: Component types must match provider schemas
- **Missing provider detection**: Clear error messages when providers are unavailable
- **Available providers**: Automatically discovered from `providers/` directory

**Example:**
```bash
$ sp lint
❌ Errors:
  • component 'my-db': provider 'terraform' not found
  • component 'my-api': type 'helm.database' not supported (available: helm.service)

Available providers:
  • helm
```

---

## Philosophy

- **Intent as code** — declare what you want, not how to build it
- **Components first** — everything is composable
- **Specs over scripts** — explicit beats implicit
- **Git as the control plane** — version everything
- **AI reads truth** — no hallucination when specs exist

---

## Status

Sourceplane CLI is under active development. Core concepts and language are stabilizing while capabilities continue to expand.

---

## Installation & Build

### Prerequisites

- Go 1.21 or later
- Git

### Build from Source

```bash
# Clone the repository
git clone https://github.com/sourceplane/cli.git
cd cli

# Build the binary
go build -o sp
# Or use make
make build

# Optionally, install to your PATH
sudo mv sp /usr/local/bin/
# Or use make
make install
```

### Verify Installation

```bash
sp --version
sp --help
```

### Quick Test

```bash
# The project includes an example intent.yaml
sp component list
sp component tree
sp lint

# Explore the Helm provider example
cd providers/helm/examples/minimal
cat intent.yaml
```

---

## Project Structure

```
├── main.go                    # Entry point
├── cmd/                       # CLI commands
│   ├── root.go               # Root command
│   ├── component.go          # Component operations
│   ├── blueprint.go          # Blueprint workflow
│   ├── lint.go               # Validation
│   └── org.go                # Org-wide operations
├── internal/
│   ├── models/               # Data structures
│   └── parser/               # YAML parsing
├── examples/                  # Usage examples
└── sourceplane.yaml          # Example repository
```

See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for detailed documentation.

---

> **Sourceplane: A language for software organizations.**
