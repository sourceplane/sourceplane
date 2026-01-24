# Sourceplane CLI - Build and Release Guide

This repository contains two independent binaries that can be built and released separately:

1. **`sp`** - The full Sourceplane CLI with all commands
2. **`thinci`** - Standalone Thin-CI planning engine

## Quick Start

### Build Both Binaries

```bash
make build
```

This creates:
- `./sp` - Sourceplane CLI
- `./thinci` - Thin-CI standalone

### Build Individual Binaries

```bash
# Build only Sourceplane CLI
make build-sp

# Build only Thin-CI
make build-thinci
```

## Installation

### Install Sourceplane CLI Only

```bash
make install
```

Installs `sp` to `/usr/local/bin/`

### Install Both Binaries

```bash
make install-all
```

Installs both `sp` and `thinci` to `/usr/local/bin/`

## Usage

### Sourceplane CLI

The full CLI includes all commands plus thin-ci as a subcommand:

```bash
# Use thin-ci as subcommand
sp thin-ci plan --github --mode=plan

# Use other commands
sp component list
sp lint
sp org analyze
```

### Thin-CI Standalone

The standalone binary only contains thin-ci functionality:

```bash
# Direct plan command
thinci plan --github --mode=plan

# No 'thin-ci' prefix needed
thinci plan --github --mode=apply --env=prod
```

## Release Builds

### Build for Multiple Platforms

```bash
make release VERSION=1.0.0
```

This creates binaries in `dist/` for:
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)
- Windows (AMD64)

Both `sp` and `thinci` are built for each platform.

### Release Artifacts

After running `make release`, you'll find:

```
dist/
├── sp-linux-amd64
├── sp-linux-arm64
├── sp-darwin-amd64
├── sp-darwin-arm64
├── sp-windows-amd64.exe
├── thinci-linux-amd64
├── thinci-linux-arm64
├── thinci-darwin-amd64
├── thinci-darwin-arm64
├── thinci-windows-amd64.exe
└── checksums.txt
```

## GitHub Actions Release

Create a new tag to trigger automated release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The GitHub Actions workflow will:
1. Build binaries for all platforms
2. Generate checksums
3. Create a GitHub release
4. Upload all binaries as release assets

## Development

### Project Structure

```
cmd/
├── sourceplane/     # Main Sourceplane CLI entry point
│   └── main.go
├── thinci/          # Thin-CI standalone entry point
│   └── main.go
├── root.go          # Shared root commands
├── thinci.go        # Thin-CI command implementation
├── component.go     # Component commands
├── lint.go          # Lint commands
└── ...

internal/
├── thinci/          # Thin-CI planning engine
├── models/          # Data models
├── parser/          # Intent parsing
└── ...
```

### How It Works

**Sourceplane CLI (`sp`)**:
- Entry point: `cmd/sourceplane/main.go`
- Calls `cmd.Execute()` which runs `rootCmd`
- Includes thin-ci via `rootCmd.AddCommand(thinCICmd)`
- Users run: `sp thin-ci plan ...`

**Thin-CI Standalone (`thinci`)**:
- Entry point: `cmd/thinci/main.go`
- Calls `cmd.ExecuteThinCI()` which runs `thinCIRootCmd`
- Only includes plan command directly
- Users run: `thinci plan ...`

Both binaries share the same core implementation in `internal/thinci/`.

## Testing

```bash
# Run all tests
make test

# Format and vet
make check

# Tidy dependencies
make tidy
```

## Clean Up

```bash
make clean
```

Removes:
- Built binaries (`sp`, `thinci`)
- Release artifacts (`dist/`)
- Test artifacts

## Makefile Targets

| Target | Description |
|--------|-------------|
| `build` | Build both binaries |
| `build-sp` | Build Sourceplane CLI only |
| `build-thinci` | Build Thin-CI only |
| `install` | Install `sp` to /usr/local/bin |
| `install-all` | Install both binaries to /usr/local/bin |
| `clean` | Remove build artifacts |
| `test` | Run tests |
| `fmt` | Format code |
| `vet` | Run go vet |
| `check` | Run fmt and vet |
| `release` | Build multi-platform binaries |
| `help` | Show help message |

## Version Management

Set version via environment variable:

```bash
# Build with custom version
VERSION=2.0.0 make build

# Release with custom version
VERSION=2.0.0 make release
```

Version defaults to `0.1.0` if not specified.

## CI/CD Integration

### GitHub Actions

The repository includes a release workflow (`.github/workflows/release.yml`) that:
- Triggers on version tags (`v*`)
- Builds for all platforms
- Creates GitHub release with binaries

### Manual Release Process

1. Update version in code/docs
2. Commit changes
3. Create and push tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions builds and publishes release

## Distribution Strategies

### Option 1: Separate Releases
- Release `sp` and `thinci` as separate downloads
- Users choose which one they need
- Smaller download size for thin-ci-only users

### Option 2: Combined Release
- Release both in same package
- Users get both binaries
- Single download for everything

### Option 3: Platform Packages
- Create platform-specific installers
- Homebrew formula (macOS)
- apt/yum packages (Linux)
- Chocolatey (Windows)

## Recommended Installation Methods

### macOS (Homebrew)

```ruby
# Future Homebrew formula
class Sourceplane < Formula
  desc "Component-driven tool for software organizations"
  homepage "https://github.com/sourceplane/cli"
  url "https://github.com/sourceplane/cli/releases/download/v1.0.0/sp-darwin-arm64"
  sha256 "..."
  
  def install
    bin.install "sp-darwin-arm64" => "sp"
  end
end

class Thinci < Formula
  desc "Deterministic CI/CD planning engine"
  homepage "https://github.com/sourceplane/cli"
  url "https://github.com/sourceplane/cli/releases/download/v1.0.0/thinci-darwin-arm64"
  sha256 "..."
  
  def install
    bin.install "thinci-darwin-arm64" => "thinci"
  end
end
```

### Linux (Install Script)

```bash
# install.sh
#!/bin/bash
VERSION="${VERSION:-latest}"
BINARY="${1:-sp}"
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

# Map architecture names
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac

URL="https://github.com/sourceplane/cli/releases/download/v${VERSION}/${BINARY}-${OS}-${ARCH}"
curl -L "$URL" -o "$BINARY"
chmod +x "$BINARY"
sudo mv "$BINARY" /usr/local/bin/
```

### Windows (Chocolatey)

```xml
<!-- Future Chocolatey package -->
<package>
  <metadata>
    <id>sourceplane</id>
    <version>1.0.0</version>
    <title>Sourceplane CLI</title>
    <authors>Sourceplane</authors>
    <description>Component-driven tool for software organizations</description>
  </metadata>
</package>
```

## FAQ

### Q: Why two separate binaries?

**A:** Some users only need thin-ci for CI/CD planning and don't need the full Sourceplane feature set. A standalone `thinci` binary:
- Is smaller and faster to download
- Has a simpler command structure
- Can be used in CI/CD environments without the full CLI

### Q: Do they share code?

**A:** Yes, both binaries use the same core implementation in `internal/`. They just have different entry points and command structures.

### Q: Which should I use?

**A:**
- Use `sp` if you want all Sourceplane features
- Use `thinci` if you only need CI planning
- Install both if you want flexibility

### Q: Can I run thin-ci from both?

**A:** Yes:
- `sp thin-ci plan ...` (as subcommand)
- `thinci plan ...` (standalone)

Both execute the same code.

## License

MIT License - see [LICENSE](LICENSE) for details
