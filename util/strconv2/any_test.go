package strconv2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestParseAny(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				val: "",
			},
			want: "",
		},
		{
			name: "2",
			args: args{
				val: "123",
			},
			want: float64(123),
		},
		{
			name: "3",
			args: args{
				val: "123.123",
			},
			want: 123.123,
		},
		{
			name: "4",
			args: args{
				val: "true",
			},
			want: true,
		},
		{
			name: "5",
			args: args{
				val: "false",
			},
			want: false,
		},
		{
			name: "6",
			args: args{
				val: "abc",
			},
			want: "abc",
		},
		{
			name: "7",
			args: args{
				val: "[foo,bar]",
			},
			want: []any{"foo", "bar"},
		},
		{
			name: "8",
			args: args{
				val: `["foo","bar"]`,
			},
			want: []any{"foo", "bar"},
		},
		{
			name: "8",
			args: args{
				val: `[1,"2",true,2.33]`,
			},
			want: []any{float64(1), "2", true, 2.33},
		},
		{
			name: "9",
			args: args{
				val: `map[a:b]`,
			},
			want: map[string]any{"a": "b"},
		},
		{
			name: "10",
			args: args{
				val: `{"foo":"bar"}`,
			},
			want: map[string]any{"foo": "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAny(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAnySlice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseAny() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatAny(t *testing.T) {
	type args struct {
		a any
	}
	type ST struct {
		A any `json:"a"`
		B any `json:"b"`
		C any `json:"c"`
		D any `json:"d"`
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "1",
			args: args{
				a: 1,
			},
			want:    "1",
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: "foo",
			},
			want:    "foo",
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: true,
			},
			want:    "true",
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []int{1, 2, 3},
			},
			want:    "[1,2,3]",
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []string{"foo", "bar"},
			},
			want:    `["foo","bar"]`,
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []any{1, "2", true, 2.33},
			},
			want:    `[1,"2",true,2.33]`,
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: map[string]any{
					"a": 1,
					"b": "2",
					"c": true,
					"d": 2.33,
				},
			},
			want:    `{"a":1,"b":"2","c":true,"d":2.33}`,
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: ST{
					A: 1,
					B: "2",
					C: true,
					D: 2.33,
				},
			},
			want:    `{"a":1,"b":"2","c":true,"d":2.33}`,
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: &ST{
					A: 1,
					B: "2",
					C: true,
					D: 2.33,
				},
			},
			want:    `{"a":1,"b":"2","c":true,"d":2.33}`,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatAny(tt.args.a)
			if !tt.wantErr(t, err, fmt.Sprintf("FormatAny(%v)", tt.args.a)) {
				return
			}
			assert.Equalf(t, tt.want, got, "FormatAny(%v)", tt.args.a)
		})
	}
}

func TestFormatLoopAny(t *testing.T) {
	type args struct {
		a any
	}
	type ST struct {
		A any `json:"a"`
		B any `json:"b"`
		C any `json:"c"`
		D any `json:"d"`
	}
	tests := []struct {
		name    string
		args    args
		want    any
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "1",
			args: args{
				a: float64(1),
			},
			want:    float64(1),
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: "foo",
			},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: true,
			},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []int{1, 2, 3},
			},
			want:    []any{float64(1), float64(2), float64(3)},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []string{"foo", "bar"},
			},
			want:    []any{"foo", "bar"},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: []any{1, "2", true, 2.33},
			},
			want:    []any{float64(1), "2", true, 2.33},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: map[string]any{
					"a": 1,
					"b": "2",
					"c": true,
					"d": 2.33,
				},
			},
			want: map[string]any{
				"a": float64(1),
				"b": "2",
				"c": true,
				"d": 2.33,
			},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: ST{
					A: 1,
					B: "2",
					C: true,
					D: 2.33,
				},
			},
			want: map[string]any{
				"a": float64(1),
				"b": "2",
				"c": true,
				"d": 2.33,
			},
			wantErr: assert.NoError,
		},
		{
			name: "1",
			args: args{
				a: &ST{
					A: 1,
					B: "2",
					C: true,
					D: 2.33,
				},
			},
			want: map[string]any{
				"a": float64(1),
				"b": "2",
				"c": true,
				"d": 2.33,
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatAny(tt.args.a)
			if !tt.wantErr(t, err, fmt.Sprintf("FormatAny(%v)", tt.args.a)) {
				return
			}
			parsed, err := ParseAny(got)
			if !tt.wantErr(t, err, fmt.Sprintf("ParseAny(%v)", tt.args.a)) {
				return
			}
			if tt.want == nil {
				assert.Equal(t, tt.args.a, parsed)
			} else {
				assert.Equal(t, tt.want, parsed)
			}
		})
	}
}
