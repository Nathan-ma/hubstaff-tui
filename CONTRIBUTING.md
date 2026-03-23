# Contributing to hubstaff-tui

Thank you for your interest in contributing!

## Development Setup

1. Install Go 1.24+
2. Clone the repository: `git clone https://github.com/Nathan-ma/hubstaff-tui.git`
3. Install dependencies: `go mod download`
4. Build: `make build`

## Making Changes

1. Fork the repository and create a branch: `git checkout -b feat/your-feature`
2. Make your changes following the coding style in existing code
3. Add tests for new functionality
4. Run quality checks:
   ```bash
   make vet    # go vet
   make test   # go test -race ./...
   make lint   # golangci-lint run
   make build  # CGO_ENABLED=0 go build
   ```
5. Commit using [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat: add new feature`
   - `fix: resolve a bug`
   - `docs: update documentation`
   - `test: add tests`
   - `chore: maintenance task`
   - `ci: CI/CD changes`
   - `refactor: code restructuring`

## Submitting a Pull Request

1. Push your branch and open a PR against `master`
2. Describe what changes you made and why
3. Wait for CI to pass (lint, build, test on macOS and Linux)
4. Request a review

## Reporting Issues

Use the GitHub issue templates:
- **Bug Report** — for unexpected behavior or crashes
- **Feature Request** — for new feature ideas

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and concise
- Write table-driven tests where appropriate
- No panics in production code — return errors instead

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
