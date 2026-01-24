# Thin-CI Planning Engine - Documentation Index

Welcome to the Sourceplane Thin-CI Planning Engine documentation. This index will guide you to the right documentation based on your needs.

## 🚀 Quick Start

**New to Thin-CI?** Start here:
- [Thin-CI Overview](THINCI_OVERVIEW.md) - Complete introduction with examples
- [Quick Start Guide](#quick-start-commands) - Get running in 5 minutes

## 📚 Documentation by Audience

### For Users
**I want to use thin-ci to generate CI plans**

Start with:
1. [THINCI_README.md](THINCI_README.md) - User guide with examples
2. [Provider Examples](#provider-examples) - Real-world use cases

### For Developers
**I want to extend or modify thin-ci**

Read:
1. [THINCI_IMPLEMENTATION.md](THINCI_IMPLEMENTATION.md) - Extension guide
2. [THINCI_ARCHITECTURE.md](THINCI_ARCHITECTURE.md) - Deep design dive

### For Architects
**I want to understand the design**

Review:
1. [THINCI_OVERVIEW.md](THINCI_OVERVIEW.md) - High-level architecture
2. [THINCI_ARCHITECTURE.md](THINCI_ARCHITECTURE.md) - Detailed design

## 📖 Core Documentation

### [THINCI_README.md](THINCI_README.md)
**User-facing guide**
- What is Thin-CI
- Quick start
- Command reference
- Plan structure
- Use cases
- FAQ

### [THINCI_OVERVIEW.md](THINCI_OVERVIEW.md)
**Complete introduction**
- Executive summary
- Architecture diagrams
- Planning pipeline
- Example walkthrough
- Design principles
- Quality attributes

### [THINCI_ARCHITECTURE.md](THINCI_ARCHITECTURE.md)
**Design deep dive**
- Architecture diagrams
- Data structures
- Algorithms (with pseudocode)
- Provider integration
- Comparison to similar systems
- Design decisions

### [THINCI_IMPLEMENTATION.md](THINCI_IMPLEMENTATION.md)
**Developer guide**
- How to extend
- Adding providers
- Adding CI targets
- Testing strategies
- Troubleshooting
- Best practices

### [THINCI_SUMMARY.md](THINCI_SUMMARY.md)
**Development summary**
- What was built
- Files created/modified
- Implementation details
- Testing approach
- Success metrics

## 🔧 Provider Documentation

### Terraform Provider
- [Examples README](../providers/terraform/examples/README.md)
- [Multi-Component Example](../providers/terraform/examples/multi-component/)
- [Provider Configuration](../providers/terraform/provider.yaml)

### Helm Provider
- [Examples README](../providers/helm/examples/README.md)
- [Microservices Example](../providers/helm/examples/microservices/)
- [Provider Configuration](../providers/helm/provider.yaml)

## 📋 Quick Reference

### Quick Start Commands

```bash
# Generate plan for GitHub Actions
sourceplane thin-ci plan --github --mode=plan

# Apply plan for production
sourceplane thin-ci plan --github --mode=apply --env=production

# Only changed components
sourceplane thin-ci plan --github --mode=plan --changed-only

# YAML output
sourceplane thin-ci plan --github --output=yaml
```

### Command Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--github` | Target GitHub Actions | - |
| `--gitlab` | Target GitLab CI | - |
| `--mode` | plan, apply, or destroy | `plan` |
| `--base` | Base git ref | `main` |
| `--head` | Head git ref | `HEAD` |
| `--changed-only` | Filter to changed components | `true` |
| `--env` | Target environment | - |
| `--output` | json or yaml | `json` |

### Example Intent

```yaml
apiVersion: v1
kind: Intent
metadata:
  name: my-app

providers:
  terraform:
    version: "0.1.0"
  helm:
    version: "0.1.0"

components:
  - name: vpc-network
    type: terraform.network
    
  - name: api-service
    type: helm.service
    relationships:
      - target: vpc-network
        type: depends_on
```

## 🎯 Learning Paths

### Path 1: User Journey
**Goal**: Generate CI plans for my repository

1. Read [THINCI_README.md](THINCI_README.md) - Overview
2. Study [Terraform Example](../providers/terraform/examples/multi-component/) or [Helm Example](../providers/helm/examples/microservices/)
3. Create your `intent.yaml`
4. Run `sourceplane thin-ci plan --github`
5. Refer to [Troubleshooting](#troubleshooting-index) if needed

### Path 2: Developer Journey
**Goal**: Add a new provider or extend thin-ci

1. Read [THINCI_OVERVIEW.md](THINCI_OVERVIEW.md) - Architecture
2. Study [THINCI_IMPLEMENTATION.md](THINCI_IMPLEMENTATION.md) - Extension guide
3. Review existing provider definitions
4. Implement your extension
5. Add tests and documentation

### Path 3: Architect Journey
**Goal**: Evaluate thin-ci for adoption

1. Read [THINCI_OVERVIEW.md](THINCI_OVERVIEW.md) - High-level design
2. Review [THINCI_ARCHITECTURE.md](THINCI_ARCHITECTURE.md) - Design decisions
3. Study [comparison section](#comparison-to-similar-systems)
4. Evaluate [examples](#provider-examples)
5. Assess [quality attributes](#quality-attributes)

## 🔍 Topic Index

### Planning Concepts
- [What is Thin-CI](THINCI_README.md#what-is-thin-ci)
- [Planning Pipeline](THINCI_OVERVIEW.md#planning-pipeline)
- [Change Detection](THINCI_ARCHITECTURE.md#stage-1-change-detection)
- [Dependency Graph](THINCI_ARCHITECTURE.md#stage-3-dependency-graph)
- [Job Generation](THINCI_ARCHITECTURE.md#stage-4-job-generation)

### Data Structures
- [Plan](THINCI_OVERVIEW.md#plan)
- [Job](THINCI_OVERVIEW.md#job)
- [ComponentChange](THINCI_ARCHITECTURE.md#componentchange)
- [DependencyNode](THINCI_ARCHITECTURE.md#dependencynode)

### Algorithms
- [Kahn's Algorithm](THINCI_ARCHITECTURE.md#stage-3-dependency-graph)
- [Topological Sort](THINCI_IMPLEMENTATION.md#dependency-graph-kahns-algorithm)
- [Change Detection](THINCI_IMPLEMENTATION.md#how-it-works)

### Provider Integration
- [Provider Contract](THINCI_ARCHITECTURE.md#provider-integration)
- [Adding Providers](THINCI_IMPLEMENTATION.md#add-a-new-provider)
- [Terraform Provider](../providers/terraform/provider.yaml)
- [Helm Provider](../providers/helm/provider.yaml)

### Examples
- [Terraform Multi-Component](../providers/terraform/examples/multi-component/)
- [Helm Microservices](../providers/helm/examples/microservices/)
- [Example Walkthrough](THINCI_OVERVIEW.md#example-walkthrough)

### Troubleshooting
- [Common Issues](THINCI_IMPLEMENTATION.md#troubleshooting)
- [No Jobs Generated](THINCI_IMPLEMENTATION.md#no-jobs-generated)
- [Circular Dependencies](THINCI_IMPLEMENTATION.md#circular-dependency-error)
- [Provider Not Found](THINCI_IMPLEMENTATION.md#provider-not-found)

## 📊 Comparison to Similar Systems

| System | Documentation |
|--------|--------------|
| **vs Nx** | [THINCI_ARCHITECTURE.md#vs-nx-task-graph](THINCI_ARCHITECTURE.md) |
| **vs Terraform** | [THINCI_ARCHITECTURE.md#vs-terraform-plan](THINCI_ARCHITECTURE.md) |
| **vs Bazel** | [THINCI_ARCHITECTURE.md#vs-bazel-action-graph](THINCI_ARCHITECTURE.md) |
| **vs Crossplane** | [THINCI_ARCHITECTURE.md#vs-crossplane](THINCI_ARCHITECTURE.md) |

## 🎓 Advanced Topics

### Extending Thin-CI
- [Add New Provider](THINCI_IMPLEMENTATION.md#add-a-new-provider)
- [Add CI Target](THINCI_IMPLEMENTATION.md#add-a-new-ci-target)
- [Custom Change Detection](THINCI_IMPLEMENTATION.md#customize-change-detection)

### Testing
- [Unit Testing](THINCI_IMPLEMENTATION.md#unit-testing)
- [Integration Testing](THINCI_IMPLEMENTATION.md#integration-testing)
- [Test Cases](THINCI_IMPLEMENTATION.md#example-test-cases)

### Best Practices
- [Component Organization](THINCI_IMPLEMENTATION.md#1-component-organization)
- [Explicit Dependencies](THINCI_IMPLEMENTATION.md#2-explicit-dependencies)
- [Provider Defaults](THINCI_IMPLEMENTATION.md#3-provider-defaults)
- [Environment Configuration](THINCI_IMPLEMENTATION.md#4-environment-specific-configuration)

## 🏗️ Implementation Details

### Code Structure
```
internal/thinci/
├── types.go       # Data structures
├── detector.go    # Change detection
└── planner.go     # Planning engine

cmd/
└── thinci.go      # CLI integration

providers/
├── terraform/     # Terraform provider
└── helm/          # Helm provider
```

### Key Files
- [types.go](../internal/thinci/types.go) - Core data structures
- [detector.go](../internal/thinci/detector.go) - Change detection logic
- [planner.go](../internal/thinci/planner.go) - Planning engine
- [thinci.go](../cmd/thinci.go) - CLI command

## 🚦 Getting Started Checklist

- [ ] Read [THINCI_README.md](THINCI_README.md)
- [ ] Review an [example](#provider-examples)
- [ ] Create your `intent.yaml`
- [ ] Run `sourceplane thin-ci plan --github`
- [ ] Understand the [output structure](#plan-structure)
- [ ] Explore [advanced features](#advanced-topics)

## 💡 FAQ Quick Links

Common questions answered in documentation:

1. **Why not generate YAML directly?** → [THINCI_README.md#faq](THINCI_README.md)
2. **How does change detection work?** → [THINCI_README.md#faq](THINCI_README.md)
3. **What about circular dependencies?** → [THINCI_README.md#faq](THINCI_README.md)
4. **Can I run specific components?** → [THINCI_README.md#faq](THINCI_README.md)
5. **How to add custom provider?** → [THINCI_IMPLEMENTATION.md](THINCI_IMPLEMENTATION.md)

## 📞 Support

- **Issues**: Report bugs or request features
- **Examples**: See `providers/*/examples/`
- **Questions**: Check FAQ sections in docs

## 🗺️ Roadmap

- [ ] Workflow rendering (JSON → YAML)
- [ ] Real git integration
- [ ] Content-based caching
- [ ] Plan visualization
- [ ] More providers (Pulumi, CDK, Ansible)
- [ ] More targets (CircleCI, Jenkins)

## 📝 Document Summary

| Document | Audience | Purpose | Length |
|----------|----------|---------|--------|
| [README](THINCI_README.md) | Users | How to use | Medium |
| [OVERVIEW](THINCI_OVERVIEW.md) | All | Complete intro | Long |
| [ARCHITECTURE](THINCI_ARCHITECTURE.md) | Architects | Design deep dive | Very Long |
| [IMPLEMENTATION](THINCI_IMPLEMENTATION.md) | Developers | Extension guide | Long |
| [SUMMARY](THINCI_SUMMARY.md) | Contributors | Dev summary | Medium |

## 🎯 Next Steps

**New Users**:
→ [THINCI_README.md](THINCI_README.md)

**Developers**:
→ [THINCI_IMPLEMENTATION.md](THINCI_IMPLEMENTATION.md)

**Architects**:
→ [THINCI_ARCHITECTURE.md](THINCI_ARCHITECTURE.md)

**Quick Reference**:
→ [Command Reference](#quick-start-commands)

---

**Navigation**: [Overview](THINCI_OVERVIEW.md) | [README](THINCI_README.md) | [Architecture](THINCI_ARCHITECTURE.md) | [Implementation](THINCI_IMPLEMENTATION.md)
