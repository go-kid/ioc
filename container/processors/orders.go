package processors

const (
	PriorityOrderLoggerAware = 1 << (iota + 1)
	PriorityOrderPropertyConfigQuoteAware
	PriorityOrderPropertyExpressionTagAware
	PriorityOrderPopulateProperties
)

const (
	OrderDependencyAware = 1 << (iota + 1)
	OrderDependencyFurtherMatching
	OrderValidate
)
