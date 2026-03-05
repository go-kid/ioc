package debug

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	controller *Controller
	collector  *Collector
	staticFS   fs.FS
	addr       string
	dryRun     bool
	sseClients map[chan []byte]struct{}
	sseMu      sync.Mutex
}

func NewServer(controller *Controller, collector *Collector, staticFS fs.FS) *Server {
	return &Server{
		controller: controller,
		collector:  collector,
		staticFS:   staticFS,
		sseClients: make(map[chan []byte]struct{}),
	}
}

func (s *Server) Start() (string, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(s.staticFS)))
	mux.HandleFunc("/api/events", s.handleSSE)
	mux.HandleFunc("/api/next", s.handleNext)
	mux.HandleFunc("/api/mode", s.handleMode)
	mux.HandleFunc("/api/breakpoint", s.handleBreakpoint)
	mux.HandleFunc("/api/components", s.handleComponents)
	mux.HandleFunc("/api/graph", s.handleGraph)
	mux.HandleFunc("/api/state", s.handleState)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("debug server listen: %w", err)
	}
	s.addr = listener.Addr().String()

	go func() {
		if err := http.Serve(listener, mux); err != nil {
			log.Printf("[ioc-debug] server error: %v", err)
		}
	}()

	return s.addr, nil
}

func (s *Server) Addr() string {
	return s.addr
}

// BroadcastEvent sends a debug event to all connected SSE clients.
func (s *Server) BroadcastEvent(event DebugEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	msg := fmt.Appendf(nil, "data: %s\n\n", data)
	s.sseMu.Lock()
	defer s.sseMu.Unlock()
	for ch := range s.sseClients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan []byte, 64)
	s.sseMu.Lock()
	s.sseClients[ch] = struct{}{}
	s.sseMu.Unlock()

	defer func() {
		s.sseMu.Lock()
		delete(s.sseClients, ch)
		s.sseMu.Unlock()
	}()

	// send existing events as initial state
	for _, ev := range s.collector.GetEvents() {
		data, _ := json.Marshal(ev)
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			w.Write(msg)
			flusher.Flush()
		}
	}
}

func (s *Server) handleNext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.controller.Next()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleMode(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"mode": int(s.controller.GetMode())})
	case http.MethodPost:
		var req struct {
			Mode int `json:"mode"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.controller.SetMode(StepMode(req.Mode))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBreakpoint(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.controller.GetBreakpoints())
	case http.MethodPost:
		var req struct {
			Component string `json:"component"`
			Enabled   bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.controller.SetBreakpoint(req.Component, req.Enabled)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleComponents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.collector.GetComponents())
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.collector.GetGraph())
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"mode":        s.controller.GetMode(),
		"breakpoints": s.controller.GetBreakpoints(),
		"components":  s.collector.GetComponents(),
		"graph":       s.collector.GetGraph(),
		"dryRun":      s.dryRun,
	})
}
