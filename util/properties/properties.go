package properties

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type Properties map[string]any

func New() Properties {
	return make(Properties)
}

func NewFromMap(m map[string]any) Properties {
	return buildProperties("", m)
}

func (p Properties) Set(key string, val any) {
	p[key] = val
}

func (p Properties) Get(key string) (any, bool) {
	v, ok := p[key]
	return v, ok
}

func checkValid(v interface{}) error {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return errors.New(typ.String() + " Must be a pointer value")
	}
	return nil
}

func buildProperties(prePath string, m map[string]any) Properties {
	tmp := make(Properties)
	for k, v := range m {
		switch v.(type) {
		case map[string]any:
			subTmp := buildProperties(path(prePath, k), v.(map[string]any))
			for subP, subV := range subTmp {
				tmp[subP] = subV
			}
		default:
			tmp[path(prePath, k)] = v
		}
	}
	return tmp
}

func path(first, second string) string {
	if first == "" {
		return second
	}
	return first + "." + second
}

func (p Properties) Expand() map[string]any {
	tmp := make(map[string]any)
	for k, v := range p {
		buildMap(k, v, tmp)
	}
	return tmp
}

func buildMap(path string, val any, tmp map[string]any) map[string]any {
	arr := strings.SplitN(path, ".", 2)
	if len(arr) > 1 {
		key := arr[0]
		next := arr[1]
		if tmp[key] == nil {
			tmp[key] = make(map[string]any)
		}
		switch sub := tmp[key]; sub.(type) {
		case map[string]any, Properties:
			tmp[key] = buildMap(next, val, sub.(map[string]any))
		default:
			tmp[key] = sub
		}
	} else {
		tmp[path] = val
	}
	return tmp
}
