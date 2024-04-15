package strconv2

import (
	"encoding/json"
	"github.com/go-kid/ioc/util/strings2"
	"github.com/pkg/errors"
)

func ParseAnySlice(val string) ([]any, error) {
	if !isSlice(val) {
		return nil, errors.Errorf("can not parse \"%s\" as slice, need [value1,value2 ...] or json array", val)
	}
	var result = make([]any, 0)
	if bytes := []byte(val); json.Valid(bytes) {
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	val = val[1 : len(val)-1]
	if val == "" {
		return result, nil
	}
	for _, v := range strings2.SplitWithConfig(val, sliceSplitConfig) {
		a, err := ParseAny(v)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

var sliceSplitConfig = &strings2.SplitConfig{
	Sep:                 ",",
	N:                   -1,
	LeftSkipCharacters:  strings2.LeftBlocks,
	RightSkipCharacters: strings2.RightBlocks,
}

// ParseStringSlice exp: "[value1,value2,value3]" -> []string{"value1", "value2", "value3"}
func ParseStringSlice(val string) ([]string, error) {
	anies, err := ParseAnySlice(val)
	if err != nil {
		return nil, err
	}
	var result = make([]string, len(anies))
	for i, a := range anies {
		f, err := FormatAny(a)
		if err != nil {
			return nil, err
		}
		result[i] = f
	}
	return result, nil
}

func ParseSlice[T any](val string) ([]T, error) {
	anies, err := ParseStringSlice(val)
	if err != nil {
		return nil, err
	}
	var result = make([]T, len(anies))
	for i, a := range anies {
		parse, err := Parse[T](a)
		if err != nil {
			return nil, err
		}
		result[i] = parse
	}
	return result, nil
}

func isSlice(val string) bool {
	return len(val) > 1 && val[:1] == "[" && val[len(val)-1:] == "]"
}
