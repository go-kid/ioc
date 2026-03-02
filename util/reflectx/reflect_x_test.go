package reflectx

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeId(t *testing.T) {
	var visited []string
	WalkField(T{TComponent: &TComponent{}}, func(parent *Node, field reflect.StructField, value reflect.Value) error {
		id := TypeId(field.Type)
		assert.NotEmpty(t, id)
		visited = append(visited, id)
		return nil
	})
	assert.NotEmpty(t, visited, "should visit at least one field")
}

func TestTypeId_Pointer(t *testing.T) {
	typ := reflect.TypeOf(&TComponent{})
	id := TypeId(typ)
	assert.True(t, strings.HasSuffix(id, "TComponent"))
	assert.False(t, strings.Contains(id, "*"))
}

func TestTypeId_NonPointer(t *testing.T) {
	typ := reflect.TypeOf(TComponent{})
	id := TypeId(typ)
	assert.True(t, strings.HasSuffix(id, "TComponent"))
}

func TestIsImplement(t *testing.T) {
	type testIface interface{ Foo() }
	type testImpl struct{}
	func() { /* testImpl does not implement testIface */ }()
	assert.False(t, IsImplement(&testImpl{}, new(testIface)))
}

func TestIsTypeImplement(t *testing.T) {
	typ := reflect.TypeOf("")
	iface := new(interface{ String() string })
	assert.False(t, IsTypeImplement(typ, iface))
}
