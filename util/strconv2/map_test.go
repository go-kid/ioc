package strconv2

import (
	"github.com/go-kid/ioc/util/strings2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseAnyMap(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "0",
			args: args{
				val: "map[b:[map[a:b],map[c:d]]]",
			},
			want: map[string]any{
				"b": []any{
					map[string]any{
						"a": "b",
					},
					map[string]any{
						"c": "d",
					},
				},
			},
		},
		{
			name: "1",
			args: args{
				val: "map[aes:map[iv:abc key:123] header:[X-Request-Id,X-Cross-Origin,X-Allowed-Method]]",
			},
			want: map[string]any{
				"aes": map[string]any{
					"iv":  "abc",
					"key": float64(123),
				},
				"header": []any{"X-Request-Id", "X-Cross-Origin", "X-Allowed-Method"},
			},
		},
		{
			name: "2",
			args: args{
				val: "map[a:map[b:map[c:[d,e]]] b:[map[a:b],map[c:d]]]",
			},
			want: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": []any{"d", "e"},
					},
				},
				"b": []any{
					map[string]any{
						"a": "b",
					},
					map[string]any{
						"c": "d",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAnyMap(tt.args.val)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "ParseAnyMap(%v)", tt.args.val)
		})
	}
}

func Test_splitMapPart(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "1",
			args: args{
				val: "aes:map[iv:abc key:123] header:[X-Request-Id X-Cross-Origin X-Allowed-Method]",
			},
			want: []string{"aes:map[iv:abc key:123]", "header:[X-Request-Id X-Cross-Origin X-Allowed-Method]"},
		},
		{
			name: "2",
			args: args{
				val: "a:map[b:map[c:[d e]]] b:[map[a:b] map[c:d]]",
			},
			want: []string{"a:map[b:map[c:[d e]]]", "b:[map[a:b] map[c:d]]"},
		},
		{
			name: "3",
			args: args{
				val: "iv:abc key:123",
			},
			want: []string{"iv:abc", "key:123"},
		},
		{
			name: "5",
			args: args{
				val: "b:[map[a:b] map[c:d]]",
			},
			want: []string{"b:[map[a:b] map[c:d]]"},
		},
		//{
		//	name: "6",
		//	args: args{
		//		val: "b: map[c:d]]",
		//	},
		//	want: []string{"b:[map[a:b] map[c:d]]"},
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, strings2.SplitWithConfig(tt.args.val, mapSplitConfig), "splitKVPairs(%v)", tt.args.val)
		})
	}
}
