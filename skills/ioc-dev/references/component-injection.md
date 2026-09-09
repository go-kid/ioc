# Component injection

## Registration

Register component pointers:

```go
ioc.Register(&Repository{}, &Service{})
a, err := ioc.Run()

// Prefer explicit options when package-global registration is undesirable.
a, err = ioc.Run(app.SetComponents(&Repository{}, &Service{}))
```

`ioc.Register` queues components for the next `ioc.Run`, which consumes and clears that queue. Duplicate component names panic. Function registration is unsupported; invoke a constructor first and register its returned pointer.

## Tagged fields

Only exported, settable fields are scanned. Embedded value structs without tags are traversed.

```go
type HandlerSet struct {
	Primary Handler   `wire:""`
	Named   Handler   `wire:"audit"`
	All     []Handler `wire:""`
	Maybe   *Metrics  `wire:",required=false"`
}
```

`wire:""` matches a pointer type, an interface, or a slice of either. A non-empty first value selects by component name and applies to pointer or interface fields. Dependencies are required by default.

Give a component an explicit name with `Naming()`:

```go
func (*AuditHandler) Naming() string { return "audit" }
```

## Multiple matches

For a non-slice field, the current selection order is:

1. A component implementing `definition.WirePrimary`.
2. A component without a `Naming()` alias.
3. A candidate from registry iteration; do not depend on its order.

Mark a primary by implementing `Primary()` or embedding `definition.WirePrimaryComponent`.

Filter matches with qualifiers:

```go
func (*EUHandler) Qualifier() string { return "eu" }

type App struct {
	EU  Handler   `wire:",qualifier=eu"`
	All []Handler `wire:",qualifier=eu apac"`
}
```

`wire:",qualifier"` matches implementations whose `Qualifier()` returns an empty string.

## Method matching

The `func` tag filters type-compatible components by a zero-argument method:

```go
type App struct {
	Default Handler   `func:"Mode"`              // Mode has no return value
	Async   Handler   `func:"Mode,returns=async"`
	All     []Handler `func:"Mode,returns=*"`    // any first return value
	Some    []Handler `func:"Mode,returns=a b"`
}
```

Method matching invokes the method when `returns` is present. Keep such methods side-effect free.

## Scope, lazy creation, and conditions

Singleton is the default scope. Return `definition.ScopePrototype` from `Scope()` to request a new instance on each lookup.

Embedding `definition.LazyInitComponent` excludes a component from eager refresh; requesting it as a dependency still creates it.

`ConditionalComponent.Condition` is evaluated while the factory builds its eager-refresh list:

```go
func (*OptionalWorker) Condition(ctx definition.ConditionContext) bool {
	return ctx.GetConfig("worker.enabled") == true
}
```

The context exposes `HasComponent(name)` and `GetConfig(key)`. A false condition skips eager creation but does not remove the definition from the registry, so another dependency can still request it.

Singleton circular dependencies use early references. Prototype circular dependencies are unsupported; proxying processors must return compatible early and final references.
