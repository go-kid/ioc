package injector

import (
	"fmt"
	"github.com/go-kid/ioc/scanner/meta"
)

func CustomizedInject(r Injector, id string, customized map[string][]*meta.Node) error {
	for tag, nodes := range customized {
		methods := r.GetByFunc(tag)
		//for _, m := range metas {
		//	values := reflectx.TryCallMethod(m.Value, tag)
		//}
		fmt.Println(id)
		for _, value := range methods {
			fmt.Println(value.Call(nil))
		}
		err := DependencyInject(r, id, nodes)
		if err != nil {
			return err
		}
	}

	return nil
}
