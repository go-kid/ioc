package post_processor

import (
	"github.com/go-kid/ioc"
	"github.com/go-kid/ioc/app"
	"github.com/stretchr/testify/assert"
	"testing"
)

type PostProcessor struct {
}

func (p *PostProcessor) PostProcessBeforeInitialization(component interface{}) error {
	if c, ok := component.(*Component); ok {
		c.BeforeInitFlag = true
	}
	return nil
}

func (p *PostProcessor) PostProcessAfterInitialization(component interface{}) error {
	if c, ok := component.(*Component); ok {
		c.AfterInitFlag = true
	}
	return nil
}

type Component struct {
	BeforeInitFlag bool
	InitFlag       bool
	AfterInitFlag  bool
}

func (c *Component) Init() error {
	c.InitFlag = true
	return nil
}

func TestPostProcessor(t *testing.T) {
	c := &Component{}
	assert.Equal(t, false, c.BeforeInitFlag)
	assert.Equal(t, false, c.InitFlag)
	assert.Equal(t, false, c.AfterInitFlag)
	ioc.RunTest(t, app.SetComponents(
		c, &PostProcessor{},
	))
	assert.Equal(t, true, c.BeforeInitFlag)
	assert.Equal(t, true, c.InitFlag)
	assert.Equal(t, true, c.AfterInitFlag)
}
