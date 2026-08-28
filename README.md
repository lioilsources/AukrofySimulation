# AukrofySimulation

Simulátor aukcí v Go, ve kterém dražitele řídí LLM. Umožňuje prohnat aukční
formát stovkami agentů s různými povahami a zjistit, jak se zachová — dřív, než
se pustí na živé lidi.

Vývojářské detaily: viz [CLAUDE.md](CLAUDE.md).

## Co to umí

- **Tři aukční formáty:** Dutch, Vickrey a penny (`internal/auction/`)
- **LLM-řízení dražitelé** s rolemi *sniper, gambler, collector, reseller,
  social, cautious, interested* a modelem emocí (`internal/bidder/`,
  `internal/llm/`, prompty v `Prompts/`)
- **Úplný event log** v SQLite — každý běh je reprodukovatelný
- **Živé sledování** přes SSE (`internal/events/`) a webové rozhraní (`web/`)
- **HTML reporty** s agregacemi a insighty (`cmd/reporter`)

## Spuštění

```bash
make build          # → bin/engine, bin/reporter
make run            # build + spuštění simulace
make test           # go test ./...
```

Konfigurace agentů (strategie, parametry, počáteční rozpočet) je v YAML;
LLM endpoint se nastavuje přes `.env` — viz `.env.example`.

---

## Investor pitch

**Jednou větou:** simulační platforma, ve které aukci prožene stovky
LLM-řízených kupujících s různými povahami, aby se dalo zjistit, jaký formát
a jaké nastavení vydělá nejvíc — dřív, než se pustí na živé lidi.

### Problém

Kdo provozuje aukce — marketplace, dražby, prodej reklamního prostoru —
rozhoduje o formátu a parametrech naslepo. A/B test na živém provozu je pomalý,
drahý a nevratný: špatně nastavená aukce odradí prodávající a ti se nevrátí.
Klasické ekonomické modely zase předpokládají racionálního agenta, což je přesně
ten předpoklad, který u penny aukcí neplatí.

### Řešení

Engine v Go se třemi implementovanými formáty (Dutch, Vickrey, penny)
a populací agentů, jejichž rozhodování řídí LLM podle přidělené role a modelu
emocí. Každý běh se ukládá do SQLite jako úplný event log, takže je
reprodukovatelný; přes SSE jde sledovat živě a reporter z něj vygeneruje HTML
analýzu.

### Proč to může vyhrát

- **Iracionalita je feature.** Hodnota není v simulaci racionálních dražitelů —
  ta se dá spočítat rovnicí. Je v tom, že LLM agent s rolí „gambler" se chová
  nepředvídatelně podobně jako člověk, a právě to dělá penny aukce ziskovými.
- **Reprodukovatelnost.** Úplný event log znamená, že se dá porovnávat
  konfigurace proti konfiguraci, ne dojem proti dojmu.
- **Náklady jako bariéra pro ostatní.** Simulace se stovkami agentů přes
  komerční API je neúnosně drahá. Nad vlastní GPU infrastrukturou (DGX Spark)
  je to smysluplně provozovatelný nástroj — tenhle projekt bez vlastní inference
  prakticky nejde dělat.

### Stav

Funkční prototyp, květen–červen 2026. Stojí auction engine s testy, bidder pool,
LLM klient s parserem, event bus se SSE, SQLite persistence, reporter a webové
rozhraní. Není to produkt: chybí zákaznická vrstva, validace proti reálným
aukčním datům a API pro třetí strany.

### Nejbližší milníky

Zkalibrovat chování agentů proti reálným datům z běžící aukční platformy. Bez
validace je to zajímavá hračka; s ní nástroj, který si marketplace koupí.
