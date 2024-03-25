package factory

import (
	"fmt"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/sync2"
)

type container struct {
	metaMaps              *sync2.Map[string, *meta.Meta]
	singletonObjects      *sync2.Map[string, *meta.Meta]
	earlySingletonObjects *sync2.Map[string, *meta.Meta]
	singletonFactories    *sync2.Map[string, FactoryMethod]
}

func (r *container) GetMetas(opts ...Option) []*meta.Meta {
	var metas = make([]*meta.Meta, 0)
	r.metaMaps.Range(func(k string, m *meta.Meta) bool {
		if Accept(m, opts...) {
			metas = append(metas, m)
		}
		return true
	})
	return metas
}

func (r *container) GetMetaByName(name string) *meta.Meta {
	if c, ok := r.metaMaps.Load(name); ok {
		return c
	}
	return nil
}

func (r *container) SetSingletonFactoryMethod(name string, method FactoryMethod) {
	r.singletonFactories.Store(name, method)
}

func (r *container) GetSingletonFactoryMethod(name string) (FactoryMethod, bool) {
	return r.singletonFactories.Load(name)
}

func (r *container) EarlyExportComponent(m *meta.Meta) {
	r.earlySingletonObjects.Store(m.Name, m)
	r.singletonFactories.Delete(m.Name)
}

func (r *container) GetEarlyExportComponent(name string) (*meta.Meta, bool) {
	return r.earlySingletonObjects.Load(name)
}

func (r *container) RemoveComponents(name string) {
	r.singletonObjects.Delete(name)
	syslog.Tracef("registry remove component %s", name)
}

func (r *container) IsComponentInited(name string) bool {
	_, loaded := r.singletonObjects.Load(name)
	return loaded
}

func (r *container) ComponentInited(name string) error {
	m, loaded := r.earlySingletonObjects.Load(name)
	if !loaded {
		return fmt.Errorf("component %s is not initiated", name)
	}
	r.singletonObjects.Store(name, m)
	r.earlySingletonObjects.Delete(name)
	//syslog.Tracef("registry update component %s to inited", name)
	return nil
}
