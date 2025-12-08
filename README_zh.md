# go-kid/ioc

`go-kid/ioc` 是一款基于 **tag + interface** 的 Go 运行时依赖注入（IoC/DI）框架。

- 通过 `wire` 标签实现 **组件依赖注入**：
  - 按指针类型注入
  - 按接口类型注入
  - 按组件名称注入
  - 接口的多个实现注入到接口数组 / 切片
- 通过 `value` / `prop` / `prefix` 等标签实现 **配置注入**：
  - 从多种配置源加载（命令行参数、文件、原始内容等）
  - 支持 `${...}` 配置占位符
  - 支持 `#{...}` 表达式计算（算术、逻辑、条件、集合操作等）

---

## 安装

```bash
go get github.com/go-kid/ioc
```

---

## 快速入门

下面以一个最简单的依赖注入示例开始：

```go
package main

import (
	"fmt"

	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
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

	// 这里应打印 "Comp-A"
	fmt.Println(a.ComponentA.GetName())
}
```

> 关键点：
> - 依赖字段必须可导出（首字母大写）且可设置；
> - 使用 `wire:""` 声明需要注入。

---

## 1. 组件依赖注入（`wire` 标签）

组件注入统一使用 tag：

```go
`wire:"<name>,arg1=xxx,arg2=yyy"`
```

其中：

- `<name>`：可选，指定按名称注入；
- `,arg=...`：可选，附加参数（如 `qualifier`）；
- 留空 `wire:""` 时表示按类型自动匹配。

### 1.1 按指针类型注入

```go
type Component struct {
	Name string
}

type App struct {
	C *Component `wire:""` // 按 *Component 类型注入
}
```

当容器中存在一个或多个 `*Component` 实例时：

- 若只有一个，则直接注入；
- 若有多个，则会进入“多实现选择规则”（见后文）。

### 1.2 按接口类型注入

更推荐用接口抽象依赖：

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

容器会查找所有实现 `IComponent` 的组件，并按策略选择一个（见 1.4）。

### 1.3 按名称注入

`go-kid/ioc` 内建 `NamingComponent` 接口，用于自定义组件名：

```go
type NamingComponent interface {
	Naming() string
}
```

实现该接口后，`Naming()` 返回值会作为组件**别名**（alias）。  
注入时可显式按名称：

```go
type ComponentA struct {
	Name          string
	componentName string
}

func (a *ComponentA) GetName() string { return a.Name }
func (a *ComponentA) Naming() string  { return a.componentName }

type IComponent interface {
	GetName() string
}

type App struct {
	ByName IComponent `wire:"comp-A"` // 按组件名注入
}

func main() {
	a := new(App)
	ioc.Register(a)

	ioc.Register(&ComponentA{
		Name:          "Comp-A",
		componentName: "comp-A",
	})

	_, _ = ioc.Run()
	fmt.Println(a.ByName.GetName()) // "Comp-A"
}
```

若 `wire:"comp-B"`，而容器无该名称组件，则在严格模式下直接报错。

### 1.4 同一接口多实现、默认选择策略

当某接口有多个实现，而字段只声明为该接口（或指针），且 `wire` 留空时：

```go
type IComponent interface{ GetName() string }

type ComponentA struct{ Name string }
func (a *ComponentA) GetName() string { return a.Name }

type ComponentB struct{ Name string }
func (b *ComponentB) GetName() string { return b.Name }

type App struct {
	C IComponent `wire:""` // 多实现，按策略选择
}
```

容器会：

1. 找出所有实现 `IComponent` 的组件；
2. 如实现数 > 1：
   - 若有实现了 `WirePrimary` 接口的组件，则优先该组件；
   - 否则优先**无别名**组件（不是通过 `Naming()` 自定义名称的）；
   - 若最终仍有多个候选，则“任选其一”（不保证顺序）。

#### 优先注入：`WirePrimary` 接口

定义：

```go
type WirePrimaryComponent struct{}
func (i *WirePrimaryComponent) Primary() {}
```

