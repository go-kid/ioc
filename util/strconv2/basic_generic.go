package strconv2

import (
	"github.com/pkg/errors"
	"strconv"
)

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
		return tVal, errors.Errorf("not supported to parse value %s as %T", val, tVal)
	}
}
