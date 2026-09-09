package container

import (
	"reflect"
	"testing"

	"github.com/go-kid/ioc/component_definition"
	"github.com/stretchr/testify/assert"
)

type optionService interface {
	Serve()
}

type optionComponent struct{}

func (*optionComponent) Serve()          {}
func (*optionComponent) Ping()           {}
func (*optionComponent) Kind() string    { return "worker" }
func (*optionComponent) Number() int     { return 42 }
func (*optionComponent) Tags() []string  { return []string{"a", "b"} }
func (*optionComponent) Nothing()        {}
func (*optionComponent) WithArg(int) int { return 1 }
func (*optionComponent) Action(int)      {}

func TestBooleanOptions(t *testing.T) {
	meta := component_definition.NewMeta(&optionComponent{})
	always := func(boolValue bool) Option {
		return func(*component_definition.Meta) bool { return boolValue }
	}

	assert.True(t, And()(meta))
	assert.False(t, Or()(meta))
	assert.True(t, And(always(true), always(true))(meta))
	assert.False(t, And(always(true), always(false))(meta))
	assert.True(t, Or(always(false), always(true))(meta))
	assert.False(t, Or(always(false), always(false))(meta))
}

func TestTypeAndInterfaceOptions(t *testing.T) {
	meta := component_definition.NewMeta(&optionComponent{})
	assert.True(t, Type(reflect.TypeOf(&optionComponent{}))(meta))
	assert.False(t, Type(reflect.TypeOf(optionComponent{}))(meta))

	interfaceType := reflect.TypeOf((*optionService)(nil)).Elem()
	assert.True(t, InterfaceType(interfaceType)(meta))
	assert.True(t, Interface(new(optionService))(meta))
}

func TestFunctionOptions(t *testing.T) {
	meta := component_definition.NewMeta(&optionComponent{})

	assert.True(t, FuncName("Ping")(meta))
	assert.False(t, FuncName("Kind")(meta))
	assert.False(t, FuncName("Action")(meta))
	assert.False(t, FuncName("Missing")(meta))

	assert.True(t, FuncNameAndResult("Kind", "worker")(meta))
	assert.True(t, FuncNameAndResult("Kind", "*")(meta))
	assert.False(t, FuncNameAndResult("Kind", "other")(meta))
	assert.True(t, FuncNameAndResult("Number", "42")(meta))
	assert.True(t, FuncNameAndResult("Tags", `["a","b"]`)(meta))
	assert.True(t, FuncNameAndResult("Nothing", "")(meta))
	assert.False(t, FuncNameAndResult("Nothing", "value")(meta))
	assert.False(t, FuncNameAndResult("WithArg", "1")(meta))
	assert.False(t, FuncNameAndResult("Missing", "")(meta))
}
