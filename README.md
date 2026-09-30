# AukrofySimulation

Simulátor aukcí v Go, ve kterém dražitele řídí LLM. Umožňuje prohnat aukční
formát stovkami agentů s různými povahami a zjistit, jak se zachová — dřív, než
se pustí na živé lidi.

Vývojářské detaily: viz [CLAUDE.md](CLAUDE.md).

## Co to umí

- **Tři aukční formáty:** Dutch, Vickrey a penny (`internal/auction/`)
- **LLM-řízení dražitelé** s rolemi *sniper, gambler, collector, reseller,
  social, cautious, interested* a modelem emocí (`internal/bidder/`,
  `internal/llm/`, prompty v `prompts/`)
- **Úplný event log** v SQLite — každý běh je reprodukovatelný
- **Živé sledování** přes SSE (`internal/events/`) a webové rozhraní (`web/`)
- **HTML reporty** s agregacemi a insighty (`internal/reporter/`)

## Spuštění

```bash
make build          # → bin/engine, bin/reporter
make run            # build + spuštění simulace
make test           # go test ./...
```

Veškerá konfigurace (LLM endpoint, defaulty simulací, pacing) se nastavuje přes
env proměnné / `.env` — viz `.env.example`. Bidder pool se definuje na každou
simulaci zvlášť ve webovém formuláři. Nasazení: `docs/DEPLOYMENT.md`.

---

## Investor pitch (English)

**In one sentence:** a simulation platform that runs an auction past hundreds of
LLM-driven buyers with different temperaments, to find out which format and
which settings earn the most — before it goes live on real people.

### Problem

Anyone who runs auctions — marketplaces, estate sales, ad inventory — decides on
format and parameters blind. A/B testing in production is slow, expensive and
irreversible: a badly configured auction drives sellers away, and they don't
come back. Classical economic models, meanwhile, assume a rational agent —
precisely the assumption that does not hold for penny auctions.

### Solution

A Go engine with three implemented formats (Dutch, Vickrey, penny) and a
population of agents whose decisions are driven by an LLM according to an
assigned role and an emotion model. Every run is stored in SQLite as a complete
event log, so it is reproducible; it can be watched live over SSE, and the
reporter turns it into an HTML analysis.

### Why it can win

- **Irrationality is the feature.** The value isn't in simulating rational
  bidders — that can be computed with an equation. It's that an LLM agent
  assigned the "gambler" role behaves unpredictably in much the way a person
  does, and that is exactly what makes penny auctions profitable.
- **Reproducibility.** A complete event log means you compare configuration
  against configuration, not impression against impression.
- **Cost as a barrier for everyone else.** Simulating hundreds of agents through
  a commercial API is prohibitively expensive. On owned GPU infrastructure (DGX
  Spark) it is a tool you can actually run — this project is essentially
  impossible without your own inference.

### Status

Working prototype, May–June 2026. The auction engine with tests, bidder pool,
LLM client with parser, event bus with SSE, SQLite persistence, reporter and web
interface are in place. It is not a product: there is no customer-facing layer,
no validation against real auction data and no third-party API.

### Next milestones

Calibrating agent behaviour against real data from a live auction platform.
Without validation it is an interesting toy; with it, a tool a marketplace will
buy.

---

## Investor pitch (česky)

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
