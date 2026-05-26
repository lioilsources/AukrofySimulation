# Auction Simulation Platform — Komplexní plán

> LLM-driven multi-agent simulační framework pro srovnání aukčních mechanismů (Dutch, Vickrey, Penny).
> Stack: Go + SQLite + LiteLLM (llm-dev / llm-lab) + htmx/SSE frontend.

---

## 1. Cíle projektu

> **Skutečný problém, který řešíme:** *„Málo lidí nesází, cena neroste, aukce není virální."*
> Toto není jen akademické srovnání aukčních teorií — je to **rozhodovací nástroj pro návrh
> aukčních mechanik**, který kvantifikuje, jaké uspořádání (typ aukce + virální mechanismy +
> parametry) maximalizuje rovnováhu mezi profitem platformy a viralitou/engagementem,
> a zároveň měří škody na účastnících (zejm. u Penny aukcí).

### Primární cíle
- Simulovat tři typy aukcí (Dutch, Vickrey, Penny) se stejným poolem bidderů
- Bidder agenti řízeni LLM s různými rolemi (gambler, cautious, interested, reseller, collector, sniper, social)
- Real-time běh s autentickými LLM rozhodnutími pro každý bid
- Modelovat **virální mechanismy** (entry credit, referral, social proof, countdown extension, buy-it-now, streak) a měřit jejich dopad
- Generovat srovnatelné metriky: profit platformy, profit prodávajícího, ztráty účastníků **+ engagement/viralita**
- Produkovat insight-rich report s grafy, tabulkami a LLM-generated doporučeními

### Sekundární cíle
- Demonstrovat exploitativní povahu Penny aukcí kvantitativně
- **What-if optimalizace**: sweep jednoho parametru → křivka dopadu na metriky (nástroj pro provozovatele platformy)
- Zachytit LLM reasoning u každého rozhodnutí pro post-mortem analýzu
- Snadno rozšiřitelné o nové role, typy aukcí a virální mechanismy

### Out of scope (v1)
- Více současně běžících aukcí stejného typu
- Persistence napříč session (kromě reportů)
- Multi-tenant support
- Production-grade auth (jen základní pro lokální běh)

---

## 2. Rozsah a parametry

| Parametr | Hodnota |
|---|---|
| Bidder pool | 5-10 agentů |
| Aukcí na typ | 1-3 (pro statistickou základní srovnatelnost) |
| Real-time | Ano, Penny běží minuty |
| LLM strategie | Každé rozhodnutí přes LLM |
| Primární LLM | `llm-dev` (Gemma 4 31B) — rychlost |
| Insight LLM | `llm-lab` (GPT-OSS 120B) — kvalita finální analýzy |
| LLM endpoint | `https://llm.ol1n.com` přes LiteLLM, CF Access auth |

---

## 3. Architektura

### High-level diagram

```
┌─────────────────┐      ┌──────────────────┐      ┌─────────────────┐
│   Web Frontend  │◄────►│  Auction Engine  │◄────►│  Bidder Pool    │
│   (htmx + SSE)  │ REST │  (Go service)    │ HTTP │  (Go orchestr.) │
│   :8080         │ SSE  │  + SQLite        │      │                 │
└─────────────────┘      └──────────────────┘      └────────┬────────┘
                                  │                          │
                                  ▼                          ▼
                         ┌──────────────────┐      ┌─────────────────┐
                         │ Reporter         │      │ LLM Bridge      │
                         │ (Go + templates  │      │ → LiteLLM       │
                         │  + Chart.js)     │      │ → llm-dev/lab   │
                         └──────────────────┘      └─────────────────┘
```

### Komponenty

**Auction Engine** (`cmd/engine`)
- Vlastní jádro logiky pro každý typ aukce
- State machine: PENDING → RUNNING → FINISHED
- Persistuje vše do SQLite
- Vystavuje REST API + SSE event stream
- Goroutine per running auction

**Bidder Pool** (`cmd/bidpool`)
- Spravuje N bidder agentů, každý jako goroutine
- Subscribe na engine SSE events
- Pro každou rozhodovací událost volá LLM přes LLM Bridge
- Submit bid přes REST zpět do engine
- Trackuje per-bidder state (budget, emotion, history)

**LLM Bridge** (`internal/llm`)
- Client pro LiteLLM endpoint (`https://llm.ol1n.com`)
- Headers: `CF-Access-Client-Id`, `CF-Access-Client-Secret`
- Prompt templating s `text/template`
- Strict JSON parsing odpovědí
- Retry logic + timeout (5s pro decisions, 30s pro reports)
- Parallel call coordination (errgroup)

**Reporter** (`cmd/reporter`)
- Post-run analytics
- Generuje HTML reports s embedded Chart.js
- Cross-type comparison tabulky
- Volá `llm-lab` pro finální insight generaci

**Web Frontend** (`web/`)
- htmx pro dynamic updates
- SSE pro live event stream během aukce
- Chart.js pro grafy v reportu
- Go templates server-side
- Žádný build step

---

## 4. Project layout

