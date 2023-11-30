package scan

import (
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/parser"
	"go/token"
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
	fset := token.NewFileSet()
	var f *ast.File
	f, err := parser.ParseFile(fset, "", testFile, parser.ParseComments)
	assert.NoError(t, err)
	registers, err := analyseToken(f)
	assert.NoError(t, err)
	assert.Equal(t, "Starter", registers[0].Name)
	assert.Equal(t, "Component", registers[0].Group)
	assert.Equal(t, "type", registers[0].Kind)
	assert.Equal(t, "test", registers[0].Pkg)
}
