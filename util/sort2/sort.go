package sort2

import "sort"

func Slice[T any](x []T, less func(i T, j T) bool) {
	sort.Slice(x, func(i, j int) bool {
		return less(x[i], x[j])
	})
}
