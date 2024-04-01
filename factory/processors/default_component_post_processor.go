package processors

type DefaultComponentPostProcessor struct {
}

func (d *DefaultComponentPostProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
	return component, nil
}

func (d *DefaultComponentPostProcessor) PostProcessAfterInitialization(component any, componentName string) (any, error) {
	return component, nil
}
