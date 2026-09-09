package draw

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrame(t *testing.T) {
	assert.Equal(t, "+---+\n|a  |\n|abc|\n+---+", Frame([]string{"a", "abc"}))
	assert.Equal(t, "++\n++", Frame(nil))
}
