package processors

import (
	"testing"

	"github.com/go-kid/ioc/component_definition"
	"github.com/go-kid/ioc/configure"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/container/support"
	"github.com/go-kid/ioc/definition"
	"github.com/go-kid/ioc/syslog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type processorTarget struct {
	Dependency *processorDependency `wire:",required=false"`
	Value      int                  `value:"42"`
	Expression int                  `value:"#{1 + 2}"`
	Logger     syslog.Logger        `logger:"processor-test"`
	Required   string
}

type processorDependency struct{}

func propertyFor(t *testing.T, meta *component_definition.Meta, fieldName string, propertyType component_definition.PropertyType, tag, value string) *component_definition.Property {
	t.Helper()
	for _, field := range meta.Fields {
		if field.StructField.Name == fieldName {
			return component_definition.NewProperty(field, propertyType, tag, value)
		}
	}
	t.Fatalf("field %s not found", fieldName)
	return nil
}

func TestDefaultProcessors(t *testing.T) {
	component := &processorTarget{}
	base := &DefaultComponentPostProcessor{}
	result, err := base.PostProcessBeforeInitialization(component, "component")
	require.NoError(t, err)
	assert.Same(t, component, result)
	result, err = base.PostProcessAfterInitialization(component, "component")
	require.NoError(t, err)
	assert.Same(t, component, result)

	instantiation := &DefaultInstantiationAwareComponentPostProcessor{}
	meta := component_definition.NewMeta(component)
	result, err = instantiation.PostProcessBeforeInstantiation(meta, "component")
	require.NoError(t, err)
	assert.Nil(t, result)
	ok, err := instantiation.PostProcessAfterInstantiation(component, "component")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestTagScanningAndValueProcessors(t *testing.T) {
	component := &processorTarget{}
	registry := support.DefaultDefinitionRegistry()
	scanner := &DefaultTagScanDefinitionRegistryPostProcessor{
		NodeType: component_definition.PropertyTypeComponent,
		Tag:      definition.InjectTag,
		Required: true,
	}
	require.NoError(t, scanner.PostProcessDefinitionRegistry(registry, component, "target"))
	properties := registry.GetMetaByName("target").GetAllProperties()
	require.Len(t, properties, 1)
	assert.Equal(t, "", properties[0].TagVal)
	assert.False(t, properties[0].IsRequired())

	meta := component_definition.NewMeta(component)
	value := propertyFor(t, meta, "Value", component_definition.PropertyTypeConfiguration, definition.ValueTag, "42")
	valueProcessor := NewValueAwarePostProcessors().(*valueAwarePostProcessors)
	_, err := valueProcessor.PostProcessProperties([]*component_definition.Property{value}, component, "target")
	require.NoError(t, err)
	assert.Equal(t, 42, component.Value)

	expression := propertyFor(t, meta, "Expression", component_definition.PropertyTypeConfiguration, definition.ValueTag, "#{1 + 2}")
	expressionProcessor := NewExpressionTagAwarePostProcessors().(*expressionTagAwarePostProcessors)
	_, err = expressionProcessor.PostProcessProperties([]*component_definition.Property{expression}, component, "target")
	require.NoError(t, err)
	assert.Equal(t, "3", expression.TagVal)
}

func TestConfigurationLoggerAndValidationProcessors(t *testing.T) {
	configuration := configure.Default()
	configuration.SetLoaders(loader.NewRawLoader([]byte("service:\n  port: 8080\n")))
	require.NoError(t, configuration.Initialize())

	component := &processorTarget{}
	meta := component_definition.NewMeta(component)
	quoted := propertyFor(t, meta, "Value", component_definition.PropertyTypeConfiguration, definition.ValueTag, "${service.port}")
	quoteProcessor := NewConfigQuoteAwarePostProcessors().(*configQuoteAwarePostProcessors)
	quoteProcessor.Configure = configuration
	_, err := quoteProcessor.PostProcessProperties([]*component_definition.Property{quoted}, component, "target")
	require.NoError(t, err)
	assert.Equal(t, "8080", quoted.TagVal)

	loggerProperty := propertyFor(t, meta, "Logger", "Logger", definition.LoggerTag, "processor-test")
	loggerProcessor := NewLoggerAwarePostProcessor().(*loggerAwarePostProcessors)
	_, err = loggerProcessor.PostProcessProperties([]*component_definition.Property{loggerProperty}, component, "target")
	require.NoError(t, err)
	assert.NotNil(t, component.Logger)

	required := propertyFor(t, meta, "Required", component_definition.PropertyTypeConfiguration, definition.ValueTag, "")
	required.SetArg(ArgValidate, "required")
	validateProcessor := NewValidateAwarePostProcessors().(*validateAwarePostProcessors)
	_, err = validateProcessor.PostProcessProperties([]*component_definition.Property{required}, component, "target")
	assert.Error(t, err)
}

func TestDependencyFiltering(t *testing.T) {
	target := &processorTarget{}
	property := propertyFor(t, component_definition.NewMeta(target), "Dependency", component_definition.PropertyTypeComponent, definition.InjectTag, "")
	property.SetArg(component_definition.ArgRequired)
	dependency := component_definition.NewMeta(&processorDependency{})

	filtered, err := filterDependencies(property, []*component_definition.Meta{nil, dependency})
	require.NoError(t, err)
	assert.Equal(t, []*component_definition.Meta{dependency}, filtered)

	property.SetArg(component_definition.ArgQualifier, "missing")
	_, err = filterDependencies(property, []*component_definition.Meta{dependency})
	assert.Error(t, err)
}
