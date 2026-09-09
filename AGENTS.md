# AGENTS.md

## Project

`go-kid/ioc` is a Go 1.21+ runtime dependency-injection framework. It uses
registered component pointers, struct tags, marker interfaces, and ordered
post-processors to provide configuration binding, dependency resolution,
lifecycle callbacks, circular-reference handling, and debugging support.

Keep changes small and aligned with the existing design. Prefer removing an
unnecessary path over adding another abstraction or parallel implementation.

## Commands

Run the checks appropriate to the change:

```bash
go test ./...                         # full test suite
go test -count=1 ./...                # full suite without test-cache reuse
go test ./path/to/package -run TestX  # focused test
go test -race ./...                   # concurrency-sensitive changes
go build ./...                        # compile all packages
go vet ./...                          # static analysis
make run_unittest                     # tracked unittest/ and util/ suites
```

Run `gofmt` on changed Go files and `git diff --check` before finishing. Do not
add a new build tool or dependency when the standard Go toolchain is sufficient.

## Release tags

Release tags use `vMAJOR.MINOR.PATCH` (for example, `v1.6.3`). Do not create or
push a release tag unless the user explicitly asks to publish a release.

- Fetch remote tags before choosing a version, then determine the latest tag
  with version-aware sorting: `git tag --list 'v[0-9]*' --sort=-version:refname`.
- When no version is specified, increment the patch component of the highest
  existing stable tag. For example, the tag after `v1.6.2` is `v1.6.3`.
- Increment minor or major only when the user explicitly requests that release
  level. Do not infer a `v2` release solely from an API change.
- Tag only a tested, committed release commit on `main`, with a clean working
  tree and the intended commit already present on the remote branch.
- Create an annotated tag whose message briefly describes the release:
  `git tag -a v1.6.3 -m "<release summary>"`.
- Never reuse, move, force-update, or silently replace an existing tag. If the
  calculated tag already exists, stop and recalculate from the current remote
  tag set.
- Verify the tag target and annotation before pushing. Push the exact tag with
  `git push origin v1.6.3`; do not use `git push --tags`.
- A published tag must have a matching GitHub Release. Before creating or
  pushing the tag, verify that the GitHub CLI is available and authenticated
  with `gh auth status`, so the release can be completed in the same workflow.
- Build release notes from the actual diff and commits since the previous tag;
  do not infer or invent changes. The notes must include a short summary and
  clearly identify applicable categories such as added features, changed
  behavior, bug fixes, and documentation or tooling updates.
- Always include a `Breaking changes` section. Write `None` when there are no
  breaking changes. Otherwise list each incompatible API or behavior change,
  affected users, and the required migration steps prominently.
- After pushing the tag, publish the matching release with a command equivalent
  to `gh release create v1.6.3 --verify-tag --title "v1.6.3" --notes-file <file>`.
  The release title and Git tag must use the same version.
- Verify the published release with `gh release view v1.6.3`. If release
  publication fails after the tag is pushed, keep the immutable tag, retry or
  report the incomplete release explicitly, and do not claim publication is
  complete until the GitHub Release exists.

## Architecture

### Startup flow

The public flow is `ioc.Register` / options -> `ioc.Run` -> `app.NewApp` ->
`App.RunWithContext`:

1. Apply explicit options and consume package-global registration options.
2. Initialize configuration loaders and the binder.
3. Prepare the factory, component definitions, and post-processors.
4. Refresh non-lazy components, populate properties, and run initialization
   callbacks.
5. Invoke `ApplicationRunner` components in order and publish the application
   started event.
6. On close, invoke registered closer components in reverse order.

### Package responsibilities

| Package | Responsibility |
| --- | --- |
| repository root (`ioc`) | Public `Register`, `Run`, debug-run, and test helpers |
| `app` | Application lifecycle and `SettingOption` configuration |
| `container` | Factory, registry, post-processor, and hook interfaces |
| `container/factory` | Component creation, population, lifecycle, scopes, and circular dependencies |
| `container/processors` | Built-in tag scanning, injection, matching, validation, and logging |
| `container/support` | Definition registry, registered components, and three-level singleton cache |
| `component_definition` | Component metadata, properties, decoding, and candidate selection |
| `definition` | Lifecycle, ordering, scope, naming, selection, and event contracts |
| `configure` | Configuration abstraction, loaders, and binders |
| `syslog` | Framework logging and `log/slog` adapter |
| `debug` | Factory event collection and debug web UI |
| `ext` | Small opt-in framework extensions |

