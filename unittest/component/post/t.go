package main

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/configure/loader"
	"github.com/go-kid/ioc/container/processors"
	"github.com/go-kid/ioc/syslog"
)

type Service interface {
	SayName()
}

type LogParent struct {
	Logger syslog.Logger `logger:",embed"`
}

type ServiceA struct {
	Name     string
	ServiceB *ServiceB `wire:""`
	ServiceC *ServiceC `wire:""`
	ServiceA Service   `wire:""`
	LogParent
}

func (s *ServiceA) Qualifier() string {
	return ""
}

func (s *ServiceA) SayName() {
	fmt.Println(s.Name)
}

type ServiceB struct {
	Name     string
	ServiceA Service `wire:""`
}

type ServiceC struct {
	Name     string
	ServiceA Service `wire:""`
}

type serviceAAOP struct {
	Service
	Enable bool
}

func (s *serviceAAOP) SayName() {
	if s.Enable {
		fmt.Printf("aop %p before say name\n", s)
	}
	s.Service.SayName()
	if s.Enable {
		fmt.Printf("aop %p after say name\n", s)
	}
}

type postProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
	//definition.LazyInitComponent
	Enable bool `value:"${aop.enable:false}"`
}

func (p *postProcessor) GetEarlyBeanReference(component any, componentName string) (any, error) {
	if s, ok := component.(*ServiceA); ok {
		return &serviceAAOP{
			Service: s,
			Enable:  p.Enable,
		}, nil
	}
	return component, nil
}

func main() {
	a := &ServiceA{Name: "service A"}
	b := &ServiceB{Name: "service B"}
	c := &ServiceC{Name: "service C"}
	run, err := ioc.Run(
		//app.LogTrace,
		app.AddConfigLoader(loader.NewRawLoader([]byte(`aop:
  enable: true`))),
		app.SetComponents(
			a,
			b,
			c,
			&postProcessor{},
		))
	if err != nil {
		panic(err)
	}
	defer run.Close()
	fmt.Println(a.ServiceB.Name)
	fmt.Println(a.ServiceC.Name)
	b.ServiceA.SayName()
	c.ServiceA.SayName()
	a.ServiceA.SayName()

	a.Logger.Info("hello")

	//fmt.Printf("a: %T\n", a.ServiceA)
	//a.ServiceA.SayName()
	//fmt.Printf("b: %T\n", b.ServiceA)
	//b.ServiceA.SayName()
	//fmt.Printf("c: %T\n", c.ServiceA)
	//c.ServiceA.SayName()
}
