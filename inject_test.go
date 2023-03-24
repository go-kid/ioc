package ioc

import (
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
	RunTest(t, "config.yaml", SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		app,
	))
	assert.Equal(t, "bImpl", app.T11.Naming())
	assert.Equal(t, "bImpl", app.T12.(NamingComponent).Naming())
	assert.Equal(t, "cImpl", app.T21.Naming())
	assert.Equal(t, "cImpl", app.T22.(NamingComponent).Naming())
}

func TestInjectByPtrType(t *testing.T) {
	var app = &struct {
		T1 *aImpl `wire:""`
		T2 *bImpl `wire:""`
		T3 *cImpl `wire:""`
	}{}
	RunTest(t, "config.yaml", SetComponents(
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
	RunTest(t, "config.yaml", SetComponents(
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
	RunTest(t, "config.yaml", SetComponents(
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
	RunTest(t, "config.yaml", SetComponents(
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

func TestCfg(t *testing.T) {
	c := &cfgAImpl{}
	RunTest(t, "config.yaml", SetComponents(c))
	assert.Equal(t, "cfgAImpl", c.GetName())
}

type arrImpl struct {
	T1  []ITest  `wire:""`
	Arr []string `prop:"a.c"`
}

func TestInjectDebug(t *testing.T) {
	var app = &struct {
		Arr *arrImpl `wire:""`
	}{}
	err := RunDebug("config.yaml", SetComponents(
		&aImpl{},
		&bImpl{},
		&cImpl{},
		&arrImpl{},
		app,
	))
	assert.NoError(t, err)
}
