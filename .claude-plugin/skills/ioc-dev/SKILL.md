---
name: ioc-dev
description: "go-kid/ioc framework development guide covering component dependency injection, configuration binding, and application lifecycle. ALWAYS use this skill when the user asks about: dependency injection, wiring components, component registration, wire tag, constructor injection, config injection, value/prop/prefix tags, placeholders ${...}, expressions #{...}, application startup, ioc.Run, app.NewApp, lifecycle methods (Init/AfterPropertiesSet/Run/Close), ApplicationRunner, graceful shutdown, or ANY go-kid/ioc development questions. Also use for questions like 'how do I inject X', 'how to load config', 'how to start my app', 'what's the startup sequence', 'how to choose between implementations', 'config value not working', or any component wiring, configuration binding, and lifecycle management. Triggers on: dependency injection, DI, IoC, wire tag, component registration, ioc.Register, ioc.Provide, app.SetComponents, constructor injection, value tag, prop tag, prefix tag, config, ${...}, #{...}, ConfigurationProperties, ioc.Run, ioc.RunWithContext, app.NewApp, ApplicationRunner, lifecycle, Init, AfterPropertiesSet, Close, startup, shutdown, Qualifier, Primary, NamingComponent, ScopeComponent, ConditionalComponent, LazyInit, Ordered, events."
---

# go-kid/ioc Development Guide

Comprehensive guide for developing applications with the **go-kid/ioc** dependency injection framework.

**Requires Go 1.21+**

---

## ⚠️ 核心约束（必读）

在使用 go-kid/ioc 时，**必须遵守**以下规则，违反会导致注入失败或运行时错误：

### 1. 依赖注入字段必须导出（首字母大写）

```go
// ❌ 错误 - 小写字段会被忽略
type Service struct {
    repo *Repository `wire:""`  // 不会被注入！
}

// ✅ 正确 - 首字母大写
type Service struct {
    Repo *Repository `wire:""`  // 会被正确注入
}
```

### 2. 配置注入必须使用专用机制

**❌ 禁止直接访问 configure 对象：**
```go
// ❌ 错误 - 不要直接注入或访问 configure
type Service struct {
    Configure configure.Configure `wire:""` // 错误！
}

func (s *Service) Init() {
    // ❌ 错误 - 不要直接调用 configure.Get()
    host := s.Configure.Get("db.host")
}
```

**✅ 必须使用以下方式之一：**
```go
// 方式 1: 使用 prop/value/prefix tag
type Service struct {
    Host string    `prop:"db.host"`           // ✅ 单个配置值
    Port int       `value:"${db.port:3306}"`  // ✅ 带默认值
    DB   *DBConfig `prefix:"database"`        // ✅ 绑定配置树
}

// 方式 2: 实现 ConfigurationProperties 接口
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
func (c *DBConfig) Prefix() string { return "database" }

// 方式 3: 构造函数参数（自动绑定）
func NewService(cfg *DBConfig) *Service {  // cfg 会自动从配置加载
    return &Service{host: cfg.Host}
}
```

### 3. 组件注册必须是指针

```go
// ❌ 错误
ioc.Register(MyService{})

// ✅ 正确
ioc.Register(&MyService{})
```

---

## Quick Start

### 1. Register Components & Start App

```go
import "github.com/go-kid/ioc"

// Register components
ioc.Register(&MyService{}, &Repository{})

// Start application
app, err := ioc.Run()
if err != nil {
    panic(err)
}
defer app.Close()
```

### 2. Inject Dependencies

```go
type MyService struct {
    Repo   *Repository `wire:""`           // inject by type
    Logger Logger      `wire:""`           // inject interface
    Config *AppConfig  `prefix:"app"`      // bind config
    Port   int         `prop:"server.port:8080"` // config value with default
}
```

### 3. Constructor Injection

```go
func NewService(repo *Repository, cfg *DBConfig) *Service {
    return &Service{repo: repo, cfg: cfg}
}

ioc.Register(NewService, &Repository{})
ioc.Run(app.SetConfig("config.yaml"))
```

---

## Topic Guide

This skill covers three core areas. For detailed information, consult the reference documents:

### 📦 Component Injection
**When to use:** Registering components, dependency wiring, selecting implementations

**Quick reference:**
- **Register**: `ioc.Register(&MyComp{})` or `ioc.Provide[T](constructor)`
- **Inject by type**: `Svc *Service \`wire:""\``
- **Inject by name**: `Svc *Service \`wire:"my-service"\``
- **Inject all**: `Svcs []Service \`wire:""\``
- **Inject by method**: `Comp *MyComp \`func:"MethodName"\``
- **Primary selection**: Embed `definition.WirePrimaryComponent`
- **Qualifier**: `wire:",qualifier=groupA"`
- **Optional**: `wire:",required=false"`
- **Scope**: Implement `ScopeComponent` → `ScopeSingleton` / `ScopePrototype`

