package reporter

import (
	"math"
	"sort"

	"github.com/ol1n/auction-sim/internal/auction"
)

// MetricStats drží statistiku jedné metriky napříč běhy.
type MetricStats struct {
	Mean     float64
	Median   float64
	Variance float64
	StdDev   float64
	Min      float64
	Max      float64
}

// AggregateReport shrnuje N běhů téhož typu aukce.
type AggregateReport struct {
	Type           auction.AuctionType
	Runs           int
	FinalPrice     MetricStats
	PlatformProfit MetricStats
	SellerProfit   MetricStats
	TotalBids      MetricStats
	Conversion     MetricStats
	ViralCost      MetricStats
	Efficiency     MetricStats
}

// Aggregate spočítá agregát z reportů jednoho typu.
func Aggregate(typ auction.AuctionType, reps []AuctionReport) AggregateReport {
	agg := AggregateReport{Type: typ, Runs: len(reps)}
	collect := func(f func(AuctionReport) float64) MetricStats {
		vals := make([]float64, len(reps))
		for i, r := range reps {
			vals[i] = f(r)
		}
		return stats(vals)
	}
	agg.FinalPrice = collect(func(r AuctionReport) float64 { return r.FinalPrice })
	agg.PlatformProfit = collect(func(r AuctionReport) float64 { return r.PlatformProfit })
	agg.SellerProfit = collect(func(r AuctionReport) float64 { return r.SellerProfit })
	agg.TotalBids = collect(func(r AuctionReport) float64 { return float64(r.TotalBidsPlaced) })
	agg.Conversion = collect(func(r AuctionReport) float64 { return r.Conversion })
	agg.ViralCost = collect(func(r AuctionReport) float64 { return r.ViralCost })
	agg.Efficiency = collect(func(r AuctionReport) float64 { return r.EfficiencyScore })
	return agg
}

func stats(vals []float64) MetricStats {
	if len(vals) == 0 {
		return MetricStats{}
	}
	var sum, min, max float64
	min = math.Inf(1)
	max = math.Inf(-1)
	for _, v := range vals {
		sum += v
		min = math.Min(min, v)
		max = math.Max(max, v)
	}
	mean := sum / float64(len(vals))
	var variance float64
	for _, v := range vals {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(vals))

	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)
	var median float64
	n := len(sorted)
	if n%2 == 1 {
		median = sorted[n/2]
	} else {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return MetricStats{Mean: mean, Median: median, Variance: variance, StdDev: math.Sqrt(variance), Min: min, Max: max}
}
