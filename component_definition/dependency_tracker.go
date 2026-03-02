package component_definition

import "github.com/go-kid/ioc/util/sync2"

type DependencyTracker struct {
	dependentSet *sync2.Map[string, struct{}]
	Dependent    []*Meta
}

func newDependencyTracker() DependencyTracker {
	return DependencyTracker{
		dependentSet: sync2.New[string, struct{}](),
	}
}

func (dt *DependencyTracker) dependOn(dependent *Meta) {
	_, loaded := dt.dependentSet.LoadOrStore(dependent.ID(), struct{}{})
	if !loaded {
		dt.Dependent = append(dt.Dependent, dependent)
	}
}

func (dt *DependencyTracker) GetDependents() (names []string) {
	for _, meta := range dt.Dependent {
		names = append(names, meta.Name())
	}
	return
}
