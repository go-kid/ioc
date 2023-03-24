package properties

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

func checkValid(v interface{}) error {
	typ := reflect.TypeOf(v)
	if typ.Kind() != reflect.Ptr {
		return errors.New(typ.String() + " Must be a pointer value")
	}
	return nil
}

func ToPropMap(m interface{}) map[string]interface{} {
	switch m.(type) {
	case map[string]interface{}:
		return buildProperties("", m.(map[string]interface{}), make(map[string]interface{}))
	default:
		tmp := make(map[string]interface{})
		bytes, _ := json.Marshal(m)
		json.Unmarshal(bytes, &tmp)
		return buildProperties("", tmp, make(map[string]interface{}))
	}
}

func buildProperties(prePath string, m, tmp map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		switch v.(type) {
		case map[string]interface{}:
			buildProperties(path(prePath, k), v.(map[string]interface{}), tmp)
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

func PropMapExpand(m map[string]interface{}) map[string]interface{} {
	tmp := make(map[string]interface{})
	for k, v := range m {
		buildMap(k, v, tmp)
	}
	return tmp
}

func buildMap(path string, val interface{}, tmp map[string]interface{}) map[string]interface{} {
	arr := strings.SplitN(path, ".", 2)
	if len(arr) > 1 {
		key := arr[0]
		next := arr[1]
		if tmp[key] == nil {
			tmp[key] = make(map[string]interface{})
		}
		tmp[key] = buildMap(next, val, tmp[key].(map[string]interface{}))
	} else {
		tmp[path] = val
	}
	return tmp
}
