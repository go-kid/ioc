package strconv2

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func ParseAny(val string) (any, error) {
	var typeVal any
	var err error
	switch true {
	case val == "":
		typeVal = ""
	case strings.ToLower(val) == "true":
		typeVal = true
	case strings.ToLower(val) == "false":
		typeVal = false
	case isNumber(val):
		typeVal, err = strconv.ParseFloat(val, 64)
	case isMap(val):
		typeVal, err = ParseAnyMap(val)
	case isSlice(val):
		typeVal, err = ParseAnySlice(val)
	case strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'"),
		strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\""):
		typeVal = val[1 : len(val)-1]
	default:
		typeVal = val
	}
	return typeVal, err
}

var (
	numberReg = regexp.MustCompile("^(-|\\+)?\\d+(\\.\\d+)?$")
)

func isNumber(v string) bool {
	return numberReg.MatchString(v)
}

func FormatAny(a any) (string, error) {
	if a == nil {
		return "<nil>", nil
	}
	switch a.(type) {
	case string:
		return a.(string), nil
	case bool:
		return strconv.FormatBool(a.(bool)), nil
	default:
		switch p := reflect.TypeOf(a); p.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array, reflect.Struct:
			bytes, err := json.Marshal(a)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		case reflect.Pointer:
			return FormatAny(reflect.ValueOf(a).Elem().Interface())
		default:
			return fmt.Sprintf("%v", a), nil
		}
	}
}
