package component_definition

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/go-kid/ioc/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type definitionDependency struct{ Name string }

type definitionConfig struct {
	Host    string        `yaml:"host"`
	Timeout time.Duration `yaml:"timeout"`
}

type definitionTarget struct {
	Dependency   *definitionDependency
	Dependencies []*definitionDependency
	Config       definitionConfig
	hidden       string
}

type namedDefinitionComponent struct{ alias string }

func (c *namedDefinitionComponent) Naming() string { return c.alias }

type primaryDefinitionComponent struct {
	definition.WirePrimaryComponent
}

type prototypeDefinitionComponent struct{}

func (*prototypeDefinitionComponent) Scope() string { return definition.ScopePrototype }

func TestTagArgParseAndMutation(t *testing.T) {
	args := make(TagArg)
	tag := args.Parse("service,required=false,qualifier=blue green,validate=min=1 max=2")

	assert.Equal(t, "service", tag)
	assert.True(t, args.Has(ArgRequired, "false"))
	assert.True(t, args.Has(ArgQualifier, "green"))
	qualifiers, ok := args.Find(ArgQualifier)
	require.True(t, ok)
	assert.Equal(t, []string{"blue", "green"}, qualifiers)

	args.Add(ArgQualifier, "canary")
	assert.True(t, args.Has(ArgQualifier, "canary"))
	args.Set("custom", "value")
	assert.Equal(t, ".Custom(value).Qualifier(blue,green,canary).Required(false).Validate(min=1,max=2)", args.String())
}

func TestMetaScansOnlySettableFieldsAndTracksScope(t *testing.T) {
	meta := NewMeta(&definitionTarget{})
	assert.Len(t, meta.Fields, 3)
	assert.Equal(t, []string{"Dependency", "Dependencies", "Config"}, []string{
		meta.Fields[0].StructField.Name,
		meta.Fields[1].StructField.Name,
		meta.Fields[2].StructField.Name,
	})
	assert.True(t, meta.IsSingleton())
	assert.False(t, meta.IsPrototype())

	prototype := NewMeta(&prototypeDefinitionComponent{})
	assert.False(t, prototype.IsSingleton())
	assert.True(t, prototype.IsPrototype())
}

func TestMetaAliasesAndProxies(t *testing.T) {
	origin := NewMeta(&namedDefinitionComponent{alias: "original"})
	assert.True(t, origin.IsAlias())
	assert.Equal(t, "original", origin.Name())

	proxyValue := &namedDefinitionComponent{alias: "proxy-alias"}
	called := false
	proxy, err := CreateProxy(origin, "original", proxyValue, func(name string, meta *Meta) error {
		called = true
		assert.Equal(t, "original", name)
		assert.Same(t, proxyValue, meta.Raw)
		return nil
	})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, "original", proxy.Name())
	assert.Same(t, origin, proxy.ProxyMeta)

	wantErr := errors.New("interceptor failed")
	_, err = CreateProxy(origin, "original", proxyValue, func(string, *Meta) error { return wantErr })
	assert.ErrorIs(t, err, wantErr)

	replacement := &namedDefinitionComponent{alias: "replacement"}
	origin.UseProxy(replacement)
	assert.Same(t, replacement, origin.Raw)
	require.NotNil(t, origin.ProxyMeta)
}

func TestSelectBestCandidate(t *testing.T) {
	aliased := NewMeta(&namedDefinitionComponent{alias: "alias"})
	plain := NewMeta(&definitionDependency{})
	primary := NewMeta(&primaryDefinitionComponent{})

	assert.Same(t, aliased, SelectBestCandidate([]*Meta{aliased}))
	assert.Same(t, plain, SelectBestCandidate([]*Meta{aliased, plain}))
	assert.Same(t, primary, SelectBestCandidate([]*Meta{aliased, plain, primary}))
}

func TestPropertyInjectsPointerAndSlice(t *testing.T) {
	target := &definitionTarget{}
	targetMeta := NewMeta(target)
	depA := NewMeta(&definitionDependency{Name: "a"})
	depB := NewMeta(&definitionDependency{Name: "b"})

	pointerProperty := NewProperty(findField(t, targetMeta, "Dependency"), PropertyTypeComponent, "wire", "")
	require.NoError(t, pointerProperty.Inject([]*Meta{depA}))
	assert.Same(t, depA.Raw, target.Dependency)
	assert.Equal(t, []string{targetMeta.Name()}, depA.GetDependents())

	sliceProperty := NewProperty(findField(t, targetMeta, "Dependencies"), PropertyTypeComponent, "wire", "")
	require.NoError(t, sliceProperty.Inject([]*Meta{depA, depB}))
	assert.Equal(t, []*definitionDependency{depA.Raw.(*definitionDependency), depB.Raw.(*definitionDependency)}, target.Dependencies)
	assert.Len(t, sliceProperty.Injects, 2)
}

func TestPropertyInjectionValidation(t *testing.T) {
	targetMeta := NewMeta(&definitionTarget{})
	field := findField(t, targetMeta, "Dependency")

	required := NewProperty(field, PropertyTypeComponent, "wire", "")
	assert.ErrorContains(t, required.Inject(nil), "not found available components")
	assert.ErrorContains(t, required.Inject([]*Meta{targetMeta}), "self inject not allowed")

	optional := NewProperty(field, PropertyTypeComponent, "wire", ",required=false")
	require.NoError(t, optional.Inject(nil))
	require.NoError(t, optional.Inject([]*Meta{targetMeta}))

	notComponent := NewProperty(field, PropertyTypeConfiguration, "prefix", "app")
	assert.ErrorContains(t, notComponent.Inject(nil), "not allowed to inject")
}

func TestPropertyUnmarshalsConfiguration(t *testing.T) {
	target := &definitionTarget{}
	meta := NewMeta(target)
	property := NewProperty(findField(t, meta, "Config"), PropertyTypeConfiguration, "prefix", "app")
	property.SetConfiguration("app", map[string]any{"host": "localhost"})

	require.NoError(t, property.Unmarshall(map[string]any{
		"host":    "localhost",
		"timeout": "2s",
	}))
	assert.Equal(t, definitionConfig{Host: "localhost", Timeout: 2 * time.Second}, target.Config)
	assert.Equal(t, map[string]any{"host": "localhost"}, property.Configurations["app"])

	meta.SetProperties(property)
	assert.Equal(t, []*Property{property}, meta.GetConfigurationProperties())
	assert.Contains(t, meta.GetAllProperties(), property)

	componentProperty := NewProperty(findField(t, meta, "Dependency"), PropertyTypeComponent, "wire", "")
	assert.ErrorContains(t, componentProperty.Unmarshall(map[string]any{}), "not allowed to unmarshall")
}

func findField(t *testing.T, meta *Meta, name string) *Field {
	t.Helper()
	for _, field := range meta.Fields {
		if field.StructField.Name == name {
			return field
		}
	}
	t.Fatalf("field %s not found in %s", name, reflect.TypeOf(meta.Raw))
	return nil
}
