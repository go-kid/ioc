# go-kid/ioc

[English](README.md) | **中文**

`go-kid/ioc` 是一款基于 **tag + interface** 的 Go 运行时依赖注入（IoC/DI）框架。

## ✨ 特性

- **组件依赖注入**（`wire` 标签）
  - 按指针类型、接口类型、组件名称注入
  - 支持多实现选择策略（Primary/Qualifier）
  - 支持注入到接口切片/数组
- **配置注入**（`value` / `prop` / `prefix` 标签）
  - 多配置源支持（命令行、文件、原始内容）
  - 配置占位符 `${...}`
  - 表达式计算 `#{...}`（算术、逻辑、条件、集合操作）
- **构造器注入**：支持函数式依赖注入（内置）
- **生命周期管理**：ApplicationRunner、CloserComponent、LazyInitComponent
- **`context.Context` 生命周期支持**：所有生命周期接口的 WithContext 变体
- **作用域机制**：Singleton/Prototype
- **条件注册**：根据运行时条件注册组件
- **应用事件机制**：发布与监听应用事件
- **泛型类型安全注册**：`ioc.Provide[T]` 在注册时验证返回类型
- **`log/slog` 适配器**：与 Go 结构化日志集成

## 📦 安装

要求 Go 1.21+

```bash
go get github.com/go-kid/ioc
```

## 🚀 快速入门

```go
package main

import (
	"fmt"
	"github.com/go-kid/ioc"
)

// 被注入的组件
type ComponentA struct {
	Name string
}

func (a *ComponentA) GetName() string {
	return a.Name
}

// 注入目标组件
type App struct {
	ComponentA *ComponentA `wire:""` // 指针 + 导出字段 + wire tag
}

func main() {
	a := new(App)

	// 注册组件
	ioc.Register(a)
	ioc.Register(&ComponentA{Name: "Comp-A"})

	// 运行框架
	_, err := ioc.Run()
	if err != nil {
		panic(err)
	}

	// 输出: "Comp-A"
	fmt.Println(a.ComponentA.GetName())
}
```

> **关键点**：依赖字段必须可导出（首字母大写），使用 `wire:""` 声明需要注入。

## 📖 文档

### 1. 组件依赖注入（`wire` 标签）

组件注入统一使用 tag：`wire:"<name>,arg1=xxx,arg2=yyy"`

- `<name>`：可选，指定按名称注入
- `,arg=...`：可选，附加参数（如 `qualifier`）
- 留空 `wire:""` 时表示按类型自动匹配

#### 按指针类型注入

```go
type Component struct {
	Name string
}

type App struct {
	C *Component `wire:""` // 按 *Component 类型注入
}
```

#### 按接口类型注入

```go
type IComponent interface {
	GetName() string
}

type ComponentA struct{ Name string }
func (a *ComponentA) GetName() string { return a.Name }

type App struct {
	C IComponent `wire:""` // 按接口类型注入
}
```

#### 按名称注入

实现 `NamingComponent` 接口自定义组件名：

```go
type ComponentA struct {
	Name          string
	componentName string
}

func (a *ComponentA) Naming() string { return a.componentName }

type App struct {
	ByName IComponent `wire:"comp-A"` // 按组件名注入
}
```

#### 多实现选择策略

当接口有多个实现时，容器按以下优先级选择：

1. 实现 `WirePrimary` 接口的组件（主实现）
2. 无别名的组件（未实现 `Naming()`）
3. 任选其一（不保证顺序）

**使用 Primary 标记：**

```go
type ComponentA struct {
	Name string
	definition.WirePrimaryComponent // 嵌入 Primary 组件
}
```

**使用 Qualifier 限定：**

```go
type ComponentA struct {
	qualifier string
}

func (a *ComponentA) Qualifier() string { return a.qualifier }

type App struct {
	C IComponent `wire:",qualifier=comp-A"` // 使用限定符
}
```

