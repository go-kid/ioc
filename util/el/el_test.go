package el

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuote_MatchString(t *testing.T) {
	q := NewQuote()
	assert.True(t, q.MatchString("${foo}"))
	assert.True(t, q.MatchString("${nested.key}"))
	assert.False(t, q.MatchString("bar"))
	assert.False(t, q.MatchString(""))
}

func TestNewQuote_FindAllContent(t *testing.T) {
	q := NewQuote()
	got := q.FindAllContent("${a} and ${b}")
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestNewQuote_ReplaceAllContent(t *testing.T) {
	q := NewQuote()
	got, err := q.ReplaceAllContent("${x}-${y}", func(content string) (string, error) {
		return strings.ToUpper(content), nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "X-Y", got)
}

func TestNewExpr_MatchString(t *testing.T) {
	e := NewExpr()
	assert.True(t, e.MatchString("#{1+2}"))
	assert.True(t, e.MatchString("#{a}"))
	assert.False(t, e.MatchString("${foo}"))
	assert.False(t, e.MatchString("plain"))
}

func TestNewExpr_FindAllContent(t *testing.T) {
	e := NewExpr()
	got := e.FindAllContent("#{a} #{b}")
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestNewExpr_ReplaceAllContent(t *testing.T) {
	e := NewExpr()
	got, err := e.ReplaceAllContent("#{x}", func(content string) (string, error) {
		return "replaced", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "replaced", got)
}

func TestNewQuote_EdgeCases(t *testing.T) {
	q := NewQuote()
	// Empty string
	assert.False(t, q.MatchString(""))
	contents := q.FindAllContent("")
	assert.Empty(t, contents)

	// No matches
	assert.False(t, q.MatchString("no placeholders here"))
	contents = q.FindAllContent("no placeholders")
	assert.Empty(t, contents)

	// Nested braces: regex [^{}]* doesn't allow { inside, so only inner ${b} matches, not outer
	contents = q.FindAllContent("${a${b}}")
	assert.Equal(t, []string{"b"}, contents)
}

func TestNewExpr_EdgeCases(t *testing.T) {
	e := NewExpr()
	// Empty string
	assert.False(t, e.MatchString(""))
	contents := e.FindAllContent("")
	assert.Empty(t, contents)

	// No matches
	assert.False(t, e.MatchString("no expressions"))
	contents = e.FindAllContent("plain text")
	assert.Empty(t, contents)
}

func TestNewQuote_ReplaceAllContent_Error(t *testing.T) {
	q := NewQuote()
	_, err := q.ReplaceAllContent("${x}", func(content string) (string, error) {
		return "", assert.AnError
	})
	assert.Error(t, err)
}
