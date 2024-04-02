package component_definition

import (
	"strings"
)

type (
	ArgType string
	TagArg  map[ArgType][]string
)

const (
	argSep    = ','
	argExpSep = "="

	ArgRequired  ArgType = "required"
	ArgQualifier ArgType = "qualifier"
)

func (m TagArg) Parse(tag string) (string, TagArg) {
	idx := argStartIndex(tag)
	if idx == -1 {
		return tag, m
	}
	exps := strings.Split(tag[idx+1:], string(argSep))
	for _, exp := range exps {
		spIdx := strings.Index(exp, argExpSep)
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

func argStartIndex(tag string) int {
	var in = 0
	for i, c := range tag {
		if c == '{' || c == '[' || c == '(' || c == '<' {
			in++
			continue
		}
		if c == '}' || c == ']' || c == ')' || c == '>' {
			in--
			continue
		}
		if c == argSep && in == 0 {
			return i
		}
	}
	return -1
}
