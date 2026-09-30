package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ol1n/auction-sim/internal/store"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func newID(prefix string) string {
	var b [6]byte
	rand.Read(b[:])
	return fmt.Sprintf("%s_%x", prefix, b)
}

func (s *Server) handleCreateSimulation(w http.ResponseWriter, r *http.Request) {
	var req SimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if len(req.AuctionTypes) == 0 || len(req.BidderPool) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "auction_types a bidder_pool jsou povinné"})
		return
	}
	id := newID("sim")

	poolJSON, _ := json.Marshal(req.BidderPool)
	types := ""
	for i, t := range req.AuctionTypes {
		if i > 0 {
			types += ","
		}
		types += string(t)
	}
	_ = s.store.CreateSimulation(store.Simulation{
		ID: id, Name: req.Name, Item: req.Item, Viral: req.Viral,
		AuctionTypes: types, RunsPerType: req.RunsPerType, Status: "RUNNING", CreatedAt: time.Now(),
	}, string(poolJSON))

	s.mu.Lock()
	s.sims[id] = &simState{req: req, status: "RUNNING"}
	s.mu.Unlock()

	go s.run(id, req)

	writeJSON(w, http.StatusCreated, map[string]string{
		"simulation_id": id,
		"status_url":    "/api/v1/simulations/" + id,
		"live_url":      "/simulations/" + id,
	})
}

func (s *Server) handleGetSimulation(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	st := s.sims[id]
	s.mu.Unlock()
	if st != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"simulation_id": id,
			"status":        st.status,
			"name":          st.req.Name,
		})
		return
	}
	// fallback do DB — simulace z běhů před restartem enginu
	sim, err := s.store.GetSimulation(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "neznámá simulace"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"simulation_id": sim.ID,
		"status":        sim.Status,
		"name":          sim.Name,
	})
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	st := s.sims[id]
	s.mu.Unlock()
	if st != nil {
		if st.status != "COMPLETED" {
			writeJSON(w, http.StatusAccepted, map[string]string{"status": "not ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"report":   st.report,
			"html_url": "/reports/" + id + ".html",
		})
		return
	}
	// fallback do DB — po restartu už není report v paměti, ale HTML na disku ano
	sim, err := s.store.GetSimulation(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "neznámá simulace"})
		return
	}
	if sim.Status != "COMPLETED" {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": sim.Status})
		return
	}
	resp := map[string]any{"status": sim.Status}
	if _, err := os.Stat(filepath.Join(s.cfg.ReportsDir, id+".html")); err == nil {
		resp["html_url"] = "/reports/" + id + ".html"
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	s.bus.Handler(w, r, r.PathValue("id"))
}
