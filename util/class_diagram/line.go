package class_diagram

import (
	"fmt"
	"github.com/go-kid/ioc/util/fas"
)

type line struct {
	FromClass string
	FromField string
	ToClass   string
	ToField   string
	Direction string
	ArrowType string
	Tag       string
}

func NewLine(from, fromField, to, toField, direction, arrowType, tag string) *line {
	return &line{
		FromClass: from,
		FromField: fromField,
		ToClass:   to,
		ToField:   toField,
		Direction: direction,
		ArrowType: arrowType,
		Tag:       tag,
	}
}

func (l *line) String() string {
	return fmt.Sprintf("\"%s%s\" %s-%s \"%s%s%s\"\n",
		l.FromClass, fas.TernaryOp(l.FromField == "", "", "::"+l.FromField),
		fas.TernaryOp(l.Direction == "", "-", "-"+l.Direction),
		fas.TernaryOp(l.ArrowType == "", ">", l.ArrowType),
		l.ToClass, fas.TernaryOp(l.ToField == "", "", "::"+l.ToField),
		l.Tag,
	)
}
