package config_exporter

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/factory/processors"
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
	GetConfig() properties.Properties
}

const (
	Append           = mode.M1
	OnlyNew          = mode.M2
	AnnotationSource = mode.M3
)

type postProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	configure configure.Configure
	pm        properties.Properties
	quoteEl   el.Helper
	exprEl    el.Helper
	mode      mode.Mode
}

func (d *postProcessor) PostProcessComponentFactory(factory factory.Factory) error {
	d.configure = factory.GetConfigure()
	return nil
}

func NewConfigExporter(mode mode.Mode) ConfigExporter {
	return &postProcessor{
		pm:      properties.New(),
		quoteEl: el.NewQuote(),
		exprEl:  el.NewExpr(),
		mode:    mode,
	}
}

func (d *postProcessor) Order() int {
	return -1
}

type configWrap struct {
	path    string
	raw     any
	rawType reflect.Type
}

func (d *postProcessor) PostProcessBeforeInstantiation(m *component_definition.Meta, componentName string) (any, error) {
	var w []*configWrap
	for _, prop := range m.GetAllProperties() {
		if prop.PropertyType != component_definition.PropertyTypeConfiguration {
			continue
		}
		if d.quoteEl.MatchString(prop.TagVal) {
			contents := d.quoteEl.FindAllContent(prop.TagVal)
			for _, content := range contents {
				spExp := strings.SplitN(content, ":", 2)
				exp := spExp[0]
				if len(spExp) != 2 {
					w = append(w, &configWrap{
						path:    exp,
						raw:     reflectx.ZeroValue(prop.Type),
						rawType: prop.Type,
					})
					continue
				}
				//parse tag default value
				if defaultVal := spExp[1]; defaultVal == "" {
					w = append(w, &configWrap{
						path:    exp,
						raw:     reflectx.ZeroValue(prop.Type),
						rawType: prop.Type,
					})
				} else {
					expVal, err := strconv2.ParseAny(defaultVal)
					if err != nil {
						return "", fmt.Errorf("parse config quote default value %s error: %v", defaultVal, err)
					}
					w = append(w, &configWrap{
						path:    exp,
						raw:     expVal,
						rawType: prop.Type,
					})
				}
			}
		} else if !d.exprEl.MatchString(prop.TagVal) && prop.Tag == definition.PropTag {
			w = append(w, &configWrap{
				path:    prop.TagVal,
				raw:     reflectx.ZeroValue(prop.Type),
				rawType: prop.Type,
			})
		}
	}

	for _, wrap := range w {
		if d.mode.Eq(AnnotationSource) {
			annoPath := "Source." + wrap.path
			if sources, ok := d.pm.Get(annoPath); ok {
				d.pm.Set(annoPath, append(sources.([]string), componentName))
			} else {
				d.pm.Set(annoPath, []string{componentName})
			}
		}
		if origin := d.configure.Get(wrap.path); origin != nil {
			if d.mode.Eq(OnlyNew) {
				continue
			}
			if d.mode.Eq(Append) {
				wrap.raw = origin
			}
		}

		switch wrap.rawType.Kind() {
		case reflect.Pointer, reflect.Struct, reflect.Map:
			subRaw, err := toMap(wrap.raw)
			if err != nil {
				return nil, err
			}
			subProp := properties.NewFromMap(subRaw)
			for p, a := range subProp {
				p := wrap.path + "." + p
				d.pm.Set(p, a)
			}
		default:
			d.pm.Set(wrap.path, wrap.raw)
		}
	}
	return m.Raw, nil
}

func (d *postProcessor) GetConfig() properties.Properties {
	return d.pm
}

func toMap(a any) (map[string]any, error) {
	bytes, err := yaml.Marshal(a)
	if err != nil {
		return nil, err
	}
	var subRaw = make(map[string]any)
	err = yaml.Unmarshal(bytes, subRaw)
	if err != nil {
		return nil, err
	}
	return subRaw, nil
}
