package config_exporter

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/factory/processors"
	"github.com/go-kid/ioc/util/el"
	"github.com/go-kid/ioc/util/properties"
	"github.com/go-kid/ioc/util/reflectx"
	"github.com/go-kid/ioc/util/strconv2"
	"strings"
)

type ConfigExporter interface {
	GetConfig() properties.Properties
}

type postProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	definition.PriorityComponent
	pm      properties.Properties
	quoteEl el.Helper
	exprEl  el.Helper
}

func NewConfigExporter() ConfigExporter {
	return &postProcessor{
		pm:      properties.New(),
		quoteEl: el.NewQuote(),
		exprEl:  el.NewExpr(),
	}
}

func (d *postProcessor) Order() int {
	return -1
}

type configWrap struct {
	path string
	raw  any
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
						path: exp,
						raw:  reflectx.ZeroValue(prop.Type),
					})
					continue
				}
				//parse tag default value
				if defaultVal := spExp[1]; defaultVal == "" {
					w = append(w, &configWrap{
						path: exp,
						raw:  reflectx.ZeroValue(prop.Type),
					})
				} else {
					expVal, err := strconv2.ParseAny(defaultVal)
					if err != nil {
						return "", fmt.Errorf("parse config quote default value %s error: %v", defaultVal, err)
					}
					w = append(w, &configWrap{
						path: exp,
						raw:  expVal,
					})
				}
			}
		} else if !d.exprEl.MatchString(prop.TagVal) && prop.Tag == definition.PropTag {
			w = append(w, &configWrap{
				path: prop.TagVal,
				raw:  reflectx.ZeroValue(prop.Type),
			})
		}
	}
	for _, wrap := range w {
		d.pm.Set(wrap.path, wrap.raw)
	}
	return m.Raw, nil
}

func (d *postProcessor) GetConfig() properties.Properties {
	return d.pm
}
