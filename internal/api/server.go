// Package api poskytuje HTTP server, orchestraci simulací a analýzy (compare, what-if).
package api

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/ol1n/auction-sim/internal/config"
	"github.com/ol1n/auction-sim/internal/events"
	"github.com/ol1n/auction-sim/internal/llm"
	"github.com/ol1n/auction-sim/internal/reporter"
	"github.com/ol1n/auction-sim/internal/store"
)

type simState struct {
	req      SimRequest
	status   string
	report   reporter.SimulationReport
	htmlPath string
}

// Server drží závislosti a běhový stav.
type Server struct {
	cfg   config.Config
	store *store.Store
	bus   *events.Bus
	llm   *llm.Client
	reg   *llm.Registry
	tmpl  *template.Template

	mu   sync.Mutex
	sims map[string]*simState
}

// New sestaví server.
func New(cfg config.Config, st *store.Store, reg *llm.Registry, webDir string) (*Server, error) {
	tmpl, err := template.ParseGlob(webDir + "/templates/*.html")
	if err != nil {
		return nil, err
	}
	client := llm.New(cfg.LLMBaseURL, cfg.LLMDecisionModel, cfg.CFAccessClientID, cfg.CFAccessSecret)
	return &Server{
		cfg:   cfg,
		store: st,
		bus:   events.NewBus(),
		llm:   client,
		reg:   reg,
		tmpl:  tmpl,
		sims:  map[string]*simState{},
	}, nil
}

// Routes vrátí http.Handler se všemi cestami.
func (s *Server) Routes(webDir string) http.Handler {
	mux := http.NewServeMux()

	// API
	mux.HandleFunc("POST /api/v1/simulations", s.handleCreateSimulation)
	mux.HandleFunc("GET /api/v1/simulations/{id}", s.handleGetSimulation)
	mux.HandleFunc("GET /api/v1/simulations/{id}/report", s.handleGetReport)
	mux.HandleFunc("GET /api/v1/auctions/{id}/events", s.handleSSE)
	mux.HandleFunc("POST /api/v1/analysis/compare", s.handleCompare)
	mux.HandleFunc("POST /api/v1/analysis/what-if", s.handleWhatIf)

	// Web stránky
	mux.HandleFunc("GET /", s.pageSetup)
	mux.HandleFunc("GET /simulations/{id}", s.pageLive)
	mux.HandleFunc("GET /simulations/{id}/report", s.pageReport)

	// Statika a reporty
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(webDir+"/static"))))
	mux.Handle("GET /reports/", http.StripPrefix("/reports/", http.FileServer(http.Dir(s.cfg.ReportsDir))))

	return mux
}
