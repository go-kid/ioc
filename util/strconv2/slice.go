package strconv2

import (
	"fmt"
	"strings"
)

// ParseStringSlice exp: "[value1,value2,value3]" -> []string{"value1", "value2", "value3"}
func ParseStringSlice(val string) ([]string, error) {
	if isSlice(val) {
		val = val[1 : len(val)-1]
		if val == "" {
			return []string{}, nil
		}
		return strings.Split(val, ","), nil
	}
	return nil, fmt.Errorf("can not parse \"%s\" as slice, need [value1 value2 ...]", val)
}

func ParseAnySlice(val string) ([]any, error) {
	slice, err := ParseStringSlice(val)
	if err != nil {
		return nil, err
	}
	var result = make([]any, len(slice))
	for i, v := range slice {
		result[i], err = ParseAny(v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func ParseSlice[T any](val string) ([]T, error) {
	slice, err := ParseStringSlice(val)
	if err != nil {
		return nil, err
	}
	var result = make([]T, len(slice))
	for i, v := range slice {
		result[i], err = Parse[T](v)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func isSlice(val string) bool {
	return len(val) > 1 && val[:1] == "[" && val[len(val)-1:] == "]"
}
