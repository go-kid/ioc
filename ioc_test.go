package ioc

import (
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/defination"
	"github.com/stretchr/testify/assert"
	"testing"
)

type ITest interface {
	GetName() string
}

type aImpl struct{}

func (a *aImpl) Init() error { return nil }

func (a *aImpl) GetName() string {
	return "aImpl"
}

type bImpl struct {
	ITest `wire:""`
}

func (a *bImpl) Init() error { return nil }

func (a *bImpl) Naming() string {
	return "bImpl"
}

func (a *bImpl) GetName() string {
	return "bImpl"
}

type cImpl struct {
	ITest `wire:""`
}

func (a *cImpl) Init() error { return nil }

func (a *cImpl) Naming() string {
	return "cImpl"
}

func TestInjectByName(t *testing.T) {
	var app = &struct {
		T11 *bImpl `wire:"bImpl"`
		T12 ITest  `wire:"bImpl"`
		T21 *cImpl `wire:"cImpl"`
		T22 ITest  `wire:"cImpl"`
	}{}
	RunTest(t, SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		app,
	))
	assert.Equal(t, "bImpl", app.T11.Naming())
	assert.Equal(t, "bImpl", app.T12.(defination.NamingComponent).Naming())
	assert.Equal(t, "cImpl", app.T21.Naming())
	assert.Equal(t, "cImpl", app.T22.(defination.NamingComponent).Naming())
}

func TestInjectByPtrType(t *testing.T) {
	var app = &struct {
		T1 *aImpl `wire:""`
		T2 *bImpl `wire:""`
		T3 *cImpl `wire:""`
	}{}
	RunTest(t, SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		app,
	))
	assert.Equal(t, "aImpl", app.T1.GetName())
	assert.Equal(t, "bImpl", app.T2.Naming())
	assert.Equal(t, "cImpl", app.T3.Naming())
}

func TestInjectInterfaceNamingPrefer(t *testing.T) {
	var app = &struct {
		T1 ITest `wire:""`
		T2 ITest `wire:"bImpl"`
		T3 ITest `wire:"cImpl"`
	}{}
	RunTest(t, SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		app,
	))
	for i := 0; i < 1000; i++ {
		assert.Equal(t, "aImpl", app.T1.GetName())
		assert.Equal(t, "bImpl", app.T2.GetName())
		assert.Equal(t, "aImpl", app.T3.GetName())
	}
}

func TestInjectByInterfaceSlice(t *testing.T) {
	var app = &struct {
		T1 []ITest `wire:""`
	}{}
	RunTest(t, SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		app,
	))
	assert.Equal(t, 3, len(app.T1))
	var countMap = make(map[string]int)
	for _, test := range app.T1 {
		countMap[test.GetName()]++
	}
	assert.Equal(t, 2, countMap["aImpl"])
	assert.Equal(t, 1, countMap["bImpl"])
	assert.Equal(t, 0, countMap["cImpl"])
}

type postImpl struct {
	count int
}

func (p *postImpl) PostProcessBeforeInitialization(component interface{}) error {
	_, ok := component.(ITest)
	if !ok {
		return nil
	}
	p.count++
	return nil
}

func (p *postImpl) PostProcessAfterInitialization(component interface{}) error {
	_, ok := component.(ITest)
	if !ok {
		return nil
	}
	p.count++
	return nil
}

func TestPostProcessor(t *testing.T) {
	p := &postImpl{}
	RunTest(t, SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		p,
	))
	assert.Equal(t, 6, p.count)
}

type cfgAImpl struct {
	Name string `prop:"a.name"`
}

func (c *cfgAImpl) GetName() string {
	return c.Name
}

//func TestCfg(t *testing.T) {
//	c := &cfgAImpl{}
//	t.Run("TestSetConfigSrc", func(t *testing.T) {
//		var _tConfig = `a:
//  name: "cfgAImpl"`
//		RunTest(t, SetComponents(c), SetConfigSrc([]byte(_tConfig), "yaml"))
//		assert.Equal(t, "cfgAImpl", c.GetName())
//	})
//	t.Run("TestSetConfigStructure", func(t *testing.T) {
//		var config = struct {
//			A struct {
//				Name string
//			}
//		}{
//			A: struct{ Name string }{Name: "cfgAImpl"},
//		}
//		RunTest(t, SetComponents(c), SetConfigStructure(config))
//		assert.Equal(t, "cfgAImpl", c.GetName())
//	})
//}

type arrImpl struct {
	T1  []ITest  `wire:""`
	Arr []string `prop:"a.c"`
}

func TestInjectDebug(t *testing.T) {
	a := &arrImpl{}
	err := RunDebug(SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		a,
	))
	assert.NoError(t, err)
}

type configA struct {
	B int   `yaml:"b"`
	C []int `yaml:"c"`
}

type configD struct {
	D1 string `yaml:"d1"`
	D2 int    `yaml:"d2"`
}

func (c *configD) Prefix() string {
	return "a.d"
}

//func TestConfig(t *testing.T) {
//	var app = &struct {
//		A *configA `prop:"a"`
//		D *configD
//	}{}
//	var _tConfig = `
//a:
//  b: 123
//  c: [ 1,2,3,4 ]
//  d:
//    d1: "abc"
//    d2: 123`
//	RunTest(t, SetComponents(app), SetConfigSrc([]byte(_tConfig), "yaml"))
//	assert.Equal(t, 123, app.A.B)
//	assert.Equal(t, "abc", app.D.D1)
//}

type port string

type producer struct {
	T    *aImpl `produce:""`
	T2   *bImpl `produce:""`
	Port *port  `produce:""`
}

func (s *producer) Init() error {
	*s.Port = "8888"
	return nil
}

type consumer struct {
	T    ITest `wire:""`
	Port *port `wire:""`
}

func TestProduce(t *testing.T) {
	p := new(producer)
	c := new(consumer)
	RunTest(t, SetComponents(p, c))
	assert.Equal(t, port("8888"), *c.Port)
	*p.Port = "9999"
	assert.Equal(t, port("9999"), *c.Port)
	assert.Equal(t, "aImpl", c.T.GetName())
}

type Base struct {
	T ITest `wire:""`
}

type Parent struct {
	Base
	T2 ITest `wire:""`
}

type Child struct {
	Parent
	T3 ITest `wire:""`
}

type Child2 struct {
	Parent
	T3 ITest `wire:""`
}

func TestEmbedComponent(t *testing.T) {
	c := &Child{}
	c2 := &Child2{}
	RunTest(t, SetComponents(c, c2, &aImpl{}))
	assert.Equal(t, "aImpl", c.T3.GetName())
	assert.Equal(t, "aImpl", c.T2.GetName())
	assert.Equal(t, "aImpl", c.T.GetName())

	assert.Equal(t, "aImpl", c2.T3.GetName())
	assert.Equal(t, "aImpl", c2.T2.GetName())
	assert.Equal(t, "aImpl", c2.T.GetName())
}