#### 注入到切片/数组

```go
type App struct {
	All []Interface `wire:""` // 所有 Interface 实现
}
```

### 2. 配置注入

#### `prefix`：绑定配置前缀

```go
type configA struct {
	B int   `yaml:"b"`
	C []int `yaml:"c"`
}

type App struct {
	A *configA `prefix:"a"` // 绑定配置前缀 "a"
}
```

配置（YAML）：
```yaml
a:
  b: 123
  c: [1,2,3,4]
```

#### `value`：字面值/占位符/表达式

```go
type T struct {
	DSN string `value:"${db.dsn}"`           // 配置占位符
	Sum int    `value:"#{1+2}"`              // 表达式
	Str string `value:"foo"`                 // 字面值
}
```

#### `prop`：`value` 的语法糖

```go
type DBConfig struct {
	DSN string `prop:"db.dsn"` // 等价于 value:"${db.dsn}"
}
```

#### 配置占位符和表达式

- **`${...}`**：配置占位符，从配置中读取值
- **`#{...}`**：表达式计算，支持算术、逻辑、条件、集合操作

示例：

```go
type T struct {
	Arithmetic  int    `value:"#{1+(1*2)}"`                      // 3
	Comparison  bool   `value:"#{1/1==1}"`                       // true
	Conditional string `value:"#{1>2?'a':'b'}"`                  // "b"
	WithConfig  int    `value:"#{${number.val1}+${number.val2}}"` // 使用配置
}
```

### 3. 构造器注入

构造器注入已内置，无需额外配置：

```go
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

ioc.Register(NewService)      // 直接注册构造函数
ioc.Register(&Repository{})
_, err := ioc.Run()
```

泛型变体（注册时验证返回类型）：

```go
ioc.Provide[Service](NewService)  // 注册时验证返回类型
```

### 4. 应用启动

**方式一：使用 `ioc.Run`/`ioc.Register`**

```go
ioc.Register(a)
ioc.Register(&ComponentA{Name: "Comp-A"})
_, err := ioc.Run()
```

**方式二：直接使用 `app.App`**

```go
application := app.NewApp()
err := application.Run(
	app.SetComponents(...),
	app.SetConfigLoader(loader.NewFileLoader("config.yaml")),
	app.SetConfigBinder(binder.NewViperBinder("yaml")),
)
```

### 5. 生命周期接口

- `ApplicationRunner`：组件刷新完毕后执行
- `CloserComponent`：应用关闭时调用 `Close()`
- `LazyInitComponent`：标记为懒加载
- `PriorityComponent`：控制后置处理器执行顺序
- `InitializeComponentWithContext` / `InitializingComponentWithContext`：带 Context 的初始化
- `ApplicationRunnerWithContext`：带 Context 的应用启动
- `CloserComponentWithContext`：带 Context 的关闭
- `ScopeComponent`：控制组件作用域（Singleton/Prototype）
- `ConditionalComponent`：条件注册
- `ApplicationEventListener` / `ApplicationEventPublisher`：事件机制

### 6. Context 支持

所有生命周期接口均提供 WithContext 变体，便于在初始化、启动、关闭等阶段使用 `context.Context`（超时控制、取消传播等）：

```go
// 带 Context 启动框架
app, err := ioc.RunWithContext(ctx)

// 设置关闭超时
_, err := ioc.Run(app.SetShutdownTimeout(30 * time.Second))
```

Context 感知的生命周期接口：

```go
// InitializeComponentWithContext（替代 InitializeComponent）
func (c *MyComp) Init(ctx context.Context) error { return nil }

// InitializingComponentWithContext（替代 InitializingComponent）
func (c *MyComp) AfterPropertiesSet(ctx context.Context) error { return nil }

// ApplicationRunnerWithContext（可与 ApplicationRunner 共存）
func (r *MyRunner) RunWithContext(ctx context.Context) error { return nil }

// CloserComponentWithContext（可与 CloserComponent 共存）
func (c *MyComp) CloseWithContext(ctx context.Context) error { return nil }
```

