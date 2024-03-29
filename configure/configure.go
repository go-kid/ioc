package configure

import (
	"fmt"
	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/syslog"
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
		c.logger().Trace("not config loaders found, skip initialize configure")
		return nil
	}
	c.logger().Info("start loading configurations...")
	err := c.loadConfigure()
	if err != nil {
		return fmt.Errorf("loading configurations: %v", err)
	}
	c.logger().Info("loading configurations finished")
	return nil
}

func (c *configure) loadConfigure() error {
	sumLoaders := len(c.loaders)
	for i, l := range c.loaders {
		c.logger().Tracef("config loader %T start loading configurations... [%d/%d]", l, i+1, sumLoaders)
		config, err := l.LoadConfig()
		if err != nil {
			return fmt.Errorf("config loader %T load config failed: %v", l, err)
		}
		if len(config) != 0 {
			c.logger().Tracef("config binder set configurations with size %d", len(config))
			err = c.Binder.SetConfig(config)
			if err != nil {
				return fmt.Errorf("config binder set configurations failed: %v\nraw configurations: %s", err, string(config))
			}
		}
	}
	return nil
}

func (c *configure) PopulateProperties(metas ...*component_definition.Meta) error {
	for _, m := range metas {
		for _, node := range m.GetConfigurationNodes() {
			for _, processor := range c.populateProcessors {
				if processor.Filter(node) {
					c.logger().Tracef("populate property %s.Value(%s)", node.ID(), node.TagVal)
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

func (c *configure) logger() syslog.Logger {
	return syslog.GetLogger().Pref("Configure")
}
