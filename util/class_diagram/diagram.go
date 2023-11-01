package class_diagram

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type Object interface {
	Name() string
	fmt.Stringer
}

type ClassDiagram interface {
	AddSetting(s string) ClassDiagram
	AddClass(c Object) ClassDiagram
	AddLine(l *line) ClassDiagram
	fmt.Stringer
}

type diagram struct {
	Settings []string
	Classes  []Object
	Lines    []*line
}

func NewClassDiagram() ClassDiagram {
	return &diagram{}
}

func (d *diagram) AddSetting(s string) ClassDiagram {
	d.Settings = append(d.Settings, s)
	return d
}

func (d *diagram) AddClass(c Object) ClassDiagram {
	contain := lo.ContainsBy(d.Classes, func(item Object) bool {
		return c.Name() == item.Name()
	})
	if !contain {
		d.Classes = append(d.Classes, c)
	}
	return d
}

func (d *diagram) AddLine(l *line) ClassDiagram {
	contain := lo.ContainsBy(d.Lines, func(item *line) bool {
		return l.String() == item.String()
	})
	if !contain {
		d.Lines = append(d.Lines, l)
	}
	return d
}

func (d *diagram) String() string {
	builder := strings.Builder{}
	builder.WriteString("\n@startuml\n")
	for _, setting := range d.Settings {
		builder.WriteString(fmt.Sprintf("%s\n", setting))
	}
	for _, c := range d.Classes {
		builder.WriteString(c.String())
	}
	for _, l := range d.Lines {
		builder.WriteString(l.String())
	}
	builder.WriteString("@enduml\n")
	return builder.String()
}
