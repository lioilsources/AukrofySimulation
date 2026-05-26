# AuctionSim Platform - AGENTS.md

> Dokumentace pro AI agenty a vývojářský tým  
> Verze: 1.0.0 | Poslední aktualizace: 2026-05-27

---

## 1. Přehled projektu

**AuctionSim** je simulační a analytická platforma pro návrh, testování a optimalizaci aukčních mechanismů. Platforma řeší konkrétní problém: *málo lidí nesází, cena neroste, aukce není virální*.

### Klíčové komponenty
- **Simulační engine** (Python) – 4 typy aukcí, 5 rolí bidderů, viralitní mechanismy
- **REST API** (FastAPI) – endpointy pro simulace, LLM integraci, analýzy
- **Web dashboard** – vizualizace výsledků, konfigurátor, AI doporučení
- **LLM integrace** – volání externího LLM pro realistické chování bidderů

---

## 2. Architektura systému

```
┌─────────────────────────────────────────────────────────────┐
│                    AUCTIONSIM PLATFORM                       │
├─────────────────────────────────────────────────────────────┤
│  WEB DASHBOARD (React/Vue)                                   │
│  ├── Live Auction View (WebSocket real-time)                │
│  ├── Simulation Configurator (drag-drop parametry)          │
│  └── Results Analyzer (AI doporučení, export PDF/CSV)     │
├─────────────────────────────────────────────────────────────┤
│  REST API (FastAPI)                                          │
│  ├── POST /api/v1/simulations       → Spuštění simulace     │
│  ├── GET  /api/v1/simulations/{id}  → Výsledky            │
│  ├── POST /api/v1/llm/decide        → LLM rozhodnutí      │
│  ├── POST /api/v1/analysis/compare  → Srovnání typů       │
│  └── POST /api/v1/analysis/what-if  → What-if analýza     │
├─────────────────────────────────────────────────────────────┤
│  SIMULATION ENGINE (Python)                                │
│  ├── Auctioneer (state machine)                           │
│  ├── BidderAgent                                          │
│  │   ├── Rule-based (benchmark)                           │
│  │   └── LLM-based (externí API volání)                   │
│  └── MarketAnalyzer (real-time metriky)                   │
├─────────────────────────────────────────────────────────────┤
│  LLM INTEGRATION                                             │
│  ├── Prompt: system + context + task                      │
│  ├── Roles: gambler, opatrny, zainteresovany, sniper      │
│  └── Output: {action, amount, reasoning, confidence}    │
├─────────────────────────────────────────────────────────────┤
│  DATA (PostgreSQL + Redis)                                  │
│  ├── Relational: aukce, bidderi, transakce              │
│  └── Cache: real-time stavy, leaderboards                │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Simulační engine

### 3.1 Typy aukcí

| Typ | Princip | Kdy použít | Viralita |
|-----|---------|------------|----------|
| **Anglická** | Cena nahoru, vítěz platí svou nabídku | Sběratelské předměty, transparentní soutěž | ⭐⭐⭐ |
| **Vickrey** | Tajné nabídky, vítěz platí druhou nejvyšší | B2B, nemovitosti, frekvence | ⭐⭐ |
| **Penny** | Pay-to-bid, cena +1 za přihození | Maximální engagement, gamifikace | ⭐⭐⭐⭐⭐ |
| **Holandská** | Cena klesá, první bere | Rychlý prodej, likvidace zásob | ⭐ |

### 3.2 Role bidderů

| Role | Charakteristika | Strategie |
|------|-----------------|-----------|
| **Gambler** | Riskuje, agresivní, emoce | Přihazuje nad valuaci, FOMO efekt |
| **Opatrný** | Čeká, analyzuje, strach z přeplacení | Nízké nabídky, čeká na konec |
| **Zainteresovaný** | Zná hodnotu, strategický | Truthful bidding, blízko true value |
| **Sniper** | Čeká na poslední vteřinu | Bid až v posledních 10s |
| **Sociální** | Přivádí lidi, viralita | Časté malé bidy, referral aktivita |

### 3.3 Viralitní mechanismy

| Mechanismus | Princip | Dopad | Implementace |
|-------------|---------|-------|--------------|
| **Entry Credit** | $5 za první přihození | Snižuje bariéru vstupu o 60% | Automatický kredit po prvním bidu |
| **Referral Bonus** | 10% z ceny pro referrer | Virální koeficient 1.3-1.8 | Tracking kód, payout po výhře |
| **Social Proof** | "X lidí přihazuje" | FOMO, +25-40% conversion | Live counter, heatmap |
| **Countdown Extension** | +15s při bidu v posledních 30s | Eliminuje sniping, udržuje napětí | Timer reset na backendu |
| **Buy It Now** | Koupě ihned za 130% true value | Jistota pro opatrné | Tlačítko v polovině aukce |
| **Streak Bonus** | Sleva za 3+ aukce za sebou | Retence, závislost | Uživatelský profil, historie |

---

## 4. REST API Reference

### 4.1 Simulace

#### `POST /api/v1/simulations`
Spustí novou simulaci aukce.

**Request:**
```json
{
  "config": {
    "auction_type": "anglicka",
    "product": {
      "name": "iPhone 15 Pro",
      "true_value": 1000,
      "starting_price": 500,
      "quantity": 1
    },
    "duration_seconds": 300,
    "tick_seconds": 5,
    "platform_fee_pct": 0.10
  },
  "bidders": [
    {
      "id": "G1",
      "role": "gambler",
      "budget": 1200,
      "valuation": 1100,
      "risk_tolerance": 0.8,
      "social_factor": 0.2
    }
  ],
  "viral_config": {
    "mechanisms": ["entry_credit", "social_proof", "countdown_ext"],
    "entry_credit_amount": 5.0,
    "social_proof_threshold": 3
  },
  "runs": 10
}
```

**Response:**
```json
{
  "simulation_id": "sim_123_4567",
  "status": "completed",
  "runs_completed": 10,
  "summary": {
    "avg_final_price": 1124.50,
    "avg_platform_profit": 108.20,
    "avg_seller_profit": 1016.30,
    "total_bids": 142,
    "new_bidders_attracted": 3,
    "viral_costs": 35.0
  },
  "recommendation": "Anglická aukce s entry_credit a countdown_extension vykazuje nejlepší poměr virality/cena."
}
```

### 4.2 LLM Rozhodování

#### `POST /api/v1/llm/decide`
Získá rozhodnutí od LLM pro konkrétního biddera.

**Request:**
```json
{
  "bidder_id": "G1",
  "role": "gambler",
  "auction_state": {
    "current_price": 750,
    "time_left": 120,
    "active_bidders": 5,
    "auction_type": "anglicka"
  },
  "bidder_state": {
    "budget": 1200,
    "valuation": 1100,
    "total_spent": 0,
    "bids_count": 0,
    "risk_tolerance": 0.8
  }
}
```

**Response:**
```json
{
  "bidder_id": "G1",
  "action": "bid",
  "amount": 820,
  "reasoning": "Jako gambler vidím příležitost – cena je stále pod mou valuací, zbývá dost času a konkurence je mírná. Přihazuji agresivně pro vyvolání FOMO.",
  "confidence": 0.85,
  "latency_ms": 150
}
```

### 4.3 Analýzy

#### `POST /api/v1/analysis/compare`
Srovná všechny 4 typy aukcí se stejnými parametry.

#### `POST /api/v1/analysis/what-if`
Testuje dopad změny jednoho parametru na metriky.

**Request:**
```json
{
  "base_config": { /* ... */ },
  "base_bidders": [ /* ... */ ],
  "parameter": "entry_credit",
  "range_from": 0,
  "range_to": 25,
  "steps": 10
}
```

---

## 5. LLM Prompt Engineering

### 5.1 System Prompt Template

```
Jsi {role} bidder v {auction_type} aukci.

