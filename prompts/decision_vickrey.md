VICKREY AUKCE (zapečetěná, druhá cena) — TVOJE ROZHODNUTÍ

Předložíš JEDEN zapečetěný bid. Vyhrává nejvyšší nabídka, ale platí se DRUHÁ nejvyšší cena.
Ostatní nabídky NEVIDÍŠ. Teoreticky optimální je nabídnout svou pravdivou valuaci.

AUKCE:
- Zboží: {{.Item.Name}} (retail: {{printf "%.0f" .Item.RetailPrice}} Kč)
- Rezervní cena (minimum): {{printf "%.0f" .CurrentPrice}} Kč
- Účastníků: {{.ActiveBidders}}

TVŮJ STAV:
- Rozpočet: {{printf "%.0f" .BudgetRemaining}} Kč
- Tvoje valuation: {{printf "%.0f" .Valuation}} Kč
- Emocionální stav: {{.Emotion}}

ROZHODNI:
1. BID s konkrétní částkou "amount" (kolik nabízíš)
2. LEAVE — neúčastníš se

ODPOVĚZ POUZE JSON:
{"action": "BID|LEAVE", "amount": <číslo Kč pokud BID, jinak null>, "reasoning": "<2-3 věty>", "confidence": 0.0-1.0}
