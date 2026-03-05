---
name: IoC Configuration System
description: This skill should be used when the user asks to "add configuration", "bind config", "use value tag", "use prop tag", "use prefix tag", "add config placeholder", "use expression tag", "load YAML config", "set default value", "use ConfigurationProperties", "add config loader", or needs guidance on the go-kid/ioc configuration binding system including placeholders, expressions, loaders, and binders.
---

# IoC Configuration System

Guide for using the configuration binding system in go-kid/ioc.

## Configuration Sources (Loaders)

Loaders provide raw configuration data. Multiple loaders can be combined:

```go
import "github.com/go-kid/ioc/configure/loader"

// From YAML file
app.SetConfigLoader(loader.NewFileLoader("config.yaml"))

// From raw YAML bytes (useful in tests)
app.SetConfigLoader(loader.NewRawLoader([]byte(`
db:
  host: localhost
  port: 5432
`)))

// From command-line args
app.SetConfigLoader(loader.NewArgsLoader())

// Multiple loaders (later ones override earlier)
app.AddConfigLoader(loader.NewFileLoader("config.yaml"))
app.AddConfigLoader(loader.NewRawLoader(overrides))
```

## Configuration Tags

### `prefix` — Bind Config Subtree

Bind an entire config section to a struct. Struct fields use `yaml` tags for mapping:

```go
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
    SSL  bool   `yaml:"ssl"`
}

type App struct {
    DB *DBConfig `prefix:"database"`
}
```

Config YAML:
```yaml
database:
  host: localhost
  port: 5432
  ssl: true
```

Dynamic prefix with placeholders:

```go
type App struct {
    Host string `prefix:"test.${env}.host"`
}
```

### `value` — Literals, Placeholders, Expressions

The `value` tag supports three modes:

**Literal values:**
```go
type T struct {
    Name string  `value:"foo"`
    Flag bool    `value:"true"`
    Num  int     `value:"123"`
    Rate float64 `value:"99.9"`
    List []int   `value:"[1,2,3]"`
    Map  map[string]any `value:"{\"key\":\"val\"}"`
}
```

**Config placeholders (`${...}`):**
```go
type T struct {
    Host string `value:"${db.host}"`                    // required
    Port int    `value:"${db.port:5432}"`               // with default
    DSN  string `value:"${db.host}:${db.port:5432}"`    // multiple placeholders
    Opt  string `value:"${missing:},required=false"`     // optional empty default
}
```

**Expressions (`#{...}`):**

Powered by `expr-lang/expr`. Supports arithmetic, logic, conditional, membership, and string operations:

```go
type T struct {
    Sum    int    `value:"#{1+(1*2)}"`                    // 3
    Check  bool   `value:"#{1/1==1}"`                     // true
    Logic  bool   `value:"#{'a' in ['a','b','c']}"`       // true
    Cond   string `value:"#{1>2?'a':'b'}"`                // "b"
    Mixed  int    `value:"#{${val1}+${val2}}"`            // placeholders in expressions
}
```

### `prop` — Sugar for Placeholders

`prop:"key"` is equivalent to `value:"${key}"`:

```go
type T struct {
    Host string `prop:"db.host"`                  // = value:"${db.host}"
    Port int    `prop:"db.port:5432"`             // = value:"${db.port:5432}"
    List []int  `prop:"items:[1,2,3]"`            // with default
}
```

`prop` also supports extra args:

```go
Port []int `prop:"port2:[1,2,3],required=true,validate=required min=3 max=20"`
```

## Default Values

Use `:` separator after the key name:

```go
`value:"${key:defaultValue}"`       // string default
`value:"${port:8080}"`              // int default
`value:"${list:[1,2,3]}"`           // slice default
`value:"${params:map[a:b]}"`        // map default
`prop:"key:defaultValue"`           // same in prop syntax
```

When no key is found and no default provided, the field is **required** by default and causes an error. Override with `required=false`:

```go
`value:"${optional_key:},required=false"`
```

## ConfigurationProperties (Constructor Pattern)

For constructor injection, implement `ConfigurationProperties` to auto-bind config:

```go
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

func (c *DBConfig) Prefix() string { return "db" }

func NewService(cfg *DBConfig) *Service {
    return &Service{host: cfg.Host, port: cfg.Port}
}

// Usage:
ioc.Register(NewService)
// cfg is auto-created and populated from config prefix "db"
```

## Processing Order

Configuration tags are processed by post-processors in this order:

1. **LoggerAware** — `logger` tag injection
2. **ConfigQuoteAware** — resolve `${...}` placeholders to config values
3. **ExpressionTagAware** — evaluate `#{...}` expressions
4. **PropertiesAware** — bind `prefix` tagged fields
5. **ValueAware** — bind `value` and `prop` tagged fields
6. **ValidateAware** — run struct validation
7. **DependencyAware** — resolve `wire` tags

This means expressions can reference placeholders: `#{${val1}+${val2}}` works because `${...}` is resolved before `#{...}`.

## Validation

Add validation using `validate` tag arg (uses `go-playground/validator`):

```go
type T struct {
    Port int `prop:"port:8080,validate=required,min=1,max=65535"`
}
```

## Binder

The default binder uses Viper (`configure/binder/viper.go`). To use a custom binder:

```go
app.SetConfigBinder(binder.NewViperBinder("yaml")) // default
```

## Testing Configuration

```go
func TestConfig(t *testing.T) {
    cfg := []byte(`
app:
  name: test-app
  debug: true
`)
    tApp := &struct {
        Name  string `prop:"app.name"`
        Debug bool   `value:"${app.debug:false}"`
    }{}
    ioc.RunTest(t,
        app.SetConfigLoader(loader.NewRawLoader(cfg)),
        app.SetComponents(tApp),
    )
    assert.Equal(t, "test-app", tApp.Name)
    assert.True(t, tApp.Debug)
}
```