TVA PROFIL:
- Budget: ${budget}
- Valuace produktu: ${valuation} (kolik je ti produkt opravdu hodnotny)
- Tolerance rizika: {risk_tolerance} (0=velmi opatrny, 1=agresivni hazarder)
- Socialni faktor: {social_factor} (0=samotar, 1=viralni influencer)

PRAVIDLA AUKCE ({auction_type}):
- Anglicka: Cena roste nahoru, posledni nabidka vyhrava, platis svou cenu
- Vickrey: Tajna nabidka, vyhravas pokud mas nejvyssi, platis druhou nejvyssi
- Penny: Kazde prihozeni stoji ${penny_cost}, cena roste o $1, posledni bid vyhrava
- Holandska: Cena klesa od shora, prvni kdo prijme, kupuje za aktualni cenu

AKTUALNI STAV:
- Cena: ${current_price}
- Zbyvajici cas: {time_left}s
- Pocet aktivnich bidderu: {active_count}
- Tve predchozi nabidky: {history}
- Tve celkove naklady: ${total_spent}

TVUJ UKOL:
Rozhodni se: PRIHOD, POCKEJ, nebo OPUSŤ aukci.
Zduvodni sve rozhodnuti z pohledu sve role ({role}).

OUTPUT FORMAT (JSON):
{
  "action": "bid" | "wait" | "leave",
  "amount": number | null,
  "reasoning": "string (max 200 chars)",
  "confidence": 0.0-1.0
}
```

### 5.2 Role-specific instrukce

**Gambler:**
```
Jsi EMOCIONALNI a IMPULZIVNI. Riskujes, casto preplacis.
- Prihazujes, kdyz citis FOMO nebo adrenalin
- Ignorujes rozum, naslouchas pocitum
- "Ta cena poroste, musim to mit!"
```

**Opatrný:**
```
Jsi ANALYTICKY a OPATRNY. Strach z preplaceni te brzdi.
- Cekas na posledni chvili
- Nabizis pod svou valuaci
- "Co kdyz to neni tak hodnotne?"
```

**Zainteresovaný:**
```
Jsi ZNALEC a STRATEG. Vis presnou hodnotu produktu.
- Nabizis blizko true value
- Nepredplacis, nepodplacis
- "Tento produkt ma hodnotu $X, neplatim vic."
```

**Sniper:**
```
Jsi TRPELIVY a PRESNY. Cekas na dokonalou chvili.
- Bidujes az v poslednich 10 sekundach
- Chces prekvapit konkurenci
- "Pockam, az ostatni odpadnou..."
```

**Sociální:**
```
Jsi SPOLECENSKY a VIRALNI. Přivadis lidi, budujes komunitu.
- Casto prihazujes pro aktivitu
- Sdilis aukci, ziskavas referraly
- "Podivejte se vsichni, tohle je super!"
```

---

## 6. Implementační roadmap

### Fáze 1: MVP (2 týdny)
- [ ] Rule-based simulace 4 typů aukcí
- [ ] REST API (FastAPI) pro spuštění a výsledky
- [ ] Základní dashboard s grafy (Plotly/Chart.js)
- [ ] Export CSV/PDF reportů

### Fáze 2: LLM integrace (2 týdny)
- [ ] Prompt engineering pro 5 rolí
- [ ] A/B test: LLM vs rule-based strategie
- [ ] Fine-tuning na historická data
- [ ] Caching LLM odpovědí (Redis)

### Fáze 3: Viralita (1 týden)
- [ ] Referral systém + tracking
- [ ] Entry bonusy + kreditní systém
- [ ] Social proof widgety (live counters)
- [ ] Countdown dramatizace

### Fáze 4: Optimalizace (ongoing)
- [ ] Auto-tuning parametrů (Bayesian optimization)
- [ ] Predikce nejlepšího typu pro produkt
- [ ] Anti-sniping mechanismy
- [ ] Regulace penny aukcí (compliance)

---

## 7. Konfigurace a prostředí

### 7.1 Požadavky
```bash
# Python 3.10+
pip install fastapi uvicorn numpy pandas matplotlib pydantic

