---
name: ioc-extension-testing
description: "go-kid/ioc framework extension and testing guide. Use when creating custom PostProcessors to extend the IoC container, implementing custom tag processing, creating custom config Loaders or Binders, or writing tests with ioc.RunTest/RunErrorTest. Triggers on: PostProcessor, ComponentPostProcessor, InstantiationAwareComponentPostProcessor, DefinitionRegistryPostProcessor, DefaultTagScanDefinitionRegistryPostProcessor, custom tag, ioc.RunTest, ioc.RunErrorTest, testing IoC components."
---

# go-kid/ioc Extension & Testing

## Custom PostProcessors

PostProcessors are the primary extension mechanism. Register them as components via `app.SetComponents`.

For detailed interface hierarchy and built-in processor patterns, see [references/postprocessor-patterns.md](references/postprocessor-patterns.md).

### Basic: `ComponentPostProcessor`

Intercept component initialization:

```go
type MyProcessor struct{}

func (p *MyProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
    if svc, ok := component.(*MyService); ok {
        svc.Instrumented = true
    }
    return component, nil
}

func (p *MyProcessor) PostProcessAfterInitialization(component any, componentName string) (any, error) {
    return component, nil
}
```

### Advanced: `InstantiationAwareComponentPostProcessor`

Intercept property injection. Embed `DefaultInstantiationAwareComponentPostProcessor` and override only what you need:

```go
import (
    "github.com/go-kid/ioc/component_definition"
    "github.com/go-kid/ioc/container/processors"
)

type MyInjector struct {
    processors.DefaultInstantiationAwareComponentPostProcessor
}

func (m *MyInjector) Order() int { return 100 }

func (m *MyInjector) PostProcessAfterInstantiation(component any, name string) (bool, error) {
    return true, nil  // true = continue processing properties
}

func (m *MyInjector) PostProcessProperties(
    properties []*component_definition.Property,
    component any, name string,
) ([]*component_definition.Property, error) {
    for _, p := range properties {
        if p.Tag == "my-tag" {
            // set value via reflection: p.Value.Set(reflect.ValueOf(...))
            p.SetArg(component_definition.ArgRequired, "false")
        }
    }
    return nil, nil
}
```

### Custom Tag Scanning: `DefinitionRegistryPostProcessor`

Use `DefaultTagScanDefinitionRegistryPostProcessor` to register a custom struct tag:

```go
type myTagScanner struct {
    processors.DefaultTagScanDefinitionRegistryPostProcessor
}

// Register:
app.SetComponents(
    &myTagScanner{
        processors.DefaultTagScanDefinitionRegistryPostProcessor{
            NodeType: "custom",    // property type identifier
            Tag:      "my-tag",    // struct tag to scan
        },
    },
    &MyInjector{},  // pair with a processor that handles "my-tag"
)
```

### Factory Access: `ComponentFactoryPostProcessor`

```go
func (p *MyProcessor) PostProcessComponentFactory(factory container.Factory) error {
    registry := factory.GetDefinitionRegistry()
    configure := factory.GetConfigure()
    return nil
}
```

## Custom Loader

Implement `configure.Loader` to add a new config source:

```go
type EnvLoader struct{}

func (l *EnvLoader) LoadConfig() ([]byte, error) {
    // return YAML/JSON bytes from any source
    return yaml.Marshal(map[string]string{"env": os.Getenv("APP_ENV")})
}

// Optionally implement definition.Ordered for load priority
func (l *EnvLoader) Order() int { return 10 }
```

Register: `app.AddConfigLoader(&EnvLoader{})`

## Testing

Import: `github.com/go-kid/ioc`

### `ioc.RunTest`

Start the IoC container and assert no error:

```go
func TestMyComponent(t *testing.T) {
    comp := &MyComponent{}
    ioc.RunTest(t, app.SetComponents(comp))
    assert.Equal(t, "expected", comp.Value)
}
```

### `ioc.RunErrorTest`

Expect the container to return an error:

```go
func TestMissingDependency(t *testing.T) {
    comp := &struct {
        Svc *NonExistent `wire:""`
    }{}
    ioc.RunErrorTest(t, app.SetComponents(comp))
}
```

### Common Test Patterns

```go
// With config
ioc.RunTest(t,
    app.SetComponents(comp),
    app.SetConfigLoader(loader.NewRawLoader([]byte(`key: value`))),
)

// With log level
ioc.RunTest(t, app.LogTrace, app.SetComponents(comp))

// Test lifecycle: Close
app := ioc.RunTest(t, app.SetComponents(comp))
app.Close()
assert.True(t, comp.Closed)

// Access config after startup
app := ioc.RunTest(t, app.SetConfigLoader(loader.NewRawLoader(cfg)))
val := app.Get("some.key")
```

### Test Organization

Follow existing project conventions:

```
unittest/
├── component/
│   ├── builtin_inject/          # wire tag: pointer, interface, slice, name
│   ├── embed_inject/            # embedded struct injection
│   ├── func_inject/             # func tag injection
│   ├── life_cycle_test/         # Init, Run, Close lifecycle
│   ├── modified_inject/         # custom PostProcessor injection
│   ├── post_processor/          # PostProcessor behavior
│   └── special_inject_condition/ # qualifier, primary, required=false
└── configure/
    ├── configuration_test.go     # prefix tag, Binder
    ├── config_quote_test.go      # ${...} placeholders
    ├── expression_tag_test.go    # #{...} expressions
    ├── value_tag_test.go         # value tag types
    └── validate_test.go          # validation
```
