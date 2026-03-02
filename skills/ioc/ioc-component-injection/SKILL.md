---
name: ioc-component-injection
description: "go-kid/ioc framework component dependency injection guide. Use when registering components, injecting dependencies with `wire` tag, selecting from multiple implementations (Primary/Qualifier), injecting into slices, using `func` tag for method-based matching, using constructor injection, or using type-safe generic registration with ioc.Provide. Triggers on: wire tag, component registration, dependency injection, ioc.Register, ioc.Provide, app.SetComponents, NamingComponent, WireQualifier, WirePrimary, constructor injection, ScopeComponent, ConditionalComponent."
---

# go-kid/ioc Component Injection

Requires **Go 1.21+**.

## Component Registration

Three styles for registering components:

```go
// Style 1: Global registration (pointer)
ioc.Register(&MyComponent{})
ioc.Run()

// Style 2: Via app options
app.NewApp().Run(app.SetComponents(&MyComponent{}, &AnotherComponent{}))

// Style 3: Constructor function
ioc.Register(NewMyComponent)  // func(...deps) *MyComponent
ioc.Run()
```

Registered values must be **pointers** (e.g. `&MyStruct{}`), or **constructor functions** returning pointers.

### Type-Safe Registration with `ioc.Provide[T]`

Use generics to validate the constructor return type at registration time:

```go
// Concrete type: constructor must return *Service
ioc.Provide[Service](func(repo *Repository) *Service {
    return &Service{repo: repo}
})

// Interface type: constructor must return a type implementing IService
ioc.Provide[IService](func(repo *Repository) *serviceImpl {
    return &serviceImpl{repo: repo}
})

// Supports (value, error) return pattern
ioc.Provide[Service](func(repo *Repository) (*Service, error) {
    return &Service{repo: repo}, nil
})
```

Panics at registration time if the return type does not match `T`.

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

Constructor injection is built into the framework. Register a constructor function that takes dependencies as parameters:

```go
func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

// Register constructor directly - no additional processor needed
ioc.Register(NewService)
ioc.Register(&Repository{})
_, err := ioc.Run()
```

### Constructor Parameter Resolution

Constructor parameters are resolved as dependencies using the same rules as `wire:""` tag injection:

- **Pointer types**: `*Repository` matches a registered `*Repository`
- **Interface types**: `Logger` matches any registered type implementing `Logger`
- **Slice types**: `[]Handler` injects all registered `Handler` implementations
- **ConfigurationProperties**: parameters implementing `ConfigurationProperties` are automatically populated from config

### Type-Safe Constructor Registration

```go
ioc.Provide[Service](NewService)  // validates return type at registration time
```

## Scope

Control component scope by implementing `ScopeComponent`:

```go
import "github.com/go-kid/ioc/definition"

type MyPrototype struct{}
func (p *MyPrototype) Scope() string { return definition.ScopePrototype }
```

- `definition.ScopeSingleton` (default): single instance, returned on every request
- `definition.ScopePrototype`: new instance created on every access

## Conditional Registration

Implement `ConditionalComponent` to decide at runtime whether a component should be created:

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct{}
func (c *MyComp) Condition(ctx definition.ConditionContext) bool {
    return ctx.HasComponent("dependency")
}
```
