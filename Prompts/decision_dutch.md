HOLANDSKÁ (DUTCH) AUKCE — TVOJE ROZHODNUTÍ

Cena KLESÁ. První, kdo přihlásí (BID), vyhrává za AKTUÁLNÍ cenu.

AUKCE:
- Zboží: {{.Item.Name}} (retail: {{printf "%.0f" .Item.RetailPrice}} Kč)
- Aktuální cena: {{printf "%.2f" .CurrentPrice}} Kč
- Klesne na: {{printf "%.2f" .NextPrice}} Kč za {{.TickRemaining}}s
- Aktivních bidderů: {{.ActiveBidders}}
{{if .BuyNowEnabled}}- Buy-it-now cena: {{printf "%.0f" .BuyNowPrice}} Kč
{{end}}
TVŮJ STAV:
- Rozpočet zbývá: {{printf "%.0f" .BudgetRemaining}} Kč
- Tvoje valuation: {{printf "%.0f" .Valuation}} Kč
- Emocionální stav: {{.Emotion}}

Pokud čekáš déle, dostaneš lepší cenu — ale někdo jiný může přihlásit dřív a vyhrát.

ROZHODNI:
1. BID — přihlásíš a vyhráváš za {{printf "%.2f" .CurrentPrice}} Kč
2. WAIT — počkáš na nižší cenu (riskuješ ztrátu)
3. LEAVE — trvale opustíš aukci
{{if .BuyNowEnabled}}4. BUY_NOW — koupíš ihned za {{printf "%.0f" .BuyNowPrice}} Kč
{{end}}
ODPOVĚZ POUZE JSON:
{"action": "BID|WAIT|LEAVE{{if .BuyNowEnabled}}|BUY_NOW{{end}}", "amount": null, "reasoning": "<2-3 věty>", "confidence": 0.0-1.0}
