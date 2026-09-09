package support

import (
	"errors"
	"testing"

	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/container"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/stretchr/testify/assert"
)

// --- test component types ---

type testComponent struct{}

type testComponentB struct{}

// --- registry tests ---

func TestRegistry_RegisterSingleton_StructPointer(t *testing.T) {
	r := NewRegistry().(*registry)
	inst := &testComponent{}
	r.RegisterSingleton(inst)

	name := framework_helper.GetComponentName(inst)
	got, err := r.GetSingleton(name)
	assert.NoError(t, err)
	assert.Same(t, inst, got)
}

func TestRegistry_RegisterSingleton_FunctionPanics(t *testing.T) {
	r := NewRegistry().(*registry)
	assert.PanicsWithValue(t,
		"constructor registration is not supported; register the constructed component pointer",
		func() { r.RegisterSingleton(func() *testComponent { return &testComponent{} }) },
	)
}

func TestRegistry_RegisterSingleton_DuplicateSameInstance(t *testing.T) {
	r := NewRegistry().(*registry)
	inst := &testComponent{}
	r.RegisterSingleton(inst)
	// same instance is OK
	assert.NotPanics(t, func() {
		r.RegisterSingleton(inst)
	})
}

func TestRegistry_RegisterSingleton_DuplicateDifferentInstance_Panics(t *testing.T) {
	// Use a struct with a field so compiler doesn't coalesce &dupComp{} to same address.
	// Empty struct literals may be optimized to the same pointer.
	type dupComp struct{ x int }
	r := NewRegistry().(*registry)
	inst1 := &dupComp{x: 1}
	inst2 := &dupComp{x: 2}
	r.RegisterSingleton(inst1)
	assert.Panics(t, func() {
		r.RegisterSingleton(inst2)
	})
}

func TestRegistry_ContainsSingleton(t *testing.T) {
	r := NewRegistry().(*registry)
	inst := &testComponent{}
	name := framework_helper.GetComponentName(inst)

	assert.False(t, r.ContainsSingleton(name))
	r.RegisterSingleton(inst)
	assert.True(t, r.ContainsSingleton(name))
}

func TestRegistry_GetSingletonNames(t *testing.T) {
	r := NewRegistry().(*registry)
	instA := &testComponent{}
	instB := &testComponentB{}
	r.RegisterSingleton(instA)
	r.RegisterSingleton(instB)

	names := r.GetSingletonNames()
	assert.Len(t, names, 2)
	assert.Contains(t, names, framework_helper.GetComponentName(instA))
	assert.Contains(t, names, framework_helper.GetComponentName(instB))
}

func TestRegistry_GetSingletonCount(t *testing.T) {
	r := NewRegistry().(*registry)
	assert.Equal(t, 0, r.GetSingletonCount())

	r.RegisterSingleton(&testComponent{})
	assert.Equal(t, 1, r.GetSingletonCount())

	r.RegisterSingleton(&testComponentB{})
	assert.Equal(t, 2, r.GetSingletonCount())
}

func TestRegistry_GetSingleton_NotFound(t *testing.T) {
	r := NewRegistry().(*registry)
	_, err := r.GetSingleton("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not exist")
}

// --- defaultDefinitionRegistry tests ---

func TestDefaultDefinitionRegistry_RegisterMeta_GetMetaByName(t *testing.T) {
	reg := DefaultDefinitionRegistry().(*defaultDefinitionRegistry)
	meta := component_definition.NewMeta(&testComponent{})
	meta.SetName("testComp")

	reg.RegisterMeta(meta)
	got := reg.GetMetaByName("testComp")
	assert.Same(t, meta, got)
}

func TestDefaultDefinitionRegistry_GetMetaByName_NotFound(t *testing.T) {
	reg := DefaultDefinitionRegistry().(*defaultDefinitionRegistry)
	got := reg.GetMetaByName("nonexistent")
	assert.Nil(t, got)
}

