# Thin-CI Run Command

The `thinci run` command allows you to execute CI jobs locally from a generated plan file. This is useful for testing and debugging CI jobs before pushing to your CI/CD platform.

## Usage

```bash
sp thinci run --plan <plan-file> --job-id <job-id> [flags]
```

Or using the standalone binary:

```bash
thinci run --plan <plan-file> --job-id <job-id> [flags]
```

## Flags

- `--plan`, `-p`: Path to the plan file (default: `plan.json`)
- `--job-id`: Job ID to execute (required)
- `--verbose`, `-v`: Enable verbose output (default: `true`)
- `--dry-run`: Dry run mode - show what would be executed without running commands
- `--github`: Running in GitHub Actions context (optional)

## Examples

### Basic Usage

Execute a job with verbose output:

```bash
sp thinci run \
    --plan plan.json \
    --job-id "hello-app-validate"
```

### Dry Run

Preview what would be executed without actually running commands:

```bash
sp thinci run \
    --plan plan.json \
    --job-id "hello-app-validate" \
    --dry-run
```

### GitHub Actions Context

Execute with GitHub Actions flag (for GitHub-specific behavior):

```bash
sp thinci run \
    --github \
    --plan plan.json \
    --job-id "hello-app-validate"
```

### Using Standalone Binary

The same commands work with the standalone `thinci` binary:

```bash
thinci run --plan plan.json --job-id "hello-app-plan" --verbose
```

## Job Execution Flow

The run command executes jobs in the following order:

1. **Pre-Steps**: Setup steps that run before the main commands
   - Environment setup
   - Dependency checks
   - Configuration validation

2. **Main Commands**: The core commands for the action
   - Build commands
   - Test commands
   - Deployment commands

3. **Post-Steps**: Cleanup and finalization steps
   - Save outputs
   - Upload artifacts
   - Send notifications

## Output Format

The command provides clear, structured output:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Executing Job: hello-app-validate
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ℹ Component: hello-app
  ℹ Action: validate

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Pre-Steps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Step 1: setup-helm
  ├─ Command: helm version
  ├─ Output:
  │ version.BuildInfo{...}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Main Commands
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Step 1: Command 1
  ├─ Command: helm lint .
  ├─ Output:
  │ ==> Linting .
  │ [INFO] Chart.yaml: icon is recommended
  │ 1 chart(s) linted, 0 chart(s) failed

  ✓ Job completed successfully in 1.2s
```

## Template Variables

Commands can include template variables that are resolved at runtime:

- `{{.component}}`: Component name
- `{{.provider}}`: Provider name
- `{{.action}}`: Action name
- `{{.releaseName}}`: Helm release name
- `{{.namespace}}`: Kubernetes namespace
- `{{.chartPath}}`: Path to Helm chart
- `{{.valuesPath}}`: Path to values file
- Any custom inputs defined in the job

Example:

```json
{
  "commands": [
    "helm template {{.releaseName}} {{.chartPath}} --values {{.valuesPath}}"
  ],
  "inputs": {
    "releaseName": "my-app",
    "chartPath": "./charts/app",
    "valuesPath": "values.prod.yaml"
  }
}
```

Resolves to:

```bash
helm template my-app ./charts/app --values values.prod.yaml
```

## Error Handling

- If a command fails, execution stops immediately
- Error output is displayed clearly
- Exit code reflects the failure
- In verbose mode, all output is streamed in real-time
- In non-verbose mode, output is only shown on error

## Common Workflows

### Local Testing

Before pushing to CI:

```bash
# Generate a plan
sp thinci plan --github -m plan > plan.json

# Test the validate job locally
sp thinci run --plan plan.json --job-id "my-app-validate" --dry-run

# Run it for real
sp thinci run --plan plan.json --job-id "my-app-validate"
```

### Debugging CI Failures

If a job fails in CI:

```bash
# Download the plan from CI artifacts
# or regenerate it locally
sp thinci plan --github -m plan > plan.json

# Run the specific failing job
sp thinci run --plan plan.json --job-id "my-app-apply" --verbose
```

### Iterative Development

```bash
# Make changes to provider configuration
vim providers/helm/provider.yaml

# Regenerate plan
sp thinci plan --github -m plan > plan.json

# Test the changes
sp thinci run --plan plan.json --job-id "my-app-validate" --dry-run
```

## See Also

- [THINCI_OVERVIEW.md](../docs/THINCI_OVERVIEW.md) - Thin-CI architecture
- [THINCI_IMPLEMENTATION.md](../docs/THINCI_IMPLEMENTATION.md) - Implementation details
- [provider.yaml](../providers/helm/provider.yaml) - Example provider configuration
