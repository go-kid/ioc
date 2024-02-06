package uml

import (
	_ "embed"
	"encoding/json"
	"fmt"
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/extension/plantuml"
	"github.com/go-kid/ioc/factory"
	"github.com/go-kid/ioc/registry"
	"github.com/go-kid/ioc/scanner/meta"
	"github.com/go-kid/ioc/util/class_diagram"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
)

type DebugSetting struct {
	DisablePackageView      bool
	DisableConfig           bool
	DisableConfigDetail     bool
	DisableDependency       bool
	DisableDependencyDetail bool
	DisableUselessClass     bool
	RawData                 io.Writer
	StartServer             bool
}

//go:embed index.html
var html []byte

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
	diagram, err := plantuml.BuildDiagram(s, plantuml.DebugSetting{
		DisablePackageView:      setting.DisablePackageView,
		DisableConfig:           setting.DisableConfig,
		DisableConfigDetail:     setting.DisableConfigDetail,
		DisableDependency:       setting.DisableDependency,
		DisableDependencyDetail: setting.DisableDependencyDetail,
		DisableUselessClass:     setting.DisableUselessClass,
		PreciseArrow:            true,
		Writer:                  setting.RawData,
	})
	if err != nil {
		return nil, err
	}

	if setting.RawData == nil {
		setting.RawData = os.Stdout
	}

	nodes := convertDiagram2AntV(diagram)
	bytes, err := json.Marshal(nodes)
	if err != nil {
		return nil, err
	}

	_, err = setting.RawData.Write(bytes)
	if err != nil {
		return nil, err
	}
	if setting.StartServer {
		StartServer(bytes)
	}
	return s, err
}

func StartServer(data []byte) error {
	http.HandleFunc("/schema", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(data)
	})
	http.HandleFunc("/index", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write(html)
	})
	return http.ListenAndServe(":8888", nil)
}

func convertDiagram2AntV(diagram class_diagram.ClassDiagram) (nodes []*Node) {
	for _, object := range diagram.Classes() {
		node := NewNode(object.Name(), fmt.Sprintf("%s %s", object.Type(), object.Name()))
		nodes = append(nodes, node)
		for _, fg := range object.FieldGroups() {
			for _, field := range fg.Fields() {
				node.AddAttr(field.Name(), field.Type()+" "+field.Arg())
			}
		}
		fmt.Println(object.String())
	}
	nodeMap := lo.SliceToMap(nodes, func(item *Node) (string, *Node) {
		return item.Id, item
	})
	for _, line := range diagram.Lines() {
		fromClass, fromField := line.From()
		toClass, toField := line.To()
		fmt.Println(line.String())
		nodeMap[fromClass].AddRel(toClass, fromField, toField)
	}
	return
}
