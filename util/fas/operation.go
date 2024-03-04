package fas

import "reflect"

// TernaryOp is ternary operation like max = a > b ? a : b
func TernaryOp[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

func TernaryOpNil[T any](condition bool, a, b func() T) T {
	if condition {
		return a()
	}
	return b()
}

type Comparable interface {
	Uint | Int | Float | ~string | ~byte | ~rune
}

type Uint interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint | ~uintptr
}

type Int interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int
}

type Float interface {
	~float32 | ~float64
}

func Max[T Comparable](a, b T) T {
	return TernaryOp(a > b, a, b)
}

func Min[T Comparable](a, b T) T {
	return TernaryOp(a < b, a, b)
}

func IsNil(a any) bool {
	if a == nil {
		return true
	}
	if reflect.ValueOf(a).IsNil() {
		return true
	}
	return false
}
