package definition

type InitializingComponent interface {
	AfterPropertiesSet() error
}

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component any, componentName string) (any, error)
	PostProcessAfterInitialization(component any, componentName string) (any, error)
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
