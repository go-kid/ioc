---
name: IoC Post-Processor Development
description: This skill should be used when the user asks to "create a post-processor", "add a ComponentPostProcessor", "implement InstantiationAwareComponentPostProcessor", "create AOP proxy", "intercept component creation", "modify component properties", "add custom tag processing", "create DefinitionRegistryPostProcessor", or needs guidance on extending the IoC container's component lifecycle pipeline.
---

# IoC Post-Processor Development

Guide for creating custom post-processors that extend the IoC container's component lifecycle.

## Post-Processor Hierarchy

The framework provides a layered post-processor hierarchy in `container` package:

```
ComponentPostProcessor
  ├── PostProcessBeforeInitialization(component, name) (any, error)
  └── PostProcessAfterInitialization(component, name) (any, error)

InstantiationAwareComponentPostProcessor (extends above)
  ├── PostProcessBeforeInstantiation(meta, name) (any, error)
  ├── PostProcessAfterInstantiation(component, name) (bool, error)
  └── PostProcessProperties(properties, component, name) ([]*Property, error)

SmartInstantiationAwareBeanPostProcessor (extends above)
  └── GetEarlyBeanReference(component, name) (any, error)
```

Additional interfaces:
- `ComponentFactoryPostProcessor` — access the Factory itself
- `DefinitionRegistryPostProcessor` — custom tag scanning/meta parsing
- `DestructionAwareComponentPostProcessor` — pre-destruction hooks

## Using Base Structs

The `processors` package provides embeddable base structs with no-op defaults:

```go
import "github.com/go-kid/ioc/container/processors"

// For simple before/after init hooks:
type MyProcessor struct {
    processors.DefaultComponentPostProcessor
}

// For property/instantiation control:
type MyProcessor struct {
    processors.DefaultInstantiationAwareComponentPostProcessor
}
```

Embed the appropriate base and override only the methods needed.

## Execution Order

Control order by implementing `definition.Ordered`:

```go
func (p *MyProcessor) Order() int { return 100 }
```

Built-in processor order constants (from `container/processors/orders.go`):

**Priority order group** (config processing):
- `PriorityOrderLoggerAware` (2)
- `PriorityOrderPropertyConfigQuoteAware` (4) — `${...}` resolution
- `PriorityOrderPropertyExpressionTagAware` (8) — `#{...}` evaluation
- `PriorityOrderPopulateProperties` (16) — `prefix`/`value`/`prop` binding

**Standard order group** (dependency processing):
- `OrderDependencyAware` (2) — `wire` tag resolution
- `OrderDependencyFurtherMatching` (4) — Primary/Qualifier disambiguation
- `OrderValidate` (8) — struct validation

Custom processors typically use orders > 16 to run after built-in processors.

## Lifecycle Callback Order

For each component during Refresh:
1. `PostProcessBeforeInstantiation` — return non-nil to replace the component entirely
2. Early singleton cache (for circular reference support)
3. `PostProcessAfterInstantiation` — return false to skip property injection
4. `PostProcessProperties` — modify/inject field values
5. `PostProcessBeforeInitialization`
6. `AfterPropertiesSet()` / `Init()` — component's own lifecycle
7. `PostProcessAfterInitialization` — final wrapping (AOP proxy)

## Common Patterns

### AOP Proxy via GetEarlyBeanReference

Create proxies for components involved in circular dependencies (see `examples/post_processor/main.go`):

```go
type AopProcessor struct {
    processors.DefaultInstantiationAwareComponentPostProcessor
    Enable bool `value:"${aop.enable:false}"`
}

func (p *AopProcessor) GetEarlyBeanReference(component any, name string) (any, error) {
    if svc, ok := component.(*TargetService); ok {
        return &ServiceProxy{Service: svc, Enable: p.Enable}, nil
    }
    return component, nil
}
```

### Custom Tag Scanner

Extend `DefaultTagScanDefinitionRegistryPostProcessor` for custom tag scanning:

```go
type MyTagProcessor struct {
    processors.DefaultTagScanDefinitionRegistryPostProcessor
    processors.DefaultInstantiationAwareComponentPostProcessor
}

func NewMyTagProcessor() *MyTagProcessor {
    return &MyTagProcessor{
        DefaultTagScanDefinitionRegistryPostProcessor: processors.DefaultTagScanDefinitionRegistryPostProcessor{
            NodeType: component_definition.PropertyTypeComponent, // or PropertyTypeConfigure
            Tag:      "mytag",
            Required: false,
        },
    }
}

func (p *MyTagProcessor) PostProcessProperties(
    properties []*component_definition.Property,
    component any, name string,
) ([]*component_definition.Property, error) {
    for _, prop := range properties {
        if prop.Tag != "mytag" { continue }
        // Process fields with `mytag:"..."` tags
    }
    return nil, nil
}
```

### Accessing the Factory

```go
type MyProcessor struct {
    processors.DefaultComponentPostProcessor
}

func (p *MyProcessor) PostProcessComponentFactory(factory container.Factory) error {
    // Access registry, configure, etc.
    registry := factory.GetDefinitionRegistry()
    return nil
}
```

## Registration

Post-processors are registered as regular components. The factory auto-detects them:

```go
ioc.Register(&MyProcessor{})
// or
app.SetComponents(&MyProcessor{})
```

## Key Rules

- Post-processors can themselves receive injections (`wire`, `value`, etc.)
- Returning a different object from `PostProcessAfterInitialization` or `GetEarlyBeanReference` creates a proxy — the container uses the proxy going forward
- `PostProcessProperties` returning `nil, nil` means "no modifications, continue normally"
- Mark post-processors with `definition.LazyInitComponent` if they should not participate in early container setup
- Implement `definition.Priority` (embed `definition.PriorityComponent`) to ensure the post-processor itself is initialized before other components
