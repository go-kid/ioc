# go-kid/ioc
go-kid/ioc是一款基于go tag和interface的运行时依赖注入框架

## 快速使用
### 安装

```shell
go get github.com/go-kid/ioc
```


### 构建一个基础依赖注入场景

main.go
```go
package main

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
)

//定义组件ComponentA作为被依赖项
type ComponentA struct {
	Name string
}

func (a *ComponentA) GetName() string {
	return a.Name
}

//定义组件App作为注入主体
type App struct {
	ComponentA *ComponentA `wire:""`
}

func main() {
	//创建组件App
	a := new(App)
	//注册组件App
    ioc.Register(a)
	//创建组件ComponentA
	compA:=&ComponentA{Name: "Comp-A"}
    //注册组件ComponentA
    ioc.Register(compA)
	//运行框架
	_, err := ioc.Run()
	if err != nil {
		panic(err)
	}
	//此时a.ComponentA被填充为compA
	//所以此处应打印出"Comp-A"
	fmt.Println(a.ComponentA.GetName())
}
```

此处展示了go-kid/ioc依赖注入的基本能力,
在组件App中,我们使用`wire`标签标注了ComponentA为一个依赖项,
注意此时ComponentA的类型为指针,并且字段名首字母大写,这样才能注入成功.

## 依赖注入

此章将详细介绍go-kid/ioc支持的各种注入场景

### 按接口类型注入

前面例子中展示了指针类型的基本注入能力, 但在开发中,
我们倾向于使用接口类型进行抽象和解耦,
下面我们将用接口对上面例子进行重构

main.go
```go
package main

//定义接口IComponent, 此时ComponentA实现了该接口
type IComponent interface {
    GetName() string
}

type App struct {
    ComponentA IComponent `wire:""`
}
```
此处App依赖ComponentA的形式从指针变为了接口,更符合设计规范.
至此你可以看到go-kid/ioc提供了基于指针类型和基于接口的注入能力.

### 为组件命名和按名称注入

容器默认会为每个组件生成组件名,格式为{package}/{组件名}.
当需要自定义组件名时, go-kid/ioc在
github.com/go-kid/ioc/definition
中提供了内置接口`NamingComponent`

def.go
```go
package main

type NamingComponent interface {
	Naming() string
}
```
组件只需实现此接口, 在Naming()方法中返回的字符串将被作为组件别名.
下面我们来让ComponentA实现Naming接口:

main.go
```go
package main

type ComponentA struct {
	Name          string
	componentName string
}

func (a *ComponentA) GetName() string {
	return a.Name
}

func (a *ComponentA) Naming() string {
	return a.componentName
}

//...///

type App struct {
	ComponentA IComponent `wire:"comp-A"`
}

func main() {
	a := new(App)
	ioc.Register(a)
	compA := &ComponentA{
		Name:          "Comp-A",
		componentName: "comp-A",
	}
	ioc.Register(compA)
	_, err := ioc.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(a.ComponentA.GetName())
}
```
上面代码中我们自定义了ComponentA的别名为comp-A,
并修改App中依赖项ComponentA的wire标签内容为comp-A,
即可实现按名称注入.
> 可以尝试修改wire标签内容为comp-B, 
> 运行代码,将看到报错
> 
> [ERROR] [Application] application run failed: inject 'main/App.Field(ComponentA).Type(Component).Tag(wire:'comp-B').Required()' not found available components
> 
> 说明容器没有找到名称为comp-B的组件,在严格模式下将立即报错.

main.go
```go
package main

//...///

//定义ComponentB,此处为简化例子代码,我们让其等于ComponentA
type ComponentB = ComponentA

//...//

type App struct {
	ComponentA IComponent   `wire:"comp-A"`
	ComponentB IComponent   `wire:"comp-B"`
}

func main() {
	a := new(App)
	ioc.Register(a)
	_, err := ioc.Run(app.SetComponents(
		&ComponentA{
			Name:          "Comp-A",
			componentName: "comp-A",
		},
		&ComponentB{
			Name:          "Comp-B",
			componentName: "comp-B",
		},
	))
	if err != nil {
		panic(err)
	}
	fmt.Println("ComponentA", a.ComponentA.GetName())
	fmt.Println("ComponentB", a.ComponentB.GetName())
}
```

这里我们定义了ComponentB,
为了简化, 我们定义其等于ComponentA, 所以ComponentB也实现了ComponentA的所有接口.
在main方法中, 可以看到我们使用app.SetComponents(...)方法替换了ioc.Register(...),其作用是相同的.

### 同一接口多个实例情况的注入

修改上面例子:

