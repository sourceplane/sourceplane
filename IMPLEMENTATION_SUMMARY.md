# Thin-CI Run Command - Implementation Summary

## What Was Implemented

A new `thinci run` command that executes CI jobs locally from a generated plan file, with full support for:

- **Pre-steps execution**: Setup and preparation steps
- **Main commands**: The core job commands
- **Post-steps execution**: Cleanup and finalization
- **Template variable resolution**: Dynamic substitution of variables like `{{.releaseName}}`, `{{.namespace}}`, etc.
- **Verbose output**: Detailed, well-formatted execution logs
- **Dry-run mode**: Preview what would execute without running commands
- **Error handling**: Clear error reporting with context

## Files Created/Modified

### New Files

1. **`internal/thinci/executor.go`** - Core execution engine
   - `Executor` struct for managing job execution
   - Template variable resolution
   - Step and command execution
   - Formatted output with clear visual sections
   - Error handling and output streaming

2. **`docs/THINCI_RUN.md`** - Comprehensive documentation
   - Usage examples
   - Flag descriptions
   - Template variable guide
   - Common workflows
   - Error handling guide

3. **`examples/plan.json`** - Example plan with helm jobs
4. **`examples/test-plan.json`** - Demo plan with safe echo commands

### Modified Files

1. **`cmd/thinci.go`**
   - Added `thinCIRunCmd` command definition
   - Added run command flags (--plan, --job-id, --verbose, --dry-run, --github)
   - Added `runThinCIRun()` function implementation

2. **`cmd/root.go`**
   - Registered run command with both `sp thinci` and standalone `thinci` binary

## Usage Examples

### Basic Usage

```bash
sp thinci run \
    --plan plan.json \
    --job-id "hello-app-validate"
```

### With GitHub Flag (as requested)

```bash
sp thinci run \
    --github \
    --plan plan.json \
    --job-id "hello-app-validate"
```

### Dry Run

```bash
sp thinci run \
    --plan plan.json \
    --job-id "hello-app-validate" \
    --dry-run
```

### Using Standalone Binary

```bash
thinci run --plan plan.json --job-id "hello-app-plan" --verbose
```

## Key Features

### 1. Template Variable Resolution

Commands can use Go template syntax:

```json
{
  "commands": [
    "helm template {{.releaseName}} {{.chartPath}} --values {{.valuesPath}}"
  ],
  "inputs": {
    "releaseName": "my-app",
    "chartPath": "./charts/app",
    "valuesPath": "values.yaml"
  }
}
```

Resolves to:
```bash
helm template my-app ./charts/app --values values.yaml
```

### 2. Structured Output

Clear visual separation of execution phases:

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
```

### 3. Error Handling

- Stops immediately on command failure
- Shows error output clearly
- Returns appropriate exit codes
- In verbose mode, streams all output in real-time
- In non-verbose mode, captures and shows output only on error

### 4. Dual Binary Support

Works with both:
- `sp thinci run` (main CLI)
- `thinci run` (standalone binary)

## Testing

Successfully tested with:

1. ✅ Dry-run mode
2. ✅ Verbose execution
3. ✅ Template variable resolution
4. ✅ Pre-steps, commands, and post-steps execution
5. ✅ Error handling
6. ✅ Both binaries (sp and thinci)
7. ✅ GitHub flag support

## Example Output

```
Sourceplane Thin-CI Job Executor
Plan: examples/test-plan.json

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Executing Job: demo-app-validate
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ℹ Component: demo-app
  ℹ Action: validate

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Pre-Steps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Step 1: setup-environment
  ├─ Command: echo "=== Setting up environment ==="
  ├─ Output:
  │ === Setting up environment ===

  ▸ Step 2: check-dependencies
  ├─ Command: echo "Checking dependencies for demo-app"
  ├─ Output:
  │ Checking dependencies for demo-app

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Main Commands
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Step 1: Command 1
  ├─ Command: echo "Validating component: demo-app"
  ├─ Output:
  │ Validating component: demo-app

  ▸ Step 2: Command 2
  ├─ Command: echo "Provider: demo | Action: validate"
  ├─ Output:
  │ Provider: demo | Action: validate

  ▸ Step 3: Command 3
  ├─ Command: echo "Running in namespace: production"
  ├─ Output:
  │ Running in namespace: production

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Post-Steps
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ▸ Step 1: save-results
  ├─ Command: echo "Saving validation results"
  ├─ Output:
  │ Saving validation results

  ✓ Job completed successfully in 26ms
```

## Integration with Provider Configuration

The run command works seamlessly with provider definitions in `provider.yaml`:

```yaml
thinCI:
  actions:
    - name: validate
      preSteps:
        - name: setup-helm
          command: helm version
      commands:
        - helm lint {{.chartPath}}
        - helm template {{.releaseName}} {{.chartPath}} --validate
      postSteps: []
      inputs:
        chartPath: "."
```

These are automatically included in the generated plan and executed by the run command.
