---
name: ioc-app-lifecycle
description: "go-kid/ioc framework application startup and component lifecycle guide. Use when setting up application entry point with ioc.Run or app.NewApp, configuring app options (SettingOption), managing component lifecycle (Init, AfterPropertiesSet, Run, Close), controlling execution order, or using lazy initialization. Triggers on: ioc.Run, app.NewApp, ApplicationRunner, CloserComponent, InitializingComponent, InitializeComponent, LazyInit, Ordered, PriorityOrdered, app.Settings, SettingOption."
---

# go-kid/ioc Application & Lifecycle

## Application Startup

### Style 1: Global API

```go
ioc.Register(&ComponentA{}, &ComponentB{})
app, err := ioc.Run()  // returns *app.App
if err != nil { panic(err) }
defer app.Close()
```

### Style 2: Explicit App

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
| `app.LogLevel(lv syslog.Lv)` | Set log level |
| `app.LogTrace` | Shortcut: trace level |
| `app.LogDebug` | Shortcut: debug level |
| `app.LogWarn` | Shortcut: warn level |
| `app.LogError` | Shortcut: error level |
| `app.Options(ops ...SettingOption)` | Combine multiple options |

## Startup Flow

```
app.Run(options...)
  -> Apply SettingOption + globalOptions
  -> initiate()            Register App + built-in PostProcessors
  -> initConfiguration()   Loader.LoadConfig() -> Binder.SetConfig()
  -> initFactory()         PrepareComponents: scan definitions
  -> refresh()             Create & inject all components
  -> callRunners()         Call ApplicationRunner.Run() in order
```

## Lifecycle Interfaces

All interfaces are in `github.com/go-kid/ioc/definition`.

### `InitializingComponent`

Called after all properties are injected, before `Init()`:

```go
func (c *MyComp) AfterPropertiesSet() error {
    // validate injected fields, set up derived state
    return nil
}
```

### `InitializeComponent`

Called after `AfterPropertiesSet()`:

```go
func (c *MyComp) Init() error {
    // initialize resources (connections, caches, etc.)
    return nil
}
```

### `ApplicationRunner`

Called after all components are fully initialized. Use `Ordered` to control execution sequence:

```go
type MyRunner struct{}

func (r *MyRunner) Run() error {
    // start HTTP server, background jobs, etc.
    return nil
}

func (r *MyRunner) Order() int { return 10 }  // lower = earlier
```

### `CloserComponent`

Called when `app.Close()` is invoked. All closers run **concurrently**:

```go
func (c *MyComp) Close() error {
    // release resources, close connections
    return nil
}
```

### `LazyInit`

Skip eager initialization during `Refresh()`. Component is created on first access:

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct {
    definition.LazyInitComponent  // embed to mark lazy
}
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
1. PostProcessBeforeInstantiation
2. Component created (instantiation)
3. PostProcessAfterInstantiation
4. PostProcessProperties (dependency injection)
5. PostProcessBeforeInitialization
6. AfterPropertiesSet()      <- InitializingComponent
7. Init()                    <- InitializeComponent
8. PostProcessAfterInitialization
9. ... (all components done) ...
10. ApplicationRunner.Run()   <- in Ordered sequence
11. ... (app.Close() called) ...
12. CloserComponent.Close()   <- concurrent
```
