package definition

type InitializingComponent interface {
	AfterPropertiesSet() error
}

type InitializeComponent interface {
	Init() error
}

type NamingComponent interface {
	Naming() string
}

type Ordered interface {
	Order() int
}

type Priority interface {
	Priority()
}

type PriorityOrdered interface {
	Priority
	Ordered
}

type ApplicationRunner interface {
	Ordered
	Run() error
}

type Configuration interface {
	Prefix() string
}

type CloserComponent interface {
	Close() error
}

type WireQualifier interface {
	Qualifier() string
}

type WirePrimary interface {
	Primary()
}

type LazyInit interface {
	LazyInit()
}
