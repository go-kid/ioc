---
name: ioc-test
description: "go-kid/ioc framework extension and testing guide. ALWAYS use this skill when the user asks about: extending IoC container, creating PostProcessors, custom tag processing, custom config loaders/binders, writing tests, testing IoC components, ioc.RunTest/RunErrorTest, slog integration, AOP/proxies, or advanced framework customization. Also use for questions like 'how to write tests for IoC', 'how to create custom tag', 'how to extend the container', 'how to add my own PostProcessor', 'how to test dependency injection', 'how to use slog with ioc', 'how to implement AOP', 'how to intercept component creation', 'how to customize property injection', or any go-kid/ioc testing, extension, and advanced customization questions. Triggers on: test, testing, ioc.RunTest, ioc.RunErrorTest, PostProcessor, ComponentPostProcessor, custom tag, extension, extend, customize, slog, logging, AOP, proxy, InstantiationAwareComponentPostProcessor, DefinitionRegistryPostProcessor, custom loader, custom binder."
---

# go-kid/ioc Extension & Testing

Requires **Go 1.21+**.

## Custom PostProcessors

PostProcessors are the primary extension mechanism. They can be registered in two ways:

1. Via `app.SetComponents` (explicit)
2. Via `ioc.Register` (any registered component implementing a PostProcessor interface is auto-discovered)

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

### Auto-Discovery

Any component registered via `ioc.Register` that implements `ComponentPostProcessor` or `DefinitionRegistryPostProcessor` is automatically picked up during `PrepareComponents()`. No special registration is needed beyond the standard `ioc.Register`.

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

## Logging

### `log/slog` Adapter

The framework provides an adapter to bridge `syslog.Logger` with Go 1.21's `log/slog`:

```go
import "github.com/go-kid/ioc/syslog"

adapter := syslog.NewSlogAdapter(slogHandler)
ioc.Run(app.SetLogger(adapter))
```

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
a := ioc.RunTest(t, app.SetComponents(comp))
a.Close()
assert.True(t, comp.Closed)

// Access config after startup
a := ioc.RunTest(t, app.SetConfigLoader(loader.NewRawLoader(cfg)))
val := a.Get("some.key")

// Test with context
ctx := context.Background()
a, err := ioc.RunWithContext(ctx, app.SetComponents(comp))
assert.NoError(t, err)

// Test constructor injection
ioc.RunTest(t, app.SetComponents(NewService, &Repository{}))

// Test type-safe registration
ioc.Provide[Service](NewService)
ioc.RunTest(t, app.SetComponents(&Repository{}))
```

### Test Organization

Follow existing project conventions:

```
unittest/
├── component/
│   ├── builtin_inject/          # wire tag: pointer, interface, slice, name
│   ├── constructor_inject/      # constructor injection
│   ├── embed_inject/            # embedded struct injection
│   ├── func_inject/             # func tag injection
│   ├── life_cycle_test/         # Init, Run, Close lifecycle
│   ├── modified_inject/         # custom PostProcessor injection
│   ├── post_processor/          # PostProcessor behavior
│   ├── refactor_test/           # refactored features and backward compatibility
│   └── special_inject_condition/ # qualifier, primary, required=false
└── configure/
    ├── configuration_test.go     # prefix tag, Binder
    ├── config_quote_test.go      # ${...} placeholders
    ├── expression_tag_test.go    # #{...} expressions
    ├── value_tag_test.go         # value tag types
    └── validate_test.go          # validation
```
