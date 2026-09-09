---
name: ioc-test
description: Test go-kid/ioc applications and extensions, including injection, configuration, lifecycle, post-processors, AOP proxies, circular dependencies, and shutdown behavior. Use for IoC-focused test design and regression coverage, not implementation guidance.
---

# go-kid/ioc Testing

Use focused tests that prove injected values, lifecycle effects, selection behavior, or returned errors. Read [references/extension-testing.md](references/extension-testing.md) when testing a custom loader/binder, post-processor, tag scanner, AOP proxy, early singleton reference, factory hook, or destruction callback. Use the [ioc-dev extension guide](../ioc-dev/references/postprocessor-extensions.md) for implementation and interface semantics.

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

## Coverage boundaries

- Test public behavior through a full application when registration and callback reachability matter; use package-local tests for isolated parsing, matching, and error branches.
- Add a regression test for every fixed lifecycle or container boundary.
- Run race tests for concurrent definition scanning, mutable proxy caches, singleton creation, and debug event collection.
- Follow the existing `unittest/component` and `unittest/configure` organization when adding framework tests.