任一组件只要实现了 `Primary()` 方法，即可被视为“主实现”（primary）。例如：

```go
type ComponentA struct {
	Name string
	definition.WirePrimaryComponent
}

type ComponentB struct{ Name string }

type App struct {
	C IComponent `wire:""`
}
```

此时 `C` 会始终注入 `ComponentA`。

> 注意：
> - `WirePrimary` 仅在**接口/指针注入**且未指定 `wire:"name"` 时生效；
> - 不建议同一接口多个实现都标记为 Primary，否则仍可能存在不确定性。

#### 限定注入：`WireQualifier` 接口 + `qualifier` 参数

当一个接口有多个实现，希望**精确标识**注入对象时，可使用限定符：

```go
type WireQualifier interface {
	Qualifier() string
}
```

实现示例：

```go
type ComponentA struct {
	Name          string
	componentName string
	qualifier     string
}

func (a *ComponentA) Qualifier() string { return a.qualifier }
```

在 `wire` 中使用：

```go
type App struct {
	C IComponent `wire:",qualifier=comp-A"`
}

func main() {
	a := new(App)
	ioc.Register(a)

	ioc.Register(&ComponentA{
		Name:          "Comp-A",
		componentName: "comp-A",
		qualifier:     "comp-A",
	})
	ioc.Register(&ComponentA{
		Name:          "Comp-B",
		componentName: "comp-B",
		qualifier:     "comp-B",
	})

	_, _ = ioc.Run()
	fmt.Println(a.C.GetName()) // 始终为 "Comp-A"
}
```

> `wire:"name,qualifier=xxx"` 中：
> - `name` 部分用于按名称查找；
> - `qualifier` 参数用于在多候选中再过滤；
> - 当同时存在 `name` 和 `qualifier` 时，两者都生效。

### 1.5 注入多实现到接口数组 / 切片

当字段声明为接口切片 / 数组，并使用 `wire:""` 时，容器会将**所有匹配实现**注入进去：

```go
type Interface interface {
	SimpleInterface()
}

type InterfaceImplComponent struct{ Component }
func (i *InterfaceImplComponent) SimpleInterface() {}

type InterfaceImplNamingComponent struct{ SpecifyNameComponent }
func (i *InterfaceImplNamingComponent) SimpleInterface() {}

type App struct {
	All []Interface `wire:""` // 所有 Interface 实现
}
```

对应单测（节选）：

```go
var tApp = &struct {
	Ts []Interface `wire:""`
}{}

ioc.RunTest(t, app.SetComponents(
	&InterfaceImplNamingComponent{SpecifyNameComponent{Component{"InterfaceImplNamingComponent"}}},
	&InterfaceImplComponent{Component{"InterfaceImplComponent"}},
	&SpecifyNameComponent{Component{"SpecifyNameComponent"}},
	&Component{"Component"},
	tApp,
))
assert.Equal(t, 2, len(tApp.Ts)) // 仅两个真正实现 Interface 的组件被注入
```

同理：

```go
type App struct {
	Ptrs []*SpecifyNameComponent `wire:""` // 所有 *SpecifyNameComponent 实例
}
```

---

## 2. 配置注入（`value` / `prop` / `prefix`）

`go-kid/ioc` 的配置系统由三部分组成：

- `configure.Loader`：从不同来源加载原始配置（YAML/JSON/命令行等）；
- `configure.Binder`：将原始配置解析为统一访问接口，默认使用 `viper`；
- 注入标签：
  - `value`：从**字面值 / 表达式 / 配置占位符**转换为字段值；
  - `prop`：是 `value` 的语法糖；
  - `prefix`：将一个配置前缀整个绑定为结构体。

### 2.1 基础配置加载

默认 `App` 使用：

```go
configure.Default():
- Loader:  loader.NewArgsLoader(os.Args)
- Binder:  binder.NewViperBinder("yaml")
```

你也可以在 `Run` 时自定义：

