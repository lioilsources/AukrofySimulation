package auction

import (
	"context"
	"time"
)

// runPenny: pay-to-bid. Každý bid stojí fee (sunk), zvýší cenu o increment, resetuje timer.
// Když timer doběhne, poslední leader vyhrává za aktuální cenu.
func (e *Engine) runPenny(ctx context.Context) Result {
	e.price = e.Params.StartPrice
	tickIv := e.Params.TickInterval
	if tickIv <= 0 {
		tickIv = time.Second
	}
	timer := e.Params.TimerReset
	if timer <= 0 {
		timer = 10 * time.Second
	}
	timerReset := timer

	maxTicks := e.Params.MaxTicks
	if maxTicks <= 0 {
		maxTicks = 2000
	}

	for tick := 0; tick < maxTicks && timer > 0; tick++ {
		if ctx.Err() != nil {
			break
		}
		if len(e.Parts.ActiveIDs()) == 0 {
			break
		}
		e.emit("timer_update", map[string]any{
			"timer_s": timer.Seconds(), "price": e.price, "leader_id": e.leaderID,
		})

		s := e.snapshot(tick, timer, e.price+e.Params.PriceIncrement)
		decisions := e.Parts.Decide(ctx, s)

		win, ok := e.pickWinningBid(decisions)
		if !ok {
			timer -= tickIv
			e.Clock.Sleep(tickIv)
			continue
		}
		if res, bought := e.tryBuyNow(win); bought {
			return res
		}
		// bid fee je nevratný; zkontroluj že na něj bidder má
		fee := e.Params.BidFee
		if !e.Parts.CanAfford(win.BidderID, fee) {
			e.recordNonBid(win)
			timer -= tickIv
			e.Clock.Sleep(tickIv)
			continue
		}
		prevLeader := e.leaderID
		e.Parts.Debit(win.BidderID, 0, fee)
		e.price += e.Params.PriceIncrement
		e.leaderID = win.BidderID
		e.applyEntryCredit(win.BidderID)
		e.recordBid(win, fee)
		e.emit("bid_placed", map[string]any{
			"bidder_id": win.BidderID, "price": e.price, "fee": fee, "reasoning": win.Reasoning,
		})
		if prevLeader != e.leaderID {
			e.emit("leader_changed", map[string]any{"previous": prevLeader, "current": e.leaderID})
		}

		// reset timeru; countdown_extension prodlouží při bidu v kritické zóně
		timer = timerReset
		if e.Viral.Has(ViralCountdownExtension) && e.Viral.CountdownExtendS > 0 {
			timer += e.Viral.CountdownExtendS
			e.emit("countdown_extended", map[string]any{"by_s": e.Viral.CountdownExtendS.Seconds()})
		}
		e.Clock.Sleep(tickIv)
	}

	if e.leaderID == "" {
		e.emit("no_sale", map[string]any{"reason": "žádný bid v Penny aukci"})
		return Result{}
	}
	// vítěz platí aktuální cenu
	e.Parts.Debit(e.leaderID, e.price, 0)
	return Result{WinnerID: e.leaderID, FinalPrice: e.price}
}
