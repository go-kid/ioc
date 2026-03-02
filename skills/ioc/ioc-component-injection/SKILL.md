---
name: ioc-component-injection
description: "go-kid/ioc framework component dependency injection guide. Use when registering components, injecting dependencies with `wire` tag, selecting from multiple implementations (Primary/Qualifier), injecting into slices, using `func` tag for method-based matching, or using constructor injection. Triggers on: wire tag, component registration, dependency injection, ioc.Register, app.SetComponents, NamingComponent, WireQualifier, WirePrimary."
---

# go-kid/ioc Component Injection

## Component Registration

Two styles for registering components:

```go
// Style 1: Global registration
ioc.Register(&MyComponent{})
ioc.Run()

// Style 2: Via app options
app.NewApp().Run(app.SetComponents(&MyComponent{}, &AnotherComponent{}))
```

All registered values must be **pointers** (e.g. `&MyStruct{}`). Constructor functions can also be registered directly.

## `wire` Tag Syntax

```
wire:"<name>,qualifier=<group>,required=<bool>"
```

- `<name>` (optional): inject by component name
- `qualifier=<group>` (optional): filter by qualifier group
- `required=false` (optional): skip error when no match found (default: required)

Target field must be **exported** (uppercase).

### Inject by Pointer Type

```go
type App struct {
    Svc *MyService `wire:""`
}
```

### Inject by Interface Type

```go
type Logger interface { Log(msg string) }

type App struct {
    Logger Logger `wire:""`  // matches any *T implementing Logger
}
```

### Inject by Name

Implement `NamingComponent` to give a component a name:

```go
type MyService struct{}
func (s *MyService) Naming() string { return "my-svc" }

type App struct {
    Svc *MyService `wire:"my-svc"`
}
```

### Inject All Implementations (Slice)

```go
type App struct {
    Handlers []Handler `wire:""`  // all Handler implementations
}
```

## Multiple Implementation Selection

When an interface has multiple implementations and the field is not a slice, selection priority:

1. **Primary**: component embeds `definition.WirePrimaryComponent` or implements `WirePrimary`
2. **Nameless preferred**: components without `Naming()` are preferred over named ones
3. **Arbitrary**: one is chosen (no guaranteed order)

### Primary

```go
import "github.com/go-kid/ioc/definition"

type PreferredImpl struct {
    definition.WirePrimaryComponent  // mark as primary
}
```

### Qualifier

Implement `WireQualifier` to group components:

```go
func (c *MyComp) Qualifier() string { return "groupA" }
```

Inject by qualifier:

```go
type App struct {
    Svc  IService   `wire:",qualifier=groupA"`         // single from groupA
    Svcs []IService `wire:",qualifier=groupA"`          // all from groupA
    Both []IService `wire:",qualifier=groupA groupB"`   // from groupA and groupB
}
```

Empty qualifier (`wire:",qualifier"`) matches components whose `Qualifier()` returns `""`.

### Optional Injection

```go
type App struct {
    Svc IService `wire:",required=false"` // nil if not found
}
```

## `func` Tag (Method-based Matching)

Match components by method name and optional return value:

```
func:"<MethodName>,returns=<value>"
```

```go
type App struct {
    Comp  *MyComp   `func:"Type"`            // has method Type() with no return
    Comp2 *MyComp   `func:"Type,returns=A"`  // Type() returns "A"
    All   []Handler `func:"Type,returns=*"`  // any return value
    Multi []Handler `func:"Type,returns=A B"` // returns "A" or "B"
}
```

## Constructor Injection

Register a constructor function that takes dependencies as parameters:

```go
func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

ioc.Run(app.SetComponents(
    &Repository{},
    NewService,  // register constructor function
    processors.NewConstructorAwarePostProcessors(),
))
```

The framework resolves constructor parameters as dependencies automatically.
