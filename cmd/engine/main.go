// Command engine spouští HTTP server aukční simulace.
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ol1n/auction-sim/internal/api"
	"github.com/ol1n/auction-sim/internal/config"
	"github.com/ol1n/auction-sim/internal/llm"
	"github.com/ol1n/auction-sim/internal/store"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := os.MkdirAll("data", 0o755); err != nil {
		log.Error("mkdir data", "err", err)
		os.Exit(1)
	}
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Error("otevření DB", "err", err)
		os.Exit(1)
	}
	defer st.Close()

	reg, err := llm.LoadRegistry("prompts")
	if err != nil {
		log.Error("načtení promptů", "err", err)
		os.Exit(1)
	}

	srv, err := api.New(cfg, st, reg, "web")
	if err != nil {
		log.Error("init serveru", "err", err)
		os.Exit(1)
	}

	addr := ":" + cfg.EnginePort
	log.Info("auction-sim engine naslouchá", "addr", addr, "pacing", cfg.TickPacing, "llm", cfg.LLMBaseURL)
	if err := http.ListenAndServe(addr, srv.Routes("web")); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}
