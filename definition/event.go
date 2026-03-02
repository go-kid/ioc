package definition

type ApplicationEvent interface {
	Source() interface{}
}

type ApplicationEventListener interface {
	OnEvent(event ApplicationEvent) error
}

type ApplicationEventPublisher interface {
	PublishEvent(event ApplicationEvent) error
}

type ComponentCreatedEvent struct {
	ComponentName string
	Component     interface{}
}

func (e *ComponentCreatedEvent) Source() interface{} { return e.Component }

type ApplicationStartedEvent struct {
	App interface{}
}

func (e *ApplicationStartedEvent) Source() interface{} { return e.App }

type ApplicationClosingEvent struct {
	App interface{}
}

func (e *ApplicationClosingEvent) Source() interface{} { return e.App }
