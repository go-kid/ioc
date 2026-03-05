# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`go-kid/ioc` is a Go runtime dependency injection (IoC/DI) framework based on struct tags and interfaces. Provides Spring-like component lifecycle management, configuration binding, and dependency resolution with circular reference support. Requires Go 1.21+.

## Build & Test Commands

```bash
go test ./...                                                    # all tests
go test ./unittest/component/builtin_inject/ -run TestInterface  # single test
go test -v -race ./...                                           # verbose + race detector
go build ./...                                                   # build check
go vet ./...                                                     # static analysis
```

No Makefile, linter config, or CI pipeline exists.

## Architecture

### Startup Flow

`ioc.Register()` / `ioc.Provide[T]()` → `ioc.Run()` → `app.NewApp()` → `app.Run()`:
1. **Configuration init** — loaders (file/args/raw) feed into binder (viper)
2. **Factory prepare** — registers singletons, discovers post-processors
3. **Refresh** — resolves all non-lazy component definitions via `doGetComponent`; handles circular references through early singleton exposure (three-level cache)
4. **Runners** — invokes `ApplicationRunner` implementations in order

### Package Layers

| Layer | Package | Role |
|-------|---------|------|
| Public API | `ioc` (root) | `Register`, `Provide[T]`, `Run`, `RunTest`/`RunErrorTest` |
| Orchestration | `app` | `App` lifecycle, `SettingOption` functional options |
| Container core | `container` | Interfaces: `Factory`, `SingletonRegistry`, `DefinitionRegistry`, post-processor hierarchy |
| Factory impl | `container/factory` | `defaultFactory` — creation, constructors, circular refs, prototype scope |
| Post-processors | `container/processors` | Built-in tag processors (see chain below) |
| Registries | `container/support` | Singleton registry (three-level cache), definition registry |
| Metadata | `component_definition` | `Meta`, `Property`, `SelectBestCandidate` |
| Contracts | `definition` | Lifecycle/marker interfaces, event types |
| Config | `configure`, `configure/loader`, `configure/binder` | `Configure` interface, loaders (file/args/raw), viper binder |
| Logging | `syslog` | Leveled logging, colored output, `log/slog` adapter |
| Debug | `debug` | Debug server with web UI for component/event inspection |

### Post-Processor Chain

Ordered via constants in `container/processors/orders.go`. **Order matters** — config must resolve before dependencies:

**Priority order group** (config processing):
1. `LoggerAware` — `logger` tag
2. `ConfigQuoteAware` — `${...}` placeholder resolution
3. `ExpressionTagAware` — `#{...}` expression evaluation (expr-lang/expr)
4. `PropertiesAware` + `ValueAware` — `prefix`, `value`, `prop` binding

**Standard order group** (dependency processing):
5. `DependencyAware` — `wire` tag type/name resolution
6. `DependencyFurtherMatching` — Primary/Qualifier disambiguation, required check
7. `DependencyFunctionAware` — `func` tag resolution
8. `ValidateAware` — struct validation (go-playground/validator)

### Struct Tags

| Tag | Purpose | Example |
|-----|---------|---------|
| `wire` | Dependency injection | `wire:""`, `wire:"name"`, `wire:",qualifier=x"`, `wire:",required=false"` |
| `value` | Config value/placeholder/expression | `value:"${db.dsn}"`, `value:"#{1+2}"`, `value:"literal"` |
| `prop` | Sugar for `value:"${...}"` | `prop:"db.dsn"`, `prop:"port:8080"` |
| `prefix` | Bind entire config subtree | `prefix:"app.db"` |
| `logger` | Logger injection | `logger:""`, `logger:",embed"` |
| `func` | Inject by method name | `func:"MethodName"`, `func:"Method,returns=val"` |

### Component Selection Logic

When multiple components match a dependency (`SelectBestCandidate` in `component_definition/selection.go`):
1. `WirePrimary` marker takes precedence
2. Non-alias components (no custom `Naming()`) preferred
3. `WireQualifier` can select explicitly via `wire:",qualifier=name"`
4. Slice fields (`[]Interface`) receive all matching components

### Constructor Support

Functions registered via `ioc.Register(NewService)` or `ioc.Provide[T](NewService)` are detected as constructors. Parameters resolved by type from the container. Supports `(T)` and `(T, error)` return signatures. `ConfigurationProperties` parameters are auto-bound from config via `Prefix()`.

## Testing Conventions

- Use `ioc.RunTest(t, ...options)` / `ioc.RunErrorTest(t, ...options)` — bootstrap a full container
- Use `app.SetComponents(...)` to register per-test components (not `ioc.Register()` which uses global state)
- Use `loader.NewRawLoader([]byte(...))` for inline YAML config
- Tests organized under `unittest/` by concern: `component/builtin_inject/`, `component/constructor_inject/`, `configure/`
- Test file names use PascalCase: `Interface_Slice_test.go`

## Code Conventions

- Error wrapping: `github.com/pkg/errors` (`errors.WithMessage`, `errors.Wrapf`, `errors.Errorf`)
- Assertions: `github.com/stretchr/testify/assert`
- Functional utilities: `github.com/samber/lo` (`lo.Filter`, `lo.Map`)
- Config parsing: viper + `gopkg.in/yaml.v3`; struct mapping via `github.com/mitchellh/mapstructure` with `yaml` tag

## Gotchas

- **Exported fields only**: `wire`/`value`/`prop`/`prefix` tags on unexported fields are silently ignored
- **Pointer registration**: Components must be registered as pointers (`&MyService{}`), not value types
- **Global state**: `ioc.Register()` appends to a package-level slice; prefer `app.SetComponents()` in tests to avoid cross-test contamination
- **Required by default**: Both `wire` and `value`/`prop` fields are required by default; use `required=false` to make optional
- **Prototype + circular**: Prototype-scoped components cannot participate in circular reference resolution
- **`prop` vs `value`**: `prop:"key"` equals `value:"${key}"`; use `prop` for simple lookups, `value` for mixed literals/expressions
