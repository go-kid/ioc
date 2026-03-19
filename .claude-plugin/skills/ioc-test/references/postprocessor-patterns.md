# PostProcessor Interface Hierarchy & Built-in Patterns

## Table of Contents

1. [Interface Hierarchy](#interface-hierarchy)
2. [Base Classes](#base-classes)
3. [Built-in PostProcessors](#built-in-postprocessors)
4. [Execution Order Constants](#execution-order-constants)
5. [Complete Custom Tag Example](#complete-custom-tag-example)

## Interface Hierarchy

All interfaces defined in `container/def.go`:

```
ComponentPostProcessor
├── PostProcessBeforeInitialization(component any, name string) (any, error)
└── PostProcessAfterInitialization(component any, name string) (any, error)

InstantiationAwareComponentPostProcessor (extends ComponentPostProcessor)
├── PostProcessBeforeInstantiation(m *Meta, name string) (any, error)
├── PostProcessAfterInstantiation(component any, name string) (bool, error)
└── PostProcessProperties(props []*Property, component any, name string) ([]*Property, error)

SmartInstantiationAwareBeanPostProcessor (extends InstantiationAware)
└── GetEarlyBeanReference(component any, name string) (any, error)

DestructionAwareComponentPostProcessor (extends ComponentPostProcessor)
├── PostProcessBeforeDestruction(component any, name string) error
└── RequireDestruction(component any) bool

DefinitionRegistryPostProcessor
└── PostProcessDefinitionRegistry(registry DefinitionRegistry, component any, name string) error

ComponentFactoryPostProcessor
└── PostProcessComponentFactory(factory Factory) error
```

## Base Classes

Package: `github.com/go-kid/ioc/container/processors`

### `DefaultComponentPostProcessor`

No-op implementation of `ComponentPostProcessor`. Embed to avoid implementing both methods:

```go
type DefaultComponentPostProcessor struct{}
// Both Before/AfterInitialization return (component, nil)
```

### `DefaultInstantiationAwareComponentPostProcessor`

No-op implementation of `InstantiationAwareComponentPostProcessor`. Embeds `DefaultComponentPostProcessor`:

```go
type DefaultInstantiationAwareComponentPostProcessor struct {
    DefaultComponentPostProcessor
}
// BeforeInstantiation returns (nil, nil) - no proxy
// AfterInstantiation returns (false, nil)
// PostProcessProperties returns (nil, nil) - no modification
```

### `DefaultTagScanDefinitionRegistryPostProcessor`

Generic tag scanner implementing `DefinitionRegistryPostProcessor`. Scans struct fields for a specified tag and generates `Property` entries:

```go
type DefaultTagScanDefinitionRegistryPostProcessor struct {
    definition.LazyInitComponent
    NodeType       component_definition.PropertyType  // property type identifier
    Tag            string                              // struct tag to scan
    ExtractHandler func(meta *Meta, field *Field) (tag, tagVal string, ok bool)
    Required       bool                                // default required for properties
}
```

Fields:
- `NodeType`: arbitrary string identifying the property type (e.g. "component", "configuration", "function")
- `Tag`: the struct tag name to look up (e.g. "wire", "value", "my-custom-tag")
- `ExtractHandler`: optional custom extraction logic when tag-based lookup is insufficient
- `Required`: if true, adds `required` arg to all scanned properties

## Built-in PostProcessors

Registered automatically in `app.App.initiate()`. Execution order controlled by priority/order constants.

| Processor | File | Type | Tag | Purpose |
|-----------|------|------|-----|---------|
| `LoggerAwarePostProcessor` | `logger_aware_post_processors.go` | InstantiationAware + DefinitionRegistry | `logger` | Inject `syslog.Logger` |
| `ConfigQuoteAwarePostProcessors` | `config_quote_aware_post_processors.go` | InstantiationAware | - | Resolve `${...}` placeholders |
| `ExpressionTagAwarePostProcessors` | `expression_tag_aware_post_processors.go` | InstantiationAware | - | Evaluate `#{...}` expressions |
| `PropertiesAwarePostProcessors` | `properties_aware_post_processors.go` | InstantiationAware + DefinitionRegistry | `prefix` | Bind config prefix to struct |
| `ValueAwarePostProcessors` | `value_aware_post_processors.go` | InstantiationAware + DefinitionRegistry | `value`, `prop` | Inject config/literal values |
| `ValidateAwarePostProcessors` | `validate_aware_post_processors.go` | InstantiationAware | - | Validate config properties |
| `DependencyAwarePostProcessors` | `dependency_aware_post_processors.go` | InstantiationAware + DefinitionRegistry + FactoryPostProcessor | `wire` | Resolve component dependencies |
| `DependencyFurtherMatchingProcessors` | `dependency_further_matching_processors.go` | InstantiationAware | - | Apply Primary/Qualifier filtering |
| `DependencyFunctionAwarePostProcessors` | `dependency_function_aware_post_processors.go` | InstantiationAware + DefinitionRegistry + FactoryPostProcessor | `func` | Method-based dependency matching |

## Execution Order Constants

From `container/processors/orders.go`:

**PriorityOrdered** (execute first, in this order):
1. `PriorityOrderLoggerAware` = 2
2. `PriorityOrderPropertyConfigQuoteAware` = 4
3. `PriorityOrderPropertyExpressionTagAware` = 8
4. `PriorityOrderPopulateProperties` = 16

**Ordered** (execute after PriorityOrdered):
1. `OrderDependencyAware` = 2
2. `OrderDependencyFurtherMatching` = 4
3. `OrderValidate` = 8

Custom processors should set `Order()` values relative to these constants.

## Complete Custom Tag Example

Full example: custom `mul` tag that injects a multiplier function.

```go
// 1. Tag scanner: registers "mul" tag as scannable
type mulScanner struct {
    processors.DefaultTagScanDefinitionRegistryPostProcessor
}

// 2. Property processor: handles "mul" properties
type mulProcessor struct {
    processors.DefaultInstantiationAwareComponentPostProcessor
}

func (m *mulProcessor) Order() int { return 100 }

func (m *mulProcessor) PostProcessAfterInstantiation(component any, name string) (bool, error) {
    return true, nil
}

func (m *mulProcessor) PostProcessProperties(
    properties []*component_definition.Property, component any, name string,
) ([]*component_definition.Property, error) {
    for _, p := range properties {
        if p.Tag != "mul" { continue }
        n, _ := strconv.ParseInt(p.TagVal, 10, 64)
        p.Value.Set(reflect.ValueOf(func(i int64) int64 { return n * i }))
        p.SetArg(component_definition.ArgRequired, "false")
    }
    return nil, nil
}

// 3. Usage
type App struct {
    Double func(int64) int64 `mul:"2"`
}

// 4. Registration
ioc.RunTest(t, app.SetComponents(
    &App{},
    &mulProcessor{},
    &mulScanner{processors.DefaultTagScanDefinitionRegistryPostProcessor{
        NodeType: "function",
        Tag:      "mul",
    }},
))
```
