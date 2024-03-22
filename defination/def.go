package defination

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

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component interface{}, componentName string) error
	PostProcessAfterInitialization(component interface{}, componentName string) error
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
