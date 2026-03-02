package component_definition

type PropertyManager struct {
	propertyGroup map[PropertyType][]*Property
}

func newPropertyManager() PropertyManager {
	return PropertyManager{
		propertyGroup: make(map[PropertyType][]*Property),
	}
}

func (pm *PropertyManager) SetProperties(properties ...*Property) {
	for _, prop := range properties {
		pm.propertyGroup[prop.PropertyType] = append(pm.propertyGroup[prop.PropertyType], prop)
	}
}

func (pm *PropertyManager) GetProperties(t PropertyType) []*Property {
	return pm.propertyGroup[t]
}

func (pm *PropertyManager) GetComponentProperties() []*Property {
	return pm.GetProperties(PropertyTypeComponent)
}

func (pm *PropertyManager) GetConfigurationProperties() []*Property {
	return pm.GetProperties(PropertyTypeConfiguration)
}

func (pm *PropertyManager) GetAllProperties() []*Property {
	var props []*Property
	for _, groupNodes := range pm.propertyGroup {
		props = append(props, groupNodes...)
	}
	return props
}
