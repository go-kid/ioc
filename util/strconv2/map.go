package strconv2

import (
	"encoding/json"
	"github.com/go-kid/ioc/util/strings2"
	"github.com/pkg/errors"
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
	if !isMap(val) {
		return nil, errors.Errorf("can not parse '%s' as map, need map[key:value ...] or json object", val)
	}
	result := make(map[string]any)
	if bytes := []byte(val); json.Valid(bytes) {
		err := json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	val = val[4 : len(val)-1]
	if val == "" {
		return result, nil
	}

	for _, part := range strings2.SplitWithConfig(val, mapSplitConfig) {
		subKV := strings.SplitN(part, ":", 2)
		if len(subKV) != 2 {
			return nil, errors.Errorf("can not parse \"%s\" as map, key value not found: \"%s\"", val, part)
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

func isMap(val string) bool {
	return (len(val) > 4 && val[:4] == "map[" && val[len(val)-1:] == "]") ||
		(val[:1] == "{" && val[len(val)-1:] == "}" && json.Valid([]byte(val)))
}

var mapSplitConfig = &strings2.SplitConfig{
	Sep:                 " ",
	N:                   -1,
	LeftSkipCharacters:  strings2.LeftBlocks,
	RightSkipCharacters: strings2.RightBlocks,
}
