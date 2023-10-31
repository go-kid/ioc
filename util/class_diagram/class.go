package class_diagram

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
)

type class struct {
	name   string
	Type   string
	Groups []*fieldGroup
}

func NewClass(name string, typ ...string) *class {
	var t = "class"
	if len(typ) > 0 {
		t = typ[0]
	}
	return &class{
		name:   name,
		Type:   t,
		Groups: nil,
	}
}

func (c *class) AddGroup(group *fieldGroup) *class {
	c.Groups = append(c.Groups, group)
	return c
}

func (c *class) Name() string {
	return c.name
}

func (c *class) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s %s {\n", c.Type, c.Name()))
	groups := lo.Filter(c.Groups, func(item *fieldGroup, _ int) bool {
		return len(item.Fields) > 0
	})
	var groupHeader = map[string]struct{}{}
	for _, group := range groups {
		if _, ok := groupHeader[group.Group]; !ok {
			builder.WriteString(fmt.Sprintf("__%s__\n", group.Group))
			groupHeader[group.Group] = struct{}{}
		}
		builder.WriteString(group.String())
	}
	builder.WriteString("}\n")
	return builder.String()
}
