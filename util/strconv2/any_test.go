package strconv2

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("ParseString", func(t *testing.T) {
		parse, err := Parse[string]("abc")
		assert.NoError(t, err)
		assert.Equal(t, "abc", parse)
	})
	t.Run("ParseInt", func(t *testing.T) {
		i, err := Parse[int]("123")
		assert.NoError(t, err)
		assert.Equal(t, 123, i)
		i8, err := Parse[int8]("2")
		assert.NoError(t, err)
		assert.Equal(t, int8(2), i8)
		ui, err := Parse[uint]("123")
		assert.NoError(t, err)
		assert.Equal(t, uint(123), ui)
		ui8, err := Parse[uint8]("123")
		assert.NoError(t, err)
		assert.Equal(t, uint8(123), ui8)
	})
	t.Run("ParseFloat", func(t *testing.T) {
		f64, err := Parse[float64]("123.456")
		assert.NoError(t, err)
		assert.Equal(t, 123.456, f64)
		f32, err := Parse[float32]("123.456")
		assert.NoError(t, err)
		assert.Equal(t, float32(123.456), f32)
	})
	t.Run("ParseBool", func(t *testing.T) {
		parse, err := Parse[bool]("true")
		assert.NoError(t, err)
		assert.Equal(t, true, parse)
	})
	t.Run("ParseBool", func(t *testing.T) {
		parse, err := Parse[bool]("false")
		assert.NoError(t, err)
		assert.Equal(t, false, parse)
	})
}

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
			want: int64(123),
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
