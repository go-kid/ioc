package configure

import (
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/framework_helper"
	"github.com/pkg/errors"
	"os"
)

type configure struct {
	Binder
	loaders []Loader
}

func NewConfigure() Configure {
	return &configure{}
}

func Default() Configure {
	c := NewConfigure()
	c.SetLoaders(loader.NewArgsLoader(os.Args))
	c.SetBinder(binder.NewViperBinder("yaml"))
	return c
}

func (c *configure) AddLoaders(loaders ...Loader) {
	c.loaders = append(c.loaders, loaders...)
}

func (c *configure) SetLoaders(loaders ...Loader) {
	c.loaders = loaders
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
		return err
	}
	c.logger().Info("loading configurations finished")
	return nil
}

func (c *configure) loadConfigure() error {
	c.loaders = framework_helper.SortOrderedComponents(c.loaders)
	sumLoaders := len(c.loaders)
	for i, l := range c.loaders {
		c.logger().Tracef("config loader %T start loading configurations... [%d/%d]", l, i+1, sumLoaders)
		config, err := l.LoadConfig()
		if err != nil {
			return errors.WithMessagef(err, "loader: %T", l)
		}
		if len(config) != 0 {
			c.logger().Tracef("config binder set configurations with size %d", len(config))
			err = c.Binder.SetConfig(config)
			if err != nil {
				return errors.WithMessagef(err, "raw configuration: %s", string(config))
			}
		}
	}
	return nil
}

func (c *configure) logger() syslog.Logger {
	return syslog.Pref("Configure")
}
