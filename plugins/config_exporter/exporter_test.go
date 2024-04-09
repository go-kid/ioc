package config_exporter

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
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
	t.Run("DefaultMode", func(t *testing.T) {
		a := &A{}
		exporter := NewConfigExporter(0)
		_, err := ioc.Run(
			app.LogWarn,
			app.SetComponents(a, exporter),
		)
		if err != nil {
			panic(err)
		}
		bytes, err := yaml.Marshal(exporter.GetConfig().Expand())
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
		assert.Equal(t, string(exampleConfig), string(bytes))
	})
	t.Run("AppendMode", func(t *testing.T) {
		cfg := []byte(`
Demo:
    a: this is a test
    b: 20
    slice:
        - "hello"
        - "world"
    array:
        - 999
        - 888
        - 777
    m:
        select: 1
app:
    configA: string
    configB: string
    configSlice:
        - a
        - b
    valueB: abc
`)
		a := &A{}
		exporter := NewConfigExporter(Append)
		_, err := ioc.Run(
			app.LogWarn,
			app.AddConfigLoader(loader.NewRawLoader(cfg)),
			app.SetComponents(a, exporter),
		)
		if err != nil {
			panic(err)
		}
		bytes, err := yaml.Marshal(exporter.GetConfig().Expand())
		if err != nil {
			panic(err)
		}

		var exampleConfig = []byte(`Demo:
    a: this is a test
    array:
        - 999
        - 888
        - 777
    b: 20
    m:
        select: 1
    slice:
        - hello
        - world
app:
    configA: string
    configB: string
    configSlice:
        - a
        - b
    valueB: abc
`)
		assert.Equal(t, string(exampleConfig), string(bytes))
	})
	t.Run("OnlyNewMode", func(t *testing.T) {
		cfg := []byte(`
Demo:
    a: this is a test
    b: 20
    slice:
        - "hello"
        - "world"
    array:
        - 999
        - 888
        - 777
    m:
        select: 1
`)
		a := &A{}
		exporter := NewConfigExporter(OnlyNew)
		_, err := ioc.Run(
			app.LogWarn,
			app.AddConfigLoader(loader.NewRawLoader(cfg)),
			app.SetComponents(a, exporter),
		)
		if err != nil {
			panic(err)
		}
		bytes, err := yaml.Marshal(exporter.GetConfig().Expand())
		if err != nil {
			panic(err)
		}

		var exampleConfig = []byte(`app:
    configA: string
    configB: string
    configSlice:
        - a
        - b
    valueB: abc
`)
		assert.Equal(t, string(exampleConfig), string(bytes))
	})
	t.Run("AnnotationSourceMode", func(t *testing.T) {
		a := &A{}
		exporter := NewConfigExporter(AnnotationSource)
		_, err := ioc.Run(
			app.LogWarn,
			app.SetComponents(a, exporter),
		)
		if err != nil {
			panic(err)
		}
		bytes, err := yaml.Marshal(exporter.GetConfig().Expand())
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
source:
    Demo:
        - github.com/go-kid/ioc/plugins/config_exporter/A
    app:
        configA:
            - github.com/go-kid/ioc/plugins/config_exporter/A
        configB:
            - github.com/go-kid/ioc/plugins/config_exporter/A
        configSlice:
            - github.com/go-kid/ioc/plugins/config_exporter/A
        valueB:
            - github.com/go-kid/ioc/plugins/config_exporter/A
`)
		assert.Equal(t, string(exampleConfig), string(bytes))
	})
}
