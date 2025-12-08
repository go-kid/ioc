# go-kid/ioc

`go-kid/ioc` is a Go runtime dependency injection (IoC/DI) framework based on **tag + interface**.

- Use the `wire` tag for **component dependency injection**:
  - Inject by pointer type
  - Inject by interface type
  - Inject by component name
  - Inject multiple implementations of an interface into an interface slice/array
- Use `value` / `prop` / `prefix` tags for **configuration injection**:
  - Load from multiple configuration sources (command-line args, files, raw content, etc.)
  - Support `${...}` configuration placeholders
  - Support `#{...}` expression evaluation (arithmetic, logical, conditional, collection operations, etc.)

---

## Installation

```bash
go get github.com/go-kid/ioc
```

---

## Quick Start

A minimal example of dependency injection:

```go
package main

import (
	"fmt"

	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
)

// Component to be injected
type ComponentA struct {
	Name string
}

func (a *ComponentA) GetName() string {
	return a.Name
}

// Target struct that depends on ComponentA
type App struct {
	ComponentA *ComponentA `wire:""` // pointer + exported field + wire tag
}

func main() {
	a := new(App)

	// Register components
	ioc.Register(a)
	ioc.Register(&ComponentA{Name: "Comp-A"})

	// Run the framework
	_, err := ioc.Run()
	if err != nil {
		panic(err)
	}

	// Should print "Comp-A"
	fmt.Println(a.ComponentA.GetName())
}
```

Key points:

- Dependency fields must be exported (start with an uppercase letter) and settable.
- Use `wire:""` to declare that a field should be injected.

---

## 1. Component Dependency Injection (`wire` tag)

Component injection is controlled via a tag:

```go
`wire:"<name>,arg1=xxx,arg2=yyy"`
```

Where:

- `<name>`: optional, inject by component name.
- `,arg=...`: optional extra arguments (e.g., `qualifier`).
- `wire:""` (empty) means inject by type.

### 1.1 Inject by Pointer Type

```go
type Component struct {
	Name string
}

type App struct {
	C *Component `wire:""` // injected by *Component type
}
```

When there is one or more `*Component` instances in the container:

- If only one exists, it will be injected directly.
- If more than one exists, the “multiple implementation resolution rules” (see below) will be applied.

### 1.2 Inject by Interface Type

Using interfaces is recommended:

```go
type IComponent interface {
	GetName() string
}

type ComponentA struct{ Name string }
func (a *ComponentA) GetName() string { return a.Name }

type App struct {
	C IComponent `wire:""` // inject by interface type
}
```

The container will find all implementations of `IComponent` and choose one based on the strategy in section 1.4.

### 1.3 Inject by Name

`go-kid/ioc` provides a built-in interface `NamingComponent` to define component names:

```go
type NamingComponent interface {
	Naming() string
}
```

If a component implements this interface, the return value of `Naming()` will be used as its alias (component name).

Injection by name:

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
	ByName IComponent `wire:"comp-A"` // inject by component name
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

If `wire:"comp-B"` is specified but no such component exists in the container, an error will be thrown in strict mode.

### 1.4 Multiple Implementations for the Same Interface

When an interface has multiple implementations and the field is declared as that interface (or pointer type), with an empty `wire` tag:

```go
type IComponent interface{ GetName() string }

type ComponentA struct{ Name string }
func (a *ComponentA) GetName() string { return a.Name }

type ComponentB struct{ Name string }
func (b *ComponentB) GetName() string { return b.Name }

type App struct {
	C IComponent `wire:""` // multiple implementations, choose by strategy
}
```

The container:

1. Collects all components implementing `IComponent`.
2. If there are more than one:
   - If one of them is marked as `WirePrimary`, that one is preferred.
   - Otherwise, prefer components **without an alias** (i.e., not implementing `Naming()` or at least not using a custom name).
   - If there are still multiple candidates, one will be chosen (order is not guaranteed).

#### Primary Injection: `WirePrimary` Interface

Conceptually:

```go
type WirePrimaryComponent struct{}
func (i *WirePrimaryComponent) Primary() {}
```

Any component that has the `Primary()` method is considered a “primary” implementation.

Example:

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

In this case, `C` will always be injected with `ComponentA`.

Notes:

- `WirePrimary` only affects injection when using interface/pointer injection and `wire` does not specify an explicit name.
- Avoid marking multiple implementations of the same interface as primary, or the result may still be non-deterministic.

#### Qualified Injection: `WireQualifier` + `qualifier` Argument

When an interface has multiple implementations and you want to **explicitly choose** which one to inject, use qualifiers:

```go
type WireQualifier interface {
	Qualifier() string
}
```

Implementation example:

```go
type ComponentA struct {
	Name          string
	componentName string
	qualifier     string
}

func (a *ComponentA) Qualifier() string { return a.qualifier }
```

