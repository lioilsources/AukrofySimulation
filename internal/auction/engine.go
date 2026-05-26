package auction

import (
	"context"
	"sort"
	"time"
)

// Clock abstrahuje časování ticků. LiveClock spí reálný čas, FastClock ne (headless batch).
type Clock interface {
	Sleep(d time.Duration)
}

type LiveClock struct{}

func (LiveClock) Sleep(d time.Duration) { time.Sleep(d) }

type FastClock struct{}

func (FastClock) Sleep(time.Duration) {}

// NewClock vrátí clock dle TICK_PACING.
func NewClock(pacing string) Clock {
	if pacing == "fast" {
		return FastClock{}
	}
	return LiveClock{}
}

// Engine řídí jeden běh jedné aukce.
type Engine struct {
	AuctionID string
	Type      AuctionType
	Item      Item
	Params    AuctionParams
	Viral     ViralConfig
	Parts     Participants
	Emit      EmitFunc
	Clock     Clock

	price    float64
	leaderID string
	bids     []BidRecord
	firstBid map[string]bool
	bidTimes []time.Time
	started  time.Time
}

// Run spustí aukci dle typu a vrátí výsledek.
func (e *Engine) Run(ctx context.Context) Result {
	e.firstBid = map[string]bool{}
	e.started = time.Now()
	if e.Emit == nil {
		e.Emit = func(Event) {}
	}
	if e.Clock == nil {
		e.Clock = LiveClock{}
	}
	e.emit("auction_started", map[string]any{
		"type": e.Type, "item": e.Item, "start_price": e.Params.StartPrice,
	})

	var res Result
	switch e.Type {
	case Vickrey:
		res = e.runVickrey(ctx)
	case Dutch:
		res = e.runDutch(ctx)
	case Penny:
		res = e.runPenny(ctx)
	}
	res.AuctionID = e.AuctionID
	res.Type = e.Type
	res.Bids = e.bids
	res.StartedAt = e.started
	res.EndedAt = time.Now()

	e.emit("auction_ended", map[string]any{
		"winner_id": res.WinnerID, "final_price": res.FinalPrice,
		"buy_now": res.BuyNow, "duration_ms": res.EndedAt.Sub(res.StartedAt).Milliseconds(),
	})
	return res
}

func (e *Engine) emit(typ string, payload map[string]any) {
	e.Emit(Event{
		AuctionID: e.AuctionID,
		Type:      typ,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

// buyNowPrice vrátí buy-it-now cenu, pokud je mechanismus aktivní.
func (e *Engine) buyNowEnabled() bool { return e.Viral.Has(ViralBuyItNow) }
func (e *Engine) buyNowPrice() float64 {
	mult := e.Viral.BuyNowMultiplier
	if mult <= 0 {
		mult = 1.3
	}
	return e.Item.RetailPrice * mult
}

// velocity vrátí počet bidů za poslední minutu (bidů/min) — pro social proof.
func (e *Engine) velocity() float64 {
	cutoff := time.Now().Add(-time.Minute)
	n := 0
	for _, t := range e.bidTimes {
		if t.After(cutoff) {
			n++
		}
	}
	return float64(n)
}

// applyEntryCredit připíše entry credit při prvním bidu biddera, pokud je mechanismus aktivní.
func (e *Engine) applyEntryCredit(id string) {
	if !e.Viral.Has(ViralEntryCredit) || e.firstBid[id] {
		return
	}
	e.firstBid[id] = true
	amt := e.Viral.EntryCreditAmount
	if amt > 0 {
		e.Parts.Credit(id, amt)
		e.emit("entry_credit", map[string]any{"bidder_id": id, "amount": amt})
	}
}

// snapshot sestaví obraz stavu pro rozhodování.
func (e *Engine) snapshot(tick int, timer time.Duration, nextPrice float64) Snapshot {
	s := Snapshot{
		AuctionID:      e.AuctionID,
		Type:           e.Type,
		Item:           e.Item,
		Params:         e.Params,
		Viral:          e.Viral,
		Tick:           tick,
		CurrentPrice:   e.price,
		NextPrice:      nextPrice,
		TimerRemaining: timer,
		LeaderID:       e.leaderID,
		ActiveBidders:  len(e.Parts.ActiveIDs()),
		BidVelocity:    e.velocity(),
	}
	if e.leaderID != "" {
		s.LeaderName = e.Parts.Name(e.leaderID)
	}
	if e.buyNowEnabled() {
		s.BuyNowEnabled = true
		s.BuyNowPrice = e.buyNowPrice()
	}
	if e.Viral.Has(ViralSocialProof) && s.ActiveBidders < e.Viral.SocialProofThreshold {
		// pod prahem se social proof signál neukazuje
		s.ActiveBidders = 0
		s.BidVelocity = 0
	}
	return s
}

// recordBid zaznamená bid a aktualizuje stav velocity.
func (e *Engine) recordBid(d Decision, fee float64) {
	now := time.Now()
	e.bids = append(e.bids, BidRecord{
		BidderID:   d.BidderID,
		Amount:     d.Amount,
		FeePaid:    fee,
		Action:     d.Action,
		Reasoning:  d.Reasoning,
		Confidence: d.Confidence,
		LatencyMs:  d.LatencyMs,
		Timestamp:  now,
	})
	e.bidTimes = append(e.bidTimes, now)
}

// recordNonBid zaznamená WAIT/LEAVE pro post-mortem (reasoning trace).
func (e *Engine) recordNonBid(d Decision) {
	e.bids = append(e.bids, BidRecord{
		BidderID:   d.BidderID,
		Action:     d.Action,
		Reasoning:  d.Reasoning,
		Confidence: d.Confidence,
		LatencyMs:  d.LatencyMs,
		Timestamp:  time.Now(),
	})
}

// pickWinningBid vybere z rozhodnutí jeden vítězný BID v rámci tiku (nejvyšší confidence).
// Vrací (decision, true) pokud nějaký BID existuje. Zpracovává i LEAVE.
func (e *Engine) pickWinningBid(ds []Decision) (Decision, bool) {
	var bidsList []Decision
	for _, d := range ds {
		switch d.Action {
		case ActionLeave:
			e.Parts.Leave(d.BidderID)
			e.recordNonBid(d)
			e.emit("bidder_left", map[string]any{"bidder_id": d.BidderID, "reasoning": d.Reasoning})
		case ActionBid, ActionBuyNow:
			bidsList = append(bidsList, d)
		default:
			e.recordNonBid(d)
		}
	}
	if len(bidsList) == 0 {
		return Decision{}, false
	}
	sort.SliceStable(bidsList, func(i, j int) bool {
		return bidsList[i].Confidence > bidsList[j].Confidence
	})
	return bidsList[0], true
}

// tryBuyNow zpracuje BUY_NOW akci. Vrací (Result, true) pokud aukce skončila koupí.
func (e *Engine) tryBuyNow(d Decision) (Result, bool) {
	if d.Action != ActionBuyNow || !e.buyNowEnabled() {
		return Result{}, false
	}
	price := e.buyNowPrice()
	if !e.Parts.CanAfford(d.BidderID, price) {
		return Result{}, false
	}
	e.Parts.Debit(d.BidderID, price, 0)
	e.recordBid(d, 0)
	e.emit("buy_now", map[string]any{"bidder_id": d.BidderID, "price": price})
	return Result{WinnerID: d.BidderID, FinalPrice: price, BuyNow: true}, true
}
