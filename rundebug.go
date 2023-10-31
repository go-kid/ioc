package ioc

import (
	"fmt"
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/util/class_diagram"
	"github.com/go-kid/ioc/util/reflectx"
	"reflect"
	"sort"
)

func RunDebug(ops ...SettingOption) (*App, error) {
	s := NewApp(append([]SettingOption{SetRegistry(registry.NewRegistry())}, ops...)...)
	err := s.Run()
	if err != nil {
		return s, err
	}
	metas := s.GetComponents()
	sort.Slice(metas, func(i, j int) bool {
		if len(metas[i].DependsBy) != len(metas[j].DependsBy) {
			return len(metas[i].DependsBy) > len(metas[j].DependsBy)
		}
		return metas[i].ID() < metas[j].ID()
	})

	diagram := class_diagram.NewClassDiagram().
		AddSetting(class_diagram.NamespaceSeparator("/")).
		AddSetting(class_diagram.GroupInheritance(2))
	for _, m := range metas {
		dependencyGroup := class_diagram.NewFieldGroup("Dependency")
		configGroup := class_diagram.NewFieldGroup("Config")
		diagram.AddClass(class_diagram.NewClass(m.Name).
			AddGroup(configGroup).
			AddGroup(dependencyGroup))
		for _, node := range m.AllDependencies() {
			dependencyGroup.AddField(node.Field.Name, node.Type.String(), string(node.Field.Tag))
			for _, ij := range node.Injects {
				diagram.AddLine(class_diagram.NewLine(ij.Name, "", m.Name, node.Field.Name, "up", "*", ""))
			}
		}
		for _, p := range m.Properties {
			configGroup.AddField(p.Field.Name, p.Type.String(), string(p.Field.Tag))
			configName := reflectx.TypeId(p.Type)
			if p.Type.Kind() == reflect.Struct || p.Type.Kind() == reflect.Pointer {
				fg := class_diagram.NewFieldGroup("Field")
				pfg := class_diagram.NewFieldGroup("Prefix")
				diagram.AddClass(class_diagram.NewClass(configName).AddGroup(pfg).AddGroup(fg))
				pfg.AddField(p.Tag, p.TagVal)
				_ = reflectx.ForEachFieldV2(p.Type, reflectx.New(p.Type), true, func(field reflect.StructField, value reflect.Value) error {
					fg.AddField(field.Name, field.Type.String())
					return nil
				})
				diagram.AddLine(class_diagram.NewLine(configName, "", m.Name, p.Field.Name, "left", "o", ""))
			}
		}
	}
	fmt.Println(diagram.String())
	return s, nil
}
