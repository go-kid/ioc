package debug

import "sync"

type Collector struct {
	mu         sync.RWMutex
	components map[string]*componentInfo
	edges      []DependencyEdge
	events     []DebugEvent
}

type componentInfo struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	State ComponentState `json:"state"`
}

func NewCollector() *Collector {
	return &Collector{
		components: make(map[string]*componentInfo),
	}
}

func (c *Collector) RegisterComponent(name, typeName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.components[name]; !ok {
		c.components[name] = &componentInfo{Name: name, Type: typeName, State: StateRegistered}
	}
}

func (c *Collector) SetState(name string, state ComponentState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if info, ok := c.components[name]; ok {
		info.State = state
	}
}

func (c *Collector) AddEdge(edge DependencyEdge) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.edges = append(c.edges, edge)
}

func (c *Collector) RecordEvent(event DebugEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, event)
}

func (c *Collector) GetGraph() GraphData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	nodes := make([]GraphNode, 0, len(c.components))
	for _, info := range c.components {
		nodes = append(nodes, GraphNode{
			Name:  info.Name,
			Type:  info.Type,
			State: info.State,
		})
	}
	edges := make([]DependencyEdge, len(c.edges))
	copy(edges, c.edges)
	return GraphData{Nodes: nodes, Edges: edges}
}

func (c *Collector) GetComponents() []componentInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]componentInfo, 0, len(c.components))
	for _, info := range c.components {
		result = append(result, *info)
	}
	return result
}

func (c *Collector) GetEvents() []DebugEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]DebugEvent, len(c.events))
	copy(result, c.events)
	return result
}
