package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Component struct {
	Name string
}

func Test_Pointer_Without_Specifying_Name(t *testing.T) {
	var tApp = &struct {
		T *Component `wire:""`
	}{}
	ioc.RunTest(t, app.SetComponents(
		&Component{"TestComponent"},
		tApp,
	))
	assert.Equal(t, "TestComponent", tApp.T.Name)
}
