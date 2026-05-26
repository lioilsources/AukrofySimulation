package api

import (
	"encoding/json"
	"net/http"

	"github.com/ol1n/auction-sim/internal/auction"
	"github.com/ol1n/auction-sim/internal/reporter"
)

// handleCompare spustí konfigurované typy aukcí se stejnými parametry a vrátí srovnání.
func (s *Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	var req SimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	reports := s.executeRuns("", req, false, true)
	rep := reporter.Build("compare", req.Name, req.Item, reports)
	writeJSON(w, http.StatusOK, map[string]any{
		"aggregates": rep.Aggregates,
		"reports":    rep.Reports,
	})
}

// WhatIfRequest definuje sweep jednoho parametru.
type WhatIfRequest struct {
	Base      SimRequest `json:"base_config"`
	Parameter string     `json:"parameter"`
	From      float64    `json:"range_from"`
	To        float64    `json:"range_to"`
	Steps     int        `json:"steps"`
}

type whatIfPoint struct {
	Value   float64                       `json:"value"`
	Metrics map[string]float64            `json:"metrics"`
	ByType  map[string]map[string]float64 `json:"by_type"`
}

// handleWhatIf provede sweep parametru přes rozsah a vrátí křivky metrik.
func (s *Server) handleWhatIf(w http.ResponseWriter, r *http.Request) {
	var req WhatIfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if req.Steps < 1 {
		req.Steps = 5
	}
	step := 0.0
	if req.Steps > 1 {
		step = (req.To - req.From) / float64(req.Steps-1)
	}

	var points []whatIfPoint
	for i := 0; i < req.Steps; i++ {
		val := req.From + step*float64(i)
		cfg := req.Base
		if !applyParam(&cfg, req.Parameter, val) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "neznámý parametr: " + req.Parameter})
			return
		}
		reports := s.executeRuns("", cfg, false, true)
		rep := reporter.Build("whatif", cfg.Name, cfg.Item, reports)

		byType := map[string]map[string]float64{}
		var sumProfit, sumConv float64
		for t, a := range rep.Aggregates {
			byType[string(t)] = map[string]float64{
				"platform_profit": a.PlatformProfit.Mean,
				"final_price":     a.FinalPrice.Mean,
				"total_bids":      a.TotalBids.Mean,
				"conversion":      a.Conversion.Mean,
				"viral_cost":      a.ViralCost.Mean,
			}
			sumProfit += a.PlatformProfit.Mean
			sumConv += a.Conversion.Mean
		}
		n := float64(len(rep.Aggregates))
		if n == 0 {
			n = 1
		}
		points = append(points, whatIfPoint{
			Value:   val,
			Metrics: map[string]float64{"avg_platform_profit": sumProfit / n, "avg_conversion": sumConv / n},
			ByType:  byType,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"parameter": req.Parameter,
		"points":    points,
	})
}

// applyParam nastaví hodnotu sweepovaného parametru. Podporuje viral i auction params.
func applyParam(req *SimRequest, param string, val float64) bool {
	switch param {
	case "entry_credit_amount":
		req.Viral.EntryCreditAmount = val
		return true
	case "referral_pct":
		req.Viral.ReferralPct = val
		return true
	case "buy_now_multiplier":
		req.Viral.BuyNowMultiplier = val
		return true
	case "penny.bid_fee":
		setParam(req, "PENNY", func(p *auction.AuctionParams) { p.BidFee = val })
		return true
	case "penny.price_increment":
		setParam(req, "PENNY", func(p *auction.AuctionParams) { p.PriceIncrement = val })
		return true
	case "dutch.decrement":
		setParam(req, "DUTCH", func(p *auction.AuctionParams) { p.Decrement = val })
		return true
	case "platform_fee_pct":
		for k := range req.AuctionParams {
			p := req.AuctionParams[k]
			p.PlatformFeePct = val
			req.AuctionParams[k] = p
		}
		return true
	}
	return false
}

func setParam(req *SimRequest, typ string, f func(*auction.AuctionParams)) {
	if req.AuctionParams == nil {
		return
	}
	p := req.AuctionParams[typ]
	f(&p)
	req.AuctionParams[typ] = p
}
