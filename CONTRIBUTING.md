# Contributing to Sourceplane CLI

Thank you for considering contributing to Sourceplane CLI! This document provides guidelines and instructions for contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/cli.git
   cd cli
   ```
3. **Add the upstream repository**:
   ```bash
   git remote add upstream https://github.com/sourceplane/cli.git
   ```

## Development Setup

### Prerequisites
- Go 1.25.6 or later
- Make (for build automation)
- Git

### Building from Source

```bash
# Install dependencies
go mod download

# Build both binaries
make build

# Run tests
make test

# Format and lint code
make check
```

## Development Workflow

### 1. Create a Branch
Create a branch for your feature or fix:
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes
- Write clean, idiomatic Go code
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes
```bash
# Run all tests
make test

# Format code
make fmt

# Run linter
make vet

# Or run all checks
make check
```

### 4. Commit Your Changes
Write clear, descriptive commit messages:
```bash
git add .
git commit -m "feat: add new provider validation"
```

**Commit Message Format:**
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for test additions/changes
- `refactor:` for code refactoring
- `chore:` for maintenance tasks

### 5. Push and Create a Pull Request
```bash
git push origin feature/your-feature-name
```

Then open a Pull Request on GitHub with:
- Clear description of changes
- Reference to any related issues
- Screenshots/examples if applicable

## Code Guidelines

### Go Style
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (run `make fmt`)
- Keep functions small and focused
- Write meaningful variable and function names
- Add comments for exported functions and complex logic

### Project Structure
```
cli/
├── cmd/                 # CLI commands (Cobra)
│   ├── sp/             # Main CLI entry point
│   └── thinci/         # Thin-CI entry point
├── internal/           # Private application code
│   ├── config/        # Configuration management
│   ├── models/        # Domain models
│   ├── parser/        # YAML parsing
│   ├── provider/      # Provider abstraction
│   ├── providers/     # Provider registry
│   ├── thinci/        # Thin-CI logic
│   ├── validator/     # Validation logic
│   └── version/       # Version management
├── providers/          # Provider schemas and examples
├── docs/              # Documentation
└── examples/          # Usage examples
```

### Testing
- Write unit tests for new functionality
- Aim for meaningful test coverage
- Use table-driven tests where appropriate
- Mock external dependencies

Example test:
```go
func TestParseIntent(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Intent
        wantErr bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Adding New Providers

To add a new provider:

1. Create a directory under `providers/`:
   ```
   providers/
   └── your-provider/
       ├── provider.yaml   # Provider metadata
       ├── schema.yaml     # JSON schema for validation
       ├── README.md       # Provider documentation
       └── examples/       # Example intent.yaml files
   ```

2. Define the provider schema following existing patterns
3. Add comprehensive examples
4. Update documentation in `docs/providers/`

## Documentation

- Update relevant documentation for any changes
- Add examples for new features
- Keep README.md up to date
- Document new CLI commands and flags

## Pull Request Process

1. **Ensure all tests pass** and code is properly formatted
2. **Update documentation** for any new features or changes
3. **Add examples** if introducing new functionality
4. **Request review** from maintainers
5. **Address feedback** promptly and respectfully
6. **Squash commits** if requested before merging

## Reporting Issues

When reporting bugs or requesting features:

1. **Search existing issues** first
2. **Use issue templates** when available
3. **Provide details:**
   - Clear description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Environment details (OS, Go version, CLI version)
   - Error messages and logs

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on the code, not the person
- Help create a welcoming community

## Questions?

If you have questions about contributing:
- Open a discussion on GitHub
- Check existing documentation
- Reach out to maintainers

---

Thank you for contributing to Sourceplane CLI! 🚀
