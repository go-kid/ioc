package instantiation_aware_component_post_processors

import "github.com/go-kid/ioc/processors"

type dependencyNameAwarePostProcessors struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (d *dependencyNameAwarePostProcessors) PostProcessAfterInstantiation(component any, componentName string) (bool, error) {
	return true, nil
}
