package debug

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-kid/ioc/container"
)

const internalModulePrefix = "github.com/go-kid/ioc/"

type debugHook struct {
	controller *Controller
	server     *Server
	collector  *Collector
	eventSeq   atomic.Int64
}

func newHook(controller *Controller, server *Server, collector *Collector) *debugHook {
	return &debugHook{
		controller: controller,
		server:     server,
		collector:  collector,
	}
}

func isInternalComponent(name string) bool {
	return name != "" && strings.HasPrefix(name, internalModulePrefix)
}

func (h *debugHook) isFilteredEvent(event container.FactoryEvent) bool {
	if isInternalComponent(event.ComponentName) {
		return true
	}
	if event.Action == "dependency_injected" {
		if dep, ok := event.Details["dependency"].(string); ok && isInternalComponent(dep) {
			return true
		}
	}
	return false
}

func (h *debugHook) OnFactoryEvent(event container.FactoryEvent) {
	de := DebugEvent{
		ID:            int(h.eventSeq.Add(1)),
		Phase:         event.Phase,
		Action:        event.Action,
		ComponentName: event.ComponentName,
		ProcessorName: event.ProcessorName,
		Details:       event.Details,
		Timestamp:     time.Now(),
	}

	if h.isFilteredEvent(event) {
		return
	}

	h.updateCollector(de)
	h.collector.RecordEvent(de)
	h.server.BroadcastEvent(de)

	if h.controller.ShouldPause(event.ComponentName) {
		h.controller.WaitForNext()
	}
}

func (h *debugHook) updateCollector(event DebugEvent) {
	switch event.Action {
	case "component_registered":
		typeName, _ := event.Details["type"].(string)
		h.collector.RegisterComponent(event.ComponentName, typeName)
	case "definition_scanned":
		h.collector.SetState(event.ComponentName, StateScanned)
	case "component_creating":
		h.collector.SetState(event.ComponentName, StateCreating)
	case "populating":
		h.collector.SetState(event.ComponentName, StatePopulating)
	case "dependency_injected":
		if dep, ok := event.Details["dependency"].(string); ok {
			fieldName, _ := event.Details["field"].(string)
			depType, _ := event.Details["depType"].(string)
			h.collector.AddEdge(DependencyEdge{
				From:      event.ComponentName,
				To:        dep,
				FieldName: fieldName,
				DepType:   depType,
			})
		}
	case "before_initialization":
		h.collector.SetState(event.ComponentName, StateInitializing)
	case "component_ready":
		h.collector.SetState(event.ComponentName, StateReady)
	}
}
