# AGENTS.MD - Plán implementace Auction Simulation System

## 1. Celkový cíl projektu
Vytvořit robustní simulátor aukcí, který umožňuje:
- Testovat různé typy aukcí (Holandská, Vickrey, Penny, Anglická)
- Simulovat realistické chování bidderů pomocí LLM agentů
- Generovat detailní analýzy, grafy a srovnání
- Sloužit jako rozhodovací nástroj pro návrh aukční platformy

---

## 2. Architektura systému (High-level)

### Hlavní komponenty
- **Simulation Engine** (`auction_simulator/`)
- **Bidder Agents** (LLM-powered role-based agents)
- **API Layer** (FastAPI)
- **Visualization & Dashboard** (Streamlit nebo React)
- **Data Layer** (SQLite/PostgreSQL)
- **Analysis Module**

### Technologický stack
- Python 3.11+
- FastAPI + Uvicorn
- SQLAlchemy + Alembic
- Plotly + Pandas
- LangChain / LlamaIndex nebo přímé API calls k LLM
- Streamlit (pro rychlý MVP dashboard)

---

## 3. Bidder Agents - Role a chování

### Definované role bidderů
1. **Gambler** — vysoká risk tolerance, silný sunk cost + FOMO
2. **Value Bidder (Opatrný)** — racionální, biduje do private_value
3. **Collector / Zainteresovaný** — velmi vysoká willingness-to-pay
4. **Sniper** — čeká na poslední momenty
5. **Casual / FOMO Responder** — reaguje na sociální signály a notifikace
6. **Budget Bidder** — omezený rozpočet, opatrný

Každý agent má:
- `private_value` (skrytá hodnota položky)
- `budget`
- `risk_tolerance`
- `emotional_profile` (sunk_cost_sensitivity, fomo_level, loss_aversion)

---

## 4. Implementační fáze (Priority order)

### Fáze 1: Core Simulation Engine (Aktuálně prioritní)
- [ ] Vytvořit `Auction` base class
- [ ] Implementovat konkrétní aukce:
  - `DutchAuction`
  - `VickreyAuction`
  - `PennyAuction`
  - `EnglishAuction`
- [ ] Event-driven tick system (asyncio)
- [ ] Logging všech událostí (`AuctionHistory`)

### Fáze 2: LLM Bidder Agents
- [ ] Vytvořit `BidderAgent` class
- [ ] Prompt templates pro každou roli + typ aukce (uložit do `prompts/`)
- [ ] `decide_action()` metoda volající LLM
- [ ] Caching rozhodnutí (pro rychlost)
- [ ] Emotional state tracking (sunk_cost, frustration atd.)

### Fáze 3: API & Persistence
- [ ] FastAPI endpoints (`/auctions/`, `/bidders/`, `/simulate/`)
- [ ] Database models (Auction, Bidder, Bid, SimulationRun)
- [ ] Background tasks pro dlouhé simulace

### Fáze 4: Visualization & Analysis
- [ ] Streamlit dashboard
  - Nastavení simulace
  - Live view (pro pomalé simulace)
  - Results page s grafy
- [ ] Srovnávací reporty (multi-simulation)
- [ ] Export do PDF/CSV

### Fáze 5: Pokročilé funkce
- [ ] Multi-item auctions
- [ ] Gamification prvky (leaderboard, notifications)
- [ ] Sensitivity analysis (měnění parametrů)
- [ ] WebSocket real-time updates
- [ ] User authentication (pro více uživatelů)

---

## 5. Struktura projektu

```
auction-simulator/
├── main.py
├── AGENTS.MD                  # tento soubor
├── config/
├── core/
│   ├── auction.py
│   ├── bidders/
│   │   ├── agent.py
│   │   └── roles.py
│   └── types/
├── api/
├── simulation/
├── analysis/
├── prompts/                    # LLM prompt šablony
├── visualizations/
├── data/
├── tests/
└── requirements.txt
```

---

## 6. Další kroky (Next Actions)

1. **Dnes / zítra**: Implementovat base `Auction` class + PennyAuction (jako nejkomplexnější)
2. Vytvořit 3 základní role bidderů s LLM prompty
3. Spustit první end-to-end simulaci (50 bidderů, iPhone)
4. Vytvořit jednoduchý Streamlit dashboard pro spouštění simulací
5. Udělat srovnání Penny vs Vickrey na stejných datech

---

## 7. Poznámky & Rizika
- LLM volání mohou být pomalé → batching + caching
- Realismu vs. rychlost simulace (trade-off)
- Etické aspekty (zvláště u Penny aukcí)
- Potřeba seedování pro reprodukovatelnost simulací

**Aktuální status:** Plán schválen. Začínáme Fází 1.

---
Vytvořeno: 27. 5. 2026
Verze: 1.0
