package framework_helper

import (
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/util/sort2"
)

func SortOrderedComponents[T any](components []T) []T {
	var (
		ordered                   = make([]T, 0, len(components))
		priorityOrderedComponents []T
		orderedComponents         []T
		noneOrderedComponents     []T
	)
	for _, component := range components {
		if oc, ok := any(component).(definition.Ordered); ok {
			if _, ok := oc.(definition.Priority); ok {
				priorityOrderedComponents = append(priorityOrderedComponents, component)
			} else {
				orderedComponents = append(orderedComponents, component)
			}
		} else {
			noneOrderedComponents = append(noneOrderedComponents, component)
		}
	}

	sort2.Slice(priorityOrderedComponents, orderedComponentComparator[T])
	sort2.Slice(orderedComponents, orderedComponentComparator[T])
	ordered = append(ordered, priorityOrderedComponents...)
	ordered = append(ordered, orderedComponents...)
	ordered = append(ordered, noneOrderedComponents...)
	return ordered
}

func orderedComponentComparator[T any](i, j T) bool {
	return any(i).(definition.Ordered).Order() < any(j).(definition.Ordered).Order()
}
