package component_definition

import (
	"github.com/go-kid/ioc/util/sort2"
	"github.com/go-kid/strings2"
	"strings"
)

type (
	ArgType string
	TagArg  map[ArgType][]string
)

const (
	argSep    = ","
	argExpSep = "="

	ArgRequired  ArgType = "Required"
	ArgQualifier ArgType = "Qualifier"
)

func (m TagArg) Parse(tag string) string {
	parts := strings2.Split(tag, argSep, strings2.DefaultSplitBlock)
	tag = parts[0]
	if len(parts) == 1 {
		return tag
	}
	exps := parts[1:]
	for _, exp := range exps {
		spIdx := strings.Index(exp, argExpSep)
		if spIdx == -1 {
			m.Set(ArgType(exp), "")
			continue
		}
		m.Set(ArgType(exp[:spIdx]), strings2.Split(exp[spIdx+1:], " ", strings2.DefaultSplitBlock)...)
	}
	return tag
}

func (m TagArg) Set(argType ArgType, val ...string) {
	if argType == "" {
		return
	}
	argType = formatArgType(argType)
	m[argType] = val
}

func (m TagArg) Add(argType ArgType, val ...string) {
	if argType == "" {
		return
	}
	argType = formatArgType(argType)
	m[argType] = append(m[argType], val...)
}

func formatArgType(argType ArgType) ArgType {
	t := string(argType)
	return ArgType(strings.ToUpper(t[:1]) + t[1:])
}

func (m TagArg) Find(argType ArgType) ([]string, bool) {
	s, ok := m[formatArgType(argType)]
	return s, ok
}

func (m TagArg) Has(argType ArgType, wants ...string) bool {
	args, ok := m[formatArgType(argType)]
	if !ok {
		return false
	}
	if len(wants) == 0 {
		return ok
	}
	return isIntersect(args, wants)
}

func isIntersect(a, b []string) bool {
	for _, a2 := range a {
		for _, b2 := range b {
			if a2 == b2 {
				return true
			}
		}
	}
	return false
}

func (m TagArg) ForEach(f func(argType ArgType, args []string)) {
	var keys []ArgType
	for argType, _ := range m {
		keys = append(keys, argType)
	}
	sort2.Slice(keys, func(i ArgType, j ArgType) bool {
		return i < j
	})
	for _, key := range keys {
		f(key, m[key])
	}
}

func (m TagArg) String() string {
	sb := strings.Builder{}
	m.ForEach(func(argType ArgType, args []string) {
		sb.WriteString("." + string(argType) + "(")
		if l := len(args); l != 0 {
			if l == 1 {
				sb.WriteString(args[0])
			} else {
				sb.WriteString(strings.Join(args, ","))
			}
		}
		sb.WriteString(")")
	})
	return sb.String()
}
