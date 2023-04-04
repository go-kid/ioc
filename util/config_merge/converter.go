package config_merge

import (
	"encoding/json"
	"github.com/go-kid/ioc/util/properties"
	"gopkg.in/yaml.v3"
)

func MergeProperties(m ...map[string]interface{}) map[string]interface{} {
	var r = make(map[string]interface{})
	for i := range m {
		for k, v := range m[i] {
			r[k] = v
		}
	}
	return r
}

func MergeMap(m ...map[string]interface{}) map[string]interface{} {
	var propMaps []map[string]interface{}
	for i := range m {
		propMaps = append(propMaps, properties.ToPropMap(m[i]))
	}
	return properties.PropMapExpand(MergeProperties(propMaps...))
}

func MergeYaml(yml ...[]byte) []byte {
	var maps []map[string]interface{}
	for i := range yml {
		var m map[string]interface{}
		yaml.Unmarshal(yml[i], &m)
		maps = append(maps, m)
	}
	mergeMap := MergeMap(maps...)
	bytes, _ := yaml.Marshal(mergeMap)
	return bytes
}

func MergeJson(jsn ...[]byte) []byte {
	var maps []map[string]interface{}
	for i := range jsn {
		var m map[string]interface{}
		json.Unmarshal(jsn[i], &m)
		maps = append(maps, m)
	}
	mergeMap := MergeMap(maps...)
	bytes, _ := json.Marshal(mergeMap)
	return bytes
}
