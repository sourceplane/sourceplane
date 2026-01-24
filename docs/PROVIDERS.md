# Providers

> **Mental Model:** Providers are compilers, not runtimes.

Providers are the extension mechanism of Sourceplane. They define what component types exist, what inputs they accept, and how intent is validated and rendered.

A provider **does not execute infrastructure**. It defines schemas, rules, and templates that Sourceplane uses to understand and plan work.

---

## Table of Contents

- [What Is a Provider?](#what-is-a-provider)
- [Why Spec-First Providers?](#why-spec-first-providers)
- [Provider Structure](#provider-structure)
- [Provider Metadata](#provider-metadata)
- [Component Schemas](#component-schemas)
- [Defaults & Conventions](#defaults--conventions)
- [Templates](#templates)
- [Versioning & Compatibility](#versioning--compatibility)
- [Testing Providers](#testing-providers)
- [Best Practices](#best-practices)
- [Philosophy](#philosophy)

---

## What Is a Provider?

A Sourceplane provider is a **spec-first package** that describes:

- **Supported component types** - What kinds of components can be created
- **Input schemas** - What parameters each component accepts
- **Validation rules** - How to validate component definitions
- **Rendering templates** - How to generate CI workflows, manifests, files, etc.
- **Optional conventions and defaults** - Sensible defaults to reduce verbosity

### What Providers Enable

- ✅ Compile-time validation
- ✅ Consistent component definitions
- ✅ Tooling, linting, and AI-friendly introspection
- ✅ Deterministic planning
- ✅ Intent as code

---

## Why Spec-First Providers?

Sourceplane intentionally separates **specification from execution**.

This architectural decision provides:

- **No provider binaries required** - Providers are pure configuration
- **Git-native** - Providers live in version control alongside your code
- **Deterministic plans** - No runtime side effects or state drift
- **AI-friendly** - Clear, parseable schemas enable better tooling
- **Safer CI/CD** - Plans are predictable and reviewable

### Inspiration

This model is inspired by:

- [Crossplane](https://crossplane.io) XRDs (Composite Resource Definitions)
- Terraform provider schemas
- OpenAPI specifications
- Kubernetes Custom Resource Definitions (CRDs)

Sourceplane providers borrow the **structure** of these systems — not their complexity.

---

## Provider Structure

## Provider Structure

A provider repository should be **self-contained, versioned, and discoverable**.

### Recommended Directory Layout

```
providers/
└── helm/
    ├── provider.yaml           # Provider metadata
    ├── README.md               # High-level overview
    ├── schemas/
    │   ├── service.yaml        # Component type schemas
    │   └── job.yaml
    ├── defaults/
    │   └── values.yaml         # Optional defaults
    ├── templates/
    │   ├── github/
    │   │   └── ci.yaml         # CI workflow templates
    │   └── manifests/
    │       └── deployment.yaml # K8s manifests
    ├── examples/
    │   └── intent.yaml         # Usage examples
    └── tests/
        └── lint.yaml           # Validation tests
```

### Key Design Goals

- **Human-readable** - Easy to understand and review
- **Machine-parseable** - Structured for tools and automation
- **AI-friendly** - Clear schemas enable intelligent assistance
- **No hidden logic** - Everything is explicit and discoverable

---

## Provider Metadata

The `provider.yaml` file defines the provider itself.

### Example: provider.yaml

```yaml
apiVersion: sourceplane.io/v1
kind: Provider
metadata:
  name: helm
  version: v0.1.0
spec:
  description: Helm-based Kubernetes services
  componentTypes:
    - helm.service
    - helm.job
```

### Fields

- **name** - Unique provider identifier
- **version** - Semantic version (e.g., `v0.1.0`, `v1.2.3`)
- **description** - Brief explanation of the provider's purpose
- **componentTypes** - Array of component types this provider supports

---

## Component Schemas

Each component type must define a **schema** that declares its interface.

### What Schemas Provide

- Declare **required vs optional** inputs
- Define **default values**
- Enable **validation and linting**
- Support **IDE autocompletion** and AI assistance

### Example: Component Schema

```yaml
apiVersion: sourceplane.io/v1
kind: ComponentSchema
metadata:
  type: helm.service
spec:
  inputs:
    chart:
      type: string
      required: true
      description: "Helm chart name"
    values:
      type: object
      required: false
      description: "Helm values override"
    replicas:
      type: integer
      required: false
      default: 1
```

### Schema Best Practices

- ✅ **Prefer explicit inputs** - Make the contract clear
- ✅ **Avoid free-form blobs** - Use structured types when possible
- ✅ **Keep schemas small and composable** - Single responsibility principle
- ✅ **Add descriptions** - Document what each field does
- ❌ **Avoid overly flexible schemas** - Type safety matters

---

## Defaults & Conventions

Providers may define **defaults** to reduce YAML verbosity and enforce conventions.

### Example: defaults.yaml

```yaml
defaults:
  helm.service:
    values:
      replicaCount: 1
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
```

### How Defaults Work

- **Applied implicitly** - Users don't need to specify them
- **Can be overridden** - Per-component customization is always possible
- **Should be documented** - Make conventions explicit in README

### When to Use Defaults

- ✅ Sensible conventions (e.g., replica count = 1)
- ✅ Security baselines (e.g., resource limits)
- ✅ Organizational standards (e.g., label conventions)
- ❌ Provider-specific magic behavior
- ❌ Values that vary by environment

---

## Templates

Templates render **intent → output**.

### Typical Outputs

- **CI workflows** - GitHub Actions, GitLab CI, etc.
- **Kubernetes manifests** - Deployments, Services, Ingress
- **Configuration files** - Helm values, environment configs
- **Infrastructure code** - Terraform, Pulumi modules

### Template Guidelines

Templates should:

- ✅ **Be deterministic** - Same input = same output
- ✅ **Depend only on validated inputs** - No external state
- ✅ **Avoid conditional logic explosion** - Keep it simple
- ✅ **Follow the provider's conventions** - Consistency matters
- ❌ **Never fetch external data at render time**
- ❌ **Never modify external state**

### Example: Template Structure

```
templates/
├── github/
│   ├── deploy.yaml          # Deployment workflow
│   └── test.yaml            # Test workflow
├── manifests/
│   ├── deployment.yaml      # K8s Deployment
│   ├── service.yaml         # K8s Service
│   └── ingress.yaml         # K8s Ingress
└── helm/
    └── values.yaml          # Helm values template
```

---

## Versioning & Compatibility

Providers **must be versioned** using [semantic versioning](https://semver.org).

### Versioning Guidelines

- **MAJOR** (v1.0.0 → v2.0.0) - Breaking schema changes
- **MINOR** (v1.0.0 → v1.1.0) - New component types or backward-compatible features
- **PATCH** (v1.0.0 → v1.0.1) - Bug fixes, documentation updates

### Breaking Changes

Examples of breaking changes:

- Removing a component type
- Making an optional field required
- Changing a field type
- Renaming fields without aliases
- Removing defaults

### Provider References in intent.yaml

```yaml
apiVersion: sourceplane.io/v1
kind: Repository
metadata:
  name: my-service
providers:
  helm:
    source: github.com/sourceplane/providers/helm
    version: ">=0.1.0, <1.0.0"  # Semver range
components:
  - type: helm.service
    name: api
```

---

## Testing Providers

Providers should include **tests** to ensure correctness.

### What to Test

- ✅ **Schema validation** - Ensure schemas are valid
- ✅ **Example intent files** - Verify examples work
- ✅ **Template rendering** - Check generated output
- ✅ **Lint rules** - Validate component definitions

### Testing Structure

```
tests/
├── valid/
│   ├── basic-service.yaml
│   └── multi-component.yaml
├── invalid/
│   ├── missing-required.yaml
│   └── wrong-type.yaml
└── expected/
    └── rendered-output.yaml
```

### Sourceplane's Validation

Sourceplane will:

- Validate schemas at **compile time**
- Fail early if provider types are **missing or invalid**
- Report **detailed error messages** with line numbers

---

## Best Practices

### Do's ✅

- **Spec first, logic last** - Define the interface before implementation
- **Small, focused component types** - Single responsibility
- **Strong schemas over flexible blobs** - Type safety prevents errors
- **Deterministic templates** - No side effects or external dependencies
- **Version everything** - Track changes with semver
- **Document conventions** - Make implicit behavior explicit
- **Provide examples** - Show how to use each component type
- **Test thoroughly** - Validate schemas and templates

### Don'ts ❌

- **No provider-specific execution engines** - Specs only, not runtimes
- **No hidden behavior** - Everything should be in config
- **No runtime side effects** - Keep plans deterministic
- **No magic conventions** - Make all behavior explicit
- **No unversioned providers** - Always use semver
- **No overly complex schemas** - Keep it simple

---

## Philosophy

> **Providers exist to make intent explicit.**

If Sourceplane is the **language**,  
providers are the **grammar**.

Providers define the vocabulary for describing infrastructure, applications, and workflows. They enable teams to:

- **Communicate clearly** about system architecture
- **Validate early** before deployment
- **Generate consistently** across environments
- **Reason systematically** about changes
- **Collaborate effectively** across teams

By keeping providers **declarative, versioned, and transparent**, Sourceplane ensures that intent remains the source of truth — not runtime state, not tribal knowledge, not hidden scripts.

---

## Related Systems

Sourceplane providers are influenced by:

- **[Crossplane](https://crossplane.io)** - XRDs & Compositions
- **[Terraform](https://terraform.io)** - Provider schemas
- **[Kubernetes](https://kubernetes.io)** - Custom Resource Definitions (CRDs)
- **[OpenAPI](https://openapis.org)** - API specifications

We borrow their **structure and clarity** while avoiding their complexity.

