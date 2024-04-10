package strconv2

import (
	"fmt"
)

// ParseStringSlice exp: "[value1,value2,value3]" -> []string{"value1", "value2", "value3"}
func ParseStringSlice(val string) ([]string, error) {
	if isSlice(val) {
		val = val[1 : len(val)-1]
		if val == "" {
			return []string{}, nil
		}
		return splitSlicePart(val), nil
	}
	return nil, fmt.Errorf("can not parse \"%s\" as slice, need [value1,value2 ...]", val)
}

// {"x-request-id":"123"},{"x-cross-origin":["*"]},{"x-allowed-method":["POST","GET"]}
func splitSlicePart(val string) []string {
	var pairs []string
	var in = 0
	var last = 0
	var i = 0
	length := len(val)
	for i < length {
		c := val[i]
		i++
		if c == '{' || c == '[' || c == '(' {
			in++
			continue
		}
		if c == '}' || c == ']' || c == ')' {
			in--
			if in == 0 {
				pairs = append(pairs, val[last:i])
				last = i + 1
			}
			continue
		}
		if c == ',' && in == 0 {
			if subLen := i - last; subLen > 1 {
				pairs = append(pairs, val[last:i-1])
			} else if subLen == 1 {
				pairs = append(pairs, "")
			}
			last = i
			continue
		}
		if i == length && in == 0 {
			pairs = append(pairs, val[last:])
			last = length
		}
	}
	return pairs
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
