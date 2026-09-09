# Post-processor and AOP extensions

Use the narrowest extension interface that matches the required phase. Register processors with `app.SetComponents` like ordinary component pointers; the factory discovers their interfaces during preparation.

## Choose the extension point

| Interface | Called | Primary use |
| --- | --- | --- |
| `ComponentFactoryPostProcessor` | Once during factory preparation, before definition scanning | Capture `Factory`, `Configure`, or `DefinitionRegistry`; prepare processor infrastructure |
| `DefinitionRegistryPostProcessor` | Per registered component during definition preparation, and again for each prototype instance | Scan fields or types and add component metadata/properties |
| `InstantiationAwareComponentPostProcessor` | Around creation and before dependency injection | Replace creation, populate custom properties, or alter an instance before initialization |
| `SmartInstantiationAwareBeanPostProcessor` | When a singleton cycle requests an early reference | Expose the same AOP proxy that circular dependents must observe |
| `DestructionAwareComponentPostProcessor` | During application close for each created singleton accepted by `RequireDestruction` | Cross-cutting cleanup or shutdown interception |

`ComponentPostProcessor` supplies the shared before/after-initialization callbacks. The instantiation-aware, smart-instantiation, and destruction-aware interfaces extend it, so those implementations must also provide the base callbacks. Embed the default implementations rather than adding empty methods.

`FactoryHook` is separate observability infrastructure used by the debug module. It receives factory events but cannot change components. It is installed on a compatible factory, not discovered as a registered post-processor.

## Actual callback flow

Factory preparation:

1. Discover registered processor interfaces.
2. Invoke all `ComponentFactoryPostProcessor` instances.
3. For each `DefinitionRegistryPostProcessor`, scan registered components concurrently. Prototype lookup later reruns definition processors for the new instance.
4. Sort only the `ComponentPostProcessor` chain: priority processors first, then ordinary ordered processors, then unordered processors.

Normal component creation:

1. Run `PostProcessBeforeInstantiation` until one processor returns a non-nil replacement.
2. If replaced, skip population and init methods, then run the after-initialization chain.
3. Otherwise call each instantiation-aware processor's `PostProcessAfterInstantiation`.
4. When that processor returns `true`, call its `PostProcessProperties`.
5. Inject the resolved component properties.
6. Run `PostProcessBeforeInitialization`, `AfterPropertiesSet`, `Init`, and `PostProcessAfterInitialization`.
7. Cache the final exposed component when its scope is singleton.

Circular singleton creation may call `GetEarlyBeanReference` between steps 4 and 6. Application shutdown invokes destruction-aware processors in reverse singleton creation order before closer components.

Important current semantics:

- `PostProcessAfterInstantiation` controls only the same processor's `PostProcessProperties`; returning `false` does not stop other processors or normal dependency injection.
- The slice returned by `PostProcessProperties` is currently ignored. Mutate the supplied properties or component and return `nil, nil`.
- Returning `nil` from a before-initialization callback stops initialization and the after-initialization chain while retaining the populated original component.
- Returning `nil` from an after-initialization callback retains the last non-nil component and stops the remaining callbacks.
- Only the component post-processor chain honors `Priority`/`Ordered`. Do not rely on ordering among factory or definition-registry processors.
- Definition scanning is concurrent across components. Mutable state used by a definition processor must be concurrency-safe.

## `ComponentFactoryPostProcessor`

Use this phase to acquire container services needed later. Do not resolve ordinary application components here: definitions and the final processor chain are still being prepared.

The processor's own `wire`/`value` fields and init callbacks have not run yet. Supply preparation dependencies directly or obtain them from the `Factory` argument.

```go
type registryAwareProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	registry  container.DefinitionRegistry
	configure configure.Configure
}

func (p *registryAwareProcessor) PostProcessComponentFactory(factory container.Factory) error {
	p.registry = factory.GetDefinitionRegistry()
	p.configure = factory.GetConfigure()
	return nil
}
```

The built-in dependency processors capture the definition registry here; configuration processors capture `Configure`.

## `DefinitionRegistryPostProcessor`

Use a definition processor to describe properties before any normal component is populated. For field tags, prefer the provided scanner:

```go
type tokenScanner struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func newTokenScanner() *tokenScanner {
	return &tokenScanner{
		DefaultTagScanDefinitionRegistryPostProcessor: processors.DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType: component_definition.PropertyTypeConfiguration,
			Tag:      "token",
			Required: false,
		},
	}
}
```

For type-level metadata or additional definitions, implement the interface directly:

