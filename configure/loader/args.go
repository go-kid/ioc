package loader

import (
	"flag"
	"github.com/go-kid/ioc/syslog"
	"github.com/go-kid/ioc/util/strconv2"
	"github.com/go-kid/properties"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strings"
)

func init() {
	flag.String("app.config", "", "used for command line configure")
}

type ArgsLoader []string

func NewArgsLoader(args []string) ArgsLoader {
	return ArgsLoader(args)
}

func (args ArgsLoader) LoadConfig() ([]byte, error) {
	p := properties.New()
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--app.config") {
			continue
		}
		cfg := strings.TrimPrefix(arg, "--app.config=")
		syslog.Pref("ArgsLoader").Tracef("detected command config: %s", cfg)
		propPair := strings.SplitN(cfg, "=", 2)
		var val string
		if len(propPair) == 2 {
			val = propPair[1]
		}
		typeVal, err := strconv2.ParseAny(val)
		if err != nil {
			return nil, errors.Wrapf(err, "parse '%s' as any", val)
		}
		p.Set(propPair[0], typeVal)
		syslog.Pref("ArgsLoader").Debugf("parse command config: %s=%s", propPair[0], typeVal)
	}
	if len(p) == 0 {
		return nil, nil
	}
	bytes, err := yaml.Marshal(p.Expand())
	if err != nil {
		return nil, errors.Wrapf(err, "marshal to YAML: %+v", p.Expand())
	}

	return bytes, nil
}
