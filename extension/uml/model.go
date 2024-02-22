package uml

import (
	"fmt"
	"github.com/samber/lo"
	"path/filepath"
	"strings"
)

type SchemaResponse struct {
	Nodes  []*Node  `json:"nodes"`
	Edges  []*Edge  `json:"edges"`
	Combos []*Combo `json:"combos,omitempty"`
}

type Node struct {
	Id       string       `json:"id"`
	Label    string       `json:"label"`
	NodeType string       `json:"nodeType"`
	ComboId  string       `json:"comboId,omitempty"`
	Attrs    []*Attribute `json:"attrs"`
}

type Attribute struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type Edge struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	SourceKey string `json:"sourceKey,omitempty"`
	TargetKey string `json:"targetKey"`
}

type Combo struct {
	Id       string `json:"id"`
	Label    string `json:"label"`
	ParentId string `json:"parentId,omitempty"`
}

func NewNode(id, typ string) *Node {
	pkg, file := filepath.Split(id)
	pkg = strings.TrimSuffix(pkg, "/")
	return &Node{
		Id:       id,
		Label:    fmt.Sprintf("%s %s", typ, file),
		NodeType: typ,
		ComboId:  pkg,
		Attrs:    []*Attribute{},
	}
}

func (n *Node) AddAttr(fieldKey, fieldType string) *Node {
	n.Attrs = append(n.Attrs, &Attribute{
		Key:  fieldKey,
		Type: fieldType,
	})
	return n
}

func NewEdge(fromClass, fromField, toClass, toField string) *Edge {
	return &Edge{
		Source:    fromClass,
		Target:    toClass,
		SourceKey: fromField,
		TargetKey: toField,
	}
}

func BuildCombos(nodes []*Node) []*Combo {
	var parent = &linkNode{}
	for _, node := range nodes {
		buildLinkNode(parent, node.ComboId, node)
	}
	combos := Group(parent)
	usedCombos := lo.SliceToMap(nodes, func(item *Node) (string, struct{}) {
		return item.ComboId, struct{}{}
	})
	usedCombos2 := lo.SliceToMap(combos, func(item *Combo) (string, struct{}) {
		return item.ParentId, struct{}{}
	})

	combos = lo.Filter(combos, func(item *Combo, index int) bool {
		_, ok := usedCombos[item.Id]
		_, ok2 := usedCombos2[item.Id]
		return ok || ok2
	})
	return combos
}

type linkNode struct {
	Val      string      `json:"val"`
	Children []*linkNode `json:"children,omitempty"`
	Contents []*Node     `json:"contents,omitempty"`
}

func Group(parent *linkNode) []*Combo {
	combos := lo.FlatMap(parent.Children, func(item *linkNode, index int) []*Combo {
		return group(item, item.Val)
	})
	combos = lo.UniqBy(combos, func(item *Combo) string {
		return item.Id
	})
	return combos
}

func buildLinkNode(parent *linkNode, path string, child *Node) {
	sp := strings.SplitN(path, "/", 2)
	var currentPath = sp[0]
	currentNode, ok := lo.Find(parent.Children, func(item *linkNode) bool {
		return item.Val == currentPath
	})
	if !ok {
		currentNode = &linkNode{
			Val: currentPath,
		}
		parent.Children = append(parent.Children, currentNode)
	}
	if len(sp) == 1 {
		currentNode.Contents = append(currentNode.Contents, child)
		return
	}
	buildLinkNode(currentNode, sp[1], child)
}

func group(parent *linkNode, path string) []*Combo {
	var combos []*Combo
	for _, child := range parent.Children {
		if len(child.Children) != 0 {
			childCombos := group(child, strings.Join([]string{path, child.Val}, "/"))
			combos = append(combos, childCombos...)
		}
		if len(child.Contents) != 0 {
			id := child.Contents[0].ComboId
			combos = append(combos, &Combo{
				Id:       id,
				Label:    filepath.Base(id),
				ParentId: path,
			})
			i := strings.Index(id, path)
			combos = append(combos, &Combo{
				Id:       path,
				Label:    path,
				ParentId: id[:i],
			})
		}
	}
	return combos
}
