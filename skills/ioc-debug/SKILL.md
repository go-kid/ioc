---
name: ioc-debug
description: Diagnose go-kid/ioc startup, registration, dependency resolution, configuration injection, lifecycle, and post-processor failures. Use when an IoC application panics, returns an error, leaves an injected field nil, or behaves differently from its tags and interfaces.
---

# go-kid/ioc Debugging

Diagnose against the exact module version in use. Start from the returned error and the smallest failing component graph; do not change application design until the failing phase is identified.

## Get evidence

Enable trace logs for registration, definition scanning, dependency matching, population, and initialization:

```go
a, err := ioc.Run(app.LogTrace, app.SetComponents(...))
```

For an interactive local graph and factory event stream, use:

```go
a, err := ioc.RunDebug(app.LogTrace, app.SetComponents(...))
```

`RunDebug` starts a loopback server on a random port and opens it in a browser. `--ioc:run_debug` makes ordinary `ioc.Run` enter debug mode. Add `--ioc:dry_run` only with debug mode to populate dependencies while skipping component initialization and runners.

For non-interactive local startup without initializers or runners, use `ext.SkipComponentInitialization()`. It also skips the after-initialization post-processor chain, so do not use it to verify AOP proxies or other final wrappers.

## Interpret the failure

- `component definition with name 'X' not found`: verify the exact `Naming()` value and registration through `app.SetComponents`/`ioc.Register`.
- Tagged field remains nil without an error: verify the field is exported and settable; unexported fields are not scanned.
- Required injection fails: `wire`, `value`, and `prefix` are required by default. Use `required=false` only when absence is valid.
- Wrong interface implementation: inspect `Naming()`, `Primary()`, and `Qualifier()` plus the tag's name/qualifier. A non-slice match is selected as Primary, then a non-aliased candidate, then an implementation-dependent candidate.
- Registration panics for a function: invoke the function first and register the returned component pointer.
- Config is missing: distinguish `SetConfigLoader` (replace loaders) from `AddConfigLoader` (append), then verify the path, binder format, placeholder, and tag default.
- Runner or closer is not called: a type with `RunWithContext` must also implement `Run`; a type with `CloseWithContext` must also implement `Close` to enter the current discovery slices.
- Post-processor property logic is not called: when embedding `DefaultInstantiationAwareComponentPostProcessor`, override `PostProcessAfterInstantiation` to return `true`.

Dependency creation failures include a `dependency resolution failed` chain. Read it from the outer component toward the final missing or failed dependency; retain wrapped causes when reporting or testing the failure.

## Circular dependencies

Singleton cycles are resolved through early singleton exposure. Prototype cycles are not supported. If a singleton cycle still fails, inspect processors that return a proxy from `GetEarlyBeanReference` or after initialization: all dependents must observe a compatible final instance.

When framework behavior is unclear, inspect `app/app.go`, `container/factory/factory.go`, and `container/factory/post_processor_registration_delegate.go` before proposing a workaround.
