# go-kid/ioc

**English** | [中文](README_zh.md)

`go-kid/ioc` is a Go runtime dependency injection (IoC/DI) framework based on **tag + interface**.

## ✨ Features

- **Component Dependency Injection** (`wire` tag)
  - Inject by pointer type, interface type, or component name
  - Multiple implementation selection strategies (Primary/Qualifier)
  - Inject into interface slices/arrays
- **Configuration Injection** (`value` / `prop` / `prefix` tags)
  - Multiple configuration sources (command-line args, files, raw content)
  - Configuration placeholders `${...}`
  - Expression evaluation `#{...}` (arithmetic, logical, conditional, collection operations)
- **Constructor Injection**: Function-based dependency injection (built-in)
- **Lifecycle Management**: ApplicationRunner, CloserComponent, LazyInitComponent
- **`context.Context` lifecycle support**: WithContext variants for all lifecycle interfaces
- **Scope mechanism**: Singleton/Prototype
- **Conditional component registration**: Register components based on runtime conditions
- **Application event mechanism**: Publish and listen for application events
- **Type-safe generic registration**: `ioc.Provide[T]` validates return type at registration time
- **`log/slog` adapter support**: Integrate with Go's structured logging

## 📦 Installation

Requires **Go 1.21+**.

```bash
go get github.com/go-kid/ioc
```

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	"github.com/go-kid/ioc"
)

// Component to be injected
type ComponentA struct {
	Name string
}

func (a *ComponentA) GetName() string {
	return a.Name
}

// Target struct that depends on ComponentA
type App struct {
	ComponentA *ComponentA `wire:""` // pointer + exported field + wire tag
}

func main() {
	a := new(App)

	// Register components
	ioc.Register(a)
	ioc.Register(&ComponentA{Name: "Comp-A"})

	// Run the framework
	_, err := ioc.Run()
	if err != nil {
		panic(err)
	}

	// Output: "Comp-A"
	fmt.Println(a.ComponentA.GetName())
}
```

> **Key points**: Dependency fields must be exported (start with uppercase) and use `wire:""` to declare injection.

## 📖 Documentation

### 1. Component Dependency Injection (`wire` tag)

Component injection is controlled via tag: `wire:"<name>,arg1=xxx,arg2=yyy"`

- `<name>`: optional, inject by component name
- `,arg=...`: optional extra arguments (e.g., `qualifier`)
- `wire:""` (empty) means inject by type

#### Inject by Pointer Type

```go
type Component struct {
	Name string
}

type App struct {
	C *Component `wire:""` // injected by *Component type
}
```

#### Inject by Interface Type

```go
type IComponent interface {
	GetName() string
}

type ComponentA struct{ Name string }
func (a *ComponentA) GetName() string { return a.Name }

type App struct {
	C IComponent `wire:""` // inject by interface type
}
```

#### Inject by Name

Implement `NamingComponent` interface to define component names:

```go
type ComponentA struct {
	Name          string
	componentName string
}

func (a *ComponentA) Naming() string { return a.componentName }

type App struct {
	ByName IComponent `wire:"comp-A"` // inject by component name
}
```

#### Multiple Implementation Selection Strategy

When an interface has multiple implementations, the container selects by priority:

1. Components implementing `WirePrimary` interface (primary)
2. Components without alias (not implementing `Naming()`)
3. One will be chosen (order not guaranteed)

**Using Primary marker:**

```go
type ComponentA struct {
	Name string
	definition.WirePrimaryComponent // embed Primary component
}
```

**Using Qualifier:**

```go
type ComponentA struct {
	qualifier string
}

func (a *ComponentA) Qualifier() string { return a.qualifier }

type App struct {
	C IComponent `wire:",qualifier=comp-A"` // use qualifier
}
```

#### Inject into Slice/Array

```go
type App struct {
	All []Interface `wire:""` // all Interface implementations
}
```

### 2. Configuration Injection

#### `prefix`: Bind Config Prefix

```go
type configA struct {
	B int   `yaml:"b"`
	C []int `yaml:"c"`
}

type App struct {
	A *configA `prefix:"a"` // bind config prefix "a"
}
```

Configuration (YAML):
```yaml
a:
  b: 123
  c: [1,2,3,4]
```

#### `value`: Literals / Placeholders / Expressions

```go
type T struct {
	DSN string `value:"${db.dsn}"`           // config placeholder
	Sum int    `value:"#{1+2}"`               // expression
	Str string `value:"foo"`                  // literal
}
```

#### `prop`: Sugar for `value`

```go
type DBConfig struct {
	DSN string `prop:"db.dsn"` // equivalent to value:"${db.dsn}"
}
```

#### Placeholders and Expressions

- **`${...}`**: Configuration placeholder, reads value from config
- **`#{...}`**: Expression evaluation, supports arithmetic, logical, conditional, collection operations

Example:

```go
type T struct {
	Arithmetic  int    `value:"#{1+(1*2)}"`                      // 3
	Comparison  bool   `value:"#{1/1==1}"`                       // true
	Conditional string `value:"#{1>2?'a':'b'}"`                  // "b"
	WithConfig  int    `value:"#{${number.val1}+${number.val2}}"` // use config
}
```

### 3. Constructor Injection

Constructor injection is built into the framework. Simply register the constructor directly:

```go
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

ioc.Register(NewService)      // register constructor directly
ioc.Register(&Repository{})
_, err := ioc.Run()
```

Type-safe variant with `ioc.Provide[T]`:

```go
ioc.Provide[Service](NewService)  // validates return type at registration time
```

### 4. Application Startup

**Style 1: Use `ioc.Run` / `ioc.Register`**

