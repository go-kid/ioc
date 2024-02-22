package class_diagram

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type Object interface {
	Name() string
	AddGroup(group FieldGroup) Object
	Type() string
	FieldGroups() []FieldGroup
	fmt.Stringer
}

type class struct {
	name        string
	typ         string
	fieldGroups []FieldGroup
}

func NewClass(name string, typ ...string) Object {
	var t = "class"
	if len(typ) > 0 {
		t = typ[0]
	}
	return &class{
		name:        name,
		typ:         t,
		fieldGroups: nil,
	}
}

func (c *class) AddGroup(group FieldGroup) Object {
	c.fieldGroups = append(c.fieldGroups, group)
	return c
}

func (c *class) Name() string {
	return c.name
}

func (c *class) Type() string {
	return c.typ
}

func (c *class) FieldGroups() []FieldGroup {
	return c.fieldGroups
}

func (c *class) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s %s {\n", c.typ, c.Name()))
	groups := lo.Filter(c.fieldGroups, func(item FieldGroup, _ int) bool {
		return len(item.Fields()) > 0
	})
	var groupHeader = map[string]struct{}{}
	for _, group := range groups {
		if _, ok := groupHeader[group.Group()]; !ok {
			builder.WriteString(fmt.Sprintf("__%s__\n", group.Group()))
			groupHeader[group.Group()] = struct{}{}
		}
		builder.WriteString(group.String())
	}
	builder.WriteString("}\n")
	return builder.String()
}