```go
ioc.Run(
    app.SetConfigLoader(
        loader.NewFileLoader("config.yaml"), // 从文件加载
    ),
    app.SetConfigBinder(
        binder.NewViperBinder("yaml"),       // 或 "json"
    ),
)
```

或者直接使用原始字节（单测中常用）：

```go
cfg := []byte(`
a:
  b: 123
  c: [1,2,3]
`)
ioc.Run(
    app.SetConfigLoader(loader.NewRawLoader(cfg)),
)
```

> `Configure` 接口本身也暴露：
>
> ```go
> Get(path string) any
> Set(path string, val any)
> ```
>
> 可在运行中直接读取/写入配置。

### 2.2 `prefix`：按配置前缀绑定结构体

`prefix` 标签会将某个配置前缀（如 `a`）下的所有配置，反序列化到字段结构体上。

简单示例（来自单测 `configuration_test.go`）：

```go
type configA struct {
	B int   `yaml:"b"`
	C []int `yaml:"c"`
}

type configD struct {
	D1 string `yaml:"d1"`
	D2 int    `yaml:"d2"`
}

func (c *configD) Prefix() string {
	return "a.d" // 支持由实现接口返回前缀
}

type configApp struct {
	A  *configA `prefix:"a"` // 结构体 tag 中声明前缀
	D  *configD             // 使用 Prefix() 方法提供前缀
	A2 configA  `prefix:"a"`
}
```

配置示例（YAML）：

```yaml
a:
  b: 123
  c: [1,2,3,4]
  d:
    d1: "abc"
    d2: 123
```

运行后：

```go
assert.Equal(t, 123, tApp.A.B)
assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
assert.Equal(t, "abc", tApp.D.D1)
assert.Equal(t, 123, tApp.D.D2)
assert.Equal(t, 123, tApp.A2.B)
assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
```

`prefix` 注入流程：

1. `PropertiesAwarePostProcessors` 扫描所有 `PropertyTypeConfiguration` 字段（通过 Tag 扫描器）；
2. 对于 `prefix` 字段，调用 `Configure.Get(prefix)` 获取配置子树；
3. 通过 `Unmarshall` 将配置反序列化到结构体字段。

### 2.3 `value`：字面值 / 配置占位 / 表达式

`value` 标签用于字段级配置注入，核心行为由 `valueAwarePostProcessors` 完成：

1. 解析标签字符串（支持 JSON / map 风格 / 基本类型 / 切片 / 指针等）；
2. 使用 `strconv2.ParseAny` 将字符串解析为 `any`；
3. 调用 `Property.Unmarshall` 将值填入字段。

#### 2.3.1 简单类型

```go
type T struct {
	A string  `value:"foo"`
	B bool    `value:"true"`
	I int     `value:"123"`
	F float64 `value:"123.321"`
}
```

#### 2.3.2 切片 / map / struct

单测中覆盖了多种情况（`value_tag_test.go`）：

```go
type T struct {
	S []string  `value:"[\"hello\",\"world\",\"foo\",\"bar\"]"`
	I []int     `value:"[1,2,3]"`
	B []bool    `value:"[true,false,false,true]"`
	F []float64 `value:"[1.1,2.2,3]"`

	MF   map[string]any `value:"map[]"`
	MF2  map[string]any `value:"map[foo:bar]"`
	MJ   map[string]any `value:"{}"`
	MNil map[string]any
	MJ2  map[string]any `value:"{\"foo\":\"bar\"}"`

	type S struct {
		Foo string `json:"foo"`
	}
	StructFromJSON S `value:"{\"foo\":\"bar\"}"`
	StructFromMap  S `value:"map[foo:bar]"`
}
```

也支持指针：

```go
type T struct {
	Ap *string  `value:"foo"`
	B  *bool    `value:"true"`
	I  *int     `value:"123"`
	F  *float64 `value:"123.321"`
}
```

#### 2.3.3 `prop`：`value` 的语法糖

在 `valueAwarePostProcessors` 中可以看到，对 `prop` 的处理是：

