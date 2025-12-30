# Repository Guidelines

READ `~/projects/personal/agent-scripts/AGENTS.MD` BEFORE ANYTHING (skip if missing).

## Project Structure & Module Organization
- `main.go`: CLI entrypoint; calls `cmd.Execute`.
- `cmd/`: Cobra commands, shared flags, and command-specific logic.
- `docs/cli-spec.md`: CLI contract (flags, exit codes, config/env precedence, examples).
- `docs/`: user-facing documentation and specs.
- Tests live alongside code as `*_test.go` in the same packages.

## Build, Test, and Development Commands
- `go build ./...`: build all packages.
- `go test ./...`: run tests (if present).
- `make lint`: run `golangci-lint` (see `Makefile`).

## Coding Style & Naming Conventions
- Format Go code with `gofmt`.
- Follow standard Go naming: exported identifiers in `CamelCase`, unexported in `camelCase`.

## Testing Guidelines
- Use Go's `testing` package.
- Place tests alongside code as `*_test.go` files in the same package.

## Commit & Pull Request Guidelines
- Keep commits scoped and descriptive; prefer one logical change per commit.
- Include key commands run (for example, `go test ./...` or `make lint`) when summarizing changes.

## Configuration & Runtime Notes
- CLI flags, env vars, and config precedence are defined in `docs/cli-spec.md`.
- When changing CLI behavior, update `docs/cli-spec.md` to keep the contract current.
