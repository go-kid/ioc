package main

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
)

type T struct {
	Env    string  `yaml:"env"`
	Server *Server `yaml:"server"`
	Ratio  float64 `yaml:"ratio"`
}

type Server struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	SPort string `yaml:"SPort"`
}

type App struct {
	Config *T `prop:""`
}

type Service interface {
	SayName()
}

type ServiceA struct {
	Name     string
	ServiceB *ServiceB `wire:""`
	ServiceC *ServiceC `wire:""`
	//ServiceA Service   `wire:""`
}

func (s *ServiceA) Qualifier() string {
	return ""
}

func (s *ServiceA) SayName() {
	fmt.Println(s.Name)
}

type ServiceB struct {
	Name string
	//ServiceA Service `wire:""`
}

type ServiceC struct {
	Name string
	//ServiceA Service `wire:""`
}

type serviceAAOP struct {
	s *ServiceA
}

func (s *serviceAAOP) SayName() {
	fmt.Println("before say")
	s.s.SayName()
	fmt.Println("after say")
}

type postProcessor struct {
}

func (p *postProcessor) PostProcessBeforeInitialization(component any, componentName string) (any, error) {
	return component, nil
}

func (p *postProcessor) PostProcessAfterInitialization(component any, componentName string) (any, error) {
	if s, ok := component.(*ServiceA); ok {
		return &serviceAAOP{s: s}, nil
	}
	return component, nil
}

func main() {
	a := &ServiceA{Name: "service A"}
	b := &ServiceB{Name: "service B"}
	c := &ServiceC{Name: "service C"}
	app.Settings(app.LogTrace)
	ioc.Register(&postProcessor{})
	run, err := ioc.Run(app.SetComponents(a, b, c))
	if err != nil {
		panic(err)
	}
	defer run.Close()
	fmt.Println(a.ServiceB.Name)
	fmt.Println(a.ServiceC.Name)
	//fmt.Printf("a: %T\n", a.ServiceA)
	//a.ServiceA.SayName()
	//fmt.Printf("b: %T\n", b.ServiceA)
	//b.ServiceA.SayName()
	//fmt.Printf("c: %T\n", c.ServiceA)
	//c.ServiceA.SayName()
}
