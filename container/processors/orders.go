package processors

const (
	PriorityOrderPropertyConfigQuoteAware = 1 << (iota + 1)
	PriorityOrderPropertyExpressionTagAware
	PriorityOrderPopulateProperties
)

const (
	OrderDependencyAware = 1 << (iota + 1)
	OrderDependencyFurtherMatching
	OrderValidate
)
