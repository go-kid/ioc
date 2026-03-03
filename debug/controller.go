package debug

import "sync"

type Controller struct {
	mu          sync.Mutex
	mode        StepMode
	breakpoints map[string]bool
	nextCh      chan struct{}
	closed      bool
}

func NewController() *Controller {
	return &Controller{
		mode:        ModeStepAll,
		breakpoints: make(map[string]bool),
		nextCh:      make(chan struct{}, 1),
	}
}

func (c *Controller) SetMode(mode StepMode) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = mode
	if mode == ModeRun {
		c.signal()
	}
}

func (c *Controller) GetMode() StepMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mode
}

func (c *Controller) SetBreakpoint(component string, enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if enabled {
		c.breakpoints[component] = true
	} else {
		delete(c.breakpoints, component)
	}
}

func (c *Controller) GetBreakpoints() map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make(map[string]bool, len(c.breakpoints))
	for k, v := range c.breakpoints {
		cp[k] = v
	}
	return cp
}

func (c *Controller) ShouldPause(componentName string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.mode {
	case ModeStepAll:
		return true
	case ModeStepWatched:
		return c.breakpoints[componentName]
	default:
		return false
	}
}

// Next signals the factory to proceed one step.
func (c *Controller) Next() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.signal()
}

// WaitForNext blocks until the user signals to proceed or mode is Run.
func (c *Controller) WaitForNext() {
	<-c.nextCh
}

func (c *Controller) signal() {
	select {
	case c.nextCh <- struct{}{}:
	default:
	}
}

func (c *Controller) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.closed {
		c.closed = true
		close(c.nextCh)
	}
}