The three-level singleton cache lives in
`container/support/singleton_component_registry.go`. Do not duplicate this
state in the registered-component registry.

## Injection model

Components are already-constructed pointers. Constructor functions are not
components: callers must invoke constructors themselves before registration,
and passing a function to `ioc.Register` or `app.SetComponents` panics. Do not
reintroduce constructor parameter resolution or a second component-creation
path without an explicit API decision.

Tagged fields must be exported and settable:

| Tag | Meaning |
| --- | --- |
| `wire` | Inject a component by type, name, qualifier, or optional requirement |
| `value` | Inject a literal, `${...}` placeholder, or `#{...}` expression |
| `prop` | Shorthand for a configuration placeholder |
| `prefix` | Decode a configuration subtree into a struct or pointer |
| `logger` | Inject a framework logger |
| `func` | Match a component by method name and optional return value |

Configuration injection belongs to the built-in processors. `prefix` and
`ConfigurationProperties` are handled by `PropertiesAwarePostProcessors`;
`value` and `prop` are handled by `ValueAwarePostProcessors`. Keep decoding in
the property/processor path instead of adding factory-specific fallbacks.

For a single dependency with multiple candidates, selection considers
`Primary`, unnamed components, and qualifiers. Slice dependencies receive all
matching candidates. Preserve the existing `required=false` behavior for
optional fields.

## Post-processors

Post-processors are the framework's extension boundary. Built-ins are
registered in `app.App.initiate`; their ordering constants are in
`container/processors/orders.go`.

- Priority processors run before ordinary ordered processors.
- Within a group, lower `Order()` values run first.
- Logger, placeholders, expressions, and configuration population must happen
  before ordinary dependency matching and validation.
- Do not rely on the relative order of processors with equal `Order()` values.
- Definition-registry processors scan components concurrently; mutable custom
  processor state must be concurrency-safe.
- `DefaultInstantiationAwareComponentPostProcessor` returns `false` from
  `PostProcessAfterInstantiation`; override it when a custom processor needs
  `PostProcessProperties` to run.

Before depending on a declared extension interface, verify that the current
application/factory flow actually invokes it and add an integration test for
new behavior.

## Testing

- Put maintained regression tests next to their package or under `unittest/`.
  The ignored top-level `test/` directory is local scratch space and is not a
  substitute for tracked tests.
- Prefer `app.SetComponents` in tests. `ioc.Register` and `app.Settings` use
  package-global queues consumed by the next run and can couple tests.
- Use `ioc.RunTest` for successful full-container startup and
  `ioc.RunErrorTest` only when checking that startup fails. Use `app.NewApp`
  directly when the exact returned error matters.
- Use `loader.NewRawLoader` for inline configuration fixtures.
- Assert observable behavior or stable error meaning, not complete wrapped
  error strings or incidental internal ordering.
- Add focused tests for bug fixes and boundary changes; run the full suite
  before finishing.

## Code conventions

- Follow existing package boundaries and local naming patterns.
- Wrap errors with useful component/property context; preserve the local use
  of standard `errors` and `github.com/pkg/errors` rather than performing an
  unrelated migration.
- Use `testify` consistently with surrounding tests.
- Avoid changing public APIs, lifecycle ordering, component selection, or
  configuration semantics as an incidental refactor.
- Preserve unrelated working-tree changes and ignored local files.

When user-visible behavior changes, update `README.md` and `README_zh.md`
together. When framework usage or extension behavior changes, also audit the
repository skills under `skills/` and validate each changed skill with the
`skill-creator` validator.

## Code review rules

- Flag any path that silently accepts a function as a component.
- Flag configuration binding implemented outside the existing property
  post-processors unless the change explicitly replaces that architecture.
- Flag unsynchronized state accessed by concurrent definition scanning.
- Flag tests that depend on leaked package-global registration state.
