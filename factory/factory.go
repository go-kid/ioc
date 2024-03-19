package factory

import (
	"fmt"
	"github.com/go-kid/ioc/injector"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"regexp"
)

type defaultFactory struct {
	postInitFunc MetaFunc
	expReg       *regexp.Regexp
}

func Default() Factory {
	return &defaultFactory{
		expReg: regexp.MustCompile("\\#\\{[\\d\\w]+(\\.[\\d\\w]+)*(:[\\d\\w]*)?\\}"),
	}
}

func (f *defaultFactory) SetIfNilPostInitFunc(fn MetaFunc) {
	if f.postInitFunc == nil {
		f.postInitFunc = fn
	}
}

func (f *defaultFactory) Initialize(r registry.Registry, i injector.Injector, metas ...*meta.Meta) error {
	for _, m := range metas {
		err := f.wire(r, i, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *defaultFactory) wire(r registry.Registry, i injector.Injector, m *meta.Meta) error {
	syslog.Tracef("factory start initialize component %s", m.ID())
	if r.IsComponentInited(m.Name) {
		syslog.Tracef("component %s is already init, skip initialize", m.ID())
		return nil
	}
	r.ComponentInited(m.Name)

	if nodes := m.GetComponentNodes(); len(nodes) > 0 {
		syslog.Tracef("factory start inject dependencies %s", m.ID())
		//for _, node := range nodes {
		//	rawTagVal := node.TagVal
		//	expParsed := false
		//	node.TagVal = f.expReg.ReplaceAllStringFunc(node.TagVal, func(s string) string {
		//		expParsed = true
		//		exp := s[2 : len(s)-1]
		//		spExp := strings.SplitN(exp, ":", 2)
		//		exp = spExp[0]
		//		expMeta := r.GetComponentByName(exp)
		//		err := f.wire(r, i, expMeta)
		//		if err != nil {
		//			syslog.Fatalf("wire path '%s' depend on '%s' initialize failed: %v", err)
		//		}
		//		expVal := c.Binder.Get(exp)
		//		if expVal == nil {
		//			if len(spExp) == 2 {
		//				return spExp[1]
		//			}
		//			syslog.Fatalf("config path '%s' used by expression tag value is missing", exp)
		//		}
		//		switch expVal.(type) {
		//		case string:
		//			return expVal.(string)
		//		default:
		//			return fmt.Sprintf("%v", expVal)
		//		}
		//	})
		//	if expParsed {
		//		syslog.Tracef("parse expression tag value '%s' -> '%s'", rawTagVal, prop.TagVal)
		//	}
		//}
		err := i.DependencyInject(r, m.ID(), nodes)
		if err != nil {
			return fmt.Errorf("factory inject dependencies failed: %v", err)
		}

		for _, node := range nodes {
			for _, inject := range node.Injects {
				err = f.wire(r, i, inject)
				if err != nil {
					return err
				}
			}
		}
	}

	err := f.postInitFunc(m)
	if err != nil {
		return fmt.Errorf("factory initialize component %s failed: %v", m.ID(), err)
	}

	syslog.Tracef("factory initialized component %s", m.ID())
	return nil
}
