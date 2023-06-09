package ioc

import (
	"fmt"
	"github.com/awalterschulze/gographviz"
	. "github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/meta"
	"github.com/go-kid/ioc/registry"
	"sort"
)

func RunDebug(ops ...SettingOption) error {
	s := NewApp(append([]SettingOption{SetRegistry(registry.NewRegistry())}, ops...)...)
	err := s.Run()
	if err != nil {
		return err
	}
	metas := s.GetComponents()
	sort.Slice(metas, func(i, j int) bool {
		if len(metas[i].DependsBy) != len(metas[j].DependsBy) {
			return len(metas[i].DependsBy) > len(metas[j].DependsBy)
		}
		return metas[i].ID() < metas[j].ID()
	})

	graphAst, _ := gographviz.ParseString("digraph G {}")
	graph := gographviz.NewGraph()
	if err := gographviz.Analyse(graphAst, graph); err != nil {
		return err
	}
	for _, m := range metas {
		err := graph.AddNode("g", meta.StringEscape(m.Name), m.DotNodeAttr())
		if err != nil {
			return err
		}
		for _, p := range m.DependsBy {
			err := graph.AddEdge(meta.StringEscape(p.Name), meta.StringEscape(m.Name), true, nil)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println(graph.String())
	return nil
}