**📖 Full documentation:** [references/component-injection.md](references/component-injection.md)

### ⚙️ Configuration Injection
**When to use:** Loading config files, binding config values, placeholders, expressions

**Quick reference:**
- **Load config**: `app.SetConfig("config.yaml")`
- **Bind struct**: `DB *DBConfig \`prefix:"database"\``
- **Config value**: `Host string \`prop:"server.host"\``
- **Placeholder**: `DSN string \`value:"${db.dsn}"\``
- **Default value**: `Port int \`value:"${port:8080}"\``
- **Expression**: `Sum int \`value:"#{1+2}"\``
- **Mixed**: `Total int \`value:"#{${price} * ${qty}}"\``
- **ConfigurationProperties**: Implement `Prefix() string`

**📖 Full documentation:** [references/config-injection.md](references/config-injection.md)

### 🔄 Application Lifecycle
**When to use:** App startup, initialization order, lifecycle hooks, shutdown, events

**Quick reference:**
- **Start app**: `ioc.Run(options...)`
- **With context**: `ioc.RunWithContext(ctx, options...)`
- **After properties set**: Implement `AfterPropertiesSet() error`
- **Init hook**: Implement `Init() error`
- **Run after startup**: Implement `ApplicationRunner` → `Run() error`
- **Graceful shutdown**: Implement `Close() error`
- **Execution order**: Implement `Order() int`
- **Lazy init**: Embed `definition.LazyInitComponent`
- **Conditional**: Implement `Condition(ctx ConditionContext) bool`
- **Events**: Implement `ApplicationEventListener`

**📖 Full documentation:** [references/lifecycle.md](references/lifecycle.md)

---

## Common Patterns

### Pattern 1: Service with Config & Dependencies

```go
type UserService struct {
    Repo   *UserRepository `wire:""`
    Cache  *Redis          `wire:""`
    Config *ServiceConfig  `prefix:"user-service"`
}

func (s *UserService) Init() error {
    // Initialize after dependencies injected
    return s.Cache.Connect(s.Config.RedisURL)
}
```

### Pattern 2: Constructor with Type-Safe Registration

```go
func NewUserService(repo *UserRepository, cfg *ServiceConfig) *UserService {
    return &UserService{repo: repo, config: cfg}
}

// Type-safe: panics if return type doesn't match
ioc.Provide[UserService](NewUserService)
```

### Pattern 3: Multiple Implementations with Qualifier

```go
// Mark implementations
type MySQLRepo struct{}
func (r *MySQLRepo) Qualifier() string { return "mysql" }

type PostgresRepo struct{}
func (r *PostgresRepo) Qualifier() string { return "postgres" }

// Select by qualifier
type App struct {
    MySQL    Repository   `wire:",qualifier=mysql"`
    Postgres Repository   `wire:",qualifier=postgres"`
    AllRepos []Repository `wire:""` // gets both
}
```

### Pattern 4: Graceful Startup & Shutdown

```go
type HTTPServer struct {
    Port   int    `prop:"server.port:8080"`
    server *http.Server
}

func (s *HTTPServer) Run() error {
    s.server = &http.Server{Addr: fmt.Sprintf(":%d", s.Port)}
    return s.server.ListenAndServe()
}

func (s *HTTPServer) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return s.server.Shutdown(ctx)
}
```

---

## Startup Flow Overview

```
ioc.Run(options...)
  1. Apply SettingOptions + global settings
  2. Load configuration (loaders → binder)
  3. Scan component definitions
  4. Resolve dependencies & create components
  5. Inject dependencies (wire, value, prop, prefix tags)
  6. Call lifecycle hooks (AfterPropertiesSet → Init)
  7. Execute ApplicationRunners (in Order)
  8. Publish ApplicationStartedEvent
```

---

## Important Gotchas

**参见文档开头的 [⚠️ 核心约束](#️-核心约束必读) 了解必须遵守的规则**

1. **Components must be pointers**: `ioc.Register(&MyService{})` not `MyService{}`
2. **Exported fields only**: `wire:""` tags on unexported fields are ignored
3. **Config injection only via tags**: Never inject `configure.Configure` directly; use `value`/`prop`/`prefix` tags or `ConfigurationProperties`
4. **Required by default**: Use `required=false` to make injection optional
5. **Circular refs**: Supported for singletons via early exposure, NOT for prototypes
6. **Order matters**: PostProcessors run in defined order (see lifecycle docs)
7. **`prop` vs `value`**: `prop:"key"` equals `value:"${key}"`

---

## Configuration

When you need detailed information on a specific topic, read the corresponding reference file. The main topics are:

- **Component wiring & injection** → [references/component-injection.md](references/component-injection.md)
- **Config loading & binding** → [references/config-injection.md](references/config-injection.md)
- **App lifecycle & startup** → [references/lifecycle.md](references/lifecycle.md)

For debugging and troubleshooting, use the **ioc-debug** skill.
For testing and extensions, use the **ioc-test** skill.
