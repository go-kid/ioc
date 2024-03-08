package builtin_inject

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Pointer_Slice(t *testing.T) {
	var tApp = &struct {
		Ts []*SpecifyNameComponent `wire:""`
	}{}
	ioc.RunTest(t, app.SetComponents(
		&SpecifyNameComponent{Component{"SpecifyNameComponent"}},
		&SpecifyNameComponent{Component{"SpecifyNameComponent2"}},
		&Component{"Component"},
		tApp,
	))
	assert.Equal(t, 2, len(tApp.Ts))
}
