---
name: ioc-app-lifecycle
description: "go-kid/ioc framework application startup and component lifecycle guide. Use when setting up application entry point with ioc.Run or app.NewApp, configuring app options (SettingOption), managing component lifecycle (Init, AfterPropertiesSet, Run, Close), controlling execution order, using lazy initialization, context support, scope, conditional registration, skipping runners, or events. Triggers on: ioc.Run, ioc.RunWithContext, app.NewApp, ApplicationRunner, ApplicationRunnerWithContext, CloserComponent, CloserComponentWithContext, InitializingComponent, InitializingComponentWithContext, InitializeComponent, InitializeComponentWithContext, LazyInit, Ordered, PriorityOrdered, app.Settings, SettingOption, SetShutdownTimeout, SkipRunners, ScopeComponent, ConditionalComponent, ApplicationEvent, ApplicationEventListener."
---

# go-kid/ioc Application & Lifecycle

Requires **Go 1.21+**.

## Application Startup

### Style 1: Global API

```go
ioc.Register(&ComponentA{}, &ComponentB{})
app, err := ioc.Run()  // returns *app.App
if err != nil { panic(err) }
defer app.Close()
```

### Style 2: Global API with Context

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

app, err := ioc.RunWithContext(ctx)
if err != nil { panic(err) }
defer app.Close()
```

### Style 3: Explicit App

```go
import "github.com/go-kid/ioc/app"

application := app.NewApp()
err := application.Run(
    app.SetComponents(&ComponentA{}),
    app.SetConfig("config.yaml"),
)
defer application.Close()
```

### Global Settings

Apply options to all `ioc.Run()` calls:

```go
func init() {
    app.Settings(app.LogDebug, app.SetConfig("config.yaml"))
}
```

## `app.SettingOption` Reference

| Option | Description |
|--------|-------------|
| `app.SetComponents(cs ...any)` | Register components |
| `app.SetConfig(path string)` | Load config from file |
| `app.SetConfigLoader(loaders ...configure.Loader)` | Set config loaders (replaces existing) |
| `app.AddConfigLoader(loaders ...configure.Loader)` | Add config loaders |
| `app.SetConfigBinder(binder configure.Binder)` | Set config binder (default: viper yaml) |
| `app.SetConfigure(c configure.Configure)` | Replace entire Configure |
| `app.SetRegistry(r container.SingletonRegistry)` | Replace singleton registry |
| `app.SetFactory(f container.Factory)` | Replace component factory |
| `app.SetLogger(l syslog.Logger)` | Set custom logger |
| `app.SetShutdownTimeout(d time.Duration)` | Set graceful shutdown timeout |
| `app.LogLevel(lv syslog.Lv)` | Set log level |
| `app.LogTrace` | Shortcut: trace level |
| `app.LogDebug` | Shortcut: debug level |
| `app.LogWarn` | Shortcut: warn level |
| `app.LogError` | Shortcut: error level |
| `app.SkipRunners()` | Skip ApplicationRunner execution |
| `app.Options(ops ...SettingOption)` | Combine multiple options |

## Startup Flow

```
app.Run(options...)
  -> Apply SettingOption + globalOptions
  -> initiate()            Register App + built-in PostProcessors
  -> initConfiguration()   Loader.LoadConfig() -> Binder.SetConfig()
  -> initFactory()         PrepareComponents: scan definitions, evaluate conditions
  -> refresh()             Create & inject all components (respecting scope)
  -> callRunners()         Call ApplicationRunner.Run() / RunWithContext() in order
  -> publishEvent()        Publish ApplicationStartedEvent
```

## Lifecycle Interfaces

All interfaces are in `github.com/go-kid/ioc/definition`.

### `InitializingComponent` / `InitializingComponentWithContext`

Called after all properties are injected, before `Init()`:

```go
// Original
func (c *MyComp) AfterPropertiesSet() error { return nil }

// Context-aware (implement one or the other, not both)
func (c *MyComp) AfterPropertiesSet(ctx context.Context) error { return nil }
```

### `InitializeComponent` / `InitializeComponentWithContext`

Called after `AfterPropertiesSet()`:

```go
// Original
func (c *MyComp) Init() error { return nil }

