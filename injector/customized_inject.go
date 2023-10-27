package injector

import (
	"fmt"
	"github.com/go-kid/ioc/scanner/meta"
)

func CustomizedInject(r Injector, id string, customized map[string][]*meta.Node) error {
	for tag, nodes := range customized {
		values := r.GetByFunc(tag)
		//for _, m := range metas {
		//	values := reflectx.TryCallMethod(m.Value, tag)
		//}
		fmt.Println(id)
		var result string
		for _, value := range values {
			results := value.MethodByName(tag).Call(nil)
			if len(results) > 0 {
				result = results[0].String()
			}
		}
		fmt.Println("result", result)
		err := DependencyInject(r, id, nodes)
		if err != nil {
			return err
		}
	}

	return nil
}
