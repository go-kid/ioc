package defination

import "github.com/go-kid/ioc/scanner/meta"

type InitializingComponent interface {
	Initializing() error
}

type ComponentPostInitializingProcessor interface {
	PostProcessBeforeInitializing(m *meta.Meta) (*meta.Meta, error)
	PostProcessAfterInitializing(m *meta.Meta) (*meta.Meta, error)
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

type ComponentPostProcessor interface {
	PostProcessBeforeInitialization(component interface{}) error
	PostProcessAfterInitialization(component interface{}) error
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
