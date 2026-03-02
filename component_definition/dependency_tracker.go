package component_definition

import (
	"github.com/go-kid/ioc/util/sync2"
	"github.com/samber/lo"
)

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

func (dt *DependencyTracker) GetDependents() []string {
	return lo.Map(dt.Dependent, func(m *Meta, _ int) string { return m.Name() })
}
