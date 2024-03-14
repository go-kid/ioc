package main

import (
	"fmt"
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"log"
	"os"
	"os/exec"
	"runtime/pprof"
	"strconv"
)

type A struct {
	LIn   *LI   `wire:"LIn"`
	ILIn  iLI   `wire:"ILIn"`
	LI    *LI   `wire:""`
	ILI   iLI   `wire:""`
	List  []*LI `wire:""`
	IList []iLI `wire:""`
}

type iLI interface {
	Item()
}

type iObj interface {
	Field()
}

type Obj struct {
	name  string
	LI    *LI   `wire:""`
	ILI   iLI   `wire:""`
	List  []*LI `wire:""`
	IList []iLI `wire:""`
}

func (o *Obj) Field() {}

func (o *Obj) Naming() string {
	return o.name
}

type Obj1 struct {
	Obj
}

type Obj2 struct {
	Obj1
}

type Obj3 struct {
	Obj2
}

type LI struct {
	name  string
	IObj  iObj   `wire:""`
	IList []iObj `wire:""`
}

func (l *LI) Naming() string {
	return l.name
}

func (l *LI) Item() {}

func main() {
	comps := []any{
		&A{},
		&LI{name: "LIn"},
		&LI{name: "ILIn"},
		&LI{name: ""},
		&LI{name: "LI1"},
		&LI{name: "LI2"},
		&LI{name: "LI3"},
		&Obj3{},
	}
	for i := 10; i < 100; i++ {
		comps = append(comps, &LI{name: "LI" + strconv.Itoa(i)})
		comps = append(comps, &Obj3{Obj2{Obj1{Obj{name: "Obj" + strconv.Itoa(i)}}}})
	}
	fmt.Println("start")
	err := Analyst(100, "cpu", func() error {
		_, err := ioc.Run(
			app.LogError,
			app.SetRegistry(registry.NewRegistry()),
			app.SetComponents(comps...),
		)
		//defer run.Close()
		return err
	})
	if err != nil {
		panic(err)
	}
}

func Analyst(times int, cat string, f func() error) error {
	switch cat {
	case "cpu":
		var fileName = "./performance_analyst/cpu.pprof"
		cpuFile, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer os.Remove(fileName)
		defer startPProf(fileName)
		defer cpuFile.Close()
		err = pprof.StartCPUProfile(cpuFile)
		if err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	case "mem":
		var fileName = "./performance_analyst/mem.pprof"
		memFile, err := os.Create("./performance_analyst/mem.pprof")
		if err != nil {
			return err
		}
		defer os.Remove(fileName)
		defer startPProf(fileName)
		defer memFile.Close()
		defer pprof.WriteHeapProfile(memFile)
	}
	for i := 0; i < times; i++ {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func startPProf(file string) {
	cmd := exec.Command("go", "tool", "pprof", "-web", file)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("start pprof error: %v, %s", err, string(output))
	}
}