```go
if tagVal, ok = field.StructField.Tag.Lookup(definition.PropTag); ok {
    // 将 `prop:"db.dsn"` 转换成 `value:"${db.dsn}"`
    tagVal = fmt.Sprintf("${%s}%s", tagVal, argstr)
}
```

因此：

```go
type DBConfig struct {
	DSN string `prop:"db.dsn"`
}
```

等价于：

```go
type DBConfig struct {
	DSN string `value:"${db.dsn}"`
}
```

`prop` 更适合简单的配置占位场景。

### 2.4 配置占位 `${...}` 与表达式 `#{...}`

#### 2.4.1 配置占位 `${...}`

由 `ConfigQuoteAwarePostProcessors` + `el.NewQuote()` 处理（代码略）：

- 匹配 `${...}` 格式；
- 使用 `Configure.Get` 读取对应配置值；
- 替换为字符串形式，再交给 `value` 解析。

例如：

```go
type T struct {
	DSN string `value:"${db.dsn}"`
}
```

配合：

```yaml
db:
  dsn: "user:pass@tcp(127.0.0.1:3306)/demo"
```

最终 `DSN` 为 `"user:pass@tcp(127.0.0.1:3306)/demo"`。

`expression_tag_test.go` 中还演示了使用 `${:1}` 等“默认配置占位”的技巧。

#### 2.4.2 表达式 `#{...}`

由 `ExpressionTagAwarePostProcessors` 负责：

