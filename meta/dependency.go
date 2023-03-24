package meta

import (
	"github.com/kid-hash/kid-ioc/defination"
	"reflect"
)

type Dependency struct {
	SpecifyName string
	Type        reflect.Type
	Value       reflect.Value
}

func (d *Dependency) Name() string {
	if d.SpecifyName != "" {
		return d.SpecifyName
	}
	if v := d.Value.Interface(); v != nil {
		return defination.GetComponentName(v)
	}
	return defination.GetComponentName(reflect.New(d.Type).Interface())
}
