package scan

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var testFile = `package test

// Starter @Component
type Starter struct {
}

// NewStarter @Component
func NewStarter() *Starter {
	return &Starter{}
}
`

func Test_analyseFile(t *testing.T) {
	registers, err := analyseFile([]byte(testFile))
	assert.NoError(t, err)
	assert.Equal(t, "Starter", registers[0].Name)
	assert.Equal(t, "Component", registers[0].Group)
	assert.Equal(t, "type", registers[0].Kind)
	assert.Equal(t, "github.com/go-kid/ioc/cmd/kioc/test", registers[0].Path)
	assert.Equal(t, "test", registers[0].Pkg)
}
