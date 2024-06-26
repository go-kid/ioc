package container

import (
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/strconv2"
	"reflect"
)

type Option func(m *component_definition.Meta) bool

func Or(opts ...Option) Option {
	return func(m *component_definition.Meta) bool {
		for _, opt := range opts {
			if opt(m) {
				return true
			}
		}
		return false
	}
}

func And(opts ...Option) Option {
	return func(m *component_definition.Meta) bool {
		for _, opt := range opts {
			if !opt(m) {
				return false
			}
		}
		return true
	}
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
		if mt, ok := m.Type.MethodByName(fn); ok {
			return mt.Type.NumOut() == 0 && m.Value.MethodByName(fn).IsValid()
		}
		return false
	}
}

func FuncNameAndResult(fn, result string) Option {
	return func(m *component_definition.Meta) bool {
		if method := m.Value.MethodByName(fn); method.IsValid() {
			if result == "*" {
				return true
			}
			results := method.Call(nil)
			if len(results) < 1 {
				return result == ""
			} else {
				parseAny, err := strconv2.ParseAny(result)
				if err != nil {
					parseAny = result
				}
				return results[0].Interface() == parseAny
			}
		}
		return false
	}
}