1. 使用 `el.NewExpr()` 匹配 `#{...}`；
2. 使用 [`expr-lang/expr`](https://github.com/expr-lang/expr) 编译并运行表达式；
3. 将结果格式化为字符串，再替换回 `TagVal`；
4. 后续由 `value` 流程完成类型转换与注入。

静态表达式示例（单测 `StaticExpression`）：

```go
type T struct {
	Arithmetic  int    `value:"#{1+(1*2)}"`                      // 3
	Comparison  bool   `value:"#{1/1==1}"`                      // true
	Logical     bool   `value:"#{(1+1)>=2||1!=1}"`              // true
	Conditional string `value:"#{1>2?'a':'b'}"`                 // "b"
	Membership  bool   `value:"#{'a' in ['a','b','c']}"`        // true
	String      bool   `value:"#{'hello'+' '+'world' contains 'o w'}"` // true
}
```

与配置占位组合使用（`ExpressionWithConfigQuote`）：

```go
type T struct {
	Arithmetic  int    `value:"#{${number.val1}+(${number.val1}*${number.val2})}"`
	Comparison  bool   `value:"#{${number.val1}/${number.val1}==${number.val1}}"`
	Logical     bool   `value:"#{(1+1)${logical.compare}2||1!=1}"`
	Conditional string `value:"#{1>2?'${character.val1}':'${character.val2}'}"`
	Membership  bool   `value:"#{'a' in ${slices}}"`
	String      bool   `value:"#{'${character.val3}'+' '+'${character.val4}' contains 'o w'}"`
}
```

配合配置：

```yaml
number:
  val1: 1
  val2: 2
logical:
  compare: ">="
character:
  val1: a
  val2: b
  val3: "hello"
  val4: "world"
slices:
  - a
  - 'a'
  - 'b'
  - c
  - 1
  - 3.14
  - true
```

---

## 3. 构造器注入（函数注入）

除了字段注入，`go-kid/ioc` 还支持**构造器函数注入**：

- 通过注册构造器函数（例如 `NewComponent`）；
- 使用 `NewConstructorAwarePostProcessors`（在 `test/constructor` 中演示）。

示例（简化自 `test/constructor/main.go`）：

```go
type Component struct {
	dependency *processors.A
}

func NewComponent(d1 *processors.A) *Component {
	return &Component{
		dependency: d1,
	}
}

type App struct {
	Component *Component `wire:""`
}

func main() {
	var a = &App{}
	pa := &processors.A{Name: "A23"}

	_, err := ioc.Run(
		app.SetComponents(
			a,
			pa,
			NewComponent,                                 // 注册构造器
			processors.NewConstructorAwarePostProcessors(), // 启用构造器感知处理器
		))
	if err != nil {
		panic(err)
	}
	fmt.Println(a.Component.dependency) // 指向 pa
}
```

构造器注入的大致流程：

1. 构造器函数作为组件被注册；
2. `ConstructorAwarePostProcessors` 扫描函数签名；
3. 对其参数同样应用依赖注入规则（按类型 / 名称 / qualifier 匹配）；
4. 调用构造器返回真正组件实例，用于后续注入。

---

## 4. 应用启动与生命周期

### 4.1 App 与顶层 `ioc` 包

常用两种启动方式：

#### 方式一：直接用 `ioc.Run`/`ioc.Register`

```go
func Register(cs ...interface{}) {
	registerHandlers = append(registerHandlers, app.SetComponents(cs...))
}

func Run(ops ...app.SettingOption) (*app.App, error) {
	s := app.NewApp()
	// ...
	return s, s.Run(append(ops, registerHandlers...)...)
}
```

适合普通应用。

#### 方式二：直接操纵 `app.App`

```go
application := app.NewApp()
err := application.Run(
    app.SetComponents(...),
    app.SetConfigLoader(...),
    app.SetConfigBinder(...),
    app.LogDebug,
)
```

### 4.2 生命周期接口

部分接口（在 `definition` 包中）用于参与容器生命周期：

- `ApplicationRunner`：在所有组件刷新完毕后执行；
- `CloserComponent`：在应用关闭时异步调用 `Close()`；
- `LazyInitComponent`：标记为懒加载，在 `Refresh()` 时跳过；
- `PriorityComponent`：控制某些后置处理器的执行顺序；
- `WirePrimary` / `WireQualifier`：参与注入选择逻辑。

`App.run()` 流程：

1. 初始化配置 `initConfiguration()`；
2. 初始化工厂 `initFactory()`（准备组件、注册后置处理器等）；
3. 刷新组件 `refresh()`（非 lazy 组件全部实例化并注入依赖）；
4. 调用 ApplicationRunners `callRunners()`；
5. 退出时 `Close()` 调用所有 `CloserComponent.Close()`。

---

## 5. 架构总览

核心模块关系（Mermaid 图）：

```mermaid
flowchart LR
    subgraph App
        A[app.App] --> B[Component Factory]
        A --> C[Configure]
    end

    subgraph Configure
        C --> C1[Loader: args/file/raw]
        C --> C2[Binder: viper]
    end

    subgraph Container
        B --> D[DefinitionRegistry]
        B --> E[SingletonRegistry]

        D -->|register| F[ComponentDefinition(Meta)]
        F -->|create| G[Instance]

        G --> H[PostProcessors]
    end

    subgraph PostProcessors
        H1[LoggerAware]
        H2[ConfigQuoteAware ${...}]
        H3[ExpressionTagAware #{...}]
        H4[PropertiesAware prefix]
        H5[ValueAware value/prop]
        H6[ValidateAware]
        H7[DependencyAware wire]
        H8[DependencyFuncAware func]
        H9[DependencyFurtherMatching qualifier/primary]

        H --> H1 & H2 & H3 & H4 & H5 & H6 & H7 & H8 & H9
    end

    C2 --> H2 & H4 & H5 & H3
```

---

## 6. 示例与测试

仓库中已有大量示例与单测，建议直接阅读/运行：

- 示例工程：
  - `test/ioc`：基础 IoC 示例
  - `test/constructor`：构造器注入
  - `test/prop` / `test/t_yaml`：配置 & `prefix` 示例
- 单元测试：
  - 组件注入相关：`unittest/component/builtin_inject/*`
  - 配置 / 表达式相关：`unittest/configure/*`

运行所有测试：

```bash
go test ./...
```

---

## 7. 贡献与许可证

- 欢迎通过 Issue / PR 进行反馈与贡献；
- 先确保 `go test ./...` 通过，再提交。

本项目采用 [MIT License](./LICENSE)。
