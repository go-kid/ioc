package constructor_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

// --- shared types ---

type DepA struct {
	Name string
}

func (d *DepA) Naming() string { return d.Name }

type IService interface {
	Serve() string
}

type serviceImpl struct {
	id string
}

func (s *serviceImpl) Serve() string  { return s.id }
func (s *serviceImpl) Naming() string { return s.id }

type primaryService struct {
	serviceImpl
}

func (p *primaryService) Primary() {}

// --- constructors ---

type ptrDepComponent struct {
	dep *DepA
}

func NewPtrDepComponent(a *DepA) *ptrDepComponent {
	return &ptrDepComponent{dep: a}
}

type ifaceDepComponent struct {
	svc IService
}

func NewIfaceDepComponent(s IService) *ifaceDepComponent {
	return &ifaceDepComponent{svc: s}
}

type ptrSliceComponent struct {
	deps []*DepA
}

func NewPtrSliceComponent(deps []*DepA) *ptrSliceComponent {
	return &ptrSliceComponent{deps: deps}
}

type ifaceSliceComponent struct {
	svcs []IService
}

func NewIfaceSliceComponent(svcs []IService) *ifaceSliceComponent {
	return &ifaceSliceComponent{svcs: svcs}
}

type noArgComponent struct {
	val string
}

func NewNoArgComponent() *noArgComponent {
	return &noArgComponent{val: "created"}
}

type multiArgComponent struct {
	dep *DepA
	svc IService
}

func NewMultiArgComponent(a *DepA, s IService) *multiArgComponent {
	return &multiArgComponent{dep: a, svc: s}
}

type errorComponent struct {
	val string
}

func NewErrorComponentOK() (*errorComponent, error) {
	return &errorComponent{val: "ok"}, nil
}

type wireableComponent struct {
	depFromCtor *DepA
	DepFromWire *DepA `wire:""`
}

func NewWireableComponent(a *DepA) *wireableComponent {
	return &wireableComponent{depFromCtor: a}
}

// --- tests ---

func TestConstructor_PointerParam(t *testing.T) {
	type T struct {
		C *ptrDepComponent `wire:""`
	}
	tt := &T{}
	dep := &DepA{Name: "hello"}
	ioc.RunTest(t, app.SetComponents(tt, dep, NewPtrDepComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "hello", tt.C.dep.Name)
}

func TestConstructor_InterfaceParam(t *testing.T) {
	type T struct {
		C *ifaceDepComponent `wire:""`
	}
	tt := &T{}
	svc := &serviceImpl{id: "svc1"}
	ioc.RunTest(t, app.SetComponents(tt, svc, NewIfaceDepComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "svc1", tt.C.svc.Serve())
}

func TestConstructor_PointerSliceParam(t *testing.T) {
	type T struct {
		C *ptrSliceComponent `wire:""`
	}
	tt := &T{}
	d1 := &DepA{Name: "a"}
	d2 := &DepA{Name: "b"}
	ioc.RunTest(t, app.SetComponents(tt, d1, d2, NewPtrSliceComponent))
	assert.NotNil(t, tt.C)
	assert.Len(t, tt.C.deps, 2)
}

func TestConstructor_InterfaceSliceParam(t *testing.T) {
	type T struct {
		C *ifaceSliceComponent `wire:""`
	}
	tt := &T{}
	s1 := &serviceImpl{id: "a"}
	s2 := &serviceImpl{id: "b"}
	ioc.RunTest(t, app.SetComponents(tt, s1, s2, NewIfaceSliceComponent))
	assert.NotNil(t, tt.C)
	assert.Len(t, tt.C.svcs, 2)
}

func TestConstructor_NoArgs(t *testing.T) {
	type T struct {
		C *noArgComponent `wire:""`
	}
	tt := &T{}
	ioc.RunTest(t, app.SetComponents(tt, NewNoArgComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "created", tt.C.val)
}

func TestConstructor_MultiArgs(t *testing.T) {
	type T struct {
		C *multiArgComponent `wire:""`
	}
	tt := &T{}
	dep := &DepA{Name: "dep"}
	svc := &serviceImpl{id: "svc"}
	ioc.RunTest(t, app.SetComponents(tt, dep, svc, NewMultiArgComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "dep", tt.C.dep.Name)
	assert.Equal(t, "svc", tt.C.svc.Serve())
}

func TestConstructor_WithErrorReturn(t *testing.T) {
	type T struct {
		C *errorComponent `wire:""`
	}
	tt := &T{}
	ioc.RunTest(t, app.SetComponents(tt, NewErrorComponentOK))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "ok", tt.C.val)
}

func TestConstructor_WiredByOtherComponents(t *testing.T) {
	type T struct {
		C *ptrDepComponent `wire:""`
	}
	tt := &T{}
	dep := &DepA{Name: "shared"}
	ioc.RunTest(t, app.SetComponents(tt, dep, NewPtrDepComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "shared", tt.C.dep.Name)
}

func TestConstructor_WithWireTags(t *testing.T) {
	type T struct {
		C *wireableComponent `wire:""`
	}
	tt := &T{}
	dep := &DepA{Name: "both"}
	ioc.RunTest(t, app.SetComponents(tt, dep, NewWireableComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "both", tt.C.depFromCtor.Name)
	assert.Equal(t, "both", tt.C.DepFromWire.Name)
}

func TestConstructor_PrimarySelection(t *testing.T) {
	type T struct {
		C *ifaceDepComponent `wire:""`
	}
	tt := &T{}
	s1 := &serviceImpl{id: "normal"}
	s2 := &primaryService{serviceImpl: serviceImpl{id: "primary"}}
	ioc.RunTest(t, app.SetComponents(tt, s1, s2, NewIfaceDepComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "primary", tt.C.svc.Serve())
}

// --- ConfigurationProperties ---

type dbConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (c *dbConfig) Prefix() string { return "db" }

type configDepComponent struct {
	cfg *dbConfig
}

func NewConfigDepComponent(cfg *dbConfig) *configDepComponent {
	return &configDepComponent{cfg: cfg}
}

func TestConstructor_ConfigurationProperties(t *testing.T) {
	type T struct {
		C *configDepComponent `wire:""`
	}
	tt := &T{}
	cfgData := []byte(`
db:
  host: "localhost"
  port: 5432
`)
	ioc.RunTest(t,
		app.SetComponents(tt, NewConfigDepComponent),
		app.SetConfigLoader(loader.NewRawLoader(cfgData)),
	)
	assert.NotNil(t, tt.C)
	assert.NotNil(t, tt.C.cfg)
	assert.Equal(t, "localhost", tt.C.cfg.Host)
	assert.Equal(t, 5432, tt.C.cfg.Port)
}

// --- ioc.Register style ---

func TestConstructor_ViaIocRegister(t *testing.T) {
	type T struct {
		C *ptrDepComponent `wire:""`
	}
	tt := &T{}
	dep := &DepA{Name: "registered"}
	ioc.RunTest(t, app.SetComponents(tt, dep), app.SetComponents(NewPtrDepComponent))
	assert.NotNil(t, tt.C)
	assert.Equal(t, "registered", tt.C.dep.Name)
}
