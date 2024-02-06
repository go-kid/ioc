package class_diagram

import (
	"fmt"
	"github.com/go-kid/ioc/util/fas"
	"strings"
)

type Field interface {
	SetHolding1(h1, h2 int)
	Name() string
	Type() string
	Arg() string
	fmt.Stringer
}

type field struct {
	name     string
	holding1 int
	typ      string
	holding2 int
	arg      string
}

func (f *field) SetHolding1(h1, h2 int) {
	f.holding1 = h1
	f.holding2 = h2
}

func (f *field) Name() string {
	return f.name
}

func (f *field) Type() string {
	return f.typ
}

func (f *field) Arg() string {
	return f.arg
}

func (f *field) String() string {
	return fmt.Sprintf("  +%s%s : %s%s %s\n", f.name, strings.Repeat(" ", f.holding1-len(f.name)), f.typ, strings.Repeat(" ", f.holding2-len(f.typ)), f.arg)
}

type FieldGroup interface {
	AddField(fieldName, fieldType string, arg ...string) FieldGroup
	Group() string
	Fields() []Field
	fmt.Stringer
}

type fieldGroup struct {
	group      string
	fields     []Field
	MaxNameLen int
	MaxTypeLen int
}

func NewFieldGroup(group string) FieldGroup {
	return &fieldGroup{
		group: group,
	}
}

func (f *fieldGroup) Group() string {
	return f.group
}

func (f *fieldGroup) Fields() []Field {
	return f.fields
}

func (f *fieldGroup) AddField(fieldName, fieldType string, arg ...string) FieldGroup {
	f.fields = append(f.fields, &field{
		name:     fieldName,
		holding1: 0,
		typ:      fieldType,
		holding2: 0,
		arg:      strings.Join(arg, " "),
	})
	return f
}

func (f *fieldGroup) String() string {
	builder := strings.Builder{}
	for _, field := range f.fields {
		f.MaxNameLen = fas.Max(len(field.Name()), f.MaxNameLen)
		f.MaxTypeLen = fas.Max(len(field.Type()), f.MaxTypeLen)
	}
	for _, field := range f.fields {
		field.SetHolding1(f.MaxNameLen, f.MaxTypeLen)
		builder.WriteString(field.String())
	}
	return builder.String()
}
