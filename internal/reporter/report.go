package reporter

import "github.com/ol1n/auction-sim/internal/auction"

// SimulationReport je kompletní report simulace.
type SimulationReport struct {
	SimulationID string
	Name         string
	Item         auction.Item
	Reports      []AuctionReport
	Aggregates   map[auction.AuctionType]AggregateReport
	Insights     string
}

// Build sestaví report ze všech běhů a spočítá agregáty per typ.
func Build(simID, name string, item auction.Item, reports []AuctionReport) SimulationReport {
	r := SimulationReport{
		SimulationID: simID,
		Name:         name,
		Item:         item,
		Reports:      reports,
		Aggregates:   map[auction.AuctionType]AggregateReport{},
	}
	byType := map[auction.AuctionType][]AuctionReport{}
	for _, rep := range reports {
		byType[rep.Type] = append(byType[rep.Type], rep)
	}
	for typ, reps := range byType {
		r.Aggregates[typ] = Aggregate(typ, reps)
	}
	return r
}
