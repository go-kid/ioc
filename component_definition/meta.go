package component_definition

import (
	"fmt"
	"reflect"

	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/go-kid/ioc/util/reflectx"
)

type PropertyType string

const (
	PropertyTypeConfiguration PropertyType = "Configuration"
	PropertyTypeComponent     PropertyType = "Component"
)

type Meta struct {
	*Base
	PropertyManager
	DependencyTracker
	ProxyMeta *Meta
	name      string
	alias     string

	Raw    interface{}
	Fields []*Field
}

func NewMeta(c any) *Meta {
	base := NewBase(c)
	name, alias := framework_helper.GetComponentNameWithAlias(c)
	m := &Meta{
		Base:              base,
		PropertyManager:   newPropertyManager(),
		DependencyTracker: newDependencyTracker(),
		name:              name,
		alias:             alias,
		Raw:               c,
	}
	m.scanFields(NewHolder(m))
	return m
}

func (m *Meta) ID() string {
	return fmt.Sprintf("%s(0x%x)", m.String(), m.Value.Pointer())
}

func (m *Meta) String() string {
	if m.IsAlias() {
		return fmt.Sprintf("%s(alias=%s)", m.name, m.alias)
	}
	return m.name
}

func (m *Meta) Name() string {
	if m.IsAlias() {
		return m.alias
	}
	return m.name
}

func (m *Meta) SetName(name string) {
	if name != m.name {
		m.alias = name
	}
}

func (m *Meta) IsAlias() bool {
	return m.alias != ""
}

func (m *Meta) IsSelf(o *Meta) bool {
	return m.Value.Pointer() == o.originAddress
}

func (m *Meta) UseProxy(origin any) {
	m.ProxyMeta = &Meta{
		Base:              m.Base,
		PropertyManager:   m.PropertyManager,
		DependencyTracker: m.DependencyTracker,
		ProxyMeta:         nil,
		name:              m.name,
		alias:             m.alias,
		Raw:               m.Raw,
		Fields:            m.Fields,
	}
	m.Base = NewBase(origin)
	m.Raw = origin
}

type Interceptor = func(name string, m *Meta) error

func CreateProxy(origin *Meta, name string, newComponent any, interceptors ...Interceptor) (*Meta, error) {
	nm := NewMeta(newComponent)
	nm.SetName(name)
	nm.ProxyMeta = origin
	for _, interceptor := range interceptors {
		err := interceptor(name, nm)
		if err != nil {
			return nil, err
		}
	}
	return nm, nil
}

func (m *Meta) scanFields(holder *Holder) {
	_ = reflectx.ForEachFieldV2(holder.Type, holder.Value, false, func(field reflect.StructField, value reflect.Value) error {
		var base = &Base{
			Type:  field.Type,
			Value: value,
		}
		//if is embed struct, find inside
		if field.Anonymous && field.Tag == "" && field.Type.Kind() == reflect.Struct {
			m.scanFields(NewEmbedHolder(base, holder))
			return nil
		}

		if !value.CanSet() {
			return nil
		}
		m.Fields = append(m.Fields, &Field{
			Base:        base,
			Holder:      holder,
			StructField: field,
		})
		return nil
	})
}

func (m *Meta) IsSingleton() bool {
	if sc, ok := m.Raw.(definition.ScopeComponent); ok {
		return sc.Scope() != definition.ScopePrototype
	}
	return true
}

func (m *Meta) IsPrototype() bool {
	if sc, ok := m.Raw.(definition.ScopeComponent); ok {
		return sc.Scope() == definition.ScopePrototype
	}
	return false
}
