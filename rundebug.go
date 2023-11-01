package ioc

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

func RunDebug(setting DebugSetting, ops ...SettingOption) (*App, error) {
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
					diagram.AddLine(class_diagram.NewLine(configName, "", metaName, toField, "o--", ""))
				}
			}
		}
		if !setting.DisableDependency {
			dependencyGroup := class_diagram.NewFieldGroup("Dependency")
			class.AddGroup(dependencyGroup)
			for _, node := range m.AllDependencies() {
				if !setting.DisableDependencyDetail {
					dependencyGroup.AddField(node.Field.Name, node.Type.String(), string(node.Field.Tag))
				}
				var interfaceName string
				if node.Type.Kind() == reflect.Interface {
					interfaceName = fas.TernaryOp(setting.DisablePackageView, node.Type.String(), reflectx.TypeId(node.Type))
				} else if node.Type.Kind() == reflect.Slice && node.Type.Elem().Kind() == reflect.Interface {
					interfaceName = fas.TernaryOp(setting.DisablePackageView, node.Type.Elem().String(), reflectx.TypeId(node.Type.Elem()))
				}
				if interfaceName != "" {
					interfaceName = strings.ReplaceAll(interfaceName, "interface {}", "any")
					diagram.AddClass(class_diagram.NewClass(interfaceName, "interface"))
				}
				for _, ij := range node.Injects {
					injectName := fas.TernaryOp(setting.DisablePackageView, ij.Type.String(), ij.Name)
					var toField string
					if setting.PreciseArrow {
						toField = node.Field.Name
					}
					if interfaceName != "" {
						diagram.AddLine(class_diagram.NewLine(interfaceName, "", injectName, "", "<|--", ""))
						diagram.AddLine(class_diagram.NewLine(interfaceName, "", metaName, toField, "*--", ""))
					} else {
						diagram.AddLine(class_diagram.NewLine(injectName, "", metaName, toField, "*--", ""))
					}
				}
			}
		}
		diagram.AddClass(class)
	}

	if setting.Writer == nil {
		setting.Writer = os.Stdout
	}
	_, err = setting.Writer.Write([]byte(diagram.String()))
	return s, err
}
