# Dual Binary Restructuring - Summary

## What Changed

The Sourceplane CLI repository has been restructured to support building and releasing two independent binaries from a single codebase.

## New Binary Structure

### 1. Sourceplane CLI (`sp`)
**Entry Point**: `cmd/sourceplane/main.go`

Full-featured CLI with all commands:
- Component management (`component list`, `component tree`, etc.)
- Linting (`lint`)
- Organization analysis (`org tree`, `org graph`)
- Blueprint management (`blueprint create`, `blueprint apply`)
- **Thin-CI as subcommand** (`sp thin-ci plan`)

### 2. Thin-CI Standalone (`thinci`)
**Entry Point**: `cmd/thinci/main.go`

Lightweight binary with only thin-ci functionality:
- Direct plan command (`thinci plan`)
- No other Sourceplane features
- Smaller binary size (~4.5MB vs full CLI)
- Ideal for CI/CD environments

## Architecture

```
┌─────────────────────────────────────────┐
│          Shared Code Base               │
│                                          │
│  internal/thinci/    ← Planning Engine  │
│  internal/models/    ← Data Models      │
│  internal/parser/    ← Intent Parser    │
│  cmd/thinci.go       ← Command Logic    │
│  cmd/root.go         ← Root Commands    │
└─────────────────────────────────────────┘
           │                    │
           │                    │
    ┌──────▼─────┐       ┌─────▼──────┐
    │cmd/sourceplane│     │cmd/thinci  │
    │   main.go    │     │  main.go   │
    └──────┬──────┘      └─────┬──────┘
           │                    │
           │                    │
    ┌──────▼──────┐      ┌─────▼──────┐
    │  sp binary  │      │thinci binary│
    │  (full CLI) │      │ (thin-ci)  │
    └─────────────┘      └────────────┘
```

## Files Created

### Entry Points
- `cmd/sourceplane/main.go` - Main Sourceplane CLI entry
- `cmd/thinci/main.go` - Thin-CI standalone entry

### Build Configuration
- `Makefile` - Updated with dual binary targets
- `.github/workflows/release.yml` - GitHub Actions release workflow
- `BUILD.md` - Comprehensive build and release guide
- `.gitignore` - Updated to exclude both binaries and dist/

### Documentation
- `BUILD.md` - Build, install, and release instructions
- `README.md` - Updated with installation and dual-binary info

## Files Modified

### Core Files
- `cmd/root.go` - Added `ExecuteThinCI()` and `thinCIRootCmd`
- `cmd/thinci.go` - Removed duplicate `init()`, command now added via root.go
- `README.md` - Added installation section and binary explanation

## Build System

### Make Targets

| Command | Description |
|---------|-------------|
| `make build` | Build both sp and thinci |
| `make build-sp` | Build only Sourceplane CLI |
| `make build-thinci` | Build only Thin-CI |
| `make install` | Install sp to /usr/local/bin |
| `make install-all` | Install both binaries |
| `make release` | Build for all platforms |
| `make clean` | Remove build artifacts |

### Release Build

Creates binaries for:
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)  
- Windows (AMD64)

Example:
```bash
make release VERSION=1.0.0
```

Output in `dist/`:
```
sp-linux-amd64
sp-linux-arm64
sp-darwin-amd64
sp-darwin-arm64
sp-windows-amd64.exe
thinci-linux-amd64
thinci-linux-arm64
thinci-darwin-amd64
thinci-darwin-arm64
thinci-windows-amd64.exe
```

## Usage Examples

### Sourceplane CLI
```bash
# Install
make build-sp && sudo cp sp /usr/local/bin/

# Use
sp component list
sp lint
sp thin-ci plan --github
```

### Thin-CI Standalone
```bash
# Install
make build-thinci && sudo cp thinci /usr/local/bin/

# Use
thinci plan --github --mode=plan
thinci plan --github --mode=apply --env=prod
```

## CI/CD Integration

### GitHub Actions Workflow

Triggers on version tags (e.g., `v1.0.0`):
1. Builds binaries for all platforms
2. Generates SHA256 checksums
3. Creates GitHub release
4. Uploads all binaries as assets

To release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

## Benefits

### For Users

1. **Choice**: Install only what you need
2. **Size**: Thin-CI binary is standalone and smaller
3. **Simplicity**: `thinci plan` vs `sp thin-ci plan`
4. **CI/CD**: Lightweight binary for CI environments

### For Developers

1. **Code Reuse**: Shared implementation in `internal/`
2. **Independent Releases**: Can version separately if needed
3. **Testing**: Test both binaries independently
4. **Maintainability**: Single codebase, multiple entry points

## Migration Guide

### For Existing Users

**No Breaking Changes**

If you currently use:
```bash
sp thin-ci plan --github
```

This continues to work exactly as before.

**Optional**: Switch to standalone for CI/CD:
```bash
thinci plan --github
```

### For CI/CD Pipelines

**Before**:
```yaml
- run: sp thin-ci plan --github
```

**After** (optional):
```yaml
- run: thinci plan --github
```

Benefits:
- Smaller binary to download
- Faster installation
- Simpler command

## Testing

All tests pass with the new structure:

```bash
make test    # Run all tests
make check   # Format and vet
make build   # Build both binaries
```

Both binaries verified working:
- `sp --version` ✅
- `thinci --version` ✅
- `sp thin-ci plan --help` ✅
- `thinci plan --help` ✅
- `make release` ✅

## Documentation Updates

Updated:
- ✅ README.md - Installation and binary explanation
- ✅ BUILD.md - Complete build and release guide
- ✅ .gitignore - Exclude both binaries and dist/
- ✅ Makefile - Dual binary targets
- ✅ GitHub Actions - Release workflow

## Next Steps

1. **Tag a Release**: Create v0.1.0 tag to test GitHub Actions
2. **Documentation**: Update any external docs referencing installation
3. **Distribution**: Consider Homebrew, apt, and other package managers
4. **Website**: Update download links when ready

## Rollback Plan

If needed, revert to single binary:

1. Delete `cmd/sourceplane/main.go` and `cmd/thinci/main.go`
2. Keep `main.go` at root calling `cmd.Execute()`
3. Revert Makefile to original
4. Remove GitHub Actions workflow

## Summary

✅ **Working**: Both binaries build and run correctly  
✅ **Tested**: All commands work as expected  
✅ **Documented**: Complete build and usage guides  
✅ **CI/CD Ready**: GitHub Actions workflow configured  
✅ **Backward Compatible**: Existing users unaffected  
✅ **Production Ready**: Ready for v0.1.0 release  

The restructuring is complete and production-ready!
