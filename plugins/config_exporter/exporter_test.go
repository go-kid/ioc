package config_exporter

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

type SubConfig struct {
	Sub string `yaml:"sub"`
}

type Config struct {
	A     string         `yaml:"a"`
	B     int            `yaml:"b"`
	Slice []string       `yaml:"slice"`
	Array [3]float64     `yaml:"array"`
	M     map[string]int `yaml:"m"`
	G     Greeting       `yaml:"-"`
}

func (c *Config) Prefix() string {
	return "Demo"
}

type MergeConfig struct {
	S     string         `yaml:"s"`
	B     bool           `yaml:"b"`
	M     map[string]int `yaml:"m"`
	Slice []float64      `yaml:"slice"`
	Sub   SubConfig      `yaml:"sub"`
	SubP  *SubConfig     `yaml:"subP"`
}

func (c *MergeConfig) Prefix() string {
	return "Merge"
}

type A struct {
	ConfigA     string   `prop:"app.configA"`
	ConfigB     string   `prop:"${app.configB}"`
	ConfigSlice []string `prop:"${app.configSlice:[a,b]}"`
	ValueA      string   `value:"abc"`
	ValueB      string   `value:"${app.valueB:abc}"`
	ValueC      string   `value:"#{'a'+'b'}"`
	Config      *Config
	Merge       *MergeConfig
	MergeS2     string            `prop:"Merge.s2"`
	MergeM2     map[string]string `prop:"Merge.m2"`
	MergeSlice2 []int64           `prop:"Merge.slice2"`
	MergeSub2   SubConfig         `prop:"Merge.sub2"`
	MergeSubP2  *SubConfig        `prop:"Merge.subP2"`
	MergeSub    *MergeConfig      `prop:"Merge.sub"`
	Greeting    Greeting          `wire:""`
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
    array:
        - 0
        - 0
        - 0
    b: 0
    m:
        string: 0
    slice:
        - string
Merge:
    b: false
    m:
        string: 0
    m2:
        string: string
    s: string
    s2: string
    slice:
        - 0
    slice2:
        - 0
    sub:
        b: false
        m:
            string: 0
        s: string
        slice:
            - 0
        sub: string
        subP:
            sub: string
    sub2:
        sub: string
    subP:
        sub: string
    subP2:
        sub: string
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
Merge:
    b: false
    m:
        string: 0
    m2:
        string: string
    s: string
    s2: string
    slice:
        - 0
    slice2:
        - 0
    sub:
        b: false
        m:
            string: 0
        s: string
        slice:
            - 0
        sub: string
        subP:
            sub: string
    sub2:
        sub: string
    subP:
        sub: string
    subP2:
        sub: string
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
		cfg := []byte(`Merge:
    b: false
    m:
        string: 0
    s: string
    slice:
        - 0
    sub:
        sub: "subSub"
        subP:
            sub: "subSubPSub"
    subP:
        sub: string
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

		var exampleConfig = []byte(`Merge:
    m2:
        string: string
    s2: string
    slice2:
        - 0
    sub2:
        sub: string
    subP2:
        sub: string
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
	t.Run("AnnotationSourceMode", func(t *testing.T) {
		var cfg = []byte(`Demo:
    a: string
    array:
        - 0
        - 0
        - 0
    b: 0
    m:
        string: 0
    slice:
        - string
Merge:
    b: false
    m:
        string: 0
    m2:
        string: string
    s: string
    s2: string
    slice:
        - 0
    slice2:
        - 0
    sub:
        b: false
        m:
            string: 0
        s: string
        slice:
            - 0
        sub: string
        subP:
            sub: string
    sub2:
        sub: string
    subP:
        sub: string
    subP2:
        sub: string
app:
    configA: string
    configB: string
    configSlice:
        - a
        - b
    valueB: abc`)
		type A2 struct {
			Config *Config
		}
		exporter := NewConfigExporter(AnnotationSource | OnlyNew)
		_, err := ioc.Run(
			app.LogWarn,
			app.SetComponents(&A{}, &A2{}, exporter),
			app.AddConfigLoader(loader.NewRawLoader(cfg)),
		)
		if err != nil {
			panic(err)
		}
		bytes, err := yaml.Marshal(exporter.GetConfig().Expand())
		if err != nil {
			panic(err)
		}

		var exampleConfig = []byte(`Source:
    Demo:
        - github.com/go-kid/ioc/plugins/config_exporter/A
        - github.com/go-kid/ioc/plugins/config_exporter/A2
    Merge:
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