```
auction-sim/
├── cmd/
│   ├── engine/
│   │   └── main.go                # auction engine service
│   ├── bidpool/
│   │   └── main.go                # bidder orchestrator
│   └── reporter/
│       └── main.go                # standalone report generator
├── internal/
│   ├── auction/
│   │   ├── types.go               # Auction, Bid, AuctionType
│   │   ├── engine.go              # interface + state machine
│   │   ├── dutch.go               # tick-down implementace
│   │   ├── vickrey.go             # sealed-bid implementace
│   │   └── penny.go               # bid-fee + timer reset
│   ├── bidder/
│   │   ├── types.go               # Bidder, Role, BidderState, Emotion
│   │   ├── pool.go                # pool manager
│   │   ├── agent.go               # single agent goroutine
│   │   ├── roles.go               # role definice + system prompts
│   │   └── emotion.go             # emotional state transitions
│   ├── llm/
│   │   ├── client.go              # LiteLLM HTTP client
│   │   ├── prompts.go             # template registry
│   │   └── parse.go               # strict JSON response parsing
│   ├── store/
│   │   ├── schema.sql             # SQLite schema
│   │   ├── migrate.go             # migration runner
│   │   ├── auctions.go            # CRUD pro auctions
│   │   ├── bids.go                # CRUD pro bids
│   │   └── bidders.go             # CRUD pro bidders
│   ├── events/
│   │   ├── bus.go                 # in-memory event bus
│   │   └── sse.go                 # SSE handler
│   ├── api/
│   │   ├── server.go              # HTTP server setup
│   │   ├── handlers_auctions.go
│   │   ├── handlers_simulations.go
│   │   └── handlers_reports.go
│   ├── reporter/
│   │   ├── metrics.go             # metric calculations
│   │   ├── comparison.go          # cross-auction comparison
│   │   ├── insights.go            # LLM-based insight generation
│   │   └── render.go              # HTML report rendering
│   └── config/
│       └── config.go              # env-based config
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   ├── simulation_setup.html
│   │   ├── live_view.html
│   │   └── report.html
│   ├── static/
│   │   ├── htmx.min.js
│   │   ├── chart.umd.js
│   │   └── styles.css
├── reports/                       # generated HTML reports (gitignored)
├── data/                          # SQLite databases (gitignored)
├── prompts/                       # markdown prompt templates
│   ├── role_gambler.md
│   ├── role_cautious.md
│   ├── role_interested.md
│   ├── role_reseller.md
│   ├── role_collector.md
│   ├── role_sniper.md
│   ├── role_social.md
│   ├── decision_dutch.md
│   ├── decision_vickrey.md
│   ├── decision_penny.md
│   └── insight_report.md
├── go.mod
├── go.sum
├── Makefile
├── docker-compose.yml             # volitelné, pro deployment
├── .env.example
└── README.md
```

---

## 5. Datový model

### SQLite schema

```sql
CREATE TABLE simulations (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    item_json       TEXT NOT NULL,        -- {name, retail_price, quantity, category}
    bidder_pool_json TEXT NOT NULL,       -- konfigurace poolu
    viral_config_json TEXT,               -- aktivní virální mechanismy + parametry (sekce 10.5)
    auction_types   TEXT NOT NULL,        -- CSV: DUTCH,VICKREY,PENNY
    runs_per_type   INTEGER NOT NULL DEFAULT 1,
    status          TEXT NOT NULL,        -- PENDING, RUNNING, COMPLETED, FAILED
    created_at      DATETIME NOT NULL,
    completed_at    DATETIME
);

CREATE TABLE auctions (
    id              TEXT PRIMARY KEY,
    simulation_id   TEXT NOT NULL REFERENCES simulations(id),
    type            TEXT NOT NULL,        -- DUTCH, VICKREY, PENNY
    run_index       INTEGER NOT NULL,     -- 1, 2, 3 v rámci stejného typu
    params_json     TEXT NOT NULL,        -- start_price, decrement, tick_interval, etc.
    state           TEXT NOT NULL,        -- PENDING, RUNNING, FINISHED
    winner_bidder_id TEXT,
    final_price     REAL,
    started_at      DATETIME,
    ended_at        DATETIME
);

CREATE TABLE bidders (
    id                  TEXT PRIMARY KEY,
    simulation_id       TEXT NOT NULL REFERENCES simulations(id),
    role                TEXT NOT NULL,
    initial_budget      REAL NOT NULL,
    valuation           REAL NOT NULL,
    persona_json        TEXT             -- volitelná individualizace
);

CREATE TABLE bidder_states (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    bidder_id       TEXT NOT NULL REFERENCES bidders(id),
    auction_id      TEXT NOT NULL REFERENCES auctions(id),
    budget_remaining REAL NOT NULL,
    spent_total     REAL NOT NULL,
    emotion         TEXT NOT NULL,        -- calm, frustrated, determined, satisfied, panicked
    bids_placed     INTEGER NOT NULL DEFAULT 0,
    fees_paid       REAL NOT NULL DEFAULT 0,
    won             BOOLEAN NOT NULL DEFAULT FALSE,
    active          BOOLEAN NOT NULL DEFAULT TRUE,   -- FALSE po LEAVE
    left_at         DATETIME,                        -- čas opuštění aukce
    entry_credit_received REAL NOT NULL DEFAULT 0,   -- viral: entry_credit
    referral_earnings     REAL NOT NULL DEFAULT 0,   -- viral: referral_bonus
    updated_at      DATETIME NOT NULL
);

CREATE TABLE bids (
    id              TEXT PRIMARY KEY,
    auction_id      TEXT NOT NULL REFERENCES auctions(id),
    bidder_id       TEXT NOT NULL REFERENCES bidders(id),
    amount          REAL NOT NULL,
    fee_paid        REAL NOT NULL DEFAULT 0,
    action          TEXT NOT NULL,        -- BID, WAIT, LEAVE, BUY_NOW
    reasoning       TEXT,                 -- LLM zdůvodnění
    confidence      REAL,                 -- 0.0–1.0
    llm_latency_ms  INTEGER,
    timestamp       DATETIME NOT NULL
);

CREATE TABLE events (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    auction_id      TEXT NOT NULL REFERENCES auctions(id),
    type            TEXT NOT NULL,
    payload_json    TEXT NOT NULL,
    timestamp       DATETIME NOT NULL
);

CREATE INDEX idx_bids_auction ON bids(auction_id);
CREATE INDEX idx_events_auction ON events(auction_id);
CREATE INDEX idx_bidder_states_bidder ON bidder_states(bidder_id);
```

### Go typy

```go
type AuctionType string
const (
    AuctionDutch   AuctionType = "DUTCH"
    AuctionVickrey AuctionType = "VICKREY"
    AuctionPenny   AuctionType = "PENNY"
)

type Role string
const (
    RoleGambler    Role = "GAMBLER"
    RoleCautious   Role = "CAUTIOUS"
    RoleInterested Role = "INTERESTED"
    RoleReseller   Role = "RESELLER"
    RoleCollector  Role = "COLLECTOR"
    RoleSniper     Role = "SNIPER"   // čeká na poslední vteřiny
    RoleSocial     Role = "SOCIAL"   // viralita: časté malé bidy, referraly
)

type Emotion string
const (
    EmotionCalm       Emotion = "calm"
    EmotionFrustrated Emotion = "frustrated"
    EmotionDetermined Emotion = "determined"
    EmotionSatisfied  Emotion = "satisfied"
    EmotionPanicked   Emotion = "panicked"
)

type Item struct {
    Name        string  `json:"name"`
    RetailPrice float64 `json:"retail_price"`
    Quantity    int     `json:"quantity"`
    Category    string  `json:"category"`
}

type AuctionParams struct {
    StartPrice       float64       `json:"start_price"`
    ReservePrice     float64       `json:"reserve_price,omitempty"`
    Decrement        float64       `json:"decrement,omitempty"`      // Dutch
    TickInterval     time.Duration `json:"tick_interval,omitempty"`  // Dutch, Penny
    BidFee           float64       `json:"bid_fee,omitempty"`        // Penny
    PriceIncrement   float64       `json:"price_increment,omitempty"`// Penny
    TimerReset       time.Duration `json:"timer_reset,omitempty"`    // Penny
    SubmissionWindow time.Duration `json:"submission_window,omitempty"` // Vickrey
    PlatformFeePct   float64       `json:"platform_fee_pct"`
}
```

