package llm

import (
	"encoding/json"
	"strings"

	"github.com/ol1n/auction-sim/internal/auction"
)

type rawDecision struct {
	Action     string  `json:"action"`
	Amount     float64 `json:"amount"`
	Reasoning  string  `json:"reasoning"`
	Confidence float64 `json:"confidence"`
}

// ParseDecision přísně parsuje JSON rozhodnutí z LLM odpovědi.
// Při jakékoli chybě vrací bezpečný fallback WAIT (neutrácí, zůstává ve hře).
func ParseDecision(raw string) auction.Decision {
	fallback := auction.Decision{Action: auction.ActionWait, Reasoning: "fallback: nečitelná odpověď", Confidence: 0.5}

	s := stripFences(raw)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start < 0 || end < 0 || end <= start {
		return fallback
	}
	var rd rawDecision
	if err := json.Unmarshal([]byte(s[start:end+1]), &rd); err != nil {
		return fallback
	}

	act := auction.Action(strings.ToUpper(strings.TrimSpace(rd.Action)))
	switch act {
	case auction.ActionBid, auction.ActionWait, auction.ActionLeave, auction.ActionBuyNow:
	default:
		return fallback
	}

	conf := rd.Confidence
	if conf < 0 {
		conf = 0
	}
	if conf > 1 {
		conf = 1
	}
	if rd.Confidence == 0 {
		conf = 0.5 // chybějící → neutrální
	}
	return auction.Decision{
		Action:     act,
		Amount:     rd.Amount,
		Reasoning:  strings.TrimSpace(rd.Reasoning),
		Confidence: conf,
	}
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// odstraň ```json ... ```
		if i := strings.Index(s, "\n"); i >= 0 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	return s
}
