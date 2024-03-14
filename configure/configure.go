package configure

import (
	"fmt"
	"github.com/go-kid/ioc/configure/binder"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/reflectx"
	"os"
	"regexp"
	"strings"
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

func (c *configure) Initialize(metas ...*meta.Meta) error {
	if len(c.loaders) == 0 {
		syslog.Trace("not found config loaders, skip init configs")
		return nil
	}
	syslog.Info("start loading configs...")
	err := c.loadingConfigure()
	if err != nil {
		return fmt.Errorf("loading configure: %v", err)
	}
	syslog.Info("loading configure finished")

	for _, m := range metas {
		err = c.parseExpressionTags(m)
		if err != nil {
			return fmt.Errorf("parse expression tags: %v", err)
		}
		err = c.Binder.PropInject(m.GetConfigurationNodes())
		if err != nil {
			return fmt.Errorf("populate properties: %v", err)
		}
	}
	return nil
}

func (c *configure) loadingConfigure() error {
	sumLoaders := len(c.loaders)
	for i, l := range c.loaders {
		syslog.Tracef("config loaders start loading config %s ...[%d/%d]", reflectx.Id(l), i+1, sumLoaders)
		config, err := l.LoadConfig()
		if err != nil {
			return fmt.Errorf("config loader load config failed: %v", err)
		}
		syslog.Tracef("config loader loading finished ...[%d/%d]", i+1, sumLoaders)
		err = c.Binder.SetConfig(config)
		if err != nil {
			return fmt.Errorf("config binder set config failed: %v", err)
		}
	}
	return nil
}

func (c *configure) parseExpressionTags(m *meta.Meta) error {
	for _, prop := range m.GetConfigurationNodes() {
		rawTagVal := prop.TagVal
		expParsed := false
		r := regexp.MustCompile("\\$\\{[\\d\\w]+(\\.[\\d\\w]+)*(:[\\d\\w]*)?\\}")
		prop.TagVal = r.ReplaceAllStringFunc(prop.TagVal, func(s string) string {
			expParsed = true
			exp := s[2 : len(s)-1]
			spExp := strings.SplitN(exp, ":", 2)
			exp = spExp[0]
			expVal := c.Binder.Get(exp)
			if expVal == nil {
				if len(spExp) == 2 {
					return spExp[1]
				}
				syslog.Fatalf("config path '%s' used by expression tag value is missing", exp)
			}
			switch expVal.(type) {
			case string:
				return expVal.(string)
			default:
				return fmt.Sprintf("%v", expVal)
			}
		})
		if expParsed {
			syslog.Tracef("parse expression tag value '%s' -> '%s'", rawTagVal, prop.TagVal)
		}
	}
	return nil
}
