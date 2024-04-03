package config_exporter

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

type Config struct {
	A     string         `yaml:"a"`
	B     int            `yaml:"b"`
	Slice []string       `yaml:"slice"`
	Array [3]float64     `yaml:"array"`
	M     map[string]int `yaml:"m"`
}

func (c *Config) Prefix() string {
	return "Demo"
}

type A struct {
	ConfigA     string   `prop:"app.configA"`
	ConfigB     string   `prop:"${app.configB}"`
	ConfigSlice []string `prop:"${app.configSlice:[a,b]}"`
	ValueA      string   `value:"abc"`
	ValueB      string   `value:"${app.valueB:abc}"`
	ValueC      string   `value:"#{'a'+'b'}"`
	Config      *Config
	Greeting    Greeting `wire:""`
}

func (a *A) Order() int {
	return 0
}

func (a *A) Run() error {
	a.Greeting.Hi()
	return nil
}

type Greeting interface {
	Hi()
}

func TestConfigExporter(t *testing.T) {
	a := &A{}
	processor := NewConfigExporter()
	_, err := ioc.Run(
		//app.LogTrace,
		app.LogWarn,
		app.SetComponents(
			a,
			processor,
		),
	)
	if err != nil {
		panic(err)
	}
	bytes, err := yaml.Marshal(processor.GetConfig().Expand())
	if err != nil {
		panic(err)
	}

	var exampleConfig = []byte(`Demo:
    a: string
    b: 0
    slice:
        - string
    array:
        - 0
        - 0
        - 0
    m:
        string: 0
app:
    configA: string
    configB: string
    configSlice:
        - a
        - b
    valueB: abc
`)
	assert.Equal(t, exampleConfig, bytes)
	fmt.Println(string(bytes))
}
