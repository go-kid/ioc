---
name: ioc-test
description: Test or extend go-kid/ioc with RunTest/RunErrorTest, custom loaders or binders, component post-processors, tag scanners, factory hooks, logging adapters, or proxies. Use for IoC-focused tests and container extension work, not ordinary application wiring.
---

# go-kid/ioc Testing and Extension

Use focused tests that prove injected values, lifecycle effects, selection behavior, or returned errors. Read [references/postprocessor-patterns.md](references/postprocessor-patterns.md) before implementing or reviewing a post-processor, custom tag, or proxy.

## Test helpers

```go
func TestServiceInjection(t *testing.T) {
	target := &struct {
		Service *Service `wire:""`
	}{}

	a := ioc.RunTest(t, app.SetComponents(target, &Service{}))
	defer a.Close()
	require.NotNil(t, target.Service)
}
```

- `ioc.RunTest` creates an `app.App`, runs the full startup path, asserts that startup returned no error, and returns the app.
- `ioc.RunErrorTest` runs the same path and asserts that startup returned an error. It does not return the error value; use `app.NewApp().Run(...)` when the exact error must be asserted.
- Call `Close`/`CloseWithContext` explicitly when testing closers or cleanup.
- Prefer `app.SetComponents` in tests. `ioc.Register` stores global options for the next `ioc.Run` and can couple tests through shared state.
- Use `loader.NewRawLoader` for inline configuration. Remember that `SetConfigLoader` replaces the default args loader; `AddConfigLoader` appends.
- Use `app.SkipRunners()` when runners are irrelevant. Use `ext.SkipComponentInitialization()` only when the test intentionally excludes `AfterPropertiesSet`, `Init`, runners, and after-initialization processors.

## Expected failure

```go
func TestMissingDependency(t *testing.T) {
	target := &struct {
		Missing *Missing `wire:""`
	}{}

	a := app.NewApp()
	err := a.Run(app.SetComponents(target))
	require.ErrorContains(t, err, "not found available components")
}
```

Assert observable behavior or stable error meaning, not complete wrapped wording, log lines, or internal ordering unless ordering is the feature under test.

## Extension boundaries

- Implement `configure.Loader` for a new config source and `configure.Binder` for a new storage/decoding model.
- Adapt a `slog.Handler` with `syslog.NewSlogAdapter` and install it through `app.SetLogger`.
- Register post-processors as ordinary components; the factory discovers the interfaces they implement.
- Prefer the provided default processor structs and override only the callbacks needed.
- Treat `container.DestructionAwareComponentPostProcessor` and `definition.ComponentCreatedEvent` as declared but not currently invoked by the application flow; do not build behavior that depends on them without changing and testing the container implementation.
- Follow the existing `unittest/component` and `unittest/configure` organization when adding framework tests.
