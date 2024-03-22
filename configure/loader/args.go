package loader

import (
	"github.com/go-kid/ioc/util/properties"
	"github.com/go-kid/ioc/util/strconv2"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

type ArgsLoader []string

func NewArgsLoader(args []string) ArgsLoader {
	return ArgsLoader(args)
}

func (args ArgsLoader) LoadConfig() ([]byte, error) {
	p := properties.New()
	reg := regexp.MustCompile("^-{2}\\S+=\\S*")
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
		typeVal, err := strconv2.ParseAny(val)
		if err != nil {
			return nil, err
		}
		p.Set(propPair[0], typeVal)
	}
	if len(p) == 0 {
		return nil, nil
	}
	return yaml.Marshal(p.Expand())
}
