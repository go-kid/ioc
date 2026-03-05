---
name: IoC Component Development
description: This skill should be used when the user asks to "create a component", "add a new service", "implement dependency injection", "add wire tags", "create a lifecycle component", "implement ApplicationRunner", "implement CloserComponent", "add naming component", "register a constructor", or needs guidance on component struct design, tag usage, or lifecycle interface implementation in the go-kid/ioc framework.
---

# IoC Component Development

Guide for creating components in the go-kid/ioc dependency injection framework.

## Core Concepts

A component is any Go struct pointer registered with the IoC container. Components declare dependencies via struct tags and participate in lifecycle management through interfaces.

## Component Registration

Two registration styles:

```go
// Style 1: Global registration (typically in init() or main)
ioc.Register(&MyService{})
ioc.Register(NewMyService)           // constructor function
ioc.Provide[MyService](NewMyService) // type-safe constructor

// Style 2: Via app options
app.SetComponents(&MyService{}, &OtherService{})
```

## Struct Tag Reference

| Tag | Purpose | Example |
|-----|---------|---------|
| `wire:""` | Inject by type | `Dep *Service \`wire:""\`` |
| `wire:"name"` | Inject by component name | `Dep *Service \`wire:"comp-a"\`` |
| `wire:",qualifier=x"` | Inject by qualifier | `Dep IService \`wire:",qualifier=svc-a"\`` |
| `wire:",required=false"` | Optional dependency | `Dep *Service \`wire:",required=false"\`` |
| `wire:""` on slice | Inject all implementations | `All []IService \`wire:""\`` |
| `value:"literal"` | Static value | `Name string \`value:"foo"\`` |
| `value:"${key}"` | Config placeholder | `DSN string \`value:"${db.dsn}"\`` |
| `value:"${key:default}"` | Placeholder with default | `Port int \`value:"${port:8080}"\`` |
| `value:"#{expr}"` | Expression | `Sum int \`value:"#{1+2}"\`` |
| `prop:"key"` | Sugar for `value:"${key}"` | `Host string \`prop:"db.host"\`` |
| `prefix:"key"` | Bind config subtree | `DB *DBConfig \`prefix:"database"\`` |
| `logger:""` | Logger injection | `Log syslog.Logger \`logger:""\`` |
| `logger:",embed"` | Embedded logger | In embedded struct with `syslog.Logger \`logger:",embed"\`` |

## Injection Rules

- Dependency fields **must be exported** (uppercase first letter)
- Pointer injection: field type must be `*ConcreteType`
- Interface injection: field type must be the interface type
- Slice injection: `[]InterfaceType` gets all implementations
- When multiple implementations match an interface:
  1. `WirePrimary` marker wins
  2. Components without custom `Naming()` preferred
  3. Use `qualifier` tag arg to select explicitly

## Lifecycle Interfaces

Implement these interfaces from `github.com/go-kid/ioc/definition`:

```go
// Called after all properties are set (before Init)
type InitializingComponent interface {
    AfterPropertiesSet() error
}

// Called after AfterPropertiesSet
type InitializeComponent interface {
    Init() error
}

// Executed after all components are refreshed
type ApplicationRunner interface {
    Run() error
}

// Called during app shutdown (concurrent execution)
type CloserComponent interface {
    Close() error
}
```

All lifecycle interfaces have `WithContext` variants accepting `context.Context`.

## Marker Interfaces

```go
// Custom component name for wire-by-name
type NamingComponent interface { Naming() string }

// Mark as primary when multiple implementations exist
type WirePrimary interface { Primary() }

// Qualifier for disambiguation
type WireQualifier interface { Qualifier() string }

// Lazy initialization — skip during Refresh phase
type LazyInit interface { LazyInit() }

// Control scope: definition.ScopeSingleton (default) or definition.ScopePrototype
type ScopeComponent interface { Scope() string }

// Conditional registration
type ConditionalComponent interface { Condition(ctx ConditionContext) bool }

// Control execution order (lower = earlier)
type Ordered interface { Order() int }
```

Embed `definition.WirePrimaryComponent`, `definition.LazyInitComponent`, or `definition.PriorityComponent` for no-op implementations of marker interfaces.

## Constructor Injection

Register functions as constructors. Parameters are resolved from the container by type:

```go
func NewService(repo *Repository, logger *Logger) *Service {
    return &Service{repo: repo, logger: logger}
}

// Supports (T, error) return signature
func NewService(repo *Repository) (*Service, error) {
    return &Service{repo: repo}, nil
}
```

Constructor parameters support: pointer types, interface types, slices of either, and `ConfigurationProperties` types (auto-bound from config).

## Event System

```go
// Publish events
type ApplicationEventPublisher interface {
    PublishEvent(event ApplicationEvent) error
}

// Listen for events
type ApplicationEventListener interface {
    OnEvent(event ApplicationEvent) error
}
```

Built-in events: `ComponentCreatedEvent`, `ApplicationStartedEvent`, `ApplicationClosingEvent`.

## Common Patterns

### Service with Dependencies

```go
type UserService struct {
    Repo   *UserRepository `wire:""`
    Cache  CacheService    `wire:""`
    Logger syslog.Logger   `logger:""`
    Config *AppConfig      `prefix:"app"`
}

func (s *UserService) AfterPropertiesSet() error {
    // validate dependencies, setup internal state
    return nil
}
```

### Ordered Runner

```go
type MigrationRunner struct {
    DB *Database `wire:""`
}
func (r *MigrationRunner) Run() error { return r.DB.Migrate() }
func (r *MigrationRunner) Order() int { return 1 } // runs first
```

### Prototype Scope

```go
type RequestHandler struct{}
func (h *RequestHandler) Scope() string { return definition.ScopePrototype }
```
