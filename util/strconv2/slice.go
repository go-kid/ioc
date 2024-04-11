package strconv2

import (
	"encoding/json"
	"fmt"
)

func ParseAnySlice(val string) ([]any, error) {
	if !isSlice(val) {
		return nil, fmt.Errorf("can not parse \"%s\" as slice, need [value1,value2 ...] or json array", val)
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
	for _, v := range splitSlicePart(val) {
		a, err := ParseAny(v)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
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
