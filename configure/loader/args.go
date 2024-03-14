package loader

import (
	"github.com/go-kid/ioc/util/properties"
	"gopkg.in/yaml.v3"
	"regexp"
	"strconv"
	"strings"
)

type ArgsLoader []string

func NewArgsLoader(args []string) ArgsLoader {
	return ArgsLoader(args)
}

func (args ArgsLoader) LoadConfig() ([]byte, error) {
	p := properties.New()
	reg := regexp.MustCompile("^-{2}\\S+=\\S*")
	intReg := regexp.MustCompile("^\\d+$")
	floatReg := regexp.MustCompile("^\\d+\\.\\d+$")
	for _, arg := range args {
		prop := reg.FindString(arg)
		if prop == "" {
			continue
		}
		propPair := strings.SplitN(prop[2:], "=", 2)
		var val string
		if len(propPair) == 2 {
			val = propPair[1]
		}

		var typeVal any
		if val == "" {
			typeVal = ""
		} else if val == "true" {
			typeVal = true
		} else if val == "false" {
			typeVal = false
		} else if intReg.MatchString(val) {
			typeVal, _ = strconv.ParseUint(val, 10, 64)
		} else if floatReg.MatchString(val) {
			typeVal, _ = strconv.ParseFloat(val, 64)
		} else {
			typeVal = val
		}
		p.Set(propPair[0], typeVal)
	}
	return yaml.Marshal(p.Expand())
}
