package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
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
	if st == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "neznámá simulace"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"simulation_id": id,
		"status":        st.status,
		"name":          st.req.Name,
	})
}

func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.mu.Lock()
	st := s.sims[id]
	s.mu.Unlock()
	if st == nil || st.status != "COMPLETED" {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "not ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"report":   st.report,
		"html_url": "/reports/" + id + ".html",
	})
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	s.bus.Handler(w, r, r.PathValue("id"))
}
