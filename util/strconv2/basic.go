package strconv2

import "strconv"

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
