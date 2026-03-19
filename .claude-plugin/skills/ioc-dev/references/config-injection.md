# Configuration Injection Reference

Complete guide for configuration loading and binding in go-kid/ioc.

**Requires Go 1.21+**

---

## Config Sources

Set config sources via `app.SettingOption`:

```go
// From file
app.SetConfig("config.yaml")

// From raw bytes
app.SetConfigLoader(loader.NewRawLoader([]byte(`key: value`)))

// From file with explicit loader
app.SetConfigLoader(loader.NewFileLoader("config.yaml"))

// JSON format (change binder)
app.SetConfigLoader(loader.NewRawLoader(jsonBytes)),
app.SetConfigBinder(binder.NewViperBinder("json"))
```

Import paths:
- `github.com/go-kid/ioc/configure/loader`
- `github.com/go-kid/ioc/configure/binder`

---

## `prefix` Tag

Bind a config subtree to a struct pointer or struct:

```go
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

type App struct {
    DB *DBConfig `prefix:"database"`  // pointer
    DB2 DBConfig `prefix:"database"`  // value type also works
}
```

Config:
```yaml
database:
  host: localhost
  port: 5432
```

### `ConfigurationProperties` Interface

Alternative to `prefix` tag -- implement the interface directly:

```go
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

func (c *DBConfig) Prefix() string { return "database" }
```

Register as a component; `prefix` is inferred from the method. No tag needed on the parent struct field.

### ConfigurationProperties with Constructor Injection

When a constructor parameter's type implements `ConfigurationProperties`, the framework automatically creates an instance and populates it from config, even if it wasn't explicitly registered:

```go
type DBConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}

func (c *DBConfig) Prefix() string { return "database" }

func NewService(cfg *DBConfig) *Service {
    return &Service{host: cfg.Host, port: cfg.Port}
}

ioc.Register(NewService)
ioc.Run(app.SetConfig("config.yaml"))
```

### Dynamic Prefix with Placeholder

```go
type App struct {
    Host string `prefix:"server.${env}.host"`  // env resolved from config
}
```

---

## `value` Tag

Inject literal values, config placeholders, or expressions:

```go
type App struct {
    // Literal values
    Name string  `value:"hello"`
    Port int     `value:"8080"`
    Flag bool    `value:"true"`
    Rate float64 `value:"0.95"`

    // Config placeholder
    DSN string `value:"${database.dsn}"`

    // Placeholder with default
    Host string `value:"${server.host:localhost}"`

    // Expression
    Sum int `value:"#{1+2}"`

    // Mixed: expression with config values
    Total int `value:"#{${price} * ${qty}}"`
}
```

Supported types: string, bool, int/int64, float64, slices, maps, structs, pointers.

### Slice / Map / Struct Literals

```go
type App struct {
    Ports   []int          `value:"[8080,9090]"`
    Params  map[string]any `value:"map[key:val]"`       // map literal
    Params2 map[string]any `value:"{\"key\":\"val\"}"`   // JSON literal
}
```

### Optional Value

```go
type App struct {
    Val string `value:"${maybe.missing:},required=false"`
}
```

---

## `prop` Tag

Syntactic sugar for `value:"${...}"`:

```go
type App struct {
    Host string `prop:"server.host"`           // same as value:"${server.host}"
    Port int    `prop:"server.port:8080"`      // with default
    Tags []int  `prop:"app.tags:[1,2,3]"`      // with default slice
}
```

`prop` supports the same default value syntax and additional args as `value`:

```go
type App struct {
    Port []int `prop:"server.port:[1,2,3],required=true,validate=required min=3 max=20"`
}
```

---

## Placeholders `${...}`

Syntax: `${config.path}` or `${config.path:default_value}`

- Resolves from config at the given path
- Supports default after `:` separator
- Can be nested in expressions or other tags
- Multiple placeholders in one value: `"https://${sub:api}.${domain:example.com}"`

---

## Expressions `#{...}`

Powered by expr-lang. Supports:

| Category | Examples |
|----------|---------|
| Arithmetic | `#{1+(1*2)}` = 3 |
| Comparison | `#{1/1==1}` = true |
| Logical | `#{(1+1)>=2 \|\| 1!=1}` = true |
| Conditional | `#{1>2?'a':'b'}` = "b" |
| Membership | `#{'a' in ['a','b','c']}` = true |
| String ops | `#{'hello world' contains 'o w'}` = true |

Combine with placeholders:

```go
type App struct {
    Total int `value:"#{${price}+${tax}}"`
}
```

---

## Runtime Config Access

After `ioc.Run()`, access config directly:

```go
app, _ := ioc.Run(app.SetConfigLoader(loader.NewRawLoader(cfg)))
val := app.Get("server.host")    // read
app.Set("server.host", "0.0.0.0") // write
```
