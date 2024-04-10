package config_exporter

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/el"
	"github.com/go-kid/ioc/util/mode"
	"github.com/go-kid/ioc/util/properties"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/strconv2"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
)

type ConfigExporter interface {
	GetConfig(mode mode.Mode) properties.Properties
	GetConfigWraps() []*ConfigWrap
}

const (
	Append           = mode.M1
	OnlyNew          = mode.M2
	AnnotationSource = mode.M3
	AnnotationArgs   = mode.M4
)

type postProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	configure configure.Configure
	quoteEl   el.Helper
	exprEl    el.Helper
	wraps     []*ConfigWrap
}

func (d *postProcessor) PostProcessComponentFactory(factory factory.Factory) error {
	d.configure = factory.GetConfigure()
	return nil
}

func NewConfigExporter() ConfigExporter {
	return &postProcessor{
		quoteEl: el.NewQuote(),
		exprEl:  el.NewExpr(),
	}
}

func (d *postProcessor) Order() int {
	return -1
}

type ConfigWrap struct {
	ComponentName string
	Property      *component_definition.Property
	Prefix        string
	RealValue     any
}

func (d *postProcessor) PostProcessBeforeInstantiation(m *component_definition.Meta, componentName string) (any, error) {
	for _, prop := range m.GetAllProperties() {
		if prop.PropertyType != component_definition.PropertyTypeConfiguration {
			continue
		}
		var (
			path string
			raw  any
		)

		if d.quoteEl.MatchString(prop.TagVal) {
			contents := d.quoteEl.FindAllContent(prop.TagVal)
			for _, content := range contents {
				spExp := strings.SplitN(content, ":", 2)
				path = spExp[0]
				if len(spExp) == 2 && spExp[1] != "" {
					defaultVal := spExp[1]
					expVal, err := strconv2.ParseAny(defaultVal)
					if err != nil {
						return "", fmt.Errorf("parse config quote default value %s error: %v", defaultVal, err)
					}
					raw = expVal
				}
				d.wraps = append(d.wraps, &ConfigWrap{
					ComponentName: componentName,
					Property:      prop,
					Prefix:        path,
					RealValue:     raw,
				})
			}
		} else if !d.exprEl.MatchString(prop.TagVal) && prop.Tag == definition.PropTag {
			path = prop.TagVal
			d.wraps = append(d.wraps, &ConfigWrap{
				ComponentName: componentName,
				Property:      prop,
				Prefix:        path,
			})
		}
	}

	return m.Raw, nil
}

func (d *postProcessor) GetConfigWraps() []*ConfigWrap {
	return d.wraps
}

func (d *postProcessor) GetConfig(mode mode.Mode) properties.Properties {
	pm := properties.New()
	for _, wrap := range d.wraps {
		value := wrap.RealValue
		if value == nil {
			value = reflectx.ZeroValue(wrap.Property.Type)
		}
		prefix := wrap.Prefix
		if mode.Eq(AnnotationArgs) {
			wrap.Property.Args().ForEach(func(argType component_definition.ArgType, args []string) {
				pm.Set(fmt.Sprintf("%s@Args.%s", prefix, argType), args)
			})
		}
		if mode.Eq(AnnotationSource) {
			annoPath := fmt.Sprintf("%s@Sources", prefix)
			if sources, ok := pm.Get(annoPath); ok {
				pm.Set(annoPath, append(sources.([]string), wrap.ComponentName))
			} else {
				pm.Set(annoPath, []string{wrap.ComponentName})
			}
		}
		if origin := d.configure.Get(prefix); origin != nil {
			if mode.Eq(OnlyNew) {
				continue
			}
			if mode.Eq(Append) {
				value = origin
			}
		}

		switch wrap.Property.Type.Kind() {
		case reflect.Pointer, reflect.Struct, reflect.Map:
			subRaw := toMap(value)
			subProp := properties.NewFromMap(subRaw)
			for p, a := range subProp {
				pm.Set(prefix+"."+p, a)
			}
		default:
			pm.Set(prefix, value)
		}
	}
	return pm
}

func toMap(a any) map[string]any {
	bytes, err := yaml.Marshal(a)
	if err != nil {
		syslog.Panicf("yaml marshal error: %#v, %+v", a, err)
	}
	var subRaw = make(map[string]any)
	err = yaml.Unmarshal(bytes, subRaw)
	if err != nil {
		syslog.Panicf("yaml unmarshal error: %s, %+v", string(bytes), err)
	}
	return subRaw
}
