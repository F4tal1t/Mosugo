# Contributing to Mosugo

Thank you for your interest in contributing to Mosugo! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Commit Guidelines](#commit-guidelines)
- [First-Time Contributors](#first-time-contributors)
- [Build & Release Process](#build--release-process)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please:

- Be respectful and considerate in your communication
- Focus on constructive feedback
- Accept different viewpoints and experiences
- Show empathy toward other community members

Unacceptable behavior will not be tolerated and may result in removal from the project.

## Getting Started

### Prerequisites

Before you begin, ensure you have:

- **Go 1.23 or later** installed ([download](https://go.dev/dl/))
- **Git** for version control
- **A C compiler**:
  - Windows: Install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64
  - Linux: `sudo apt-get install gcc libgl1-mesa-dev x11-dev`
  - macOS: Xcode Command Line Tools (`xcode-select --install`)
- **golangci-lint** for code quality checks: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- **Fyne CLI** (optional, for packaging): `go install fyne.io/fyne/v2/cmd/fyne@latest`

### Forking the Repository

1. Fork the repository on GitHub: https://github.com/F4tal1t/Mosugo
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/Mosugo.git
   cd Mosugo
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/F4tal1t/Mosugo.git
   ```

## Development Setup

### Building the Project

```powershell
# Download dependencies
go mod download

# Run in development mode
go run cmd/mosugo/main.go

# Build executable
go build -o mosugo.exe cmd/mosugo/main.go
```

### Running Tests

```powershell
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Running Linters

```powershell
# Run all configured linters
golangci-lint run

# Run with increased timeout for slower machines
golangci-lint run --timeout 5m

# Auto-fix issues where possible
golangci-lint run --fix
```

## Project Structure

Understanding the codebase structure will help you navigate and contribute more effectively:

```
Mosugo/
├── cmd/
│   └── mosugo/
│       └── main.go           # Application entry point, window setup
│
├── internal/                 # Internal packages (not importable externally)
│   ├── canvas/
│   │   ├── canvas.go         # Main MosugoCanvas widget, coordinate transforms
│   │   └── grid.go           # Grid pattern rendering
│   ├── cards/
│   │   └── mosu.go           # MosuWidget card implementation
│   ├── storage/
│   │   └── storage.go        # JSON persistence, workspace save/load
│   ├── theme/
│   │   └── theme.go          # Custom Fyne theme, colors, fonts
│   ├── tools/
│   │   ├── animation.go      # Bounce easing animation utilities
│   │   ├── state.go          # ToolType enum and String()
│   │   └── tools.go          # Tool interface, Select/Card/Draw/Erase implementations
│   └── ui/
│       ├── calendar.go       # Calendar widget for date navigation
│       └── canvas_container.go # Metaball border with SDF rendering
│
├── assets/                   # Embedded resources
│   ├── *.svg                 # Tool icons (select, card, draw, eraser)
│   ├── Mosugo_Icon.png       # Application icon
│   └── Comic.ttf             # Comic Sans font
│
├── .golangci.yml             # Linter configuration
├── go.mod                    # Go module definition
├── Mosugo.toml               # Fyne packaging metadata
└── README.md                 # User documentation
```

### Key Concepts

- **World vs Screen Space**: Canvas uses two coordinate systems:
  - **World Space**: Infinite canvas coordinates (cards/strokes positions)
  - **Screen Space**: Viewport pixels (what you see on screen)
  - Transforms handled by `ScreenToWorld()` and `WorldToScreen()` in canvas

- **Tool State Machine**: Tools implement the `Tool` interface and handle mouse events independently

- **Persistence Model**: One JSON file per day (`YYYY-MM-DD.mosugo`)

## Coding Standards

### General Guidelines

1. **Follow Go idioms**: Write idiomatic Go code as per [Effective Go](https://go.dev/doc/effective_go)
2. **Use `gofmt`**: All code must be formatted with `gofmt` (automatically checked by CI)
3. **Package documentation**: Every package must have a package-level comment explaining its purpose
4. **Exported symbols**: All exported types, functions, and constants must have GoDoc comments
5. **Error handling**: Always handle errors explicitly; use `%w` for error wrapping

### Code Style

```go
// Good: Clear function name, GoDoc comment, proper error handling
// LoadWorkspace loads the workspace state from a JSON file.
// If the file doesn't exist, returns an empty workspace state without error.
func LoadWorkspace(date time.Time) (WorkspaceState, error) {
    filePath, err := getWorkspaceFilePath(date)
    if err != nil {
        return WorkspaceState{}, fmt.Errorf("get file path: %w", err)
    }
    // ... implementation
}

//  Bad: No comment, unclear error handling
func LoadWS(d time.Time) (WorkspaceState, error) {
    p := getPath(d)
    // ...
}
```

### Naming Conventions

- **Packages**: Short, lowercase, single-word names (`canvas`, `cards`, not `canvas_pkg`)
- **Interfaces**: End with `-er` suffix when behavior-focused (`Tool`, `Renderer`)
- **Getters**: Omit `Get` prefix (`Offset()` not `GetOffset()`)
- **Setters**: Use `Set` prefix (`SetOffset()`)
- **Private fields**: Start with lowercase (`isDragging`, `strokeWidth`)
- **Exported types**: Start with uppercase (`MosugoCanvas`, `ToolType`)

### Linting Requirements

All pull requests must pass the following linters (configured in `.golangci.yml`):

- `gofmt`, `goimports` – Code formatting and import organization
- `govet` – Standard Go vet checks
- `errcheck` – All errors must be handled
- `staticcheck` – Advanced static analysis
- `unused` – No unused code
- `revive` – Exported symbols must be documented
- `gosimple`, `ineffassign` – Code simplifications

**Before submitting a PR**, run:
```powershell
golangci-lint run
```

Fix all issues reported.

## Testing Requirements

### Test Coverage Goals

We aim for **60-70% test coverage** on critical packages:

- `internal/storage` – File I/O, JSON marshaling
- `internal/canvas` – Coordinate transformations, stroke simplification
- `internal/tools` – Tool behavior and state transitions

### Writing Tests

- Use **table-driven tests** for testing multiple scenarios
- Place test files alongside source: `canvas.go` → `canvas_test.go`
- Use `testdata/` directory for test fixtures (sample JSON files)

Example:

```go
func TestScreenToWorld(t *testing.T) {
    tests := []struct {
        name     string
        screen   fyne.Position
        offset   fyne.Position
        scale    float32
        expected fyne.Position
    }{
        {
            name:     "no offset, 1:1 scale",
            screen:   fyne.NewPos(100, 100),
            offset:   fyne.NewPos(0, 0),
            scale:    1.0,
            expected: fyne.NewPos(100, 100),
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

### Running Tests

```powershell
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/storage

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Pull Request Process

### Before You Start

1. **Check existing issues**: Look for related issues or create a new one
2. **Discuss major changes**: For large features, discuss your approach in an issue first
3. **Keep PRs focused**: One feature/fix per pull request

### Creating a Pull Request

1. **Create a feature branch** from `master`:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b bugfix/issue-number-short-description
   ```

2. **Make your changes**:
   - Follow coding standards
   - Add tests for new functionality
   - Update documentation if needed

3. **Commit your changes** (see [Commit Guidelines](#commit-guidelines))

4. **Keep your branch up-to-date**:
   ```bash
   git fetch upstream
   git rebase upstream/master
   ```

5. **Run checks locally**:
   ```powershell
   # Format code
   go fmt ./...
   
   # Run tests
   go test ./...
   
   # Run linters
   golangci-lint run
   ```

6. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Open a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Link to related issue(s)
   - Screenshots/GIFs for UI changes

### PR Checklist

Before submitting, ensure:

- [ ] Code follows project coding standards
- [ ] All tests pass (`go test ./...`)
- [ ] Linting passes (`golangci-lint run`)
- [ ] New code has appropriate test coverage
- [ ] Documentation is updated (README, package docs, architecture docs)
- [ ] Commit messages follow conventional commits format
- [ ] PR description clearly explains the changes

### Review Process

1. **Automated checks** run on every PR (linting, tests)
2. **Code review** by maintainers (expect feedback and iteration)
3. **Approval and merge** by project maintainer

## Commit Guidelines

We follow **Conventional Commits** format for clear, semantic commit history:

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat:` – New feature
- `fix:` – Bug fix
- `docs:` – Documentation changes
- `style:` – Code style changes (formatting, no logic change)
- `refactor:` – Code refactoring (no feature change or bug fix)
- `test:` – Adding or updating tests
- `chore:` – Build process, tooling, dependencies

### Scopes (optional)

Use package names: `canvas`, `cards`, `storage`, `tools`, `ui`, `theme`

### Examples

```bash
feat(canvas): add grid snapping to card positioning

fix(storage): handle corrupted JSON files gracefully

docs: update CONTRIBUTING with commit guidelines

test(tools): add tests for Douglas-Peucker simplification

refactor(ui): extract calendar date formatting to helper
```

## First-Time Contributors

Welcome! Here's how to get started:

### Finding Issues

Look for issues labeled:
- `good first issue` – Small, well-defined tasks perfect for newcomers
- `help wanted` – Maintainers are looking for contributors
- `documentation` – Improve docs without deep codebase knowledge

### Good First Contributions

- Fix typos in documentation
- Improve error messages
- Add unit tests to existing code
- Enhance code comments
- Create examples or tutorials

### Getting Help

- **Questions?** Open a [Discussion](https://github.com/F4tal1t/Mosugo/discussions)
- **Stuck?** Comment on the issue you're working on
- **Want guidance?** Tag maintainers in your PR

## Build & Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- `MAJOR.MINOR.PATCH` (e.g., `1.2.3`)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Creating a Release (Maintainers Only)

1. **Update version** in:
   - `cmd/mosugo/main.go` (Version constant)
   - `Mosugo.toml` (Version field)
   - `CHANGELOG.md` (add release notes)

2. **Create annotated tag**:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions** automatically:
   - Builds Windows executable
   - Creates GitHub Release
   - Uploads binary as release asset

### Manual Build for Release

```powershell
# Package with Fyne (includes metadata and icon)
fyne package -os windows -name Mosugo -icon assets/Mosugo_Icon.png

# Or build manually
go build -ldflags="-s -w" -o mosugo.exe cmd/mosugo/main.go
```

---

## Questions?

If you have any questions not covered here:

- Open a [GitHub Discussion](https://github.com/F4tal1t/Mosugo/discussions)
- Create an issue for bugs or feature requests
- Reach out to maintainers in an existing issue

Thank you for contributing to Mosugo! 🎉
