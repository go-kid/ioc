package strconv2

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	intReg   = regexp.MustCompile("^\\d+$")
	floatReg = regexp.MustCompile("^\\d+\\.\\d+$")
)

func ParseAny(val string) (any, error) {
	var typeVal any
	var err error
	switch true {
	case val == "":
		typeVal = ""
	case val == "true":
		typeVal = true
	case val == "false":
		typeVal = false
	case intReg.MatchString(val):
		typeVal, err = strconv.ParseInt(val, 10, 64)
	case floatReg.MatchString(val):
		typeVal, err = strconv.ParseFloat(val, 64)
	case isMap(val):
		typeVal, err = ParseAnyMap(val)
	case isSlice(val):
		typeVal, err = ParseAnySlice(val)
	default:
		typeVal = val
	}
	return typeVal, err
}

func Parse[T any](val string) (T, error) {
	var tVal T
	switch any(tVal).(type) {
	case bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return tVal, err
		}
		return any(b).(T), nil
	case int64:
		i, err := ParseInt64(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case int:
		i, err := ParseInt(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case int8:
		i, err := ParseInt8(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case int16:
		i, err := ParseInt16(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case int32:
		i, err := ParseInt32(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case uint:
		i, err := ParseUint(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case uint64:
		i, err := ParseUint64(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case uint8:
		i, err := ParseUint8(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case uint16:
		i, err := ParseUint16(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case uint32:
		i, err := ParseUint32(val, 10)
		if err != nil {
			return tVal, err
		}
		return any(i).(T), nil
	case float32:
		f, err := ParseFloat32(val)
		if err != nil {
			return tVal, err
		}
		return any(f).(T), nil
	case float64:
		f, err := ParseFloat64(val)
		if err != nil {
			return tVal, err
		}
		return any(f).(T), nil
	case string, any:
		return any(val).(T), nil
	default:
		return tVal, fmt.Errorf("not supported to parse value %s as %T", val, tVal)
	}
}

func ParseInt(s string, base int) (int, error) {
	i, err := strconv.ParseInt(s, base, 64)
	return int(i), err
}

func ParseInt8(s string, base int) (int8, error) {
	i, err := strconv.ParseInt(s, base, 8)
	return int8(i), err
}

func ParseInt16(s string, base int) (int16, error) {
	i, err := strconv.ParseInt(s, base, 16)
	return int16(i), err
}

func ParseInt32(s string, base int) (int32, error) {
	i, err := strconv.ParseInt(s, base, 32)
	return int32(i), err
}

func ParseInt64(s string, base int) (int64, error) {
	return strconv.ParseInt(s, base, 64)
}

func ParseUint(s string, base int) (uint, error) {
	i, err := strconv.ParseUint(s, base, 64)
	return uint(i), err
}

func ParseUint8(s string, base int) (uint8, error) {
	i, err := strconv.ParseUint(s, base, 8)
	return uint8(i), err
}

func ParseUint16(s string, base int) (uint16, error) {
	i, err := strconv.ParseUint(s, base, 16)
	return uint16(i), err
}

func ParseUint32(s string, base int) (uint32, error) {
	i, err := strconv.ParseUint(s, base, 32)
	return uint32(i), err
}

func ParseUint64(s string, base int) (uint64, error) {
	return strconv.ParseUint(s, base, 64)
}

func ParseFloat32(s string) (float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return float32(f), err
}

func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
