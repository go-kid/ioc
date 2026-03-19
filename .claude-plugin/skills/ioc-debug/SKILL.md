---
name: ioc-debug
description: "go-kid/ioc framework dependency injection debugging guide. ALWAYS use this skill when the user encounters ANY errors, failures, or issues with go-kid/ioc, especially: injection errors, component not found, circular dependency, startup panics, wire tag not working, nil dependencies, missing components, constructor parameter resolution failures, config values not injected, or post-processor not applied. Also use for debugging questions like 'why is my component nil', 'injection failed error', 'component X not found', 'how to enable debug mode', 'how to trace dependency resolution', 'my IoC app won't start', 'panic in ioc.Run', 'circular reference error', or ANY go-kid/ioc troubleshooting and error diagnosis. Triggers on: error, failed, panic, not found, nil, not working, missing, debug, debugging, trace, app.LogTrace, RunDebug, injection failed, component not found, circular dependency, dependency resolution failed, wire tag not working, required property empty, constructor parameter not resolved."
---

# IoC Dependency Injection Debugging

Guide for diagnosing and fixing dependency injection issues in the go-kid/ioc framework.

## Diagnostic Approach

### Step 1: Enable Trace Logging

Add `app.LogTrace` to the Run call to see the full container lifecycle:

```go
ioc.Run(app.LogTrace, ...)
// or in tests
ioc.RunTest(t, app.LogTrace, app.SetComponents(...))
```

Trace output shows: component registration, definition scanning, dependency resolution order, and injection details.

### Step 2: Read the Error Chain

The framework provides dependency chain formatting in errors:

```
dependency resolution failed:
  ComponentA
    -> ComponentB
      -> ComponentC (not found or creation failed)
```

This trace shows the full dependency path that led to the failure.

## Common Issues and Solutions

### 1. "component definition with name 'X' not found"

**Cause**: A dependency is declared via `wire` tag but the component is not registered.

**Checklist**:
- Verify the component is registered via `ioc.Register()` or `app.SetComponents()`
- Ensure the component is registered as a **pointer** (`&MyService{}`, not `MyService{}`)
- For constructor injection, verify the constructor function is registered (`ioc.Register(NewService)`)

### 2. Unexported Field Not Injected

**Cause**: Dependency fields must be exported (start with uppercase).

```go
// WRONG - field is unexported, silently ignored
type App struct {
    service *Service `wire:""` // lowercase 's' — won't be injected
}

// CORRECT
type App struct {
    Service *Service `wire:""` // uppercase 'S'
}
```

### 3. Interface Injection Gets Wrong Implementation

**Cause**: Multiple implementations registered; selection order unclear.

**Solutions** (in priority order):
1. Implement `WirePrimary` on the preferred component: `func (s *Preferred) Primary() {}`
2. Use qualifier: `wire:",qualifier=specific-name"` + implement `Qualifier() string`
3. Use wire-by-name: `wire:"component-name"` + implement `Naming() string`

### 4. Circular Dependency Error

**Cause**: Component A depends on B which depends on A (directly or transitively).

The framework supports circular references for singletons via early singleton exposure (three-level cache). If it still fails:

- Check if one component is Prototype scope — prototype components cannot participate in circular reference resolution
- Check if a `SmartInstantiationAwareBeanPostProcessor` is wrapping a component involved in the cycle — the wrapped version might not match
- Look for the log message: "eagerly caching bean 'X' to allow for resolving potential circular references"

### 5. "required property X is empty" / Value Not Injected

**Cause**: Config placeholder `${key}` resolved to empty and the field is required by default.

**Solutions**:
- Add a default value: `value:"${key:defaultValue}"`
- Mark as optional: `value:"${key},required=false"`
- Verify the config loader is set up: `app.SetConfigLoader(loader.NewFileLoader("config.yaml"))`
- Check YAML key spelling and nesting

### 6. Panic at Startup: "missing configure" / "missing registry"

**Cause**: `app.NewApp()` was not used, or critical app fields were overwritten to nil.

**Solution**: Use `ioc.Run()` or `app.NewApp()` — they properly initialize all defaults.

### 7. Constructor Parameter Not Resolved

**Cause**: Constructor function parameter type not found in container.

**Check**:
- Constructor parameters must be pointer types, interface types, or slices thereof
- For `ConfigurationProperties` parameters, ensure the struct implements `Prefix() string` and config is loaded
- Error message includes: "no component found for constructor parameter[N] type X"

### 8. Post-Processor Not Applied

**Cause**: Post-processor registered but not taking effect.

**Check**:
- Post-processor must implement the correct interface (`ComponentPostProcessor`, `InstantiationAwareComponentPostProcessor`, etc.)
- Verify `Order()` return value — it might be running before dependencies are ready
- If post-processor itself needs injection, ensure it doesn't have `LazyInit` marker (lazy components skip the initial Refresh)

## Debug Mode

Run with debug flag to start a web-based debug server:

```go
ioc.RunDebug(...)
// or
ioc.RunDebugWithContext(ctx, ...)
```

The debug server provides a web UI for inspecting component states, factory events, and the dependency graph.

## Key Log Messages to Watch

| Log Message | Meaning |
|------------|---------|
| `"refresh component with name 'X'"` | Component X is being resolved |
| `"eagerly caching bean 'X'"` | Circular reference handling active |
| `"returning eagerly cached instance"` | Using cached early reference |
| `"component 'X' population finished"` | All dependencies injected for X |
| `"skip conditional component 'X'"` | ConditionalComponent returned false |
| `"creating new prototype instance"` | Prototype scope creating new instance |
| `"invoking constructor X for component 'Y'"` | Constructor being called |

## Quick Debugging Template

```go
// Minimal reproduction test
func TestDebugIssue(t *testing.T) {
    tApp := &struct {
        // Put the failing injection here
        Problem *MyService `wire:""`
    }{}
    ioc.RunTest(t,
        app.LogTrace,  // full trace logging
        app.SetComponents(
            tApp,
            // Register all required components
        ),
    )
}
```