```go
func (p *metadataProcessor) PostProcessDefinitionRegistry(
	registry container.DefinitionRegistry,
	component any,
	componentName string,
) error {
	meta := registry.GetMetaOrRegister(componentName, component)
	// Inspect meta.Type/meta.Fields and update this component's metadata.
	_ = meta
	return nil
}
```

Do not initialize components, perform network I/O per field, or depend on component iteration order in this phase.

Definition processors run before their ordinary component lifecycle. A combined definition/property processor should obtain `Configure` or the registry through `ComponentFactoryPostProcessor`, as the built-ins do. `DefaultTagScanDefinitionRegistryPostProcessor` embeds `LazyInitComponent` so its scanner can participate without forcing early self-creation.

## `InstantiationAwareComponentPostProcessor`

Pair a custom property handler with its definition scanner:

```go
type tokenProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (*tokenProcessor) PostProcessAfterInstantiation(any, string) (bool, error) {
	return true, nil // enables this processor's PostProcessProperties
}

func (*tokenProcessor) PostProcessProperties(
	properties []*component_definition.Property,
	component any,
	componentName string,
) ([]*component_definition.Property, error) {
	for _, property := range properties {
		if property.Tag == "token" {
			property.Value.SetString(resolveToken(property.TagVal))
		}
	}
	return nil, nil
}

a, err := ioc.Run(app.SetComponents(
	newTokenScanner(),
	&tokenProcessor{},
	&Consumer{},
))
```

Use `PostProcessBeforeInstantiation` only when a non-nil replacement should bypass normal population, `AfterPropertiesSet`, and `Init`. For ordinary decoration after a component is ready, use `PostProcessAfterInitialization` instead.

## AOP and `SmartInstantiationAwareBeanPostProcessor`

An after-initialization proxy is enough when no singleton cycle exists. With a cycle, dependents can request the target before initialization finishes, so `GetEarlyBeanReference` must expose a compatible proxy.

Under the current factory flow, cache one early proxy and return the original component from `PostProcessAfterInitialization` when that early proxy already exists; the factory then promotes the early reference as the final exposed component.

```go
type Greeter interface {
	Greet(name string) string
}

type tracedGreeter struct {
	next Greeter
}

func (g *tracedGreeter) Greet(name string) string {
	log.Printf("before Greet")
	result := g.next.Greet(name)
	log.Printf("after Greet")
	return result
}

type tracingProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	proxies sync.Map // original pointer address -> proxy
}

func (p *tracingProcessor) proxy(component any) (any, bool) {
	target, ok := component.(Greeter)
	if !ok {
		return component, false
	}
	key := reflect.ValueOf(component).Pointer()
	proxy, _ := p.proxies.LoadOrStore(key, &tracedGreeter{next: target})
	return proxy, true
}

func (p *tracingProcessor) GetEarlyBeanReference(component any, _ string) (any, error) {
	proxy, _ := p.proxy(component)
	return proxy, nil
}

func (p *tracingProcessor) PostProcessAfterInitialization(component any, _ string) (any, error) {
	key := reflect.ValueOf(component).Pointer()
	if _, exposedEarly := p.proxies.Load(key); exposedEarly {
		return component, nil // factory promotes the cached early proxy
	}
	proxy, _ := p.proxy(component)
	return proxy, nil
}

a, err := ioc.Run(app.SetComponents(
	&tracingProcessor{},
	&GreeterImpl{},
	&GreeterConsumer{}, // inject Greeter through an interface field
))
```

Prefer interface-typed injection for wrapped services. A replacement proxy must be assignable to every field receiving it; a wrapper implementing `Greeter` cannot be assigned to a field of concrete type `*GreeterImpl`.

Do not create one proxy in `GetEarlyBeanReference` and another in `PostProcessAfterInitialization`. Circular dependents would retain a different instance and the factory can reject the component. Keep proxy cache state concurrency-safe.

## `DestructionAwareComponentPostProcessor`

Use destruction processing when cleanup is cross-cutting. For cleanup owned by one component, implementing `CloserComponent` is simpler.

```go
type flushable interface {
	Flush() error
}

type flushProcessor struct {
	processors.DefaultComponentPostProcessor
}

func (*flushProcessor) RequireDestruction(component any) bool {
	_, ok := component.(flushable)
	return ok
}

func (*flushProcessor) PostProcessBeforeDestruction(component any, _ string) error {
	return component.(flushable).Flush()
}
```

Only created singletons are visited automatically. Lazy singletons that were never created and caller-owned prototype instances are excluded. All matching destruction processors are attempted; their errors are aggregated and logged by application shutdown.
