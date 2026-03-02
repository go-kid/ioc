package ioc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type provideTestComponent struct {
	Value string
}

type provideTestIface interface {
	GetValue() string
}

type provideTestImpl struct {
	Val string
}

func (p *provideTestImpl) GetValue() string { return p.Val }

func TestProvide_ConcreteType(t *testing.T) {
	assert.NotPanics(t, func() {
		Provide[provideTestComponent](func() *provideTestComponent {
			return &provideTestComponent{Value: "ok"}
		})
	})
	registerHandlers = nil
}

func TestProvide_ConcreteTypeWithError(t *testing.T) {
	assert.NotPanics(t, func() {
		Provide[provideTestComponent](func() (*provideTestComponent, error) {
			return &provideTestComponent{Value: "ok"}, nil
		})
	})
	registerHandlers = nil
}

func TestProvide_InterfaceType(t *testing.T) {
	assert.NotPanics(t, func() {
		Provide[provideTestIface](func() *provideTestImpl {
			return &provideTestImpl{Val: "ok"}
		})
	})
	registerHandlers = nil
}

func TestProvide_NotAFunction(t *testing.T) {
	assert.PanicsWithValue(t, "ioc.Provide: constructor must be a function", func() {
		Provide[provideTestComponent](&provideTestComponent{})
	})
}

func TestProvide_WrongReturnType(t *testing.T) {
	assert.Panics(t, func() {
		Provide[provideTestComponent](func() *provideTestImpl {
			return &provideTestImpl{}
		})
	})
}

func TestProvide_TooManyReturns(t *testing.T) {
	assert.Panics(t, func() {
		Provide[provideTestComponent](func() (*provideTestComponent, error, int) {
			return nil, nil, 0
		})
	})
}

func TestProvide_SecondReturnNotError(t *testing.T) {
	assert.Panics(t, func() {
		Provide[provideTestComponent](func() (*provideTestComponent, string) {
			return nil, ""
		})
	})
}

func TestProvide_InterfaceNotImplemented(t *testing.T) {
	assert.Panics(t, func() {
		Provide[provideTestIface](func() *provideTestComponent {
			return &provideTestComponent{}
		})
	})
}
