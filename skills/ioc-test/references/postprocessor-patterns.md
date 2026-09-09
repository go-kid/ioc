# Post-processor patterns

## Interfaces and current support

Interfaces live in `container/def.go`:

```text
ComponentPostProcessor
  PostProcessBeforeInitialization
  PostProcessAfterInitialization

InstantiationAwareComponentPostProcessor
  ComponentPostProcessor
  PostProcessBeforeInstantiation
  PostProcessAfterInstantiation
  PostProcessProperties

SmartInstantiationAwareBeanPostProcessor
  InstantiationAwareComponentPostProcessor
  GetEarlyBeanReference

DefinitionRegistryPostProcessor
ComponentFactoryPostProcessor
DestructionAwareComponentPostProcessor
FactoryHook
```

The factory invokes component, instantiation-aware, smart-instantiation, definition-registry, factory, and factory-hook callbacks. `DestructionAwareComponentPostProcessor` is declared and detected but is not called by the current `App.Close` flow.

## Default implementations

Embed `processors.DefaultComponentPostProcessor` for before/after initialization hooks.

Embed `processors.DefaultInstantiationAwareComponentPostProcessor` for instantiation or property callbacks. Its `PostProcessAfterInstantiation` returns `false`, so a processor implementing `PostProcessProperties` must override it:

```go
func (*MyProcessor) PostProcessAfterInstantiation(any, string) (bool, error) {
	return true, nil
}
```

In the current delegate, that boolean controls only the same processor's `PostProcessProperties` call; it does not cancel other processors. The returned property slice is currently ignored, so mutate the supplied properties or component and return `nil, nil`.

`processors.DefaultTagScanDefinitionRegistryPostProcessor` scans exported fields and adds properties for a configured tag. It embeds `definition.LazyInitComponent`, allowing the raw scanner to participate in definition preparation without being initialized first.

## Callback semantics

- A non-nil result from `PostProcessBeforeInstantiation` replaces normal creation and then enters the after-initialization processor chain; population and init methods are skipped.
- Returning `nil` from any `PostProcessBeforeInitialization` stops that chain, skips init methods and all after-initialization processors, and retains the populated original instance.
- Returning a different object after initialization creates a proxy meta used by later lookups.
- Returning `nil` from an after-initialization processor stops the remaining chain while retaining the last non-nil result.
- `GetEarlyBeanReference` must produce a proxy compatible with the final proxy when singleton cycles are possible.

Definition-registry processors scan components concurrently. Make mutable scanner state thread-safe and avoid depending on component iteration order.

## Ordering

`definition.Priority` components are sorted before ordinary `definition.Ordered` components. Within each group, lower `Order()` values run first.

Built-in priority orders resolve logger/config placeholders/expressions/population. Ordinary orders then resolve dependencies, qualifiers/Primary, and validation. Use the constants in `container/processors/orders.go` when a custom processor must run relative to a built-in phase; otherwise avoid coupling to exact numeric values.

## Custom field tag

Use one scanner to create properties and one instantiation-aware processor to handle them:

```go
type tokenScanner struct {
	processors.DefaultTagScanDefinitionRegistryPostProcessor
}

func newTokenScanner() *tokenScanner {
	return &tokenScanner{
		DefaultTagScanDefinitionRegistryPostProcessor: processors.DefaultTagScanDefinitionRegistryPostProcessor{
			NodeType: component_definition.PropertyTypeConfiguration,
			Tag:      "token",
		},
	}
}

type tokenProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (*tokenProcessor) PostProcessAfterInstantiation(any, string) (bool, error) {
	return true, nil
}

func (*tokenProcessor) PostProcessProperties(
	properties []*component_definition.Property,
	component any,
	name string,
) ([]*component_definition.Property, error) {
	for _, property := range properties {
		if property.Tag == "token" {
			property.Value.SetString(resolveToken(property.TagVal))
		}
	}
	return nil, nil
}

ioc.Run(app.SetComponents(
	newTokenScanner(),
	&tokenProcessor{},
	&Consumer{},
))
```

Keep the handler responsible only for the custom tag. Use an existing built-in tag when its semantics already match.

## Factory access and proxies

Implement `ComponentFactoryPostProcessor` when setup needs the definition registry or configure object before refresh:

```go
func (p *Processor) PostProcessComponentFactory(factory container.Factory) error {
	p.registry = factory.GetDefinitionRegistry()
	p.configure = factory.GetConfigure()
	return nil
}
```

For AOP around singleton cycles, implement `SmartInstantiationAwareBeanPostProcessor` and keep `GetEarlyBeanReference` consistent with `PostProcessAfterInitialization`. Use `examples/post_processor/main.go` as the repository-local integration example.

For dependency-population-only startup, prefer the maintained `ext.SkipComponentInitialization()` extension over recreating its short-circuit processor.