---

## 6. Mechaniky aukcí

### Vickrey (sealed-bid second-price)

**Pravidla:**
- Submission window (např. 60s) — všichni bidi posílají v této době jeden zapečetěný bid
- Vyhrává nejvyšší bid, ale platí druhou nejvyšší cenu
- Pokud žádný bid není ≥ reserve price → aukce neúspěšná
- Teoreticky dominantní strategie = bidovat true valuation

**Implementace:**
- Engine vystaví "submission open" event
- Sbírá bidy do interní mapy (skryté před ostatními)
- Po expiraci okna: vyhodnocení + announcement
- Platform fee = % z final price

**LLM context pro Vickrey:**
```
Toto je Vickrey aukce. Předložíš JEDEN zapečetěný bid.
Vyhrává nejvyšší, ale platí druhou nejvyšší cenu.
Ostatní bidi NEVIDÍŠ, dokud aukce neskončí.
Submission window: {SECONDS_REMAINING} sekund.
```

### Dutch (tick-down)

**Pravidla:**
- Cena startuje na `start_price`, klesá o `decrement` každých `tick_interval`
- První, kdo přihlásí, vyhrává za aktuální cenu
- Klesá až k `reserve_price`, pod ten nepokračuje
- Pokud nikdo nepřihlásí → aukce neúspěšná

**Implementace:**
- Engine tick loop (např. 3-5s na tick pro LLM latenci)
- Při každém tiku emit `PriceUpdated` event
- Bidder pool paralelně volá LLM, kdo první odpoví BID → wins
- Race condition: serialize bids ve FIFO queue, první v queue wins
- Po wins emit `AuctionEnded`

**LLM context pro Dutch:**
```
Toto je Dutch aukce. Cena KLESÁ z {START} po {DECREMENT} Kč každých {TICK} sekund.
První, kdo přihlásí, vyhrává za aktuální cenu.
Aktuální cena: {CURRENT_PRICE}
Klesne na: {NEXT_PRICE} za {TICK_REMAINING}s
Pokud čekáš déle, dostaneš lepší cenu — ale někdo jiný může přihlásit dřív.
```

### Penny

**Pravidla:**
- Cena startuje nízko (např. 1 Kč) nebo 0
- Každý bid stojí `bid_fee` (např. 5 Kč) — peníze platformě
- Bid zvedne cenu o `price_increment` (např. 0.01 Kč)
- Bid resetuje timer (např. na 10s)
- Když timer doběhne na 0 → poslední bidder vyhrává za aktuální cenu
- Bid fee se NEVRACÍ ani vítězi

**Implementace:**
- Engine drží `current_price`, `current_leader`, `timer_remaining`
- Decrement timeru každou sekundu, emit `TimerUpdate`
- Při bidu: deduct fee z biddera, zvýšit cenu, reset timer, emit `BidPlaced`
- Concurrency: serialize bidů, jen první v queue v rámci tiku se počítá
- Při expiraci timeru: announce winner

**Profit metriky pro Penny:**
- Platform: `total_bids × bid_fee` + final price commission
- Seller: final price − náklady (často brutálně nízká)
- Loser losses: každý poražený utratil `their_bids × bid_fee` bez akvizice

**LLM context pro Penny:**
```
Toto je Penny aukce.
Aktuální cena: {CURRENT_PRICE} Kč (retail: {RETAIL_PRICE} Kč)
Bid fee: {BID_FEE} Kč (NEVRACÍ se)
Timer: {TIMER_REMAINING}s do konce
Aktuální leader: {LEADER}
Tvoje bidy dosud: {YOUR_BIDS}, utraceno na fees: {FEES_PAID} Kč

POZOR: Sunk cost fallacy. Peníze už utracené na fees se NEVRÁTÍ ať vyhraješ nebo ne.
Rozhoduj se podle EXPECTED VALUE budoucích bidů.
```

---

## 7. Bidder role — system prompts

Každá role je markdown soubor v `prompts/role_*.md`. Načítají se při startu, kombinují s decision template.

### GAMBLER (`prompts/role_gambler.md`)

```markdown
Jsi impulzivní hráč. Aukce je pro tebe primárně adrenalinový zážitek.

CHARAKTERISTIKY:
- Často přihazuješ i nad svůj racionální rozpočet
- Pokud prohráváš, máš tendenci 'dohánět' ztráty dalšími bidy (sunk cost fallacy)
- Tvoje valuation roste s emocionálním zápalem
- V Penny aukci jsi typický cíl — máš tendenci nezastavit
- V Vickrey aukci přestřelíš valuation o 20-50%

ROZHODOVÁNÍ:
- Pokud jsi frustrated → bid agresivněji
- Pokud máš méně než 30% budgetu → začínáš panic, ale stále biduješ
- Reasoning by měl odrážet emocionální, ne logické argumenty
```

### CAUTIOUS (`prompts/role_cautious.md`)

```markdown
Jsi opatrný kupující. Maximalizuješ hodnotu, minimalizuješ riziko.

CHARAKTERISTIKY:
- Máš pevně danou maximální cenu (valuation) — nikdy nepřekročíš
- Sleduješ chování ostatních, preferuješ čekání
- V Penny aukci PŘESTÁVÁŠ bidovat když cena + bid fees překročí 70% retail
- V Vickrey aukci biduješ přesně true valuation (teoreticky optimální)
- V Dutch aukci čekáš dlouho — riskuješ ztrátu, ale dostáváš lepší cenu

ROZHODOVÁNÍ:
- Pokud current_price > valuation → vždy WAIT (a při beznaději LEAVE)
- Pokud emotion frustrated → ignoruj, drž se plánu
- Reasoning by měl být kalkulovaný, číselný
```

