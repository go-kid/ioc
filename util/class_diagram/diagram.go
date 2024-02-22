package class_diagram

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type ClassDiagram interface {
	AddSetting(s string) ClassDiagram
	AddClass(c Object) ClassDiagram
	AddLine(l Line) ClassDiagram
	Lines() []Line
	Classes() []Object
	fmt.Stringer
}

type diagram struct {
	settings []string
	classes  []Object
	lines    []Line
}

func NewClassDiagram() ClassDiagram {
	return &diagram{}
}

func (d *diagram) AddSetting(s string) ClassDiagram {
	d.settings = append(d.settings, s)
	return d
}

func (d *diagram) AddClass(c Object) ClassDiagram {
	contain := lo.ContainsBy(d.classes, func(item Object) bool {
		return c.Name() == item.Name()
	})
	if !contain {
		d.classes = append(d.classes, c)
	}
	return d
}

func (d *diagram) AddLine(l Line) ClassDiagram {
	contain := lo.ContainsBy(d.lines, func(item Line) bool {
		return l.String() == item.String()
	})
	if !contain {
		d.lines = append(d.lines, l)
	}
	return d
}

func (d *diagram) Lines() []Line {
	return d.lines
}

func (d *diagram) Classes() []Object {
	return d.classes
}

func (d *diagram) String() string {
	builder := strings.Builder{}
	builder.WriteString("\n@startuml\n")
	for _, setting := range d.settings {
		builder.WriteString(fmt.Sprintf("%s\n", setting))
	}
	for _, c := range d.classes {
		builder.WriteString(c.String())
	}
	for _, l := range d.lines {
		builder.WriteString(l.String())
	}
	builder.WriteString("@enduml\n")
	return builder.String()
}
