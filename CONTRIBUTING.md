# Contributing to SendLog

First off, thank you for considering contributing to SendLog! 🎉

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

- **Go 1.25+** — [Install Go](https://go.dev/doc/install)

That's it. SendLog is a single static Go binary with no webview, frontend, or
system dependencies, so there is nothing else to install.

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Y3LCorp/Send-Log-TCP.git
   cd Send-Log-TCP
   ```

2. **Build**
   ```bash
   go build -o sendlog ./cmd/sendlog
   ```

3. **Run the interactive terminal UI**
   ```bash
   ./sendlog
   ```

4. **Use the CLI for automation**
   ```bash
   ./sendlog send -a 127.0.0.1 -p 514 -t udp -m "hello"
   ./sendlog help
   ```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues. When creating a bug report, include:

- **Clear title** describing the issue
- **Steps to reproduce** the behavior
- **Expected behavior** vs actual behavior
- **Environment details** (OS, version, `sendlog version`)

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.md).

### Suggesting Features

Feature suggestions are welcome! Please use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.md) and include:

- **Use case** — Why is this feature needed?
- **Proposed solution** — How should it work?
- **Alternatives considered** — Other approaches you've thought of

### Your First Contribution

Look for issues labeled:
- `good first issue` — Simple tasks for newcomers
- `help wanted` — Issues where we need community help

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
- Prefer zero comments; let names carry intent. Comment only a genuine
  non-obvious *why*, an external constraint, or a complex algorithm.
- Error messages should be lowercase and not end with punctuation

```go
// Good
return fmt.Errorf("failed to connect: %w", err)

// Bad
return fmt.Errorf("Failed to connect: %w.", err)
```

### Project Layout

```
├── cmd/sendlog/     # Entry point: TUI when bare, CLI subcommands otherwise
├── internal/
│   ├── cli/         # Plain CLI (send/connect/stress) for automation
│   ├── tui/         # Bubble Tea terminal UI
│   ├── syslog/      # Syslog message building and sending
│   ├── framing/     # TCP framing (octet-counting, non-transparent)
│   ├── stress/      # Continuous send engine
│   ├── storage/     # Local persistence
│   └── logimport/   # Import messages from .json/.csv/.txt
└── *_test.go        # Tests live next to the code they cover
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
feat(tui): add stress screen with live stats

fix(connection): handle TLS timeout properly

docs: update installation instructions

test(framing): add benchmarks for batch processing
```

## Questions?

- Open a [Discussion](https://github.com/yllada/Send-Log-TCP/discussions) for general questions
- Check existing issues and documentation

---

Thank you for contributing! 🚀
