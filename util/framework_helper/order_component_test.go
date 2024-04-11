package framework_helper

import (
	"reflect"
	"testing"
)

type C interface {
	c()
}

type POC int

func (c POC) c() {}

func (c POC) Priority() {
}

func (c POC) Order() int {
	return int(c)
}

type OC int

func (c OC) c() {}
func (c OC) Order() int {
	return int(c)
}

type NC int

func (c NC) c() {}

func TestSortOrderedComponents(t *testing.T) {
	type args[T any] struct {
		components []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[C /* TODO: Insert concrete types here */]{
		{
			name: "",
			args: args[C]{
				components: []C{NC(0), POC(1), OC(99), POC(98), OC(0)},
			},
			want: []C{POC(1), POC(98), OC(0), OC(99), NC(0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SortOrderedComponents(tt.args.components); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SortOrderedComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
