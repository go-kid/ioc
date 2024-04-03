package el

import (
	"regexp"
	"strings"
)

type Helper interface {
	MatchString(s string) bool
	FindAllContent(s string) (contents []string)
	ReplaceAllContent(s string, f func(content string) (string, error)) (string, error)
}

type elHelper struct {
	*regexp.Regexp
	pre, suf int
}

func newEl(reg *regexp.Regexp, pre, suf int) Helper {
	return &elHelper{
		Regexp: reg,
		pre:    pre,
		suf:    suf,
	}
}

func (e *elHelper) FindAllContent(s string) (contents []string) {
	for _, s := range e.FindAllString(s, -1) {
		contents = append(contents, e.content(s))
	}
	return
}

func (e *elHelper) content(elr string) string {
	return elr[e.pre : len(elr)-e.suf]
}

func (e *elHelper) ReplaceAllContent(s string, f func(content string) (string, error)) (string, error) {
	var result = s
	for _, elr := range e.FindAllString(s, -1) {
		r, err := f(e.content(elr))
		if err != nil {
			return "", err
		}
		result = strings.Replace(result, elr, r, 1)
	}
	return result, nil
}

func NewQuote() Helper {
	return newEl(regexp.MustCompile("\\${[^{}]*}"), 2, 1)
}

func NewExpr() Helper {
	return newEl(regexp.MustCompile("#{[^{}]*}"), 2, 1)
}
