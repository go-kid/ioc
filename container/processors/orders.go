package processors

const (
	PriorityOrderLoggerAware = 1 << (iota + 1)
	PriorityOrderPropertyConfigQuoteAware
	PriorityOrderPropertyExpressionTagAware
	PriorityOrderPopulateProperties
)

const (
	OrderConstructorAware = 1 << (iota + 1)
	OrderDependencyAware
	OrderDependencyFurtherMatching
	OrderValidate
)