```go
ioc.Register(a)
ioc.Register(&ComponentA{Name: "Comp-A"})
_, err := ioc.Run()
```

**Style 2: Work with `app.App` Directly**

```go
application := app.NewApp()
err := application.Run(
	app.SetComponents(...),
	app.SetConfigLoader(loader.NewFileLoader("config.yaml")),
	app.SetConfigBinder(binder.NewViperBinder("yaml")),
)
```

### 5. Lifecycle Interfaces

- `ApplicationRunner`: executed after all components are refreshed
- `CloserComponent`: `Close()` called when app stops
- `LazyInitComponent`: marked as lazy initialization
- `PriorityComponent`: controls post-processor execution order
- `InitializeComponentWithContext` / `InitializingComponentWithContext`: context-aware Init/AfterPropertiesSet
- `ApplicationRunnerWithContext`: context-aware Run
- `CloserComponentWithContext`: context-aware Close
- `ScopeComponent`: control component scope (singleton/prototype)
- `ConditionalComponent`: conditional registration based on runtime conditions
- `ApplicationEventListener` / `ApplicationEventPublisher`: event mechanism

### 6. Context Support

All lifecycle interfaces have WithContext variants. Components can implement them to receive a `context.Context` for timeout control, cancellation propagation, etc.

```go
// Run the framework with a context
app, err := ioc.RunWithContext(ctx)

// Set a shutdown timeout
_, err := ioc.Run(app.SetShutdownTimeout(30 * time.Second))
```

Context-aware lifecycle interfaces:

```go
// InitializeComponentWithContext (replaces InitializeComponent)
func (c *MyComp) Init(ctx context.Context) error { return nil }

// InitializingComponentWithContext (replaces InitializingComponent)
func (c *MyComp) AfterPropertiesSet(ctx context.Context) error { return nil }

// ApplicationRunnerWithContext (can coexist with ApplicationRunner)
func (r *MyRunner) RunWithContext(ctx context.Context) error { return nil }

// CloserComponentWithContext (can coexist with CloserComponent)
func (c *MyComp) CloseWithContext(ctx context.Context) error { return nil }
```

> **Note**: `Init(ctx)` and `AfterPropertiesSet(ctx)` have the same method names as their base interfaces but different signatures, so a component implements one or the other. `RunWithContext` and `CloseWithContext` use distinct method names, allowing a component to implement both the original and context-aware versions.

### 7. Scope

Control component scope by implementing `ScopeComponent`:

- **Singleton** (default): single instance in the container
- **Prototype**: new instance created on every access

```go
import "github.com/go-kid/ioc/definition"

type MyPrototype struct{}

func (p *MyPrototype) Scope() string { return definition.ScopePrototype }
```

### 8. Conditional Registration

Implement `ConditionalComponent` to decide at runtime whether a component should be created:

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct{}

func (c *MyComp) Condition(ctx definition.ConditionContext) bool {
	return ctx.HasComponent("dependency")
}
```

### 9. Events

Publish and listen for application events to enable loose coupling between components:

```go
import "github.com/go-kid/ioc/definition"

// Listen for events
type MyListener struct{}

func (l *MyListener) OnEvent(event definition.ApplicationEvent) error {
	// handle event
	return nil
}
```

Built-in events: `ComponentCreatedEvent`, `ApplicationStartedEvent`, `ApplicationClosingEvent`.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          app.App                                 │
│  ┌──────────────────┐              ┌──────────────────┐         │
│  │ Component Factory│              │    Configure     │         │
│  └────────┬─────────┘              └────────┬─────────┘         │
│           │                                  │                   │
│           │                                  │                   │
│  ┌────────▼─────────┐              ┌────────▼─────────┐       │
│  │ DefinitionRegistry│              │      Loader       │       │
│  │ SingletonRegistry│              │  (args/file/raw)  │       │
│  └────────┬─────────┘              └────────┬─────────┘       │
│           │                                  │                   │
│           │ register                         │                   │
│  ┌────────▼─────────┐              ┌────────▼─────────┐       │
│  │ComponentDefinition│              │      Binder       │       │
│  └────────┬─────────┘              │      (viper)       │       │
│           │                         └───────────────────┘       │
│           │ create                                               │
│  ┌────────▼─────────┐                                           │
│  │     Instance     │                                           │
│  └────────┬─────────┘                                           │
│           │                                                      │
│  ┌────────▼──────────────────────────────────────────┐          │
│  │            PostProcessors                          │          │
│  │  ┌────────────────────────────────────────────┐   │          │
│  │  │ • LoggerAware                               │   │          │
│  │  │ • ConfigQuoteAware ${...}                   │   │          │
│  │  │ • ExpressionTagAware #{...}                  │   │          │
│  │  │ • PropertiesAware prefix                    │   │          │
│  │  │ • ValueAware value/prop                     │   │          │
│  │  │ • DependencyAware wire                      │   │          │
│  │  │ • ConstructorAware func                      │   │          │
│  │  └────────────────────────────────────────────┘   │          │
│  └──────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘

Config Flow: Binder → ConfigQuoteAware, PropertiesAware, ValueAware, ExpressionTagAware
```

> **Note**: The constructor processor is now built into the framework; no separate registration is required.

## 📚 Examples & Tests

- **Example projects**: `test/ioc`, `test/constructor`, `test/prop`, `test/t_yaml`
- **Post processor example**: `examples/post_processor/`
- **Performance tools**: `cmd/performance_analyst/`
- **Unit tests**: `unittest/component/builtin_inject/*`, `unittest/configure/*`

Run tests:

```bash
go test ./...
```

## 📄 License

[MIT License](./LICENSE)

## 🤝 Contributing

Contributions via Issues / PRs are welcome. Please ensure `go test ./...` passes before submitting.
