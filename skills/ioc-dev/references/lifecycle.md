# Application lifecycle

## `definition` interface map

Choose the narrowest interface that expresses the behavior:

| Interface | Use |
| --- | --- |
| `NamingComponent` | Override the default package-qualified component name |
| `InitializingComponent` / `InitializingComponentWithContext` | Run `AfterPropertiesSet` after population and before `Init` |
| `InitializeComponent` / `InitializeComponentWithContext` | Initialize one populated component |
| `ApplicationRunner` / `ApplicationRunnerWithContext` | Run application-level startup work after all eager components refresh |
| `CloserComponent` / `CloserComponentWithContext` | Release resources during application shutdown |
| `Ordered` and `Priority` | Order runners, loaders, and component post-processors where that flow performs sorting |
| `WirePrimary` and `WireQualifier` | Select among multiple dependency candidates |
| `LazyInit` | Exclude a singleton from eager refresh |
| `ScopeComponent` | Select singleton or prototype scope |
| `ConditionalComponent` | Decide whether a component participates in eager refresh |
| `ConfigurationProperties` | Supply the configuration prefix for an untagged field |
| `ApplicationEventListener` / `ApplicationEventPublisher` | Receive or publish synchronous application events |

Embed `definition.PriorityComponent`, `definition.WirePrimaryComponent`, or `definition.LazyInitComponent` when their marker behavior is sufficient. Implement `Order`, `Qualifier`, `Naming`, `Scope`, and `Condition` explicitly because their return values carry application-specific behavior.

## Startup APIs

```go
// Package-level API
a, err := ioc.Run(app.SetComponents(&Service{}))

// Context propagation
a, err = ioc.RunWithContext(ctx, app.SetComponents(&Service{}))

// Explicit application
a = app.NewApp()
err = a.RunWithContext(ctx, app.SetComponents(&Service{}))
```

`ioc.Register` queues components for the next package-level `ioc.Run`. `app.Settings` queues global options for the next application run. Both queues are consumed, and global options are applied after options passed directly to `Run`.

Common options include:

| Option | Effect |
| --- | --- |
| `SetComponents` | Register component pointers |
| `SetConfig` / `AddConfigLoader` | Append configuration sources |
| `SetConfigLoader` | Replace configuration sources |
| `SetConfigBinder` | Replace the binder |
| `SetConfigure`, `SetRegistry`, `SetFactory` | Replace container infrastructure |
| `SetLogger`, `LogLevel`, `LogTrace` | Configure the package-global logger |
| `SetShutdownTimeout` | Add a deadline to the context passed to context-aware closers |
| `SkipRunners` | Do not invoke application runners |
| `Options` | Combine options |

## Startup order

1. Apply direct and global options.
2. Register the app and built-in processors.
3. Initialize configuration loaders and binder.
4. Prepare definitions and post-processors.
5. Eagerly refresh non-lazy singleton components whose conditions pass.
6. Invoke application runners unless skipped.
7. Publish `ApplicationStartedEvent`.

For each normally created component, population (`wire`, `func`, `value`, `prop`, `prefix`) completes before initialization:

1. `PostProcessBeforeInstantiation`.
2. Use the registered component instance.
3. `PostProcessAfterInstantiation` and `PostProcessProperties`.
4. Inject resolved dependencies.
5. `PostProcessBeforeInitialization`.
6. `AfterPropertiesSet(ctx)` or `AfterPropertiesSet()`.
7. `Init(ctx)` or `Init()`.
8. `PostProcessAfterInitialization`.
9. Publish `ComponentCreatedEvent` with the final exposed instance.

## Initialization interfaces

Initialization supports either the context or non-context form:

```go
func (s *Service) AfterPropertiesSet(ctx context.Context) error { return nil }
func (s *Service) Init(ctx context.Context) error               { return nil }
```

If a type implements both forms of the same hook, the context form is called.

## Runners and closers

The current app discovers runners through `[]definition.ApplicationRunner` and closers through `[]definition.CloserComponent`. Therefore context-aware implementations must also provide the base method:

```go
func (r *Runner) Run() error { return nil }
func (r *Runner) RunWithContext(ctx context.Context) error { return nil }

func (c *Resource) Close() error { return nil }
func (c *Resource) CloseWithContext(ctx context.Context) error { return nil }
```

When both methods exist, the context form is used. Runners execute sequentially after refresh; `definition.Ordered` sorts lower values first, and `definition.Priority` places a component before ordinary ordered components.

`App.Close` publishes `ApplicationClosingEvent`, invokes `DestructionAwareComponentPostProcessor` callbacks for created singletons in reverse creation order, closes factory-owned resources such as the debug server, then runs closer components concurrently. Prototype instances are caller-owned and are not included in automatic destruction. `SetShutdownTimeout` supplies a deadline to `CloseWithContext`; it does not forcibly stop a closer or make `Close` return early.

## Events

Register `definition.ApplicationEventListener` components to receive `ComponentCreatedEvent`, `ApplicationStartedEvent`, and `ApplicationClosingEvent`. `ComponentCreatedEvent` is published synchronously after initialization and contains the final exposed component. A listener only receives events published after that listener has itself been created and injected into the app. `App.PublishEvent` is also available for custom events.

## Skip selected work

- `app.SkipRunners()` skips only application runners.
- `ext.SkipComponentInitialization()` preserves population but skips `AfterPropertiesSet`, `Init`, all runners, and the after-initialization processor chain. It is intended for local wiring/config inspection when external dependencies are unavailable.
- Debug mode with `--ioc:dry_run` has equivalent broad skip semantics plus the debug UI.

Because initialization skipping also suppresses after-initialization wrapping, it is not suitable for validating AOP proxies or final decorated instances.

## Lazy and conditional components

Embedding `definition.LazyInitComponent` skips eager refresh. `ConditionalComponent.Condition` controls inclusion in the eager-refresh list. Either component can still be created when another component explicitly resolves it.
