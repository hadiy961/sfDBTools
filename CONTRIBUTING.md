# Contributing Guide

## Development Workflow

### Setup Development Environment

1. **Clone repository**
   ```bash
   git clone https://github.com/hadiy961/sfDBTools_new.git
   cd sfDBTools_new
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build and test**
   ```bash
   go build -o sfdbtools main.go
   ./sfdbtools --help
   ```

### Making Changes

1. **Create feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test**
   ```bash
   # Make your changes
   go test ./...
   go build -o sfdbtools main.go
   ```

3. **Commit with conventional format**
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   # or
   git commit -m "fix: resolve issue description"
   # or  
   git commit -m "docs: update documentation"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   # Create Pull Request on GitHub
   ```

### Commit Message Convention

Use conventional commits for better changelog generation:

- `feat:` - New features
- `fix:` - Bug fixes  
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code refactoring
- `test:` - Adding/updating tests
- `chore:` - Build process, dependencies, etc.

### Release Process

Only maintainers can create releases:

1. **Automated Release (Recommended)**
   ```bash
   ./release.sh 1.2.3
   ```

2. **Manual Release**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

### Testing

- **Run all tests**: `go test ./...`
- **Run with coverage**: `go test -cover ./...`
- **Run specific package**: `go test ./internal/config`
- **Run with race detection**: `go test -race ./...`

### Code Quality

The CI pipeline will check:
- Go vet
- Staticcheck
- Tests pass
- Code builds successfully

Make sure all checks pass before submitting PR.

## Project Structure

```
.
├── main.go                 # Entry point
├── cmd/                    # CLI commands (Cobra)
├── internal/               # Private application code
│   ├── config/            # Configuration management
│   ├── core/              # Core business logic
│   ├── logger/            # Logging utilities
│   └── version/           # Version information
├── utils/                  # Utility functions
├── config/                 # Configuration files
├── docs/                   # Documentation
└── .github/               # GitHub Actions workflows
```

## Configuration

- Configuration files go in `config/`
- Use `internal/config` for configuration loading logic
- Sensitive configs should be encrypted or use environment variables

## Logging

- Use the logger from `internal/logger`
- Log levels: debug, info, warn, error
- Structured logging with key-value pairs

## Dependencies

- Use `go mod` for dependency management
- Pin major versions where stability is important
- Keep dependencies minimal and well-maintained

## Documentation

- Update README.md for user-facing changes
- Add code comments for complex logic
- Update docs/ for detailed documentation
- Include examples in documentation
