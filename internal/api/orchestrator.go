package api

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ol1n/auction-sim/internal/auction"
	"github.com/ol1n/auction-sim/internal/bidder"
	"github.com/ol1n/auction-sim/internal/reporter"
)

// PoolSpec popisuje skupinu bidderů jedné role.
type PoolSpec struct {
	Role         string  `json:"role"`
	Count        int     `json:"count"`
	BudgetMin    float64 `json:"budget_min"`
	BudgetMax    float64 `json:"budget_max"`
	ValuationMin float64 `json:"valuation_min"`
	ValuationMax float64 `json:"valuation_max"`
}

// SimRequest je tělo požadavku na vytvoření simulace.
type SimRequest struct {
	Name          string                              `json:"name"`
	Item          auction.Item                        `json:"item"`
	BidderPool    []PoolSpec                          `json:"bidder_pool"`
	AuctionTypes  []auction.AuctionType               `json:"auction_types"`
	AuctionParams map[string]auction.AuctionParams    `json:"auction_params"`
	Viral         auction.ViralConfig                 `json:"viral_config"`
	RunsPerType   int                                 `json:"runs_per_type"`
}

// buildBidders sestaví bidery z pool specifikace (budgety/valuace náhodně v rozsahu).
func (s *Server) buildBidders(req SimRequest, rng *rand.Rand) []*bidder.Bidder {
	var bidders []*bidder.Bidder
	var socials []string
	idx := 0
	for _, spec := range req.BidderPool {
		role := bidder.Role(spec.Role)
		for i := 0; i < spec.Count; i++ {
			idx++
			id := fmt.Sprintf("%s-%d", role, idx)
			budget := pick(rng, spec.BudgetMin, spec.BudgetMax, req.Item.RetailPrice*1.5)
			val := pick(rng, spec.ValuationMin, spec.ValuationMax, req.Item.RetailPrice*bidder.DefaultValuationMult(role))
			b := &bidder.Bidder{
				ID:        id,
				Name:      id,
				Role:      role,
				Budget:    budget,
				Valuation: val,
				Persona:   bidder.DefaultPersona(role),
			}
			bidders = append(bidders, b)
			if role == bidder.Social {
				socials = append(socials, id)
			}
		}
	}
	// referral_bonus: přiřaď nesociálním bidderům referrera ze SOCIAL role
	if req.Viral.Has(auction.ViralReferralBonus) && len(socials) > 0 {
		for _, b := range bidders {
			if b.Role != bidder.Social {
				b.ReferrerID = socials[rng.Intn(len(socials))]
			}
		}
	}
	return bidders
}

func pick(rng *rand.Rand, min, max, def float64) float64 {
	if max <= 0 {
		return def
	}
	if min <= 0 {
		min = max * 0.7
	}
	if max < min {
		min, max = max, min
	}
	return min + rng.Float64()*(max-min)
}

// executeRuns provede všechny typy × runs a vrátí reporty.
// persist=true zapisuje do DB a publikuje události; forceFast vynutí fast clock (analýzy).
func (s *Server) executeRuns(simID string, req SimRequest, persist, forceFast bool) []reporter.AuctionReport {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	bidders := s.buildBidders(req, rng)
	pool := bidder.NewPool(bidders, s.llm, s.reg, s.cfg.LLMDecisionModel, 0.9, s.cfg.DecisionTimeout)
	pacing := s.cfg.TickPacing
	if forceFast {
		pacing = "fast"
	}
	clock := auction.NewClock(pacing)

	streak := map[string]int{}
	var reports []reporter.AuctionReport

	runs := req.RunsPerType
	if runs <= 0 {
		runs = 1
	}

	for _, typ := range req.AuctionTypes {
		params := req.AuctionParams[string(typ)]
		params.PlatformFeePct = nonZero(params.PlatformFeePct, s.cfg.PlatformFeePct)
		for run := 1; run <= runs; run++ {
			pool.ResetForAuction()
			auctionID := fmt.Sprintf("%s-%s-%d", simID, typ, run)

			eng := &auction.Engine{
				AuctionID: auctionID,
				Type:      typ,
				Item:      req.Item,
				Params:    params,
				Viral:     req.Viral,
				Parts:     pool,
				Clock:     clock,
				Emit: func(e auction.Event) {
					if persist {
						s.bus.Publish(e)
						s.bus.PublishTo(simID, e) // simulační kanál pro živý pohled
						_ = s.store.SaveEvent(e)
					}
				},
			}
			res := eng.Run(context.Background())

			s.applyReferral(req, pool, res)
			s.applyStreak(req, pool, res, streak)

			if persist {
				_ = s.store.SaveAuction(simID, auctionID, typ, run, params, res)
				_ = s.store.SaveBids(auctionID, res.Bids)
				_ = s.store.SaveBidderStates(auctionID, pool.Bidders())
			}

			reports = append(reports, reporter.Compute(reporter.RunInput{
				Type: typ, RunIndex: run, Item: req.Item, Params: params,
				Result: res, Bidders: snapshotBidders(pool),
			}))
		}
	}
	return reports
}

