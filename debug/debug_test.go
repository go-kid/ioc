package debug

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kid/ioc/app"
	"github.com/go-kid/ioc/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorSnapshots(t *testing.T) {
	collector := NewCollector()
	collector.RegisterComponent("service", "*example.Service")
	collector.RegisterComponent("service", "ignored")
	collector.SetState("service", StateReady)
	collector.SetState("missing", StateCreating)
	edge := DependencyEdge{From: "service", To: "repository", FieldName: "Repo", DepType: "pointer"}
	event := DebugEvent{ID: 1, Action: "component_ready", ComponentName: "service"}
	collector.AddEdge(edge)
	collector.RecordEvent(event)

	components := collector.GetComponents()
	require.Len(t, components, 1)
	assert.Equal(t, componentInfo{Name: "service", Type: "*example.Service", State: StateReady}, components[0])
	assert.Equal(t, GraphData{
		Nodes: []GraphNode{{Name: "service", Type: "*example.Service", State: StateReady}},
		Edges: []DependencyEdge{edge},
	}, collector.GetGraph())
	assert.Equal(t, []DebugEvent{event}, collector.GetEvents())

	components[0].State = StateCreating
	events := collector.GetEvents()
	events[0].Action = "changed"
	assert.Equal(t, StateReady, collector.GetComponents()[0].State)
	assert.Equal(t, "component_ready", collector.GetEvents()[0].Action)
}

func TestControllerModesAndClose(t *testing.T) {
	controller := NewController()
	assert.Equal(t, ModeStepAll, controller.GetMode())
	assert.True(t, controller.ShouldPause("service"))

	controller.SetMode(ModeStepWatched)
	assert.False(t, controller.ShouldPause("service"))
	controller.SetBreakpoint("service", true)
	assert.True(t, controller.ShouldPause("service"))
	assert.Equal(t, map[string]bool{"service": true}, controller.GetBreakpoints())
	controller.SetBreakpoint("service", false)
	assert.Empty(t, controller.GetBreakpoints())

	controller.Next()
	controller.WaitForNext()
	controller.SetMode(ModeRun)
	assert.False(t, controller.ShouldPause("service"))
	controller.WaitForNext()

	controller.Close()
	assert.NotPanics(t, controller.Close)
	assert.NotPanics(t, controller.Next)
	assert.NotPanics(t, func() { controller.SetMode(ModeRun) })
}

func TestDebugHookCollectsAndFiltersEvents(t *testing.T) {
	controller := NewController()
	controller.SetMode(ModeRun)
	collector := NewCollector()
	server := NewServer(controller, collector, nil)
	hook := newHook(controller, server, collector)

	hook.OnFactoryEvent(container.FactoryEvent{
		Phase:         "prepare",
		Action:        "component_registered",
		ComponentName: "service",
		Details:       map[string]any{"type": "*example.Service"},
	})
	hook.OnFactoryEvent(container.FactoryEvent{
		Phase:         "refresh",
		Action:        "dependency_injected",
		ComponentName: "service",
		Details: map[string]any{
			"dependency": "repository",
			"field":      "Repo",
			"depType":    "pointer",
		},
	})
	hook.OnFactoryEvent(container.FactoryEvent{Phase: "refresh", Action: "component_ready", ComponentName: "service"})
	hook.OnFactoryEvent(container.FactoryEvent{
		Phase:         "prepare",
		Action:        "component_registered",
		ComponentName: internalModulePrefix + "internal",
	})

	components := collector.GetComponents()
	require.Len(t, components, 1)
	assert.Equal(t, StateReady, components[0].State)
	assert.Equal(t, []DependencyEdge{{From: "service", To: "repository", FieldName: "Repo", DepType: "pointer"}}, collector.GetGraph().Edges)
	assert.Len(t, collector.GetEvents(), 3)
	assert.True(t, isInternalComponent(internalModulePrefix+"factory"))
	assert.False(t, isInternalComponent("service"))
}

func TestServerJSONHandlers(t *testing.T) {
	controller := NewController()
	collector := NewCollector()
	collector.RegisterComponent("service", "*example.Service")
	server := NewServer(controller, collector, nil)

	t.Run("mode", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		server.handleMode(recorder, httptest.NewRequest(http.MethodPost, "/api/mode", bytes.NewBufferString(`{"mode":2}`)))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, ModeRun, controller.GetMode())

		recorder = httptest.NewRecorder()
		server.handleMode(recorder, httptest.NewRequest(http.MethodGet, "/api/mode", nil))
		var response map[string]int
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		assert.Equal(t, int(ModeRun), response["mode"])
	})

	t.Run("breakpoint", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		server.handleBreakpoint(recorder, httptest.NewRequest(http.MethodPost, "/api/breakpoint", bytes.NewBufferString(`{"component":"service","enabled":true}`)))
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.True(t, controller.GetBreakpoints()["service"])

		recorder = httptest.NewRecorder()
		server.handleBreakpoint(recorder, httptest.NewRequest(http.MethodGet, "/api/breakpoint", nil))
		assert.Contains(t, recorder.Body.String(), "service")
	})

	t.Run("components graph and state", func(t *testing.T) {
		for _, handler := range []func(http.ResponseWriter, *http.Request){server.handleComponents, server.handleGraph, server.handleState} {
			recorder := httptest.NewRecorder()
			handler(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.True(t, json.Valid(recorder.Body.Bytes()))
		}
	})

	t.Run("method validation", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		server.handleNext(recorder, httptest.NewRequest(http.MethodGet, "/api/next", nil))
		assert.Equal(t, http.StatusMethodNotAllowed, recorder.Code)

		recorder = httptest.NewRecorder()
		server.handleMode(recorder, httptest.NewRequest(http.MethodPost, "/api/mode", bytes.NewBufferString("{")))
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}

func TestCLIFlagsAndDryRunOption(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"command", "--ioc:run_debug", "--ioc:dry_run"}
	assert.True(t, HasRunDebugFlag())
	assert.True(t, hasCLIFlag("--ioc:dry_run"))
	assert.False(t, hasCLIFlag("--missing"))

	df := newDebugFactory(WithDryRun())
	t.Cleanup(df.Close)
	assert.True(t, df.dryRun)
	assert.True(t, df.server.dryRun)
}

func TestAppCloseStopsDebugServer(t *testing.T) {
	df := newDebugFactory()
	df.controller.SetMode(ModeRun)
	addr, err := df.server.Start()
	require.NoError(t, err)

	a := app.NewApp()
	require.NoError(t, a.Run(app.SetFactory(df)))
	response, err := http.Get("http://" + addr + "/api/state")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	a.Close()
	_, err = http.Get("http://" + addr + "/api/state")
	assert.Error(t, err)
	assert.NoError(t, df.server.Close(), "closing the server twice should be safe")
}