# Pro LLM integraci
pip install openai anthropic

# Pro databázi
pip install sqlalchemy asyncpg redis
```

### 7.2 Environment variables
```bash
# API
AUCTIONSIM_HOST=0.0.0.0
AUCTIONSIM_PORT=8000
AUCTIONSIM_DEBUG=false

# LLM
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-...
LLM_MODEL=gpt-4-turbo
LLM_TIMEOUT=2000

# Database
DATABASE_URL=postgresql://user:pass@localhost/auctionsim
REDIS_URL=redis://localhost:6379/0

# Simulace
DEFAULT_RUNS=10
MAX_BIDDERS=50
```

---

## 8. Testování

### 8.1 Unit testy
```python
# test_auctioneer.py
def test_english_auction_winner():
    config = AuctionConfig(auction_type=AuctionType.ENGLISH, ...)
    bidders = create_test_pool()
    sim = AuctionSimulator(config, bidders)
    result = sim.run()
    assert result.winner is not None
    assert result.final_price >= config.product.starting_price
```

### 8.2 Integrační testy
```python
# test_api.py
def test_simulation_endpoint(client):
    response = client.post("/api/v1/simulations", json={...})
    assert response.status_code == 200
    assert "simulation_id" in response.json()
```

### 8.3 A/B testování LLM
```python
# Porovnání rule-based vs LLM strategií
def ab_test_strategies(n_runs=100):
    rule_results = [run_simulation(rule_based=True) for _ in range(n_runs)]
    llm_results = [run_simulation(rule_based=False) for _ in range(n_runs)]

    return {
        "rule_avg_price": np.mean([r.final_price for r in rule_results]),
        "llm_avg_price": np.mean([r.final_price for r in llm_results]),
        "rule_avg_bids": np.mean([r.total_bids for r in rule_results]),
        "llm_avg_bids": np.mean([r.total_bids for r in llm_results])
    }