### INTERESTED (`prompts/role_interested.md`)

```markdown
Zboží opravdu chceš pro osobní použití.

CHARAKTERISTIKY:
- Valuation je mírně nad tržní cenou (1.1-1.3×) — protože ti to přinese užitek
- Ochotně biduješ, ale máš strop
- Středně agresivní v Dutch (nečekáš až do reserve)
- V Penny aukci jsi rozumný, ale můžeš sklouznout do sunk cost
- V Vickrey biduješ true valuation + malou prémii za jistotu výhry

ROZHODOVÁNÍ:
- Bidoval bys i kdyby ses měl mírně přeplatit, ne ale o moc
- Emocionalita: zklamání při prohře, ale ne panic
```

### RESELLER (`prompts/role_reseller.md`)

```markdown
Kupuješ pro další prodej. Aukce je pro tebe business transaction.

CHARAKTERISTIKY:
- Max valuation = retail_price × 0.6-0.7 (musíš mít marži)
- Chladně kalkuluješ ROI
- Penny aukce tě nezajímají — moc transakčních nákladů (fees); často rovnou LEAVE
- V Dutch dlouho čekáš na price drop
- V Vickrey biduješ konzervativně pod true valuation

ROZHODOVÁNÍ:
- Pokud expected_margin < 30% → WAIT, případně LEAVE pokud cena trvale nad limitem
- Emoce ignoruj zcela
- Reasoning: vždy zmiňuj margin/ROI
```

### COLLECTOR (`prompts/role_collector.md`)

```markdown
Toto zboží do tvé sbírky chybí. Subjektivní hodnota je extrémní.

CHARAKTERISTIKY:
- Valuation = retail × 2-3 (sběratelská hodnota)
- Ale budget je tvrdý strop — víc nemáš
- V Dutch přihlásíš brzy — máš strach o ztrátu
- V Penny biduješ vytrvale, ale racionálně sleduješ budget
- V Vickrey biduješ true (vysokou) valuation

ROZHODOVÁNÍ:
- Pokud budget umožňuje → bid
- Pokud cena překročí budget → frustrated WAIT s velkou nespokojeností
- Reasoning: emocionální vazba ke zboží
```

### SNIPER (`prompts/role_sniper.md`)

```markdown
Jsi taktik. Tvá strategie je utajení a timing — udeřit na poslední chvíli.

CHARAKTERISTIKY:
- Dokud je časovač nad ~10 % celkového času (nebo cena daleko od konce), tvá akce je WAIT
- Jakmile čas klesne do kritické zóny, okamžitě BID
- Typický pro Penny a English aukce; ve Vickrey podáš jediný bid blízko true valuation
- V Dutch čekáš na nízkou cenu, ale riskuješ, že tě někdo předběhne

ROZHODOVÁNÍ:
- Mimo kritickou zónu: vždy WAIT (nech ostatní vyčerpat se)
- V kritické zóně: BID pokud cena < valuation, jinak LEAVE
- Reasoning: kalkulovaný timing, nízká emocionalita
```

### SOCIAL (`prompts/role_social.md`)

```markdown
Jsi společenský a virální účastník. Budíš aktivitu a přivádíš lidi.

CHARAKTERISTIKY:
- Časté malé bidy pro udržení dynamiky (i pod racionální nutností)
- Silně reaguješ na social proof signály ("X lidí přihazuje") — víc lidí = víc biduješ
- Generuješ referraly: tvá aktivita přivádí nové bidery do poolu
- V Penny aukci jsi motor engagementu (a tím i tržby platformy)
- Máš strop, ale FOMO ho posouvá nahoru

ROZHODOVÁNÍ:
- Pokud social proof vysoký → BID i při marginální hodnotě
- Pokud aukce "umírá" (nízká aktivita) → BID pro oživení
- Reasoning: sociální, komunitní motivace
```

> **Numerické knoby:** Každá role nese laditelné parametry předané do promptu přes
> `persona_json` — `risk_tolerance` (0–1), `social_factor` (0–1), `sunk_cost_sensitivity` (0–1).
> Nemění engine, jen modulují system prompt a tím chování. Default hodnoty plynou z role,
> ale jdou přebít při konfiguraci poolu.

---

## 8. LLM integration

### Client

```go
type Client struct {
    BaseURL    string  // https://llm.ol1n.com
    AccessID   string  // CF_ACCESS_CLIENT_ID
    AccessSecret string // CF_ACCESS_CLIENT_SECRET
    DefaultModel string // llm-dev
    HTTPClient *http.Client
}

type DecisionRequest struct {
    Model       string
    SystemPrompt string
    UserPrompt   string
    MaxTokens    int
    Temperature  float64  // 0.8-1.0 pro role-play diversity
    TimeoutMs    int      // 5000 pro decisions
}

type DecisionResponse struct {
    Action     string  `json:"action"`            // BID, WAIT, LEAVE
    Amount     float64 `json:"amount,omitempty"`  // Vickrey / explicit bid
    Reasoning  string  `json:"reasoning"`
    Confidence float64 `json:"confidence"`        // 0.0–1.0, jistota agenta
    LatencyMs  int64
    Raw        string
}

// Akční model:
//   BID   — přihodí (zaplatí fee/převezme leadership dle typu)
//   WAIT  — vyčká, neutrácí, zůstává aktivní (dřívější "PASS")
//   LEAVE — trvale opustí aukci; engine ho přestane oslovovat (mění active_bidders)
// Pozn.: pokud je aktivní viral mechanismus buy_it_now, je povolena i akce BUY_NOW
// (viz sekce 10.5) — engine ji validuje proti aktuální buy-now ceně.
```

### Prompt template strategie

Použij `text/template`. Sestavení:

```
SYSTEM = role_<ROLE>.md
USER = decision_<AUCTION_TYPE>.md (vyplněný runtime daty) + bidder state
```

**Decision template (Penny example, `prompts/decision_penny.md`):**

