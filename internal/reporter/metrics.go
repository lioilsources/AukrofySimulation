// Package reporter počítá metriky aukcí, agreguje batch běhy a renderuje HTML report.
package reporter

import (
	"github.com/ol1n/auction-sim/internal/auction"
)

// BidderEnd je snímek biddera na konci jednoho běhu.
type BidderEnd struct {
	ID               string
	Name             string
	Role             string
	Valuation        float64
	Budget           float64
	BudgetRemaining  float64
	Spent            float64
	FeesPaid         float64
	BidsPlaced       int
	Won              bool
	Active           bool
	Emotion          string
	EntryCredit      float64
	ReferralEarnings float64
}

// RunInput jsou data jednoho běhu aukce předaná reportéru.
type RunInput struct {
	Type     auction.AuctionType
	RunIndex int
	Item     auction.Item
	Params   auction.AuctionParams
	Result   auction.Result
	Bidders  []BidderEnd
}

// BidderReport je výsledek jednoho biddera.
type BidderReport struct {
	ID         string
	Name       string
	Role       string
	BidsPlaced int
	FeesPaid   float64
	AmountPaid float64
	NetOutcome float64
	BudgetUsed float64
	EndEmotion string
	Won        bool
}

// AuctionReport agreguje metriky jednoho běhu.
type AuctionReport struct {
	Type       auction.AuctionType
	RunIndex   int
	WinnerID   string
	WinnerName string
	FinalPrice float64
	BuyNow     bool

	PlatformProfit   float64
	SellerProfit     float64
	WinnerSurplus    float64
	TotalLoserLosses float64
	EfficiencyScore  float64

	TotalBidsPlaced     int
	UniqueBidders       int
	UniqueActiveBidders int
	Conversion          float64
	ViralCost           float64

	PerBidder []BidderReport
}

// Compute spočítá AuctionReport z dat běhu.
func Compute(in RunInput) AuctionReport {
	r := AuctionReport{
		Type:       in.Type,
		RunIndex:   in.RunIndex,
		WinnerID:   in.Result.WinnerID,
		FinalPrice: in.Result.FinalPrice,
		BuyNow:     in.Result.BuyNow,
	}
	feePct := in.Params.PlatformFeePct / 100.0

	var totalFees, viralCost float64
	var maxVal float64
	var maxValBidderID string
	for _, b := range in.Bidders {
		totalFees += b.FeesPaid
		viralCost += b.EntryCredit + b.ReferralEarnings
		if b.Valuation > maxVal {
			maxVal = b.Valuation
			maxValBidderID = b.ID
		}
		bidsCount := countBids(in.Result.Bids, b.ID, auction.ActionBid) + countBids(in.Result.Bids, b.ID, auction.ActionBuyNow)
		net := -b.Spent - b.FeesPaid
		amountPaid := 0.0
		if b.Won {
			amountPaid = b.Spent
			net = b.Valuation - b.Spent - b.FeesPaid
		}
		used := 0.0
		if b.Budget > 0 {
			used = (b.Budget - b.BudgetRemaining) / b.Budget * 100
		}
		if b.Won {
			r.WinnerName = b.Name
			r.WinnerSurplus = b.Valuation - b.Spent
		}
		if !b.Won {
			r.TotalLoserLosses += b.Spent + b.FeesPaid
		}
		r.TotalBidsPlaced += b.BidsPlaced
		r.UniqueBidders++
		if b.BidsPlaced > 0 {
			r.UniqueActiveBidders++
		}
		r.PerBidder = append(r.PerBidder, BidderReport{
			ID: b.ID, Name: b.Name, Role: b.Role, BidsPlaced: b.BidsPlaced,
			FeesPaid: b.FeesPaid, AmountPaid: amountPaid, NetOutcome: net,
			BudgetUsed: used, EndEmotion: b.Emotion, Won: b.Won,
		})
		_ = bidsCount
	}

	r.ViralCost = viralCost
	r.PlatformProfit = in.Result.FinalPrice*feePct + totalFees - viralCost
	r.SellerProfit = in.Result.FinalPrice * (1 - feePct)
	if r.UniqueBidders > 0 {
		r.Conversion = float64(r.UniqueActiveBidders) / float64(r.UniqueBidders)
	}
	if in.Result.WinnerID != "" && in.Result.WinnerID == maxValBidderID {
		r.EfficiencyScore = 1
	}
	return r
}

func countBids(bids []auction.BidRecord, bidderID string, act auction.Action) int {
	n := 0
	for _, b := range bids {
		if b.BidderID == bidderID && b.Action == act {
			n++
		}
	}
	return n
}