```

---

## 9. Bezpečnost a compliance

### 9.1 Penny aukce – regulace
- **Max 30 bidů/osoba** – proti závislosti
- **Refundace 50%** neúspěšných bidů jako kredit
- **Transparentní kalkulátor** – zobrazit "Celkové náklady = cena + bid fees"
- **Buy-it-now možnost** – úniková cesta
- **Age verification** – 18+ pro pay-to-bid

### 9.2 Anti-manipulace
- **Rate limiting** – max 1 bid/sec na uživatele
- **Bot detection** – CAPTCHA, behavior analysis
- **Cooldown period** – 24h mezi aukcemi pro nové účty
- **Audit log** – všechny transakce logovány

---

## 10. Troubleshooting

### Problém: Aukce nemá žádné nabídky
**Diagnostika:**
- Kontrola starting_price vs true_value (moc vysoká start?)
- Kontrola bidder poolu (málo účastníků?)
- Kontrola viralitních mechanismů (jsou aktivní?)

**Řešení:**
- Snížit starting_price na 40-50% true_value
- Přidat entry credit ($3-5)
- Aktivovat social proof

### Problém: Cena neroste nad true_value
**Diagnostika:**
- Kontrola bidder valuací (jsou příliš nízké?)
- Kontrola role rozložení (moc opatrných?)

**Řešení:**
- Přidat gamblery do poolu
- Aktivovat countdown extension (vytváří napětí)
- Snížit tick_seconds (rychlejší dynamika)

### Problém: LLM odpovědi jsou pomalé (>2s)
**Řešení:**
- Implementovat Redis caching
- Použít menší model (GPT-3.5-turbo)
- Batch processing více bidderů najednou
- Fallback na rule-based při timeoutu

---

## 11. Kontakty a zdroje

- **Repozitář:** https://github.com/auctionsim/platform
- **Dokumentace:** https://docs.auctionsim.io
- **API Playground:** https://api.auctionsim.io/docs
- **Slack:** #auctionsim-dev

---

*Generováno automaticky z Jupyter notebooku AuctionSim v1.0*
