package definition

const (
	//System Components tag
	LoggerTag = "logger"

	//components tag
	InjectTag = "wire"
	FuncTag   = "func"

	//configuration tag
	ValueTag  = "value"
	PropTag   = "prop" //`prop` tag is alias to `value:"${prop_value}"`
	PrefixTag = "prefix"
)