Usage in `wire`:

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
	fmt.Println(a.C.GetName()) // always "Comp-A"
}
```

In `wire:"name,qualifier=xxx"`:

- `name` is used to look up components by alias/name.
- `qualifier` is used to further narrow down candidates.
- If both are provided, both are applied.

### 1.5 Inject Multiple Implementations into Interface Slice/Array

When a field is an interface slice/array and uses `wire:""`, the container will inject **all matching implementations**:

```go
type Interface interface {
	SimpleInterface()
}

type InterfaceImplComponent struct{ Component }
func (i *InterfaceImplComponent) SimpleInterface() {}

type InterfaceImplNamingComponent struct{ SpecifyNameComponent }
func (i *InterfaceImplNamingComponent) SimpleInterface() {}

type App struct {
	All []Interface `wire:""` // all Interface implementations
}
```

Test snippet:

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
assert.Equal(t, 2, len(tApp.Ts)) // only the two real Interface implementations are injected
```

Similarly:

```go
type App struct {
	Ptrs []*SpecifyNameComponent `wire:""` // all *SpecifyNameComponent instances
}
```

---

## 2. Configuration Injection (`value` / `prop` / `prefix`)

The configuration system in `go-kid/ioc` consists of:

- `configure.Loader`: loads raw configuration from various sources (YAML/JSON files, command-line args, raw bytes, etc.).
- `configure.Binder`: parses the raw configuration into a unified access layer, by default using `viper`.
- Injection tags:
  - `value`: convert **literal values / placeholders / expressions** into field values.
  - `prop`: a shorthand for `value` with `${...}`.
  - `prefix`: bind a configuration subtree to a struct.

### 2.1 Basic Configuration Loading

By default `App` uses:

```go
configure.Default():
- Loader:  loader.NewArgsLoader(os.Args)
- Binder:  binder.NewViperBinder("yaml")
```

You can customize the configuration loader and binder in `Run`:

```go
ioc.Run(
    app.SetConfigLoader(
        loader.NewFileLoader("config.yaml"), // load from file
    ),
    app.SetConfigBinder(
        binder.NewViperBinder("yaml"),       // or "json"
    ),
)
```

Or use raw bytes (frequently used in tests):

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

The `Configure` interface itself also exposes:

```go
Get(path string) any
Set(path string, val any)
```

So you can read/write configuration at runtime.

### 2.2 `prefix`: Bind Struct from Config Prefix

The `prefix` tag binds the configuration subtree under a specific prefix (e.g., `a`) into the struct field.

A simple example (from `configuration_test.go`):

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
	return "a.d" // dynamic prefix via method
}

type configApp struct {
	A  *configA `prefix:"a"` // prefix declared in tag
	D  *configD             // prefix from Prefix() method
	A2 configA  `prefix:"a"`
}
```

Configuration (YAML):

```yaml
a:
  b: 123
  c: [1,2,3,4]
  d:
    d1: "abc"
    d2: 123
```

After injection:

```go
assert.Equal(t, 123, tApp.A.B)
assert.Equal(t, []int{1, 2, 3, 4}, tApp.A.C)
assert.Equal(t, "abc", tApp.D.D1)
assert.Equal(t, 123, tApp.D.D2)
assert.Equal(t, 123, tApp.A2.B)
assert.Equal(t, []int{1, 2, 3, 4}, tApp.A2.C)
```

`prefix` injection flow:

1. `PropertiesAwarePostProcessors` scans fields of `PropertyTypeConfiguration` (discovered by tag scanner).
2. For `prefix` fields, call `Configure.Get(prefix)` to obtain the subtree.
3. Use `Unmarshall` to deserialize the subtree into the struct field.

### 2.3 `value`: Literals / Placeholders / Expressions

The `value` tag is used for field-level configuration injection. The main logic is in `valueAwarePostProcessors`:

1. Parse the tag string (supports JSON, map syntax, basic types, slices, pointers, etc.).
2. Use `strconv2.ParseAny` to parse the string into `any`.
3. Use `Property.Unmarshall` to assign the value to the field.

#### 2.3.1 Simple Types

```go
type T struct {
	A string  `value:"foo"`
	B bool    `value:"true"`
	I int     `value:"123"`
	F float64 `value:"123.321"`
}
```

#### 2.3.2 Slices / Maps / Structs

From `value_tag_test.go`:

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

Pointers are also supported:

```go
type T struct {
	Ap *string  `value:"foo"`
	B  *bool    `value:"true"`
	I  *int     `value:"123"`
	F  *float64 `value:"123.321"`
}
```

#### 2.3.3 `prop`: Sugar for `value` with `${...}`

In `valueAwarePostProcessors`, `prop` is essentially transformed into `value`:

```go
if tagVal, ok = field.StructField.Tag.Lookup(definition.PropTag); ok {
    // convert `prop:"db.dsn"` to `value:"${db.dsn}"`
    tagVal = fmt.Sprintf("${%s}%s", tagVal, argstr)
}
```

So:

```go
type DBConfig struct {
	DSN string `prop:"db.dsn"`
}
```

is equivalent to:

```go
type DBConfig struct {
	DSN string `value:"${db.dsn}"`
}
```

`prop` is convenient for simple configuration placeholders.

### 2.4 Placeholders `${...}` and Expressions `#{...}`

