package debug

import (
	"math"

	"github.com/go-kid/ioc/definition"
)

type dryRunPostProcessor struct {
	definition.PriorityComponent
}

func (d *dryRunPostProcessor) Order() int {
	return math.MinInt32
}

func (d *dryRunPostProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
	return nil, nil
}

func (d *dryRunPostProcessor) PostProcessAfterInitialization(component any, componentName string) (any, error) {
	return component, nil
}
