---
name: ioc-dev
description: Develop Go applications and framework extensions with go-kid/ioc, including injection, configuration loaders/binders, lifecycle hooks, post-processors, and AOP proxies. Use for code that imports github.com/go-kid/ioc; use ioc-debug for failures and ioc-test for testing.
---

# go-kid/ioc Application and Extension Development

Use the repository's `go.mod`, exported interfaces, and tests as the source of truth. The current module requires Go 1.21 or later.

## Route by task

- Read [references/component-injection.md](references/component-injection.md) for registration, `wire`/`func` tags, naming, qualifiers, scopes, and conditional components.
- Read [references/config-injection.md](references/config-injection.md) for built-in or custom loaders/binders, `value`/`prop`/`prefix` tags, placeholders, expressions, validation, and `ConfigurationProperties`.
- Read [references/lifecycle.md](references/lifecycle.md) for the `definition` interface map, startup options, lifecycle/context interfaces, runners, shutdown, events, and initialization-skipping options.
- Read [references/postprocessor-extensions.md](references/postprocessor-extensions.md) for container post-processors, custom field processing, AOP proxies, early singleton references, and destruction interception.

Read only the references needed for the request.

## Minimal application

```go
type Repository struct{}

type Service struct {
	Repo *Repository `wire:""`
}

func main() {
	a, err := ioc.Run(app.SetComponents(&Service{}, &Repository{}))
	if err != nil {
		panic(err)
	}
	defer a.Close()
}
```

## Implementation constraints

- Register component pointers. Invoke constructors before registration; registering a function is unsupported and panics.
- Tagged fields must be exported/settable. Unexported fields are not scanned.
- `wire`, `value`, and `prefix` dependencies are required unless the tag includes `required=false`.
- `ioc.Register` and `app.Settings` use package-global queues consumed by the next `ioc.Run`; prefer `app.SetComponents` and explicit options when isolation matters.
- Do not assume an interface is active merely because it is declared. In particular, context-aware runners and closers must also implement their non-context base interfaces to be discovered by the current `app.App` fields.

Keep changes aligned with existing APIs and tests; avoid adding wrappers or abstractions when a tag, interface, or `app.SettingOption` already expresses the requirement.
