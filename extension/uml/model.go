package uml

type Node struct {
	Id    string       `json:"id"`
	Label string       `json:"label"`
	Attrs []*Attribute `json:"attrs"`
}

type Attribute struct {
	Key      string      `json:"key"`
	Type     string      `json:"type"`
	Relation []*Relation `json:"relation,omitempty"`
}

type Relation struct {
	Key    string `json:"key"`
	NodeId string `json:"nodeId"`
}

func NewNode(id, label string) *Node {
	return &Node{
		Id:    id,
		Label: label,
		Attrs: make([]*Attribute, 0),
	}
}

func (n *Node) AddAttr(fieldKey, fieldType string, rel ...*Relation) *Node {
	n.Attrs = append(n.Attrs, &Attribute{
		Key:      fieldKey,
		Type:     fieldType,
		Relation: rel,
	})
	return n
}

func (n *Node) AddRel(nodeId, fieldKeySrc, fieldKeyDst string) {
	for _, attr := range n.Attrs {
		if attr.Key == fieldKeySrc {
			attr.Relation = append(attr.Relation, &Relation{
				Key:    fieldKeyDst,
				NodeId: nodeId,
			})
		}
	}
}