func TestDefaultDefinitionRegistry_GetMetaOrRegister(t *testing.T) {
	reg := DefaultDefinitionRegistry().(*defaultDefinitionRegistry)
	name := "myComponent"

	first := reg.GetMetaOrRegister(name, &testComponent{})
	assert.NotNil(t, first)
	assert.Equal(t, name, first.Name())

	second := reg.GetMetaOrRegister(name, &testComponent{})
	assert.Same(t, first, second)
}

// --- defaultSingletonComponentRegistry tests ---

func TestDefaultSingletonComponentRegistry_AddSingleton_GetSingleton(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	meta := component_definition.NewMeta(&testComponent{})
	meta.SetName("comp1")

	reg.AddSingleton("comp1", meta)
	got, err := reg.GetSingleton("comp1", false)
	assert.NoError(t, err)
	assert.Same(t, meta, got)
}

func TestDefaultSingletonComponentRegistry_GetSingleton_NotFound(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	got, err := reg.GetSingleton("nonexistent", false)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestDefaultSingletonComponentRegistry_IsSingletonCurrentlyInCreation(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	assert.False(t, reg.IsSingletonCurrentlyInCreation("any"))
}

func TestDefaultSingletonComponentRegistry_RemoveSingleton(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	meta := component_definition.NewMeta(&testComponent{})
	meta.SetName("toRemove")

	reg.AddSingleton("toRemove", meta)
	got, _ := reg.GetSingleton("toRemove", false)
	assert.Same(t, meta, got)

	reg.RemoveSingleton("toRemove")
	got, _ = reg.GetSingleton("toRemove", false)
	assert.Nil(t, got)
}

func TestDefaultSingletonComponentRegistry_GetSingletonOrCreateByFactory(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	createCount := 0
	factory := container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
		createCount++
		m := component_definition.NewMeta(&testComponent{})
		m.SetName("factoryComp")
		return m, nil
	})

	first, err := reg.GetSingletonOrCreateByFactory("factoryComp", factory)
	assert.NoError(t, err)
	assert.NotNil(t, first)
	assert.Equal(t, 1, createCount)

	second, err := reg.GetSingletonOrCreateByFactory("factoryComp", factory)
	assert.NoError(t, err)
	assert.Same(t, first, second)
	assert.Equal(t, 1, createCount)
}

func TestDefaultSingletonComponentRegistry_FactoryErrorClearsCreationState(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	factoryErr := errors.New("create failed")

	got, err := reg.GetSingletonOrCreateByFactory("factoryComp", container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
		return nil, factoryErr
	}))
	assert.Nil(t, got)
	assert.ErrorIs(t, err, factoryErr)
	assert.False(t, reg.IsSingletonCurrentlyInCreation("factoryComp"))

	want := component_definition.NewMeta(&testComponent{})
	got, err = reg.GetSingletonOrCreateByFactory("factoryComp", container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
		return want, nil
	}))
	assert.NoError(t, err)
	assert.Same(t, want, got)
}

func TestDefaultSingletonComponentRegistry_EarlyReferenceCreatedOnce(t *testing.T) {
	reg := DefaultSingletonComponentRegistry().(*defaultSingletonComponentRegistry)
	want := component_definition.NewMeta(&testComponent{})
	createCount := 0
	reg.AddSingletonFactory("early", container.FuncSingletonFactory(func() (*component_definition.Meta, error) {
		createCount++
		return want, nil
	}))

	got, err := reg.GetSingleton("early", false)
	assert.NoError(t, err)
	assert.Nil(t, got)
	assert.Zero(t, createCount)

	got, err = reg.GetSingleton("early", true)
	assert.NoError(t, err)
	assert.Same(t, want, got)
	assert.Equal(t, 1, createCount)

	got, err = reg.GetSingleton("early", true)
	assert.NoError(t, err)
	assert.Same(t, want, got)
	assert.Equal(t, 1, createCount)
}