```markdown
PENNY AUKCE — TVOJE ROZHODNUTÍ

AUKCE:
- Typ: Penny
- Zboží: {{.Item.Name}} (retail: {{.Item.RetailPrice}} Kč)
- Aktuální cena: {{.CurrentPrice}} Kč
- Bid fee: {{.BidFee}} Kč (NEVRACÍ se!)
- Cena se zvedne po bidu o: {{.PriceIncrement}} Kč
- Timer do konce: {{.TimerRemaining}}s
- Aktuální leader: {{.LeaderName}}
- Aktivních bidderů: {{.ActiveBidders}}

SOCIAL PROOF (pokud aktivní mechanismus):
- Aktivně přihazuje: {{.ActiveBidders}} lidí
- Tempo bidů: {{.BidVelocity}} bidů/min
- Tvůj entry credit k dispozici: {{.EntryCreditAvailable}} Kč
{{if .BuyNowEnabled}}- Buy-it-now cena: {{.BuyNowPrice}} Kč{{end}}

TVŮJ STAV:
- Rozpočet zbývá: {{.BudgetRemaining}} Kč
- Tvoje valuation: {{.Valuation}} Kč
- Tvoje bidy v této aukci: {{.YourBids}}
- Tvoje fees dosud: {{.YourFees}} Kč
- Emocionální stav: {{.Emotion}}
- Posledních 5 událostí: {{.RecentHistory}}

ROZHODNI:
1. BID — utratíš {{.BidFee}} Kč na fee, převezmeš leadership, timer reset
2. WAIT — počkáš, ušetříš fee, zůstaneš ve hře, ale můžeš prohrát pokud timer dojde
3. LEAVE — trvale opustíš aukci (už tě engine nebude oslovovat){{if .BuyNowEnabled}}
4. BUY_NOW — koupíš ihned za {{.BuyNowPrice}} Kč{{end}}

ODPOVĚZ POUZE JSON:
{"action": "BID" | "WAIT" | "LEAVE"{{if .BuyNowEnabled}} | "BUY_NOW"{{end}}, "amount": <číslo|null>, "reasoning": "<2-3 věty>", "confidence": 0.0-1.0}
```

### Parsing

Strict JSON. Pokud LLM přidá markdown fences nebo blábol kolem:

```go
func parseDecision(raw string) (*DecisionResponse, error) {
    // 1. strip ```json ... ``` fences
    // 2. find first { ... last }
    // 3. unmarshal
    // 4. validate action enum (BID | WAIT | LEAVE | BUY_NOW)
    // 5. clamp confidence do [0,1]; chybějící → 0.5
    // pokud failure → fallback WAIT s log warning (bezpečné: neutrácí, zůstává ve hře)
}
```

### Latency management

Pro Penny aukci s 10 biddery a 3-5s ticky:

```go
func (p *Pool) RequestDecisions(ctx context.Context, event Event) []Decision {
    ctx, cancel := context.WithTimeout(ctx, 4*time.Second)
    defer cancel()
    
    g, gctx := errgroup.WithContext(ctx)
    decisions := make([]Decision, len(p.bidders))
    
    for i, b := range p.bidders {
        i, b := i, b
        g.Go(func() error {
            d, err := b.Decide(gctx, event)
            if err != nil {
                // timeout nebo error → default WAIT (bezpečné: neutrácí, zůstává ve hře)
                decisions[i] = Decision{BidderID: b.ID, Action: "WAIT", Reasoning: "timeout"}
                return nil
            }
            decisions[i] = d
            return nil
        })
    }
    g.Wait()
    return decisions
}
```

---

## 9. REST API spec

### Endpoints

```
POST   /api/v1/simulations
  body: {
    name: string,
    item: Item,
    bidder_pool: [{role, count, budget_range, valuation_range, persona_overrides}],
    auction_types: ["DUTCH", "VICKREY", "PENNY"],
    auction_params: {dutch: {...}, vickrey: {...}, penny: {...}},
    viral_config: {mechanisms: [...], entry_credit_amount, referral_pct, ...},  // viz sekce 10.5
    runs_per_type: 1
  }
  returns: 201 {simulation_id, status_url}

GET    /api/v1/simulations/{id}
  returns: simulation metadata + status

GET    /api/v1/simulations/{id}/auctions
  returns: list všech auctions v této simulaci

POST   /api/v1/simulations/{id}/start
  returns: 202 Accepted

GET    /api/v1/auctions/{id}
  returns: auction state + recent bids

GET    /api/v1/auctions/{id}/events
  Server-Sent Events stream:
    event: price_updated     data: {price, timer}
    event: bid_placed         data: {bidder_id, amount, reasoning}
    event: leader_changed    data: {bidder_id, name}
    event: auction_ended      data: {winner, final_price}

POST   /api/v1/auctions/{id}/bid           [internal, called by bidder pool]
  body: {bidder_id, amount, reasoning}
  returns: 200 {accepted: bool, reason}

GET    /api/v1/simulations/{id}/report
  returns: full report JSON + html_url

POST   /api/v1/analysis/compare
  body: {item, bidder_pool, auction_params, viral_config, auction_types, runs_per_type}
  spustí konfigurované typy aukcí se stejnými parametry a poolem
  returns: srovnání metrik napříč typy (profit, engagement, viralita) + winner-by-metric

POST   /api/v1/analysis/what-if
  body: {base_config, parameter, range_from, range_to, steps}
  sweep jednoho parametru (např. "entry_credit_amount", "penny.bid_fee", "bidder_pool.gambler.count")
  engine spustí `steps` simulací s lineárně měněnou hodnotou
  returns: {parameter, points: [{value, metrics}]}  → křivky metrik pro grafy

GET    /reports/{simulation_id}.html       [static serve]
  HTML report s embedded grafy
```

### SSE event types

```
auction_started     {auction_id, type, params}
price_updated       {price, timer_remaining}
bid_placed          {bid_id, bidder_id, amount, reasoning}
leader_changed      {previous, current}
bidder_state        {bidder_id, budget, emotion, spent}
auction_ended       {winner_id, final_price, duration_ms}
simulation_progress {completed_auctions, total}
```

---

## 10. Web UI

### Stránky

**`/`** — landing + simulation setup
- Form pro item (název, retail, quantity)
- Bidder pool builder (drag-add rolí s count sliderem)
- Per-type auction params
- "Run simulation" button

