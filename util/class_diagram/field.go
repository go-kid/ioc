package class_diagram

import (
	"fmt"
	"github.com/go-kid/ioc/util/fas"
	"strings"
)

type field struct {
	Name     string
	Holding1 int
	Type     string
	Holding2 int
	Arg      string
}

func (f *field) String() string {
	return fmt.Sprintf("  +%s%s : %s%s %s\n", f.Name, strings.Repeat(" ", f.Holding1-len(f.Name)), f.Type, strings.Repeat(" ", f.Holding2-len(f.Type)), f.Arg)
}

type fieldGroup struct {
	Group      string
	Fields     []*field
	MaxNameLen int
	MaxTypeLen int
}

func NewFieldGroup(group string) *fieldGroup {
	return &fieldGroup{
		Group: group,
	}
}

func (f *fieldGroup) AddField(fieldName, fieldType string, arg ...string) *fieldGroup {
	f.Fields = append(f.Fields, &field{
		Name:     fieldName,
		Holding1: 0,
		Type:     fieldType,
		Holding2: 0,
		Arg:      strings.Join(arg, " "),
	})
	return f
}

func (f *fieldGroup) String() string {
	builder := strings.Builder{}
	for _, field := range f.Fields {
		f.MaxNameLen = fas.Max(len(field.Name), f.MaxNameLen)
		f.MaxTypeLen = fas.Max(len(field.Type), f.MaxTypeLen)
	}
	for _, field := range f.Fields {
		field.Holding1 = f.MaxNameLen
		field.Holding2 = f.MaxTypeLen
		builder.WriteString(field.String())
	}
	return builder.String()
}
