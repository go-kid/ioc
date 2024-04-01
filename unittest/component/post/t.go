package main

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/factory/processors"
)

type Service interface {
	SayName()
}

type ServiceA struct {
	Name     string
	ServiceB *ServiceB `wire:""`
	ServiceC *ServiceC `wire:""`
	ServiceA Service   `wire:""`
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
}

func (s *serviceAAOP) SayName() {
	fmt.Println("aop before say name")
	s.Service.SayName()
	fmt.Println("aop after say name")
}

type postProcessor struct {
	processors.DefaultInstantiationAwareComponentPostProcessor
}

func (p *postProcessor) GetEarlyBeanReference(component any, componentName string) (any, error) {
	if s, ok := component.(*ServiceA); ok {
		return &serviceAAOP{s}, nil
	}
	return component, nil
}

func main() {
	a := &ServiceA{Name: "service A"}
	b := &ServiceB{Name: "service B"}
	c := &ServiceC{Name: "service C"}
	run, err := ioc.Run(app.LogTrace, app.SetComponents(
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

	//fmt.Printf("a: %T\n", a.ServiceA)
	//a.ServiceA.SayName()
	//fmt.Printf("b: %T\n", b.ServiceA)
	//b.ServiceA.SayName()
	//fmt.Printf("c: %T\n", c.ServiceA)
	//c.ServiceA.SayName()
}
