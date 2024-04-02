package configure

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExpressionTag(t *testing.T) {
	t.Run("StaticExpression", func(t *testing.T) {
		type T struct {
			Arithmetic  int    `value:"#{1+(1*2)}"`
			Comparison  bool   `value:"#{1/1==1}"`
			Logical     bool   `value:"#{(1+1)>=2||1!=1}"`
			Conditional string `value:"#{1>2?'a':'b'}"`
			Membership  bool   `value:"#{'a' in ['a','b','c']}"`
			String      bool   `value:"#{'hello'+' '+'world' contains 'o w'}"`
		}
		var t2 = &T{}
		ioc.RunTest(t, app.LogDebug, app.SetComponents(t2))
		assert.Equal(t, 3, t2.Arithmetic)
		assert.True(t, t2.Comparison)
		assert.True(t, t2.Logical)
		assert.Equal(t, "b", t2.Conditional)
		assert.True(t, t2.Membership)
		assert.True(t, t2.String)
	})
	t.Run("ExpressionWithDefaultConfigQuote", func(t *testing.T) {
		type T struct {
			Arithmetic  int    `value:"#{${:1}+(${:1}*${:2})}"`
			Comparison  bool   `value:"#{${:1}/${:1}==${:1}}"`
			Logical     bool   `value:"#{(1+1)${:>=}2||1!=1}"`
			Conditional string `value:"#{1>2?'${:'a'}':'${:'b'}'}"`
			Membership  bool   `value:"#{'a' in ${:[a,'a','b',c,1,3.14,true]}}"`
			String      bool   `value:"#{'${:'hello'}'+' '+'${:'world'}' contains 'o w'}"`
		}
		var t2 = &T{}
		ioc.RunTest(t, app.LogDebug, app.SetComponents(t2))
		assert.Equal(t, 3, t2.Arithmetic)
		assert.True(t, t2.Comparison)
		assert.True(t, t2.Logical)
		assert.Equal(t, "b", t2.Conditional)
		assert.True(t, t2.Membership)
		assert.True(t, t2.String)
	})
	t.Run("ExpressionWithConfigQuote", func(t *testing.T) {
		type T struct {
			Arithmetic  int    `value:"#{${number.val1}+(${number.val1}*${number.val2})}"`
			Comparison  bool   `value:"#{${number.val1}/${number.val1}==${number.val1}}"`
			Logical     bool   `value:"#{(1+1)${logical.compare}2||1!=1}"`
			Conditional string `value:"#{1>2?'${character.val1}':'${character.val2}'}"`
			Membership  bool   `value:"#{'a' in ${slices}}"`
			String      bool   `value:"#{'${character.val3}'+' '+'${character.val4}' contains 'o w'}"`
		}
		var t2 = &T{}
		ioc.RunTest(t, app.LogDebug, app.SetComponents(t2),
			app.AddConfigLoader(loader.NewRawLoader([]byte(`
number:
  val1: 1
  val2: 2
logical:
  compare: ">="
character:
  val1: a
  val2: b
  val3: "hello"
  val4: "world"
slices:
  - a
  - 'a'
  - 'b'
  - c
  - 1
  - 3.14
  - true
`))))
		assert.Equal(t, 3, t2.Arithmetic)
		assert.True(t, t2.Comparison)
		assert.True(t, t2.Logical)
		assert.Equal(t, "b", t2.Conditional)
		assert.True(t, t2.Membership)
		assert.True(t, t2.String)
	})
}
