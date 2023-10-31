package class_diagram

import (
	"fmt"
	"testing"
)

func TestDiagram(t *testing.T) {
	d := NewClassDiagram()
	d.AddSetting("namespaceSeparator /")
	d.AddClass(NewClass("org/Person", NewFieldGroup().AddField("var", "name", "string").AddField("var", "age", "integer")))
	d.AddClass(NewClass("org/IDCard", NewFieldGroup().AddField("var", "id", "integer").AddField("var", "name", "string")))
	d.AddLine(NewLine("IDCard", "name", "Person", "name", "down", "o", "associate"))
	fmt.Println(d.String())
}
