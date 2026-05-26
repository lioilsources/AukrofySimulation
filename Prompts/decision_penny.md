PENNY AUKCE — TVOJE ROZHODNUTÍ

AUKCE:
- Zboží: {{.Item.Name}} (retail: {{printf "%.0f" .Item.RetailPrice}} Kč)
- Aktuální cena: {{printf "%.2f" .CurrentPrice}} Kč
- Bid fee: {{printf "%.2f" .BidFee}} Kč (NEVRACÍ se!)
- Cena se zvedne po bidu o: {{printf "%.2f" .PriceIncrement}} Kč
- Timer do konce: {{.TimerRemaining}}s
- Aktuální leader: {{.LeaderName}}
- Aktivně přihazuje: {{.ActiveBidders}} lidí (tempo {{printf "%.1f" .BidVelocity}} bidů/min)
{{if gt .EntryCreditAvailable 0.0}}- Tvůj entry credit k dispozici: {{printf "%.2f" .EntryCreditAvailable}} Kč
{{end}}{{if .BuyNowEnabled}}- Buy-it-now cena: {{printf "%.0f" .BuyNowPrice}} Kč
{{end}}
TVŮJ STAV:
- Rozpočet zbývá: {{printf "%.0f" .BudgetRemaining}} Kč
- Tvoje valuation: {{printf "%.0f" .Valuation}} Kč
- Tvoje bidy v této aukci: {{.YourBids}}
- Tvoje fees dosud: {{printf "%.2f" .YourFees}} Kč
- Emocionální stav: {{.Emotion}}

POZOR: Sunk cost fallacy. Peníze už utracené na fees se NEVRÁTÍ ať vyhraješ nebo ne.
Rozhoduj se podle EXPECTED VALUE budoucích bidů.

ROZHODNI:
1. BID — utratíš {{printf "%.2f" .BidFee}} Kč na fee, převezmeš leadership, timer reset
2. WAIT — počkáš, ušetříš fee, zůstaneš ve hře, ale můžeš prohrát pokud timer dojde
3. LEAVE — trvale opustíš aukci
{{if .BuyNowEnabled}}4. BUY_NOW — koupíš ihned za {{printf "%.0f" .BuyNowPrice}} Kč
{{end}}
ODPOVĚZ POUZE JSON:
{"action": "BID|WAIT|LEAVE{{if .BuyNowEnabled}}|BUY_NOW{{end}}", "amount": null, "reasoning": "<2-3 věty>", "confidence": 0.0-1.0}
