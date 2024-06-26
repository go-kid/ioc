package configure

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"testing"
)

func TestValueTagValidate(t *testing.T) {
	t.Run("Validate", func(t *testing.T) {
		type C struct {
			S string `yaml:"s" validate:"eq=abc"`
			B bool   `validate:"required"`
		}
		type T struct {
			C  *C      `value:"map[s:abc b:true],validate"`
			S  string  `value:"abc,validate=eq=abc"`
			SP *string `value:"abc,validate=eq=abc"`
			B  bool    `value:"true,validate=required"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			//app.LogTrace,
			app.SetComponents(t2),
		)
		fmt.Println(*t2.C)
	})
	t.Run("Multi-Vars", func(t *testing.T) {
		type T struct {
			S string `value:"123,validate=eq=123 number"`
		}
		t2 := &T{}
		ioc.RunTest(t,
			app.LogTrace,
			app.SetComponents(t2),
		)
	})
	t.Run("ValidateFailed", func(t *testing.T) {
		t.Run("Struct", func(t *testing.T) {
			type C struct {
				S string `json:"s" validate:"eq=abcd"`
			}
			type T struct {
				C *C `value:"{\"s\":\"abc\"},validate"`
			}
			t2 := &T{}
			ioc.RunErrorTest(t, app.SetComponents(t2))
		})
		t.Run("Var", func(t *testing.T) {
			type T struct {
				S string `value:"abc,validate=eq=abcd"`
			}
			t2 := &T{}
			ioc.RunErrorTest(t, app.SetComponents(t2))
		})
		t.Run("Multi-Vars", func(t *testing.T) {
			type T struct {
				S string `value:"abc,validate=eq=abc number=true"`
			}
			t2 := &T{}
			ioc.RunErrorTest(t, app.SetComponents(t2))
		})
	})
}
