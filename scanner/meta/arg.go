package meta

import (
	"strings"
)

type (
	ArgType string
	TagArg  map[ArgType][]string
)

const (
	ArgRequired  ArgType = "required"
	ArgQualifier ArgType = "qualifier"
)

func (m TagArg) Parse(tag string) (string, TagArg) {
	idx := strings.Index(tag, ",")
	if idx == -1 {
		return tag, m
	}
	exps := strings.Split(tag[idx+1:], ",")
	for _, exp := range exps {
		spIdx := strings.Index(exp, "=")
		if spIdx == -1 {
			m[ArgType(exp)] = []string{""}
			continue
		}
		m[ArgType(exp[:spIdx])] = strings.Split(exp[spIdx+1:], " ")
	}
	return tag[:idx], m
}

func (m TagArg) Find(argType ArgType) ([]string, bool) {
	s, ok := m[argType]
	return s, ok
}

func (m TagArg) Has(argType ArgType, want string) bool {
	for _, s := range m[argType] {
		if s == want {
			return true
		}
	}
	return false
}
