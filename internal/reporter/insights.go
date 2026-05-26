package reporter

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ol1n/auction-sim/internal/llm"
)

const insightSystem = `Jsi expert na aukční teorii a behaviorální ekonomii. Analyzuj výsledky simulace
a vytvoř insight report v češtině. Zohledni nejen profit platformy, ale i viralitu/engagement
a škody na účastnících (zejména u Penny aukcí). Struktura:
1. Executive summary (cca 150 slov)
2. 3-5 klíčových zjištění s oporou v datech
3. Doporučení pro: provozovatele platformy / kupujícího / regulátora
4. Etické zhodnocení Penny aukce na základě dat`

// GenerateInsights zavolá llm-lab nad agregovanými daty reportu. Při chybě vrátí fallback text.
func GenerateInsights(ctx context.Context, client *llm.Client, model string, rep SimulationReport, timeout time.Duration) string {
	if client == nil {
		return "(LLM insight přeskočen — není nakonfigurován klient.)"
	}
	payload, _ := json.MarshalIndent(struct {
		Item       any `json:"item"`
		Aggregates any `json:"aggregates"`
	}{rep.Item, rep.Aggregates}, "", "  ")

	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, _, err := client.Complete(cctx, model, insightSystem, string(payload), 2000, 0.7)
	if err != nil {
		return "(LLM insight selhal: " + err.Error() + ")"
	}
	return out
}
