package class_diagram

import (
	"fmt"
	"github.com/go-kid/ioc/util/fas"
)

type Line interface {
	From() (class, field string)
	To() (class, field string)
	Tag() string
	fmt.Stringer
}

type line struct {
	fromClass string
	fromField string
	toClass   string
	toField   string
	arrowType string
	tag       string
}

func (l *line) From() (class, field string) {
	return l.fromClass, l.fromField
}

func (l *line) To() (class, field string) {
	return l.toClass, l.toField
}

func (l *line) Tag() string {
	return l.tag
}

func NewLine(from, fromField, to, toField, arrowType, tag string) Line {
	return &line{
		fromClass: from,
		fromField: fromField,
		toClass:   to,
		toField:   toField,
		arrowType: arrowType,
		tag:       tag,
	}
}

func (l *line) String() string {
	return fmt.Sprintf("\"%s%s\" %s \"%s%s%s\"\n",
		l.fromClass,
		fas.TernaryOp(l.fromField == "", "", "::"+l.fromField),
		fas.TernaryOp(l.arrowType == "", "--", l.arrowType),
		l.toClass, fas.TernaryOp(l.toField == "", "", "::"+l.toField),
		l.tag,
	)
}