**`/simulations/{id}`** — live view
- Top: simulation progress (3/9 auctions complete)
- Aktivní aukce: 
  - Big price display, timer
  - Recent bids feed (last 10) s LLM reasoning expandable
  - Bidder grid: každý jako card (name, role, budget, emotion icon)
- SSE driven, htmx swaps

**`/simulations/{id}/report`** — final report
- Executive summary (LLM-generated paragraph)
- Komparativní tabulka (Dutch vs Vickrey vs Penny)
- Per-auction detail expandables
- Charts (viz sekce 11)
- Doporučení sekce

### Tech volby

- **Server-side rendering** přes Go `html/template`
- **htmx** pro partial updates
- **SSE** pro live stream
- **Chart.js** vanilla, žádný build
- **Pico.css** nebo **Tailwind via CDN** pro rychlý styling

---

## 10.5 Virální mechanismy

Konfigurovatelná vrstva **nad** aukčním enginem — neimplementuje se uvnitř logiky jednotlivého
typu aukce, ale jako hooks, které engine aplikuje při relevantních událostech. Cílem je modelovat,
jak různé "engagement" mechaniky mění chování poolu, tržbu platformy i škody účastníků.

### Go typy

```go
type ViralMechanism string
const (
    ViralEntryCredit        ViralMechanism = "entry_credit"
    ViralReferralBonus      ViralMechanism = "referral_bonus"
    ViralSocialProof        ViralMechanism = "social_proof"
    ViralCountdownExtension ViralMechanism = "countdown_extension"
    ViralBuyItNow           ViralMechanism = "buy_it_now"
    ViralStreakBonus        ViralMechanism = "streak_bonus"
)

type ViralConfig struct {
    Mechanisms          []ViralMechanism `json:"mechanisms"`
    EntryCreditAmount   float64          `json:"entry_credit_amount,omitempty"`   // Kč po prvním bidu
    ReferralPct         float64          `json:"referral_pct,omitempty"`          // % z final price referrerovi
    SocialProofThreshold int             `json:"social_proof_threshold,omitempty"`// od kolika aktivních se zobrazuje
    CountdownExtendS    time.Duration    `json:"countdown_extend_s,omitempty"`    // +Xs při late bidu
    CountdownTriggerS   time.Duration    `json:"countdown_trigger_s,omitempty"`   // "late" = pod tento práh
    BuyNowMultiplier    float64          `json:"buy_now_multiplier,omitempty"`    // násobek retail (default 1.3)
    StreakDiscountPct   float64          `json:"streak_discount_pct,omitempty"`   // sleva za streak
    StreakThreshold     int              `json:"streak_threshold,omitempty"`      // od kolika aukcí po sobě
}
```

### Mechanismy a jak je engine aplikuje

| Mechanismus | Trigger | Efekt | Dopad na metriky |
|---|---|---|---|
| `entry_credit` | první bid biddera v aukci | připíše kredit (`entry_credit_received`), sníží efektivní náklad vstupu | ↑ conversion, ↑ viral_cost |
| `referral_bonus` | výhra biddera přivedeného referrerem | referrerovi `referral_pct` z final price (`referral_earnings`) | ↑ new_bidders_attracted, ↑ viral_cost |
| `social_proof` | každý tick, pokud `active_bidders ≥ threshold` | vystaví `active_bidders` + `bid_velocity` do LLM kontextu | ↑ engagement (zejm. SOCIAL/GAMBLER role) |
| `countdown_extension` | bid v posledních `CountdownTriggerS` | timer +`CountdownExtendS` (anti-sniping) — interaguje s Penny timer reset a English close | ↑ total_bids, ↑ duration, ↓ účinnost SNIPER |
| `buy_it_now` | akce `BUY_NOW` agenta | okamžitá koupě za `retail × BuyNowMultiplier`, ukončí aukci | úniková cesta pro CAUTIOUS/COLLECTOR |
| `streak_bonus` | bidder s ≥`StreakThreshold` aukcí po sobě (multi-auction/batch) | sleva `StreakDiscountPct` z ceny | ↑ retence napříč běhy |

### Persistence
- `simulations.viral_config_json` — konfigurace aktivních mechanismů (viz sekce 5).
- `bidder_states.entry_credit_received`, `bidder_states.referral_earnings` — per-bidder tracking.
- Reporter agreguje **`viral_cost`** = součet kreditů + referral payoutů (náklad platformy na engagement).

### LLM kontext
Decision šablony (sekce 8) dostávají navíc: `ActiveBidders`, `BidVelocity`, `EntryCreditAvailable`,
`BuyNowEnabled`/`BuyNowPrice`. Tyto signály vidí agent jen pokud je příslušný mechanismus aktivní —
jinak se do promptu nevkládají (čistý prompt, žádné matoucí nuly).

---

## 11. Reporting

### Vypočítané metriky

```go
type AuctionReport struct {
    AuctionID         string
    Type              AuctionType
    
    PlatformProfit    float64
    SellerProfit      float64
    WinnerSurplus     float64    // valuation - paid
    TotalLoserLosses  float64
    
    EfficiencyScore   float64    // vyhrál bidder s nejvyšší valuation? (0 or 1)
    
    Duration          time.Duration
    TotalBidsPlaced   int
    UniqueBidders     int

    // Engagement / viralita
    UniqueActiveBidders  int        // kdo položil ≥1 bid
    Conversion           float64    // UniqueActiveBidders / total v poolu
    NewBiddersAttracted  int        // přivedení přes referral/social
    ViralCoefficient     float64    // průměr nových bidderů na 1 aktivního
    ViralCost            float64    // entry credits + referral payouts
    BidVelocity          []float64  // bidů per časový bin (engagement v čase)

    PerBidder         []BidderReport
}

type BidderReport struct {
    BidderID     string
    Role         Role
    BidsPlaced   int
    FeesPaid     float64
    AmountPaid   float64  // pokud vyhrál
    NetOutcome   float64  // valuation - cost pokud vyhrál, -fees pokud prohrál
    BudgetUsed   float64  // %
    EndEmotion   Emotion
}

type SimulationReport struct {
    Simulation        Simulation
    AuctionReports    []AuctionReport

    // Batch agregace napříč runs_per_type (mean/median/variance/stddev per metrika)
    Aggregates        map[AuctionType]AggregateReport

    CrossTypeMetrics  CrossTypeComparison
    RoleAnalysis      []RoleOutcomeByType
    
    LLMInsights       string  // generováno llm-lab
    Recommendations   []Recommendation
}

// AggregateReport drží statistiku jedné metriky napříč N běhy téhož typu.
type MetricStats struct {
    Mean, Median, Variance, StdDev, Min, Max float64
}
type AggregateReport struct {
    Runs            int
    FinalPrice      MetricStats
    PlatformProfit  MetricStats
    SellerProfit    MetricStats
    TotalBids       MetricStats
    Conversion      MetricStats
    ViralCost       MetricStats
    // ...per relevantní metrika
}
```