> **注意**：`Init(ctx)` 和 `AfterPropertiesSet(ctx)` 与基础接口方法名相同但签名不同，组件只能实现其中之一。`RunWithContext` 和 `CloseWithContext` 使用不同的方法名，允许组件同时实现原始和 Context 感知版本。

### 7. 作用域

通过实现 `ScopeComponent` 接口控制组件作用域：

- **Singleton**（默认）：容器内单例，每次获取同一实例
- **Prototype**：每次获取新建实例

```go
import "github.com/go-kid/ioc/definition"

type MyPrototype struct{}

func (p *MyPrototype) Scope() string { return definition.ScopePrototype }
```

### 8. 条件注册

实现 `ConditionalComponent` 接口，根据运行时条件决定是否注册组件：

```go
import "github.com/go-kid/ioc/definition"

type MyComp struct{}

func (c *MyComp) Condition(ctx definition.ConditionContext) bool {
	return ctx.HasComponent("dependency")
}
```

### 9. 事件机制

发布和监听应用事件，实现组件间解耦通信：

```go
import "github.com/go-kid/ioc/definition"

// 监听事件
type MyListener struct{}

func (l *MyListener) OnEvent(event definition.ApplicationEvent) error {
	// 处理事件
	return nil
}
```

内置事件：`ComponentCreatedEvent`、`ApplicationStartedEvent`、`ApplicationClosingEvent`。

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                          app.App                                 │
│  ┌──────────────────┐              ┌──────────────────┐         │
│  │ Component Factory│              │    Configure    │         │
│  └────────┬─────────┘              └────────┬─────────┘         │
│           │                                  │                   │
│           │                                  │                   │
│  ┌────────▼─────────┐              ┌────────▼─────────┐       │
│  │ DefinitionRegistry│              │      Loader       │       │
│  │ SingletonRegistry│              │  (args/file/raw)   │       │
│  └────────┬─────────┘              └────────┬─────────┘       │
│           │                                  │                   │
│           │ register                         │                   │
│  ┌────────▼─────────┐              ┌────────▼─────────┐       │
│  │ComponentDefinition│              │      Binder       │       │
│  └────────┬─────────┘              │      (viper)       │       │
│           │                         └───────────────────┘       │
│           │ create                                               │
│  ┌────────▼─────────┐                                           │
│  │     Instance     │                                           │
│  └────────┬─────────┘                                           │
│           │                                                      │
│  ┌────────▼──────────────────────────────────────────┐          │
│  │            PostProcessors                          │          │
│  │  ┌────────────────────────────────────────────┐   │          │
│  │  │ • LoggerAware                               │   │          │
│  │  │ • ConfigQuoteAware ${...}                   │   │          │
│  │  │ • ExpressionTagAware #{...}                 │   │          │
│  │  │ • PropertiesAware prefix                   │   │          │
│  │  │ • ValueAware value/prop                    │   │          │
│  │  │ • DependencyAware wire                      │   │          │
│  │  │ • ConstructorAware func                     │   │          │
│  │  └────────────────────────────────────────────┘   │          │
│  └──────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘

配置流向: Binder → ConfigQuoteAware, PropertiesAware, ValueAware, ExpressionTagAware
```

## 📚 示例与测试

- **示例工程**：`test/ioc`、`test/constructor`、`test/prop`、`test/t_yaml`
- **PostProcessor 示例**：`examples/post_processor/`
- **性能分析工具**：`cmd/performance_analyst/`
- **单元测试**：`unittest/component/builtin_inject/*`、`unittest/configure/*`

运行测试：

```bash
go test ./...
```

## 📄 许可证

[MIT License](./LICENSE)

## 🤝 贡献

欢迎通过 Issue / PR 进行反馈与贡献。提交前请确保 `go test ./...` 通过。
