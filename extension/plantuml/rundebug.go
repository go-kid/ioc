package plantuml

import (
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/class_diagram"
	"github.com/go-kid/ioc/util/fas"
	"github.com/go-kid/ioc/util/reflectx"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
)

type DebugSetting struct {
	DisablePackageView      bool
	DisableConfig           bool
	DisableConfigDetail     bool
	DisableDependency       bool
	DisableDependencyDetail bool
	DisableUselessClass     bool
	PreciseArrow            bool
	Writer                  io.Writer
}

func Run(setting DebugSetting, ops ...SettingOption) (*App, error) {
	s := NewApp(append([]SettingOption{
		SetRegistry(registry.NewRegistry()),
		SetFactory(func() factory.Factory {
			var df = &factory.DefaultFactory{}
			df.SetIfNilPostInitFunc(func(m *meta.Meta) error {
				return nil
			})
			return df
		}()),
		DisableApplicationRunner()}, ops...)...)
	diagram, err := BuildDiagram(s, setting)
	if err != nil {
		return nil, err
	}

	if setting.Writer == nil {
		setting.Writer = os.Stdout
	}
	_, err = setting.Writer.Write([]byte(diagram.String()))
	return s, err
}

func BuildDiagram(s *App, setting DebugSetting) (class_diagram.ClassDiagram, error) {
	err := s.Run()
	if err != nil {
		return nil, err
	}
	metas := s.GetComponents()
	sort.Slice(metas, func(i, j int) bool {
		if len(metas[i].DependsBy) != len(metas[j].DependsBy) {
			return len(metas[i].DependsBy) > len(metas[j].DependsBy)
		}
		return metas[i].ID() < metas[j].ID()
	})

	diagram := class_diagram.NewClassDiagram().
		AddSetting(class_diagram.GroupInheritance(2)).
		AddSetting(class_diagram.NamespaceSeparator("/"))
	if setting.DisableUselessClass {
		diagram.AddSetting("remove @unlinked")
	}

	for _, m := range metas {
		metaName := fas.TernaryOp(setting.DisablePackageView, m.Type.String(), m.Name)
		class := class_diagram.NewClass(metaName)
		if !setting.DisableConfig {
			configGroup := class_diagram.NewFieldGroup("Config")
			class.AddGroup(configGroup)
			for _, p := range m.Properties {
				if !setting.DisableConfigDetail {
					configGroup.AddField(p.Field.Name, p.Type.String(), string(p.Field.Tag))
				}

				configName := fas.TernaryOp(setting.DisablePackageView, p.Type.String(), reflectx.TypeId(p.Type))
				if p.Type.Kind() == reflect.Struct || p.Type.Kind() == reflect.Pointer {
					fg := class_diagram.NewFieldGroup("Field")
					pfg := class_diagram.NewFieldGroup("Prefix")
					diagram.AddClass(class_diagram.NewClass(configName, "struct").AddGroup(pfg).AddGroup(fg))
					if !setting.DisableConfigDetail {
						pfg.AddField(p.Tag, p.TagVal)
						_ = reflectx.ForEachFieldV2(p.Type, reflectx.New(p.Type), true, func(field reflect.StructField, value reflect.Value) error {
							fg.AddField(field.Name, field.Type.String())
							return nil
						})
					}
					var toField string
					if setting.PreciseArrow {
						toField = p.Field.Name
					}
					diagram.AddLine(class_diagram.NewLine(metaName, toField, configName, "", "--o", ""))
				}
			}
		}
		if !setting.DisableDependency {
			for _, node := range m.AllDependencies() {
				var sourceClass = class
				dependencyGroup := class_diagram.NewFieldGroup("Dependency")
				//find if is interface type
				var interfaceType = _interfaceType(node.Type)
				var interfaceName string
				if interfaceType != nil {
					interfaceName = fas.TernaryOp(setting.DisablePackageView, interfaceType.String(), reflectx.TypeId(interfaceType))
					interfaceName = strings.ReplaceAll(interfaceName, "interface {}", "any")
					methodFg := class_diagram.NewFieldGroup("Method")
					for i := 0; i < interfaceType.NumMethod(); i++ {
						method := interfaceType.Method(i)
						methodFg.AddField(method.Name, method.Type.String())
					}
					diagram.AddClass(class_diagram.NewClass(interfaceName, "interface").AddGroup(methodFg))
				}

				var sourceName = metaName
				if source := node.Source; source != nil && source.IsAnonymous {
					embedGroup := class_diagram.NewFieldGroup("Embed")
					embedGroup.AddField(source.Type.Name(), source.Type.String())
					sourceClass.AddGroup(embedGroup)
					sourceName = fas.TernaryOp(setting.DisablePackageView, source.Type.String(), reflectx.TypeId(source.Type))
					sourceClass = class_diagram.NewClass(sourceName, "abstract")
					diagram.AddClass(sourceClass)
					diagram.AddLine(class_diagram.NewLine(metaName, source.Type.Name(), sourceName, "", "-->", ""))
				}
				if !setting.DisableDependencyDetail {
					dependencyGroup.AddField(node.Field.Name, node.Type.String(), string(node.Field.Tag))
				}

				for _, ij := range node.Injects {
					injectName := fas.TernaryOp(setting.DisablePackageView, ij.Type.String(), ij.Name)
					var toField string
					if setting.PreciseArrow {
						toField = node.Field.Name
					}
					if interfaceType != nil {
						for _, n := range ij.AllDependencies() {
							if n.Source == nil || !n.Source.IsAnonymous {
								continue
							}
							if reflect.New(n.Source.Type).Type().Implements(interfaceType) {
								diagram.AddLine(class_diagram.NewLine(fas.TernaryOp(setting.DisablePackageView, n.Source.Type.String(), reflectx.TypeId(n.Source.Type)), "", interfaceName, "", "-->", ""))
							} else {
								diagram.AddLine(class_diagram.NewLine(injectName, "", interfaceName, "", "-->", ""))
							}
						}
						diagram.AddLine(class_diagram.NewLine(sourceName, toField, interfaceName, "", "--*", ""))
					} else {
						diagram.AddLine(class_diagram.NewLine(sourceName, toField, injectName, "", "--*", ""))
					}
				}
				sourceClass.AddGroup(dependencyGroup)
			}
		}
		diagram.AddClass(class)
	}
	return diagram, nil
}

func _interfaceType(p reflect.Type) reflect.Type {
	if p.Kind() == reflect.Interface {
		return p
	} else if p.Kind() == reflect.Slice && p.Elem().Kind() == reflect.Interface {
		return p.Elem()
	}
	return nil
}
