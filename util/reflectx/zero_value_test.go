package reflectx

import (
	"encoding/json"
	"reflect"
	"testing"
)

type testStruct struct {
	Name string
	Self *testStruct
	Sub  *sub
}

func (t *testStruct) String() string {
	bytes, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

type sub struct {
	SubName string
}

func TestZeroValue(t *testing.T) {

	type args struct {
		p            reflect.Type
		interceptors []ZeroValueInterceptor
	}

	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "string",
			args: args{
				p: reflect.TypeOf(""),
			},
			want: "string",
		},
		{
			name: "int",
			args: args{
				p: reflect.TypeOf(1),
			},
			want: 0,
		},
		{
			name: "float",
			args: args{
				p: reflect.TypeOf(1.1),
			},
			want: 0.0,
		},
		{
			name: "bool",
			args: args{
				p: reflect.TypeOf(true),
			},
			want: false,
		},
		{
			name: "array",
			args: args{
				p: reflect.TypeOf([1]string{"a"}),
			},
			want: [1]string{"string"},
		},
		{
			name: "any array",
			args: args{
				p: reflect.TypeOf([1]any{"a"}),
			},
			want: [1]any{nil},
		},
		{
			name: "slice",
			args: args{
				p: reflect.TypeOf([]string{"a"}),
			},
			want: []string{"string"},
		},
		{
			name: "any slice",
			args: args{
				p: reflect.TypeOf([]any{"a"}),
			},
			want: []any{nil},
		},
		{
			name: "map",
			args: args{
				p: reflect.TypeOf(map[string]int{"a": 123}),
			},
			want: map[string]int{"string": 0},
		},
		{
			name: "any key map",
			args: args{
				p: reflect.TypeOf(map[any]int{"a": 123}),
			},
			want: map[any]int{},
		},
		{
			name: "any value map",
			args: args{
				p: reflect.TypeOf(map[string]any{"a": 123}),
			},
			want: map[string]any{},
		},
		{
			name: "any map",
			args: args{
				p: reflect.TypeOf(map[any]any{"a": 123}),
			},
			want: map[any]any{},
		},
		{
			name: "struct",
			args: args{
				p: reflect.TypeOf(testStruct{Name: "abc"}),
			},
			want: testStruct{
				Name: "string",
				Self: &testStruct{
					Name: "string",
				},
				Sub: &sub{
					SubName: "string",
				},
			},
		},
		{
			name: "pointer struct",
			args: args{
				p: reflect.TypeOf(&testStruct{Name: "abc"}),
			},
			want: &testStruct{
				Name: "string",
				Self: &testStruct{
					Name: "string",
				},
				Sub: &sub{SubName: "string"},
			},
		},
		{
			name: "interceptor json.RawMessage",
			args: args{
				p: reflect.TypeOf(struct {
					M json.RawMessage `json:"m"`
				}{}),
				interceptors: []ZeroValueInterceptor{
					JsonZero,
				},
			},
			want: struct {
				M json.RawMessage `json:"m"`
			}{
				M: json.RawMessage("{}"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ZeroValue(tt.args.p, tt.args.interceptors...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ZeroValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
