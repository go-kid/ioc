# Testing container extensions

Use the [ioc-dev extension guide](../../ioc-dev/references/postprocessor-extensions.md) to choose and implement extension interfaces. This reference covers how to prove the extension is actually reachable and correct through the container.

## Choose the test level

| Extension | Direct test | Required integration assertion |
| --- | --- | --- |
| `configure.Loader` | Returned bytes, empty result, source errors | Loader ordering and Binder result through `Configure.Initialize` |
| `configure.Binder` | Merge, path lookup, runtime `Set`, malformed payload | `value`/`prop`/`prefix` receives the expected decoded value |
| `ComponentFactoryPostProcessor` | Captures the supplied factory or returns a wrapped error | Registered processor is invoked during `App.Run` |
| `DefinitionRegistryPostProcessor` | Metadata/properties created for one component | Custom tag reaches the target field through full population |
| `InstantiationAwareComponentPostProcessor` | Each callback's return/error branch | Registration activates the callback at the intended lifecycle phase |
| `SmartInstantiationAwareBeanPostProcessor` | Proxy cache returns one compatible wrapper | Circular dependents receive the same final proxy |
| `DestructionAwareComponentPostProcessor` | Filtering and error propagation | Created singleton is visited on `App.Close`; unused lazy/prototype components are not |
| `FactoryHook` | Event-to-state mapping | Real preparation/population emits the expected stable event fields |

## Full-container pattern

```go
func TestCustomTag(t *testing.T) {
	target := &Consumer{}
	a := app.NewApp()
	require.NoError(t, a.Run(app.SetComponents(
		newTokenScanner(),
		&tokenProcessor{},
		target,
	)))
	t.Cleanup(a.Close)

	assert.Equal(t, "resolved", target.Token)
}
```

Use `ioc.RunTest` when success alone is enough. Use `app.NewApp` when the returned error or a later explicit shutdown assertion matters. Prefer `app.SetComponents`; package-global `ioc.Register` and `app.Settings` queues can leak between tests.

## Configuration extensions

- Use `loader.NewRawLoader` for integration fixtures around a custom Binder.
- With multiple loaders, assert both ordering and merge/override behavior; a direct Loader test cannot prove orchestration.
- Assert loader and binder errors with `ErrorIs` plus a stable contextual fragment, not the complete wrapped string.
- For a Binder, cover `Get("")` only if the implementation intentionally supports whole-document access.

## Definition and property processors

- Directly test tag parsing and property mutation with a `component_definition.Meta` when covering small branches.
- Add one full application test proving the scanner was registered before the property processor ran.
- Definition processors run concurrently across components. Exercise shared mutable state with multiple components and `go test -race`.
- Verify `PostProcessAfterInstantiation` returns `true` in the integration path when `PostProcessProperties` is expected.
- Assert observable target fields rather than the incidental order of properties returned from map-backed metadata.

## AOP proxies

Test the object observed by consumers, not merely that proxy callbacks ran:

```go
func TestTracingProxyWithCycle(t *testing.T) {
	consumer := &Consumer{}
	a := ioc.RunTest(t, app.SetComponents(
		&tracingProcessor{},
		&GreeterImpl{},
		consumer,
	))
	t.Cleanup(a.Close)

	proxy, ok := consumer.Greeter.(*tracedGreeter)
	require.True(t, ok)
	assert.Same(t, proxy, consumer.Peer.Greeter)
}
```

Include both a normal singleton and a circular singleton graph. Assert:

- interface fields receive an assignable proxy;
- all circular dependents observe the same proxy instance;
- init methods execute exactly once;
- the raw registration template is not accidentally substituted into a dependent;
- mutable proxy caches pass `go test -race`.

Also test a concrete pointer field when the extension claims to support it. Interface wrappers normally cannot replace `*Concrete` fields.

## Destruction and shutdown

Call `Close` or `CloseWithContext` explicitly. Verify matching, non-matching, and error cases. If order is part of the behavior, record component names and assert reverse singleton creation order without depending on registry iteration order.

Add boundary cases for:

- a lazy singleton that is never requested;
- a lazy singleton that is requested once;
- a prototype requested multiple times;
- a component implementing both closer and destruction-related behavior;
- multiple destruction processors where one returns an error.

## Failure and short-circuit paths

Cover non-nil errors from every implemented callback. When a callback intentionally returns `nil`, assert the documented short circuit: skipped population/init, skipped after-initialization chain, or preservation of the last non-nil component as applicable.

Use `ioc.RunErrorTest` for a simple expected startup failure. Use `app.NewApp().Run` when checking error identity or wrapped context.