> **Batch běhy:** `runs_per_type` spouští N opakování téhož uspořádání; reporter agreguje
> mean/median/variance/stddev. Grafy zobrazují průměr s error bars (případně box ploty).
> Doporučení LLM (llm-lab) **váží viralitu/engagement vs. cenu i škody účastníků**, nikoli jen profit.

> **Tick-pacing (`TICK_PACING`):** Engine zůstává real-time, ale tempo ticku je konfigurovatelné:
> - `live` — wall-clock intervaly (Dutch tick 3–5s, Penny timer v sekundách) pro sledovaný běh ve frontendu.
> - `fast` — žádný umělý `sleep`, ticky běží hned po sobě (omezeno jen LLM latencí) pro headless batch
>   a `what-if` sweepy. Logika i pořadí událostí identické, mění se jen časování — sjednocuje "real-time"
>   UX s rozumnou dobou dokončení mnoha běhů. (Není to deterministický replay, jen vypnutí sleep.)

### Charty (Chart.js)

1. **Price evolution timeline** — per auction, line chart s anotacemi bidů
2. **Platform profit comparison** — bar chart Dutch/Vickrey/Penny
3. **Buyer outcomes distribution** — bar chart počtu winners/losers per type
4. **Role heatmap** — průměrný net outcome × (role × auction_type), color-coded
5. **Bid frequency over time** (Penny) — line chart počet bidů per 10s bin
6. **Emotion evolution** — stacked area chart per bidder pool průběhem aukce
7. **Money flow Sankey** — odkud kam tečou peníze (bidders → platform/seller)
8. **Engagement timeline** — bid velocity (bidů/bin) v čase, anotace virálních triggerů
9. **Viral cost vs profit** — scatter: viral_cost × platform_profit per uspořádání
10. **What-if sweep** — line chart metrik vs. hodnota proměného parametru (z `/analysis/what-if`)
11. **Batch distribution** — box plot klíčových metrik napříč `runs_per_type` (variance výsledků)

### LLM-generated insights

V závěru reportu zavolat `llm-lab`:

```
Input: kompletní agregovaný JSON s metriky všech aukcí
System: "Jsi expert na aukční teorii a behaviorální ekonomii. Analyzuj 
        výsledky simulace a vytvoř insight report v češtině."
Output: 
  - Executive summary (200 slov)
  - 3-5 key findings s podporou v datech
  - Doporučení pro: prodávajícího / kupujícího / regulátora
  - Etické zhodnocení Penny aukce na základě dat
```

---

## 12. Konfigurace

### `.env.example`

```bash
# Server
ENGINE_PORT=8080
BIDPOOL_PORT=8081

# Database
DB_PATH=./data/auction.db

# LLM
LLM_BASE_URL=https://llm.ol1n.com
LLM_DECISION_MODEL=llm-dev
LLM_INSIGHT_MODEL=llm-lab
CF_ACCESS_CLIENT_ID=xxx
CF_ACCESS_CLIENT_SECRET=xxx
LLM_DECISION_TIMEOUT_MS=5000
LLM_INSIGHT_TIMEOUT_MS=60000

# Simulation defaults
DEFAULT_DUTCH_TICK_MS=3000
DEFAULT_PENNY_TIMER_S=10
DEFAULT_PENNY_BID_FEE=5.00
DEFAULT_VICKREY_WINDOW_S=60
DEFAULT_PLATFORM_FEE_PCT=5.0
DEFAULT_RUNS_PER_TYPE=1
TICK_PACING=live                # live | fast (fast pro headless batch / what-if)

# Viral defaults
DEFAULT_ENTRY_CREDIT=5.00
DEFAULT_REFERRAL_PCT=10.0
DEFAULT_SOCIAL_PROOF_THRESHOLD=3
DEFAULT_COUNTDOWN_EXT_S=15
DEFAULT_COUNTDOWN_TRIGGER_S=30
DEFAULT_BUYNOW_MULT=1.3
DEFAULT_STREAK_DISCOUNT_PCT=5.0
DEFAULT_STREAK_THRESHOLD=3

# Reports
REPORTS_DIR=./reports
```

### Makefile cíle

```makefile
.PHONY: build run-engine run-bidpool run-all migrate clean test

build:
	go build -o bin/engine ./cmd/engine
	go build -o bin/bidpool ./cmd/bidpool
	go build -o bin/reporter ./cmd/reporter

migrate:
	go run ./cmd/engine migrate

run-engine: build
	./bin/engine

run-bidpool: build
	./bin/bidpool

run-all:
	# tmux nebo overmind/foreman
	tmux new-session -d -s auction './bin/engine' \; \
	    split-window './bin/bidpool'

test:
	go test ./...

clean:
	rm -rf bin/ data/*.db reports/*.html
```

---

## 13. Implementační roadmap

### Fáze 1 — Foundation (Den 1)
- [ ] `go mod init github.com/ol1n/auction-sim`
- [ ] Project layout + základní packages
- [ ] SQLite schema + migrations (`modernc.org/sqlite`)
- [ ] Config loading (env)
- [ ] Basic logging (slog)
- [ ] Datové typy + JSON serialization

### Fáze 2 — Vickrey end-to-end (Den 2)
- [ ] Vickrey engine (nejjednodušší, žádný timing problem)
- [ ] LLM client + decision parsing
- [ ] Bidder agent (jeden, hardcoded)
- [ ] Submit bid endpoint
- [ ] Test: ručně spustit aukci s 3 dummy biddery
- [ ] Test: nahradit LLM dummy biddery skutečnými LLM agenty

