package factory

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
)

type Option func(m *component_definition.Meta) bool

func Accept(m *component_definition.Meta, opts ...Option) bool {
	for _, opt := range opts {
		if !opt(m) {
			return false
		}
	}
	return true
}

func Type(typ reflect.Type) Option {
	return func(m *component_definition.Meta) bool {
		return m.Value.Type() == typ
	}
}

func InterfaceType(typ reflect.Type) Option {
	return func(m *component_definition.Meta) bool {
		return m.Value.Type().Implements(typ)
	}
}

func Interface(a any) Option {
	return func(m *component_definition.Meta) bool {
		return reflectx.IsTypeImplement(m.Value.Type(), a)
	}
}

func FuncName(fn string) Option {
	return func(m *component_definition.Meta) bool {
		return m.Value.MethodByName(fn).IsValid()
	}
}

func FuncNameAndResult(fn, result string) Option {
	return func(m *component_definition.Meta) bool {
		if method := m.Value.MethodByName(fn); method.IsValid() {
			if result == "-" {
				return true
			}
			results := method.Call(nil)
			if len(results) < 1 {
				return result == ""
			} else {
				return results[0].String() == result
			}
		}
		return false
	}
}
