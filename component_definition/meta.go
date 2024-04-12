package component_definition

import (
	"fmt"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/sync2"
	"reflect"
)

type PropertyType string

const (
	PropertyTypeConfiguration PropertyType = "configuration"
	PropertyTypeComponent     PropertyType = "component"
)

type Meta struct {
	*Base
	ProxyMeta *Meta
	name      string
	alias     string

	Raw    interface{}
	Fields []*Field

	dependentSet *sync2.Map[string, struct{}]
	Dependent    []*Meta

	propertyGroup map[PropertyType][]*Property
}

func NewMeta(c any) *Meta {
	base := NewBase(c)
	name, alias := framework_helper.GetComponentNameWithAlias(c)
	m := &Meta{
		Base:          base,
		name:          name,
		alias:         alias,
		Raw:           c,
		dependentSet:  sync2.New[string, struct{}](),
		propertyGroup: make(map[PropertyType][]*Property),
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

func (m *Meta) dependOn(dependent *Meta) {
	_, loaded := m.dependentSet.LoadOrStore(dependent.ID(), struct{}{})
	if !loaded {
		m.Dependent = append(m.Dependent, dependent)
	}
}

func (m *Meta) GetDependents() (names []string) {
	for _, meta := range m.Dependent {
		names = append(names, meta.Name())
	}
	return
}

func (m *Meta) SetProperties(properties ...*Property) {
	for _, prop := range properties {
		m.propertyGroup[prop.PropertyType] = append(m.propertyGroup[prop.PropertyType], prop)
	}
}

func (m *Meta) GetProperties(t PropertyType) []*Property {
	return m.propertyGroup[t]
}

func (m *Meta) GetComponentProperties() []*Property {
	return m.GetProperties(PropertyTypeComponent)
}

func (m *Meta) GetConfigurationProperties() []*Property {
	return m.GetProperties(PropertyTypeConfiguration)
}

func (m *Meta) GetAllProperties() []*Property {
	var props []*Property
	for _, groupNodes := range m.propertyGroup {
		props = append(props, groupNodes...)
	}
	return props
}

func (m *Meta) UseProxy(origin any) {
	m.ProxyMeta = &Meta{
		Base:          m.Base,
		ProxyMeta:     nil,
		name:          m.name,
		alias:         m.alias,
		Raw:           m.Raw,
		Fields:        m.Fields,
		dependentSet:  m.dependentSet,
		Dependent:     m.Dependent,
		propertyGroup: m.propertyGroup,
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
	return true
}
