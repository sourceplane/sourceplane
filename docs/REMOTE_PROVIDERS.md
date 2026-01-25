# Remote Provider Support

Sourceplane Thin-CI now supports fetching providers from remote Git repositories automatically.

## How It Works

When you specify a provider with a remote source in your `intent.yaml`, Thin-CI will:

1. **Detect Remote Source**: Check if the provider source is a remote URL
2. **Fetch Provider**: Clone the provider repository from the remote source
3. **Cache Locally**: Store the provider in `~/.sourceplane/providers/` for reuse
4. **Auto-Update**: Pull latest changes on subsequent runs

## Configuration

In your `intent.yaml`, specify providers with a `source` field:

```yaml
apiVersion: sourceplane.io/v1
kind: Intent

metadata:
  name: my-app
  description: My application

providers:
  helm:
    source: github.com/sourceplane/provider-helm
    version: ">=0.1.0"
    defaults:
      namespace: default

components:
  - name: my-service
    type: helm.service
    spec:
      chart:
        path: ./charts/my-service
```

## Supported Source Formats

- GitHub: `github.com/org/repo`
- GitLab: `gitlab.com/org/repo`  
- Bitbucket: `bitbucket.org/org/repo`
- HTTPS URLs: `https://github.com/org/repo`
- Git SSH: `git@github.com:org/repo.git`

## Provider Repository Structure

Remote provider repositories must contain a `provider.yaml` file at the root:

```
provider-repo/
├── provider.yaml      # Required: Provider configuration
├── README.md
├── schema.yaml        # Optional: Schema definitions
└── examples/          # Optional: Example configurations
```

## Cache Location

Providers are cached at:
```
~/.sourceplane/providers/<provider-name>/
```

## Behavior

### First Run
```bash
$ sp thinci plan --github -m plan

Fetching provider helm from github.com/sourceplane/provider-helm...
Cloning into '/Users/user/.sourceplane/providers/helm'...
Loading provider: helm from /Users/user/.sourceplane/providers/helm/provider.yaml
...
```

### Subsequent Runs
```bash
$ sp thinci plan --github -m plan

Updating provider helm from github.com/sourceplane/provider-helm...
Already up to date.
Loading provider: helm from /Users/user/.sourceplane/providers/helm/provider.yaml
...
```

## Local vs Remote Providers

You can mix local and remote providers:

```yaml
providers:
  # Remote provider
  helm:
    source: github.com/sourceplane/provider-helm
    version: ">=0.1.0"
  
  # Local provider (no source specified)
  # Will be loaded from ./providers/custom/ directory
  custom:
    version: "1.0.0"
    defaults:
      region: us-west-2
```

## Fallback Behavior

If no providers are specified in the intent file, Thin-CI will fall back to loading providers from the local `providers/` directory (legacy behavior).

## Version Constraints

While version constraints can be specified, the current implementation fetches the latest version from the default branch. Full version constraint support is planned for future releases.

## Troubleshooting

### Provider Not Found

**Error:**
```
Error: failed to fetch provider helm: git clone failed
```

**Solution:**
- Verify the source URL is correct and accessible
- Ensure you have git installed and configured
- Check network connectivity
- For private repos, ensure your git credentials are configured

### Invalid Provider Structure

**Error:**
```
Error: provider.yaml not found in /Users/user/.sourceplane/providers/helm
```

**Solution:**
- Ensure the provider repository has a `provider.yaml` file at the root
- Verify the repository structure matches the expected format

### Cache Issues

To clear the provider cache:
```bash
rm -rf ~/.sourceplane/providers
```

The providers will be re-fetched on the next run.

## Examples

### GitHub Provider
```yaml
providers:
  helm:
    source: github.com/sourceplane/provider-helm
    version: ">=0.1.0"
```

### GitLab Provider
```yaml
providers:
  terraform:
    source: gitlab.com/myorg/provider-terraform
    version: ">=1.0.0"
```

### HTTPS URL
```yaml
providers:
  custom:
    source: https://github.com/myorg/provider-custom.git
    version: "latest"
```

## See Also

- [THINCI_OVERVIEW.md](THINCI_OVERVIEW.md) - Thin-CI architecture
- [THINCI_RUN.md](THINCI_RUN.md) - Running CI jobs locally
- [PROVIDERS.md](PROVIDERS.md) - Provider development guide
