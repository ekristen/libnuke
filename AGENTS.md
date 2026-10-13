# Agents

## Project Overview

libnuke is a Go library that provides the core framework for cloud resource cleanup tools (aws-nuke, azure-nuke, gcp-nuke). It handles resource discovery, filtering, queuing, and removal orchestration.

## Language & Build

- Go 1.24+
- Module: `github.com/ekristen/libnuke`
- No main package — this is a library

## Project Structure

```
pkg/
├── nuke/       Core orchestrator (validate → prompt → scan → filter → remove)
├── queue/      Resource processing queue with 9-state lifecycle
├── scanner/    Parallel resource discovery
├── filter/     Multi-type filtering (exact, glob, regex, date, etc.)
├── registry/   Global resource type registry with dependency sorting
├── resource/   Resource interfaces
├── config/     YAML configuration parsing
├── settings/   Per-resource-type settings
├── types/      Properties collection and type utilities
├── errors/     Custom error types
├── log/        Colored logging utilities
├── utils/      Helpers (UniqueID, Prompt)
├── slices/     Generic slice utilities
├── unique/     Unique key generation
└── docs/       Documentation generation
```

## Commands

- **Test:** `go test ./...`
- **Test single package:** `go test ./pkg/nuke/...`
- **Lint:** `golangci-lint run`
- **Tidy:** `go mod tidy`

## Code Style & Conventions

- Follow idiomatic Go conventions
- Use `context.Context` propagation throughout
- Use `logrus` for structured logging
- Lint rules: gocyclo limit 15, funlen limit 100 (relaxed in test files)
- Test files use `testify/assert` and `testify/suite`
- Resources implement optional interfaces to opt into capabilities (Filter, PropertyGetter, SettingsGetter, etc.)
- Register resource types via the global registry in `pkg/registry/`

## Key Interfaces

- `resource.Resource` — Base: `Remove(ctx) error`
- `resource.Filter` — Per-resource filtering: `Filter() error`
- `resource.PropertyGetter` — `Properties() types.Properties`
- `registry.Lister` — `List(ctx, opts) ([]resource.Resource, error)`

## Testing

- Every package should have corresponding `_test.go` files
- Use table-driven tests where appropriate
- Mock resources and listers are defined in `pkg/nuke/testsuite_test.go` and `pkg/scanner/testsuite_test.go`
- Do not disable linting rules in non-test code