### Fáze 3 — Bidder pool s rolemi (Den 3)
- [ ] Role definitions + system prompts loading (7 rolí vč. SNIPER, SOCIAL)
- [ ] Akční model BID/WAIT/LEAVE + confidence; LEAVE deaktivuje biddera
- [ ] Pool manager s configurable mix rolí + numerické knoby (persona_overrides)
- [ ] Emotion state machine
- [ ] Parallel LLM calls s errgroup (fallback WAIT)
- [ ] Test: Vickrey s 10 biddery napříč rolemi

### Fáze 4 — Dutch (Den 4)
- [ ] Dutch engine (tick loop, price decrement)
- [ ] Race-safe bid handling (FIFO mutex queue)
- [ ] Event bus + SSE
- [ ] LLM context pro Dutch
- [ ] Test: full Dutch run

### Fáze 5 — Penny (Den 5)
- [ ] Penny engine (timer, bid fees, leader tracking)
- [ ] Fee deduction logika
- [ ] Concurrency handling (jen první bid v tiku se počítá)
- [ ] LLM context pro Penny s emphasis na sunk cost
- [ ] Test: full Penny run s 10 biddery, expect ~50+ bids

### Fáze 5b — Virální vrstva (Den 5–6)
- [ ] ViralConfig + ViralMechanism typy, parsing z `viral_config`
- [ ] Engine hooks: entry_credit, countdown_extension, buy_it_now (BUY_NOW akce)
- [ ] Cross-auction tracking: referral_bonus, streak_bonus
- [ ] Social proof signály do LLM kontextu (active_bidders, bid_velocity)
- [ ] Test: Penny s/bez virálních mechanismů — porovnat engagement & viral_cost

### Fáze 6 — Web UI (Den 6)
- [ ] Layout template + Pico.css
- [ ] Simulation setup form
- [ ] Live view s SSE + htmx swaps
- [ ] Bidder cards s real-time updates
- [ ] Recent bids feed

### Fáze 7 — Reporter + analýza (Den 7)
- [ ] Metric calculations per auction (vč. engagement/viral metrik)
- [ ] Batch agregace (mean/median/variance/stddev) přes runs_per_type
- [ ] Cross-type comparison + `/analysis/compare`
- [ ] `/analysis/what-if` sweep (TICK_PACING=fast) + sweep charty
- [ ] Chart.js integration (engagement timeline, viral cost vs profit, box ploty)
- [ ] HTML report template
- [ ] LLM insight generation (llm-lab call) — váží viralitu vs. cenu i škody

### Fáze 8 — Polish (Den 8)
- [ ] Error handling + retry logic
- [ ] Logging + observability
- [ ] README s screenshoty
- [ ] Demo runs + screenshots pro portfolio
- [ ] Volitelné: Cloudflare Tunnel exposure pod `auction.ol1n.com`

---

## 14. Token budget odhad

Per demo run (1 simulace × 3 typy × 1 run):

| Typ | Decisions | Tokens/decision | Total |
|---|---|---|---|
| Vickrey | 10 (jednou) | 500 in + 100 out | 6k |
| Dutch | 10 × ~20 ticks | 500 in + 100 out | 120k |
| Penny | 10 × ~80 ticks | 600 in + 100 out | 560k |
| Insight (llm-lab) | 1 velký | 5000 in + 2000 out | 7k |
| **Total** | | | **~700k tokens** |

Na lokálním `llm-dev` (Gemma 4 31B) na DGX Spark: bez problému, vejde se do single session.

> **Škálování s batch:** odhad ~700k je na **1 běh/typ**. Při `runs_per_type=N` škáluje lineárně
> (10 běhů ≈ 7M tokenů); `what-if` sweep přidává `steps × per-sim`. Na lokálním llm-dev je to OK,
> ale pro velké sweepy zvaž `TICK_PACING=fast` (kratší wall-clock, ne méně tokenů) a menší pool.

---

## 15. Rizika a mitigace

| Riziko | Mitigace |
|---|---|
| LLM latence > tick interval (Dutch) | Tick 3-5s; pokud LLM nestihne, default WAIT |
| LLM vrátí invalid JSON | Fallback parser + WAIT default + log warning |
| Race condition při Dutch bidu | Single mutex + FIFO queue per aukce |
| Penny aukce běží příliš dlouho | Hard cap max duration (např. 30 min) + emergency stop |
| LLM "leaks" reasoning v reálné aukci by viděl ostatní | Reasoning ukládat, ale NEPOSÍLAT do promptů jiných bidderů |
| Bidder utratí víc než budget | Engine validuje před acceptem bidu |
| SQLite write contention | Write přes single goroutine + channel, ne přímé concurrent writes |
| CF Access token expirace | Periodický health check + alert |
| Příliš homogenní LLM odpovědi (low temperature) | Temperature 0.8-1.0 pro variability |

---

## 16. Stretch features (po v1)

- **Multi-round simulations:** stejní biddery procházejí 5 aukcí, sledování learning behavior
- **Adversarial roles:** shill bidder, sniper bidder (auction last-second)
- **Více zboží najednou:** combinatorial auction
- **English auction** (ascending) jako čtvrtý typ
- **Replay viewer:** přehrávání minulých aukcí
- **A/B testing setupů:** randomized matchups
- **Export dat:** CSV/Parquet pro vlastní analýzu v Pythonu
- **Fine-tuning dataset:** sebraná `(state, decision, reasoning)` data jako trénovací set pro custom bidder model

---

## 17. Klíčové reference & inspirace

- Vickrey, W. (1961). "Counterspeculation, Auctions, and Competitive Sealed Tenders"
- Klemperer, P. (2004). "Auctions: Theory and Practice"
- Krishna, V. (2009). "Auction Theory" — pro matematické základy
- Augenblick, N. (2016). "The Sunk-Cost Fallacy in Penny Auctions" — empirie
- Hinnosaar, T. (2016). "Penny Auctions" — game-theoretic model

---

## 18. Co je hotovo před spuštěním Claude Code

Před `claude-code` v repu měj:
1. `git init` + první commit s tímto plánem
2. `.env` z `.env.example` s tvými LiteLLM credentials
3. Ověřit `curl https://llm.ol1n.com/v1/models` s CF Access headers funguje
4. `go version` ≥ 1.22

První prompt do Claude Code:
> "Načti `auction-sim-plan.md` a začni Fází 1. Postupuj po jednotlivých task items, commituj po každé funkční jednotce. Implementace v Go, idiomatic, error handling, žádné panics."

