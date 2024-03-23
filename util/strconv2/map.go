package strconv2

import (
	"fmt"
	"strings"
)

// ParseAnyMap "map[aes:map[iv:abc key:123] header:[X-Request-Id X-Cross-Origin X-Allowed-Method]]" ->
//
//	map[string]any{
//	 "aes": map[string]any{
//	   "iv":  "abc",
//	   "key": int64(123),
//	 },
//	 "header": []any{"X-Request-Id", "X-Cross-Origin", "X-Allowed-Method"},
//	}
func ParseAnyMap(val string) (map[string]any, error) {
	if isMap(val) {
		val = val[4 : len(val)-1]
		if val == "" {
			return map[string]any{}, nil
		}
		result := make(map[string]any)
		for _, part := range splitPart(val) {
			subKV := strings.SplitN(part, ":", 2)
			if len(subKV) != 2 {
				return nil, fmt.Errorf("can not parse \"%s\" as map, key value not found: \"%s\"", val, part)
			}
			var (
				subK, subV = subKV[0], subKV[1]
				sub        any
				err        error
			)
			switch true {
			case isMap(subV):
				sub, err = ParseAnyMap(subV)
			case isSlice(subV):
				sub, err = ParseAnySlice(subV)
			default:
				sub, err = ParseAny(subV)
			}
			if err != nil {
				return result, err
			}
			result[subK] = sub
		}
		return result, nil
	}
	return nil, fmt.Errorf("can not parse \"%s\" as map, need map[key:value ...]", val)
}

func isMap(val string) bool {
	return len(val) > 4 && val[:4] == "map[" && val[len(val)-1:] == "]"
}

//aes:map[iv:abc key:123] header:[X-Request-Id X-Cross-Origin X-Allowed-Method]
func splitPart(val string) []string {
	var pairs []string
	var in = 0
	var last = 0
	var i = 0
	var q = false
	for i < len(val) {
		c := val[i]
		i++
		if c == ':' {
			q = true
			continue
		}
		if q {
			if c == '[' {
				in++
				continue
			}
			if c == ']' {
				in--
				if in == 0 {
					pairs = append(pairs, val[last:i])
					last = i + 1
					q = false
				}
				continue
			}
			if c == ' ' && in == 0 && (i-1-last) > 1 {
				pairs = append(pairs, val[last:i-1])
				last = i
				q = false
				continue
			}
			if i == len(val)-1 && in == 0 {
				pairs = append(pairs, val[last:])
				last = len(val)
			}
		}
	}
	if len(val) > last && val[last:] != "" {
		pairs = append(pairs, val[last:])
	}
	return pairs
}