main.go
```go
package main

//...//

type App struct {
	Component  IComponent   `wire:""`
}

func main() {
	a := new(App)
	ioc.Register(a)
	_, err := ioc.Run(app.SetComponents(
		&ComponentA{
			Name:          "Comp-A",
			componentName: "comp-A",
		},
		&ComponentB{
			Name:          "Comp-B",
			componentName: "comp-B",
		},
	))
	if err != nil {
		panic(err)
	}
	fmt.Println("Component", a.Component.GetName())
}
```

我们向容器中注册了两个实现了IComponent接口的实例,
在App中, 前两项将按指定的名称comp-A和comp-B查找依赖项,
而第三项Component未指定名称, 容器只能随机选择一个实例注入,
多次运行程序, 可以看到ComponentA和ComponentB的GetName()方法返回值始终固定,
而Component的GetName()将每次随机.
大部分时候随机注入将不会是我们希望的, 该问题有下面几种解决方法.

#### 限定注入

当一个接口有多个实例, 且需要精确注入时, 将需要内置的WireQualifier接口进行限定.
修改ComponentA组件实现WireQualifier接口, 该接口返回一个string作为限定符:

main.go
```go
package main

//...///

type ComponentA struct {
	Name          string
	componentName string
	qualifier     string
}

func (a *ComponentA) Qualifier() string {
	return a.qualifier
}
```

修改组件App, 添加qualifier参数,使其等于需要的组件的限定符.

main.go
```go
package main

//...//

type App struct {
	Component IComponent `wire:",qualifier=comp-A"`
}

func main() {
	a := new(App)
	ioc.Register(a)
	_, err := ioc.Run(app.SetComponents(
		&ComponentA{
			Name:          "Comp-A",
			componentName: "comp-A",
			qualifier:     "comp-A",
		},
		&ComponentB{
			Name:          "Comp-B",
			componentName: "comp-B",
			qualifier:     "comp-B",
		},
	))
	if err != nil {
		panic(err)
	}
	fmt.Println("Component", a.Component.GetName())
}
```

例子中我们将App.Component限定为`qualifier=comp-A`, 
此时a.Component将永远只注入Comp-A组件.

> wire标签的值中的 ',' 符号声明该属性启用了解析参数, 后面会介绍到各种参数

#### 优先注入

除了使用限定符接口来精确注入, 还可以使用内置的WirePrimary接口来指定优先注入的接口.
该接口没有返回值, 当组件实现WirePrimary接口后, 将被容器优先选择.

修改组件ComponentA实现Primary接口:

main.go
```go
package main

//...//

type ComponentA struct {
	Name          string
	componentName string
}

func (a *ComponentA) Primary() {}
```

修改App组件

main.go
```go
package main

type App struct {
	Component IComponent `wire:""`
}
```

由于先前代码 `type ComponentB = ComponentA`, 使得ComponentB此时同样获得了Primary接口,
所以此处应重写ComponentB, 之后运行程序, a.Component将永远为Comp-A实例.

> - WirePrimary接口的优先级低于WireQualifier接口, 
> - 实现WirePrimary接口的同类实例不应出现多个, 否则同样会出现随机注入.

### 按接口注入多个实现

需要获取同一个接口的多个不同实例时的注入方法

main.go
```go
package main

//...//

type App struct {
	ComponentA IComponent   `wire:"comp-A"`
	ComponentB IComponent   `wire:"comp-B"`
	Components []IComponent `wire:""`
}

func main() {
	a := new(App)
	ioc.Register(a)
	_, err := ioc.Run(app.SetComponents(
		&ComponentA{
			Name:          "Comp-A",
			componentName: "comp-A",
		},
		&ComponentB{
			Name:          "Comp-B",
			componentName: "comp-B",
		},
	))
	if err != nil {
		panic(err)
	}
	fmt.Println("ComponentA", a.ComponentA.GetName())
	fmt.Println("ComponentB", a.ComponentB.GetName())
	fmt.Println("Components")
	for i, component := range a.Components {
		fmt.Println(i, component.GetName())
	}
}
```
修改上面代码, 将App中的Component修改为[]Component类型, 且`wire`不指定名称,
运行代码, 将看到输出

```text
ComponentA Comp-A
ComponentB Comp-B
Components
0 Comp-B
1 Comp-A
```

从3-5行可以看出, 接口接片类型的字段Components里注入了两个组件.
这里体现了go-kid/ioc的接口切片类型注入的能力. 该能力在编写二次开发框架时将很有用.

## 配置填充

go-kid/ioc框架除了可以代理容器的依赖关系, 还可以代理配置的读取.