#### 2.4.1 Configuration Placeholders `${...}`

Handled by `ConfigQuoteAwarePostProcessors` + `el.NewQuote()`:

- Match patterns of the form `${...}`.
- Use `Configure.Get` to retrieve the configuration value.
- Replace the placeholder with its string representation.
- Pass the result onward to `value` parsing.

Example:

```go
type T struct {
	DSN string `value:"${db.dsn}"`
}
```

With configuration:

```yaml
db:
  dsn: "user:pass@tcp(127.0.0.1:3306)/demo"
```

`DSN` will be `"user:pass@tcp(127.0.0.1:3306)/demo"`.

The tests (`expression_tag_test.go`) also include “default placeholder” usage such as `${:1}`.

#### 2.4.2 Expressions `#{...}`

Handled by `ExpressionTagAwarePostProcessors`:

1. Use `el.NewExpr()` to match `#{...}`.
2. Use [`expr-lang/expr`](https://github.com/expr-lang/expr) to compile and evaluate the expression.
3. Convert the result to string and substitute it back into `TagVal`.
4. Then the usual `value` injection pipeline handles type conversion and assignment.

Static expression examples (from `StaticExpression` tests):

```go
type T struct {
	Arithmetic  int    `value:"#{1+(1*2)}"`                       // 3
	Comparison  bool   `value:"#{1/1==1}"`                        // true
	Logical     bool   `value:"#{(1+1)>=2||1!=1}"`                // true
	Conditional string `value:"#{1>2?'a':'b'}"`                   // "b"
	Membership  bool   `value:"#{'a' in ['a','b','c']}"`          // true
	String      bool   `value:"#{'hello'+' '+'world' contains 'o w'}"` // true
}
```

Combined with configuration placeholders (`ExpressionWithConfigQuote`):

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

With configuration:

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

## 3. Constructor Injection (Function Injection)

Besides field injection, `go-kid/ioc` also supports **constructor function injection**:

- Register constructor functions (e.g., `NewComponent`).
- Enable `NewConstructorAwarePostProcessors` (see `test/constructor`).

Simplified example (based on `test/constructor/main.go`):

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
			NewComponent,                                   // register constructor
			processors.NewConstructorAwarePostProcessors(), // enable constructor-aware processor
		))
	if err != nil {
		panic(err)
	}
	fmt.Println(a.Component.dependency) // points to pa
}
```

Constructor injection flow overview:

1. The constructor function itself is registered as a component.
2. `ConstructorAwarePostProcessors` scan function signatures.
3. Parameters are subject to the same DI rules (by type/name/qualifier).
4. The constructor is invoked and the returned instance is registered as the real component.

---

## 4. Application Startup & Lifecycle

### 4.1 App vs Top-Level `ioc` Package

There are two common startup styles.

#### Style 1: Use `ioc.Run` / `ioc.Register`

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

Good for simple applications.

#### Style 2: Work with `app.App` Directly

```go
application := app.NewApp()
err := application.Run(
    app.SetComponents(...),
    app.SetConfigLoader(...),
    app.SetConfigBinder(...),
    app.LogDebug,
)
```

### 4.2 Lifecycle Interfaces

Several interfaces in the `definition` package participate in the container lifecycle:

- `ApplicationRunner`: executed after all components are refreshed.
- `CloserComponent`: components implementing `Close()` will be closed asynchronously when the app stops.
- `LazyInitComponent`: marked as lazy; skipped during `Refresh()` and created on first use.
- `PriorityComponent`: controls ordering for some post-processors.
- `WirePrimary` / `WireQualifier`: affect DI resolution.

`App.run()` roughly proceeds as follows:

1. Initialize configuration: `initConfiguration()`.
2. Initialize factory: `initFactory()` (prepare components, register post-processors, etc.).
3. Refresh components: `refresh()` (instantiate non-lazy components and inject dependencies).
4. Invoke application runners: `callRunners()`.
5. On exit, `Close()` calls `Close()` on all `CloserComponent`s.

---

## 5. Architecture Overview

High-level module relationships (Mermaid):

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

## 6. Examples & Tests

There are plenty of examples and tests in the repository; it’s recommended to read/run them:

- Example projects:
  - `test/ioc`: basic IoC examples
  - `test/constructor`: constructor injection
  - `test/prop` / `test/t_yaml`: configuration & `prefix` examples
- Unit tests:
  - Component injection: `unittest/component/builtin_inject/*`
  - Configuration / expressions: `unittest/configure/*`

Run all tests:

```bash
go test ./...
```

---

## 7. Contributing & License

- Contributions via Issues / PRs are welcome.
- Please ensure `go test ./...` passes before submitting.

This project is licensed under the [MIT License](./LICENSE).
