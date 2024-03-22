package configure

import (
	"fmt"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"os"
	"sort"
)

type configure struct {
	Binder
	loaders            []Loader
	populateProcessors []PopulateProcessor
}

func NewConfigure() Configure {
	return &configure{}
}

func Default() Configure {
	c := NewConfigure()
	c.SetLoaders(loader.NewArgsLoader(os.Args))
	c.SetBinder(binder.NewViperBinder("yaml"))
	c.AddPopulateProcessors(
		new(executeExpressionPopulation),
		new(propPopulation),
		new(valuePopulation),
	)
	return c
}

func (c *configure) AddLoaders(loaders ...Loader) {
	c.loaders = append(c.loaders, loaders...)
}

func (c *configure) SetLoaders(loaders ...Loader) {
	c.loaders = loaders
}

func (c *configure) AddPopulateProcessors(processors ...PopulateProcessor) {
	c.populateProcessors = append(c.populateProcessors, processors...)
	sort.Slice(c.populateProcessors, func(i, j int) bool {
		return c.populateProcessors[i].Order() < c.populateProcessors[j].Order()
	})
}

func (c *configure) SetBinder(binder Binder) {
	c.Binder = binder
}

func (c *configure) Initialize() error {
	if len(c.loaders) == 0 {
		syslog.Trace("not found config loaders, skip init configs")
		return nil
	}
	syslog.Info("start loading configs...")
	err := c.loadConfigure()
	if err != nil {
		return fmt.Errorf("loading configure: %v", err)
	}
	syslog.Info("loading configure finished")
	return nil
}

func (c *configure) loadConfigure() error {
	sumLoaders := len(c.loaders)
	for i, l := range c.loaders {
		syslog.Tracef("config loaders start loading config %s ...[%d/%d]", reflectx.Id(l), i+1, sumLoaders)
		config, err := l.LoadConfig()
		if err != nil {
			return fmt.Errorf("config loader load config failed: %v", err)
		}
		if len(config) != 0 {
			err = c.Binder.SetConfig(config)
			if err != nil {
				return fmt.Errorf("config binder set config failed: %v", err)
			}
		}
		syslog.Tracef("config loader loading finished ...[%d/%d]", i+1, sumLoaders)
	}
	return nil
}

func (c *configure) PopulateProperties(metas ...*meta.Meta) error {
	for _, m := range metas {
		for _, node := range m.GetConfigurationNodes() {
			for _, processor := range c.populateProcessors {
				if processor.Filter(node) {
					syslog.Tracef("populate property %s.Value(%s)", node.ID(), node.TagVal)
					err := processor.Populate(c.Binder, node)
					if err != nil {
						return fmt.Errorf("populate config properties %s.Value(%s) error: %v", node.ID(), node.TagVal, err)
					}
				}
			}
		}
	}
	return nil
}
