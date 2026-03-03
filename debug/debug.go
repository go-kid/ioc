package debug

import "time"

type StepMode int

const (
	ModeStepAll     StepMode = iota // Pause at every event
	ModeStepWatched                 // Only pause at breakpoint components
	ModeRun                         // Run without pausing
)

type ComponentState string

const (
	StateRegistered   ComponentState = "registered"
	StateScanned      ComponentState = "scanned"
	StateCreating     ComponentState = "creating"
	StatePopulating   ComponentState = "populating"
	StateInitializing ComponentState = "initializing"
	StateReady        ComponentState = "ready"
)

type DebugEvent struct {
	ID            int            `json:"id"`
	Phase         string         `json:"phase"`
	Action        string         `json:"action"`
	ComponentName string         `json:"componentName,omitempty"`
	ProcessorName string         `json:"processorName,omitempty"`
	Details       map[string]any `json:"details,omitempty"`
	Timestamp     time.Time      `json:"timestamp"`
}

type DependencyEdge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	FieldName string `json:"fieldName"`
	DepType   string `json:"depType"` // "pointer" or "interface"
}

type GraphData struct {
	Nodes []GraphNode      `json:"nodes"`
	Edges []DependencyEdge `json:"edges"`
}

type GraphNode struct {
	Name  string         `json:"name"`
	Type  string         `json:"type"`
	State ComponentState `json:"state"`
}
