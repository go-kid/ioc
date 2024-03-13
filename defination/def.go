package defination

type InitializeComponent interface {
	Init() error
}

type NamingComponent interface {
	Naming() string
}

type ApplicationRunner interface {
	Order() int
	Run() error
}

type Configuration interface {
	Prefix() string
}

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component interface{}) error
	PostProcessAfterInitialization(component interface{}) error
}

type CloserComponent interface {
	Close() error
}
