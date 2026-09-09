# Configuration injection

## Loaders and binder

The default configuration contains a Viper YAML binder and an args loader. Configure additional sources through application options:

```go
ioc.Run(
	app.SetConfig("config.yaml"), // appends a file loader
	app.AddConfigLoader(loader.NewRawLoader(overrides)),
)
```

`SetConfigLoader` replaces all existing loaders, including the default args loader. `AddConfigLoader` and `SetConfig` append. Loaders are ordered and each payload is merged by the binder.

Use `app.SetConfigBinder(binder.NewViperBinder("json"))` when the payload format is JSON. A custom source implements `configure.Loader`; a custom store/decoder implements `configure.Binder`.

### Custom loader

`LoadConfig` returns one payload in the format expected by the active binder. Returning an empty payload skips the binder call. Errors stop configuration initialization and are wrapped with the loader type.

```go
type remoteLoader struct {
	client *ConfigClient
}

func (l *remoteLoader) LoadConfig() ([]byte, error) {
	return l.client.Fetch("my-service") // for example, YAML bytes
}

// Optional: loaders use the same Priority/Ordered rules as processors.
func (*remoteLoader) Order() int { return 10 }

a, err := ioc.Run(
	app.AddConfigLoader(&remoteLoader{client: client}),
	app.SetComponents(...),
)
```

Priority loaders run before ordinary ordered loaders; lower `Order()` values run first within a group. Loaders without ordering run last. The framework does not define relative ordering for equal values.

### Custom binder

A binder receives each non-empty loader payload through `SetConfig`. When multiple loaders are configured, `SetConfig` must merge successive payloads if layered configuration is desired. `Get` defines path lookup semantics used by `${...}`, `prop`, `prefix`, conditions, and runtime access; `Set` handles runtime overrides.

```go
type validatingBinder struct {
	delegate configure.Binder
}

func (b *validatingBinder) SetConfig(raw []byte) error {
	if err := validateDocument(raw); err != nil {
		return err
	}
	return b.delegate.SetConfig(raw)
}

func (b *validatingBinder) Get(path string) any {
	return b.delegate.Get(path)
}

func (b *validatingBinder) Set(path string, value any) {
	b.delegate.Set(path, value)
}

custom := &validatingBinder{
	delegate: binder.NewViperBinder("yaml"),
}
a, err := ioc.Run(app.SetConfigBinder(custom), app.SetComponents(...))
```

Install a full `configure.Configure` with `app.SetConfigure` only when loader orchestration itself must change. For a new source or decoding policy, prefer a Loader or Binder rather than replacing the whole configuration pipeline.

## `prefix`

Bind a configuration subtree into a struct or pointer field:

```go
type DBConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type App struct {
	DB *DBConfig `prefix:"database"`
}
```

`prefix` is required by default. Use `prefix:"database,required=false"` when the whole subtree may be absent.

A field type implementing `definition.ConfigurationProperties` can supply its prefix without an explicit field tag:

```go
func (*DBConfig) Prefix() string { return "database" }

type App struct {
	DB *DBConfig
}
```

## `value` and `prop`

`value` accepts literals, placeholders, and expressions. `prop` wraps its first value as a placeholder.

```go
type Server struct {
	Name    string `value:"api"`
	Host    string `value:"${server.host:localhost}"`
	Port    int    `prop:"server.port:8080"`
	Workers int    `value:"#{${cpu:2} * 2}"`
	Maybe   string `value:"${optional},required=false"`
}
```

- Placeholder: `${path}` or `${path:default}`.
- Expression: `#{...}`, evaluated after placeholders.
- Multiple placeholders can appear in a string.
- Values are decoded into scalar, slice, map, struct, or pointer fields through the framework's weakly typed mapper.

`prop:"server.port:8080"` is equivalent to `value:"${server.port:8080}"`.

## Validation and decode options

Tag arguments follow the first comma. Space-separate multiple validation rules so they remain the value of the `validate` argument:

```go
Port int `prop:"server.port:8080,validate=required min=1 max=65535"`
```

For structured decoding, `mapper=<tag-name>` changes the mapstructure field tag and `timeLayout=<layout>` adds time parsing.

## Runtime access

The returned `*app.App` embeds `configure.Configure`, so runtime reads and writes are supported:

```go
a, err := ioc.Run(...)
host := a.Get("server.host")
a.Set("server.host", "127.0.0.1")
```

Prefer field binding for declared component dependencies; use runtime access when the value is genuinely dynamic.
