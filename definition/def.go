package definition

import "context"

type InitializingComponent interface {
	AfterPropertiesSet() error
}

type InitializingComponentWithContext interface {
	AfterPropertiesSet(ctx context.Context) error
}

type InitializeComponent interface {
	Init() error
}

type InitializeComponentWithContext interface {
	Init(ctx context.Context) error
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
	Run() error
}

type ApplicationRunnerWithContext interface {
	RunWithContext(ctx context.Context) error
}

type ConfigurationProperties interface {
	Prefix() string
}

type CloserComponent interface {
	Close() error
}

type CloserComponentWithContext interface {
	CloseWithContext(ctx context.Context) error
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

const (
	ScopeSingleton = "singleton"
	ScopePrototype = "prototype"
)

type ScopeComponent interface {
	Scope() string
}

type ConditionContext interface {
	HasComponent(name string) bool
	GetConfig(key string) interface{}
}

type ConditionalComponent interface {
	Condition(ctx ConditionContext) bool
}
