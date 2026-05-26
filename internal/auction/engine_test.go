package auction

import (
	"context"
	"testing"
	"time"
)

// fakePool je deterministický stub Participants pro unit testy logiky enginu (bez LLM).
type fakePool struct {
	budget map[string]float64
	active map[string]bool
	order  []string
	decide func(s Snapshot, id string) Decision
}

func newFakePool(ids []string, budget float64, decide func(Snapshot, string) Decision) *fakePool {
	p := &fakePool{budget: map[string]float64{}, active: map[string]bool{}, order: ids, decide: decide}
	for _, id := range ids {
		p.budget[id] = budget
		p.active[id] = true
	}
	return p
}

func (p *fakePool) ActiveIDs() []string {
	var out []string
	for _, id := range p.order {
		if p.active[id] {
			out = append(out, id)
		}
	}
	return out
}
func (p *fakePool) Name(id string) string { return id }
func (p *fakePool) Decide(_ context.Context, s Snapshot) []Decision {
	var out []Decision
	for _, id := range p.ActiveIDs() {
		d := p.decide(s, id)
		d.BidderID = id
		out = append(out, d)
	}
	return out
}
func (p *fakePool) CanAfford(id string, cost float64) bool { return p.budget[id] >= cost }
func (p *fakePool) Debit(id string, amount, fee float64)   { p.budget[id] -= amount + fee }
func (p *fakePool) Credit(id string, amount float64)       { p.budget[id] += amount }
func (p *fakePool) Leave(id string)                        { p.active[id] = false }

func TestVickreySecondPrice(t *testing.T) {
	bids := map[string]float64{"A": 100, "B": 80, "C": 60}
	pool := newFakePool([]string{"A", "B", "C"}, 1000, func(_ Snapshot, id string) Decision {
		return Decision{Action: ActionBid, Amount: bids[id], Confidence: 0.9}
	})
	eng := &Engine{
		AuctionID: "t", Type: Vickrey, Clock: FastClock{}, Parts: pool,
		Params: AuctionParams{SubmissionWindow: time.Second},
	}
	res := eng.Run(context.Background())
	if res.WinnerID != "A" {
		t.Fatalf("vítěz = %q, čekáno A", res.WinnerID)
	}
	if res.FinalPrice != 80 {
		t.Fatalf("final price = %v, čekáno 80 (druhá cena)", res.FinalPrice)
	}
}

func TestVickreyReserveNoSale(t *testing.T) {
	pool := newFakePool([]string{"A", "B"}, 1000, func(_ Snapshot, id string) Decision {
		return Decision{Action: ActionBid, Amount: 50, Confidence: 0.5}
	})
	eng := &Engine{
		AuctionID: "t", Type: Vickrey, Clock: FastClock{}, Parts: pool,
		Params: AuctionParams{ReservePrice: 100},
	}
	res := eng.Run(context.Background())
	if res.WinnerID != "" {
		t.Fatalf("čekán no-sale pod rezervou, ale vyhrál %q", res.WinnerID)
	}
}

func TestDutchFirstBidWins(t *testing.T) {
	// A přihodí jakmile cena klesne na/pod 70.
	pool := newFakePool([]string{"A", "B"}, 1000, func(s Snapshot, id string) Decision {
		if id == "A" && s.CurrentPrice <= 70 {
			return Decision{Action: ActionBid, Confidence: 0.9}
		}
		return Decision{Action: ActionWait, Confidence: 0.5}
	})
	eng := &Engine{
		AuctionID: "t", Type: Dutch, Clock: FastClock{}, Parts: pool,
		Params: AuctionParams{StartPrice: 100, Decrement: 10, ReservePrice: 0, MaxTicks: 20},
	}
	res := eng.Run(context.Background())
	if res.WinnerID != "A" {
		t.Fatalf("vítěz = %q, čekáno A", res.WinnerID)
	}
	if res.FinalPrice != 70 {
		t.Fatalf("final price = %v, čekáno 70", res.FinalPrice)
	}
}

func TestPennyFeesAndWinner(t *testing.T) {
	// A přihazuje 3×, pak už nikdo → timer doběhne, A vyhrává. Cena = start + 3*increment.
	count := 0
	pool := newFakePool([]string{"A", "B"}, 1000, func(_ Snapshot, id string) Decision {
		if id == "A" && count < 3 {
			count++
			return Decision{Action: ActionBid, Confidence: 0.9}
		}
		return Decision{Action: ActionWait, Confidence: 0.5}
	})
	eng := &Engine{
		AuctionID: "t", Type: Penny, Clock: FastClock{}, Parts: pool,
		Params: AuctionParams{StartPrice: 1, BidFee: 5, PriceIncrement: 1,
			TimerReset: 3 * time.Second, TickInterval: time.Second, MaxTicks: 50},
	}
	res := eng.Run(context.Background())
	if res.WinnerID != "A" {
		t.Fatalf("vítěz = %q, čekáno A", res.WinnerID)
	}
	if res.FinalPrice != 4 { // 1 + 3*1
		t.Fatalf("final price = %v, čekáno 4", res.FinalPrice)
	}
	// A zaplatil 3 fees (15) + final price (4) = 19 → zbývá 981
	if pool.budget["A"] != 1000-15-4 {
		t.Fatalf("budget A = %v, čekáno %v", pool.budget["A"], 1000-19)
	}
}

func TestBuyNowEndsAuction(t *testing.T) {
	pool := newFakePool([]string{"A"}, 100000, func(s Snapshot, id string) Decision {
		if s.BuyNowEnabled {
			return Decision{Action: ActionBuyNow, Confidence: 1}
		}
		return Decision{Action: ActionWait}
	})
	eng := &Engine{
		AuctionID: "t", Type: Dutch, Clock: FastClock{}, Parts: pool,
		Item:   Item{RetailPrice: 1000},
		Params: AuctionParams{StartPrice: 1200, Decrement: 10, MaxTicks: 5},
		Viral:  ViralConfig{Mechanisms: []ViralMechanism{ViralBuyItNow}, BuyNowMultiplier: 1.3},
	}
	res := eng.Run(context.Background())
	if !res.BuyNow || res.FinalPrice != 1300 {
		t.Fatalf("čekán buy-now za 1300, dostal buy=%v price=%v", res.BuyNow, res.FinalPrice)
	}
}
