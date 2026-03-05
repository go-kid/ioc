---
name: IoC Testing
description: This skill should be used when the user asks to "write a test", "add unit test", "test component injection", "test configuration", "test lifecycle", "use RunTest", "use RunErrorTest", or needs guidance on writing tests for go-kid/ioc components, configuration binding, or dependency injection scenarios.
---

# IoC Testing

Guide for writing tests in the go-kid/ioc framework using the built-in test helpers.

## Test Helpers

The root `ioc` package provides two test helpers:

```go
// RunTest — bootstraps a full IoC container; asserts no error
func RunTest(t *testing.T, ops ...app.SettingOption) *app.App

// RunErrorTest — bootstraps a container; asserts an error occurred
func RunErrorTest(t *testing.T, ops ...app.SettingOption) *app.App
```

Both create a real `app.App`, run the full lifecycle (config init → factory prepare → refresh → runners), and return the app instance.

## Basic Test Pattern

```go
func TestMyComponent(t *testing.T) {
    // 1. Define inline test struct with wire tags
    tApp := &struct {
        Service *MyService `wire:""`
    }{}

    // 2. Register components and run
    ioc.RunTest(t, app.SetComponents(
        tApp,
        &MyService{},
        &Dependency{},
    ))

    // 3. Assert injected state
    assert.NotNil(t, tApp.Service)
}
```

## Common SettingOptions

```go
app.SetComponents(components...)           // register components
app.SetConfigLoader(loader.NewRawLoader(yamlBytes))  // inline YAML config
app.SetConfigLoader(loader.NewFileLoader("config.yaml"))  // file config
app.AddConfigLoader(loader.NewRawLoader(yamlBytes))  // add additional config
app.LogTrace / app.LogDebug / app.LogError // set log level
app.SkipRunners()                          // skip ApplicationRunner execution
```

## Testing Scenarios

### Dependency Injection by Type

```go
func TestPointerInjection(t *testing.T) {
    tApp := &struct {
        Dep *MyDep `wire:""`
    }{}
    dep := &MyDep{Name: "test"}
    ioc.RunTest(t, app.SetComponents(tApp, dep))
    assert.Equal(t, "test", tApp.Dep.Name)
}
```

### Interface Injection

```go
func TestInterfaceInjection(t *testing.T) {
    tApp := &struct {
        Svc IService `wire:""`
    }{}
    ioc.RunTest(t, app.SetComponents(tApp, &ServiceImpl{}))
    assert.NotNil(t, tApp.Svc)
}
```

### Slice Injection (All Implementations)

```go
func TestSliceInjection(t *testing.T) {
    tApp := &struct {
        All []IService `wire:""`
    }{}
    ioc.RunTest(t, app.SetComponents(tApp, &ImplA{}, &ImplB{}))
    assert.Len(t, tApp.All, 2)
}
```

### Configuration Binding

```go
func TestConfigBinding(t *testing.T) {
    config := []byte(`
db:
  host: localhost
  port: 5432
`)
    tApp := &struct {
        Host string `value:"${db.host}"`
        Port int    `prop:"db.port"`
    }{}
    ioc.RunTest(t,
        app.SetConfigLoader(loader.NewRawLoader(config)),
        app.SetComponents(tApp),
    )
    assert.Equal(t, "localhost", tApp.Host)
    assert.Equal(t, 5432, tApp.Port)
}
```

### Prefix Binding

```go
func TestPrefixBinding(t *testing.T) {
    type DBConfig struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    }
    config := []byte(`db:
  host: localhost
  port: 5432`)
    tApp := &struct {
        DB *DBConfig `prefix:"db"`
    }{}
    ioc.RunTest(t,
        app.SetConfigLoader(loader.NewRawLoader(config)),
        app.SetComponents(tApp),
    )
    assert.Equal(t, "localhost", tApp.DB.Host)
}
```

### Expression Tags

```go
func TestExpression(t *testing.T) {
    tApp := &struct {
        Sum  int  `value:"#{1+(1*2)}"`
        Bool bool `value:"#{1/1==1}"`
    }{}
    ioc.RunTest(t, app.SetComponents(tApp))
    assert.Equal(t, 3, tApp.Sum)
    assert.True(t, tApp.Bool)
}
```

### Constructor Injection

```go
func TestConstructor(t *testing.T) {
    tApp := &struct {
        Svc *MyService `wire:""`
    }{}
    dep := &Repository{}
    ioc.RunTest(t, app.SetComponents(tApp, dep, NewMyService))
    assert.NotNil(t, tApp.Svc)
}
```

### Expected Errors

```go
func TestMissingDependency(t *testing.T) {
    tApp := &struct {
        Dep *NonExistentService `wire:""`
    }{}
    ioc.RunErrorTest(t, app.SetComponents(tApp))
}
```

### Primary / Qualifier Selection

```go
func TestPrimary(t *testing.T) {
    tApp := &struct {
        Svc IService `wire:""`
    }{}
    normal := &NormalImpl{}
    primary := &PrimaryImpl{} // implements WirePrimary
    ioc.RunTest(t, app.SetComponents(tApp, normal, primary))
    // primary is selected
}
```

## Test Organization

Tests are organized under `unittest/` by concern:
- `unittest/component/builtin_inject/` — pointer, interface, slice injection
- `unittest/component/special_inject_condition/` — Primary, Qualifier, optional, self-inject
- `unittest/component/constructor_inject/` — constructor patterns
- `unittest/component/life_cycle_test/` — CloserComponent, etc.
- `unittest/component/post_processor/` — custom post-processors
- `unittest/configure/` — value, prop, prefix, expression, config quote

Test file names use PascalCase describing the scenario (e.g., `Interface_Slice_test.go`).

## Key Rules

- Always use `ioc.RunTest` / `ioc.RunErrorTest` — they handle `testing.Init()` and full container bootstrap
- Define test structs inline with `wire` tags rather than reusing production structs — keeps tests isolated
- Use `app.SetComponents()` to register all components for a test, not `ioc.Register()` (which uses global state)
- Use `loader.NewRawLoader([]byte(...))` for inline YAML config in tests
- Use `app.LogTrace` or `app.LogDebug` for debugging test failures
