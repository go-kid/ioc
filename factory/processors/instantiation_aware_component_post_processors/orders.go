package instantiation_aware_component_post_processors

const (
	PriorityOrderPropertyConfigQuoteAware = iota
	PriorityOrderPropertyExpressionTagAware
	PriorityOrderPopulateProperties
)

const (
	OrderDependencyAware = iota
	OrderDependencyFurtherMatching
	OrderValidate
)
