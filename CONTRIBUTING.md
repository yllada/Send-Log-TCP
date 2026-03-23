# Contributing to SendLog Syslog

First off, thank you for considering contributing to SendLog Syslog! 🎉

This document provides guidelines and steps for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Commit Messages](#commit-messages)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- **Go 1.22+** - [Install Go](https://go.dev/doc/install)
- **Node.js 20+** - [Install Node.js](https://nodejs.org/)
- **pnpm** - `npm install -g pnpm`
- **Wails CLI** - `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Platform-Specific Dependencies

<details>
<summary><b>Linux (Ubuntu/Debian)</b></summary>

```bash
# Ubuntu 24.04+
sudo apt install -y libgtk-3-dev libwebkit2gtk-4.1-dev

# Ubuntu 22.04
sudo apt install -y libgtk-3-dev libwebkit2gtk-4.0-dev
```
</details>

<details>
<summary><b>macOS</b></summary>

```bash
xcode-select --install
```
</details>

<details>
<summary><b>Windows</b></summary>

Install [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) (usually pre-installed on Windows 10/11).
</details>

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Y3LCorp/Send-Log-TCP.git
   cd Send-Log-TCP
   ```

2. **Install frontend dependencies**
   ```bash
   cd frontend
   pnpm install
   cd ..
   ```

3. **Run in development mode**
   ```bash
   wails dev
   ```
   This starts the app with hot-reload enabled.

4. **Build for production**
   ```bash
   wails build
   ```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title** describing the issue
- **Steps to reproduce** the behavior
- **Expected behavior** vs actual behavior
- **Screenshots** if applicable
- **Environment details** (OS, version, etc.)

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md).

### Suggesting Features

Feature suggestions are welcome! Please use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- **Use case** - Why is this feature needed?
- **Proposed solution** - How should it work?
- **Alternatives considered** - Other approaches you've thought of

### Your First Contribution

Look for issues labeled:
- `good first issue` - Simple tasks for newcomers
- `help wanted` - Issues where we need community help

## Pull Request Process

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding guidelines
3. **Add tests** for new functionality
4. **Ensure tests pass**: `go test ./...`
5. **Update documentation** if needed
6. **Submit a pull request** using our [PR template](.github/PULL_REQUEST_TEMPLATE.md)

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated if needed
- [ ] Commit messages follow conventional commits
- [ ] No new linting errors

## Coding Guidelines

### Go Code

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting
- Keep functions focused (Single Responsibility Principle)
- Add comments for exported functions
- Error messages should be lowercase and not end with punctuation

```go
// Good
return fmt.Errorf("failed to connect: %w", err)

// Bad
return fmt.Errorf("Failed to connect: %w.", err)
```

### TypeScript/React Code

- Use TypeScript strict mode
- Follow React hooks best practices
- Use shadcn/ui components when possible
- Keep components small and focused

### File Organization

```
├── *.go              # Backend services (one per concern)
├── *_test.go         # Tests for each service
├── frontend/
│   └── src/
│       ├── components/   # React components
│       ├── hooks/        # Custom hooks
│       └── wailsjs/      # Generated Wails bindings
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./...
```

### Writing Tests

- Use table-driven tests for multiple cases
- Test both success and error paths
- Name tests descriptively: `TestFunctionName_Scenario`

```go
func TestFrame_EmptyMessage(t *testing.T) {
    framer := NewFramer(DefaultFramingConfig())
    _, err := framer.Frame("")
    if err == nil {
        t.Error("expected error for empty message")
    }
}
```

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Code style (formatting, etc.) |
| `refactor` | Code refactoring |
| `perf` | Performance improvement |
| `test` | Adding/updating tests |
| `chore` | Maintenance tasks |

### Examples

```
feat(profiles): add export to CSV functionality

fix(connection): handle TLS timeout properly

docs: update installation instructions for Ubuntu 24.04

test(framing): add benchmarks for batch processing
```

## Questions?

- Open a [Discussion](https://github.com/yllada/Send-Log-TCP/discussions) for general questions
- Join our community (link TBD)
- Check existing issues and documentation

---

Thank you for contributing! 🚀
