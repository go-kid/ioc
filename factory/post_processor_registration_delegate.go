package factory

import "github.com/go-kid/ioc/factory/processors"

type PostProcessorRegistrationDelegate struct {
	componentPostProcessors                       []processors.ComponentPostProcessor
	instantiationAwareComponentPostProcessor      []processors.InstantiationAwareComponentPostProcessor
	smartInstantiationAwareBeanPostProcessor      []processors.SmartInstantiationAwareBeanPostProcessor
	defaultTagScanDefinitionRegistryPostProcessor []processors.DefaultTagScanDefinitionRegistryPostProcessor
	hasInstantiationAwareComponentPostProcessor   bool
	hasDestructionAwareComponentPostProcessor     bool
}

func (p *PostProcessorRegistrationDelegate) RegisterComponentPostProcessors(ps processors.ComponentPostProcessor) {

	if _, ok := ps.(processors.InstantiationAwareComponentPostProcessor); ok {
		p.hasInstantiationAwareComponentPostProcessor = true
	}
	if _, ok := ps.(processors.DestructionAwareComponentPostProcessor); ok {
		p.hasDestructionAwareComponentPostProcessor = true
	}
	p.componentPostProcessors = append(p.componentPostProcessors, ps.(processors.ComponentPostProcessor))
}

func (p *PostProcessorRegistrationDelegate) invokeBeanFactoryPostProcessors(factory Factory, processors []ComponentFactoryPostProcessor) error {
	for _, processor := range processors {
		err := processor.PostProcessComponentFactory(factory)
		if err != nil {
			return err
		}
	}

	for name, component := range factory.GetRegisteredComponents() {
		for _, processor := range factory.GetDefinitionRegistryPostProcessors() {
			err := processor.PostProcessDefinitionRegistry(factory.GetDefinitionRegistry(), component, name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
