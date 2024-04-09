package properties

import (
	"reflect"
	"testing"
)

func TestNewFromMap(t *testing.T) {
	type args struct {
		m map[string]any
	}
	tests := []struct {
		name string
		args args
		want Properties
	}{
		{
			name: "1",
			args: args{
				m: NewFromMap(map[string]any{
					"a": map[string]any{
						"b": map[string]any{
							"c":  123,
							"c2": "foo",
						},
						"b2": "bar",
					},
				}),
			},
			want: map[string]any{
				"a.b.c":  123,
				"a.b.c2": "foo",
				"a.b2":   "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFromMap(tt.args.m); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProperties_Expand(t *testing.T) {
	type C struct {
		A string
	}
	tests := []struct {
		name string
		p    Properties
		want map[string]any
	}{
		{
			name: "1",
			p: Properties{
				"a.b.c":  123,
				"a.b.c2": "foo",
				"a.b.c3": &C{A: "abc"},
				"a.b2":   "bar",
				"a.c":    &C{A: "abc"},
				"c":      &C{A: "abc"},
			},
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c":  123,
						"c2": "foo",
						"c3": &C{A: "abc"},
					},
					"b2": "bar",
					"c":  &C{A: "abc"},
				},
				"c": &C{A: "abc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Expand(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProperties_Set_Get(t *testing.T) {
	type args struct {
		key string
		val any
	}
	properties := New()
	tests := []struct {
		name string
		p    Properties
		args args
	}{
		{
			name: "1",
			p:    properties,
			args: args{
				key: "a.b.c",
				val: 123,
			},
		},
		{
			name: "1",
			p:    properties,
			args: args{
				key: "a.b.c2",
				val: "foo",
			},
		},
		{
			name: "1",
			p:    properties,
			args: args{
				key: "a.b2",
				val: "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Set(tt.args.key, tt.args.val)
		})
	}

	tests2 := []struct {
		name  string
		p     Properties
		args  args
		want  any
		want1 bool
	}{
		{
			name:  "1",
			p:     properties,
			args:  tests[0].args,
			want:  tests[0].args.val,
			want1: true,
		},
		{
			name:  "2",
			p:     properties,
			args:  tests[1].args,
			want:  tests[1].args.val,
			want1: true,
		},
		{
			name:  "3",
			p:     properties,
			args:  tests[2].args,
			want:  tests[2].args.val,
			want1: true,
		},
		{
			name: "4",
			p:    properties,
			args: args{
				key: "abc.def",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.p.Get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
