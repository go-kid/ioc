package component_definition

import (
	"fmt"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"reflect"
)

type Property struct {
	*Field
	PropertyType   PropertyType
	Tag            string
	TagVal         string
	Injects        []*Meta
	Configurations map[string]any
	args           TagArg
}

func NewProperty(field *Field, propType PropertyType, tag, tagVal string) *Property {
	parsedTagVal, arg := defaultPropertyArgs().Parse(tagVal)
	return &Property{
		Field:          field,
		PropertyType:   propType,
		Tag:            tag,
		TagVal:         parsedTagVal,
		Configurations: make(map[string]any),
		args:           arg,
	}
}

func defaultPropertyArgs() TagArg {
	return TagArg{
		ArgRequired: {"true"},
	}
}

func filter(metas []*Meta, f func(m *Meta) bool) []*Meta {
	var result = make([]*Meta, 0, len(metas))
	for _, m := range metas {
		if f(m) {
			result = append(result, m)
		}
	}
	return result
}

func (n *Property) ID() string {
	return fmt.Sprintf("%s.Tag(%s:'%s').Type(%s)", n.Field.ID(), n.Tag, n.TagVal, n.PropertyType)
}

func (n *Property) String() string {
	return n.ID()
}

func (n *Property) Inject(metas []*Meta) error {
	if n.PropertyType != PropertyTypeComponent {
		return errors.Errorf("property '%s' is not allowed to inject", n.ID())
	}
	isRequired := n.args.Has(ArgRequired, "true")
	if len(metas) == 0 {
		if isRequired {
			return errors.Errorf("inject %s: not found available components", n.ID())
		}
		return nil
	}

	//remove self-inject
	metas = filter(metas, func(m *Meta) bool {
		return !n.Holder.Meta.IsSelf(m)
	})
	if len(metas) == 0 {
		if isRequired {
			return errors.Errorf("inject %s:%s: self inject not allowed", n.ID(), n.Holder.Stack())
		}
		return nil
	}

	switch n.Type.Kind() {
	case reflect.Slice, reflect.Array:
		n.Value.Set(reflect.MakeSlice(n.Type, len(metas), len(metas)))
		for i, m := range metas {
			n.Value.Index(i).Set(m.Value)
			m.dependOn(n.Holder.Meta)
		}
	default:
		m := metas[0]
		n.Value.Set(m.Value)
		m.dependOn(n.Holder.Meta)
	}

	n.Injects = metas
	return nil
}

const (
	unmarshallArgTagName    = "mapper"
	unmarshallArgTimeLayout = "timeLayout"
)

func (n *Property) SetConfiguration(path string, configValue any) {
	n.Configurations[path] = configValue
}

func (n *Property) Unmarshall(configValue any) error {
	if n.PropertyType != PropertyTypeConfiguration {
		return errors.Errorf("property '%s' is not allowed to unmarshall configuration value", n.ID())
	}
	if configValue == nil {
		return nil
	}
	var hooks = []mapstructure.DecodeHookFunc{
		mapstructure.StringToTimeDurationHookFunc(),
	}
	if args, ok := n.Args().Find(unmarshallArgTimeLayout); ok {
		hooks = append(hooks, mapstructure.StringToTimeHookFunc(args[0]))
	}
	err := reflectx.SetValue(n.Value, func(a any) error {
		config := newDecodeConfig(a, hooks)
		if args, ok := n.Args().Find(unmarshallArgTagName); ok {
			config.TagName = args[0]
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return fmt.Errorf("create mapstructure decoder error: %v", err)
		}
		err = decoder.Decode(configValue)
		if err != nil {
			return fmt.Errorf("mapstructure decode %+v error: %v", configValue, err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "unmarshall property configuration failed")
	}
	return nil
}

func newDecodeConfig(v any, hooks []mapstructure.DecodeHookFunc) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook:           mapstructure.ComposeDecodeHookFunc(hooks...),
		ErrorUnused:          false,
		ErrorUnset:           false,
		ZeroFields:           false,
		WeaklyTypedInput:     true,
		Squash:               false,
		Metadata:             nil,
		Result:               v,
		TagName:              "yaml",
		IgnoreUntaggedFields: false,
		MatchName:            nil,
	}
}

func (n *Property) Args() TagArg {
	return n.args
}

func (n *Property) SetArg(t ArgType, val []string) {
	n.args[t] = val
}
func (n *Property) AppendArg(t ArgType, val []string) {
	n.args[t] = append(n.args[t], val...)
}

func (n *Property) SetArgs(a TagArg) {
	for argType, val := range a {
		n.SetArg(argType, val)
	}
}

func (n *Property) AppendArgs(a TagArg) {
	for argType, val := range a {
		n.AppendArg(argType, val)
	}
}