// Context-aware (implement one or the other, not both)
func (c *MyComp) Init(ctx context.Context) error { return nil }
```

### `ApplicationRunner` / `ApplicationRunnerWithContext`

Called after all components are fully initialized. Use `Ordered` to control execution sequence:

```go
// Original
func (r *MyRunner) Run() error { return nil }

// Context-aware (can coexist with Run; RunWithContext is preferred if both exist)
func (r *MyRunner) RunWithContext(ctx context.Context) error { return nil }

func (r *MyRunner) Order() int { return 10 }  // lower = earlier
```

### `CloserComponent` / `CloserComponentWithContext`

Called when `app.Close()` is invoked. All closers run **concurrently**:

```go
// Original
func (c *MyComp) Close() error { return nil }

// Context-aware (can coexist with Close; CloseWithContext is preferred if both exist)
func (c *MyComp) CloseWithContext(ctx context.Context) error { return nil }
```

### `LazyInit`

Skip eager initialization during `Refresh()`. Component is created on first access:

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct {
    definition.LazyInitComponent  // embed to mark lazy
}
```

### `ScopeComponent`

Control component scope:

```go
import "github.com/go-kid/ioc/definition"

type MyPrototype struct{}

func (p *MyPrototype) Scope() string { return definition.ScopePrototype }
```

- `definition.ScopeSingleton` (default): single instance in the container
- `definition.ScopePrototype`: new instance created on every access

### `ConditionalComponent`

Decide at runtime whether a component should be created:

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct{}

func (c *MyComp) Condition(ctx definition.ConditionContext) bool {
    return ctx.HasComponent("dependency")
}
```

`ConditionContext` provides:
- `HasComponent(name string) bool`
- `GetConfig(key string) interface{}`

## Event Mechanism

### Interfaces

```go
// Listener
type ApplicationEventListener interface {
    OnEvent(event ApplicationEvent) error
}

// Publisher (injected by the framework)
type ApplicationEventPublisher interface {
    PublishEvent(event ApplicationEvent) error
}
```

### Built-in Events

- `ComponentCreatedEvent`: fired when a component is created
- `ApplicationStartedEvent`: fired after all runners complete
- `ApplicationClosingEvent`: fired when the application is shutting down

### Usage

Register a listener as a component:

```go
type MyListener struct{}

func (l *MyListener) OnEvent(event definition.ApplicationEvent) error {
    switch e := event.(type) {
    case *definition.ApplicationStartedEvent:
        // app started
    }
    return nil
}

ioc.Register(&MyListener{})
```

## Execution Order Control

### `Ordered`

Controls execution order for `ApplicationRunner` and PostProcessors:

```go
func (r *MyRunner) Order() int { return 10 }
```

Lower value = higher priority (executed first).

### `PriorityOrdered`

Higher priority than `Ordered`. Components implementing `Priority()` execute before regular ordered components:

```go
import "github.com/go-kid/ioc/definition"

type HighPriorityRunner struct {
    definition.PriorityComponent  // embed for Priority()
}

func (r *HighPriorityRunner) Order() int { return 1 }
func (r *HighPriorityRunner) Run() error { return nil }
```

## Lifecycle Execution Order

```
1. Evaluate ConditionalComponent.Condition()
2. PostProcessBeforeInstantiation
3. Component created (instantiation / constructor invocation)
4. PostProcessAfterInstantiation
5. PostProcessProperties (dependency injection)
6. PostProcessBeforeInitialization
7. AfterPropertiesSet()      <- InitializingComponent / WithContext
8. Init()                    <- InitializeComponent / WithContext
9. PostProcessAfterInitialization
10. Publish ComponentCreatedEvent
11. ... (all components done) ...
12. ApplicationRunner.Run() / RunWithContext()  <- in Ordered sequence
13. Publish ApplicationStartedEvent
14. ... (app.Close() called) ...
15. Publish ApplicationClosingEvent
16. CloserComponent.Close() / CloseWithContext()  <- concurrent, with timeout
```
