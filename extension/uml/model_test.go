package uml

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNode(t *testing.T) {
	nodeP := NewNode("person", "Person")
	nodeP.AddAttr("Id", "int")
	nodeP.AddAttr("Name", "string")
	nodeP.AddAttr("Gender", "bool")
	nodeP.AddAttr("DeptId", "int")
	nodeP.AddRel("dept", "DeptId", "Id")

	nodeD := NewNode("dept", "Department")
	nodeD.AddAttr("Id", "int")
	nodeD.AddAttr("ParentId", "int", &Relation{
		Key:    "Id",
		NodeId: "dept",
	})

	var nodes = []*Node{nodeP, nodeD}
	marshal, err := json.Marshal(nodes)
	assert.NoError(t, err)
	fmt.Println(string(marshal))
}