// run provede celou simulaci (async) a vygeneruje + uloží report.
func (s *Server) run(simID string, req SimRequest) {
	reports := s.executeRuns(simID, req, true, false)
	rep := reporter.Build(simID, req.Name, req.Item, reports)
	rep.Insights = reporter.GenerateInsights(context.Background(), s.llm, s.cfg.LLMInsightModel, rep, s.cfg.InsightTimeout)
	htmlPath, _ := reporter.Render(rep, s.cfg.ReportsDir)

	now := time.Now()
	_ = s.store.UpdateSimulationStatus(simID, "COMPLETED", &now)
	s.mu.Lock()
	st := s.sims[simID]
	st.status = "COMPLETED"
	st.report = rep
	st.htmlPath = htmlPath
	s.mu.Unlock()
}

// applyReferral připíše referrerovi % z final price, pokud jím přivedený bidder vyhrál.
func (s *Server) applyReferral(req SimRequest, pool *bidder.Pool, res auction.Result) {
	if !req.Viral.Has(auction.ViralReferralBonus) || res.WinnerID == "" {
		return
	}
	winner := pool.Get(res.WinnerID)
	if winner == nil || winner.ReferrerID == "" {
		return
	}
	ref := pool.Get(winner.ReferrerID)
	if ref == nil {
		return
	}
	bonus := res.FinalPrice * req.Viral.ReferralPct / 100.0
	ref.ReferralEarnings += bonus
}

// applyStreak sleduje účast v po sobě jdoucích bězích a aplikuje slevu vítězi se streakem.
func (s *Server) applyStreak(req SimRequest, pool *bidder.Pool, res auction.Result, streak map[string]int) {
	if !req.Viral.Has(auction.ViralStreakBonus) {
		return
	}
	for _, b := range pool.Bidders() {
		if b.BidsPlaced > 0 {
			streak[b.ID]++
		} else {
			streak[b.ID] = 0
		}
	}
	if res.WinnerID != "" && streak[res.WinnerID] >= req.Viral.StreakThreshold {
		w := pool.Get(res.WinnerID)
		if w != nil {
			discount := w.Spent * req.Viral.StreakDiscountPct / 100.0
			w.BudgetRemaining += discount
			w.Spent -= discount
		}
	}
}

func snapshotBidders(pool *bidder.Pool) []reporter.BidderEnd {
	var out []reporter.BidderEnd
	for _, b := range pool.Bidders() {
		out = append(out, reporter.BidderEnd{
			ID: b.ID, Name: b.Name, Role: string(b.Role),
			Valuation: b.Valuation, Budget: b.Budget, BudgetRemaining: b.BudgetRemaining,
			Spent: b.Spent, FeesPaid: b.FeesPaid, BidsPlaced: b.BidsPlaced,
			Won: b.Won, Active: b.Active, Emotion: string(b.Emotion),
			EntryCredit: b.EntryCredit, ReferralEarnings: b.ReferralEarnings,
		})
	}
	return out
}

func nonZero(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}
