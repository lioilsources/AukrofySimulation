Zde je kompletní obsah souboru `AGENTS.MD`. Můžeš ho zkopírovat a uložit jako soubor s tímto názvem.

```markdown
# AGENTS.MD – Auction Simulation Platform

This file serves as the master plan for building a simulation platform that runs different auction types (Dutch, Vickrey, Penny/bidding-fee) with LLM-powered bidders playing distinct roles.

## 1. Project Overview
We are building a web application that allows users to configure, run, and compare auction simulations. The core innovation is the use of external LLMs to mimic human-like bidding behavior. The platform must support:
- Three auction types: Dutch, Vickrey, Penny
- Pluggable bidder roles (gambler, cautious, interested, sniper, random) driven by system prompts
- Real-time monitoring via WebSockets
- REST API for configuration and execution
- A frontend dashboard with live auction view, results summary, and comparative analytics

## 2. Technology Stack
- **Backend:** Python 3.11+, FastAPI, asyncio
- **Database:** PostgreSQL (primary), Redis (caching, pub/sub for live updates, task queue)
- **Task Queue:** Celery with Redis broker (for long-running simulations)
- **LLM Integration:** HTTP calls to OpenAI-compatible API (configurable endpoint and key)
- **Frontend:** React, TypeScript, Vite, Chart.js/Recharts, WebSocket client
- **Deployment:** Docker Compose (backend, frontend, postgres, redis)

## 3. High-Level Architecture
```
┌─────────────┐      REST/WS      ┌─────────────────────┐
│  React SPA  │◄─────────────────►│  FastAPI Backend    │
└─────────────┘                    │  - Auction CRUD     │
                                    │  - SimulationEngine │
                                    │  - LLMInterface     │
                                    └──────┬──────┬──────┘
                                           │      │
                                    ┌──────▼──┐ ┌─▼──────┐
                                    │PostgreSQL│ │ Redis  │
                                    └──────────┘ └────────┘
```
Simulations run as background tasks (Celery). Live events are pushed via Redis Pub/Sub → WebSocket to the client.

## 4. Implementation Phases (ordered by priority)

### Phase 1: Core Models & Database
- Define SQLAlchemy models: `Auction`, `Item`, `BidderConfig`, `SimulationRun`, `BidderState`, `AuctionEventLog`
- Set up Alembic for migrations
- Seed initial bidder roles (prompts and default parameters)

### Phase 2: Simulation Engine (pure Python, without LLM)
- Implement three auction state machines as async classes:
  - `DutchAuction`: decreasing price loop, accept on first bid
  - `VickreyAuction`: sealed bids, deterministic resolver
  - `PennyAuction`: countdown timer, bid queue, bid fee accounting
- Create a `BaseAuctionEngine` with common lifecycle methods: `initialize()`, `step()`, `end()`
- Test with deterministic dummy agents (random or rule-based)

### Phase 3: LLM Interface
- Build `LLMClient` (async HTTP) with configurable base URL, model, API key, timeouts
- Design prompt templates for each role and auction type
- Implement fallback strategy: if LLM call fails, agent passes (or uses predefined heuristic)
- Parse LLM responses to extract actions (`bid`, `pass`, `accept`)

### Phase 4: Agent Integration
- Implement `LLMBidder` class that wraps bidder state and communicates with LLM
- Support concurrent calls with asyncio for Dutch/Penny auctions
- Ensure race conditions are handled (e.g., in Penny auction only first successful bid updates state)

### Phase 5: REST API (Backend)
- `/auctions` – CRUD for auction configurations
- `/auctions/{id}/run` – trigger simulation (returns task_id)
- `/tasks/{task_id}` – status polling
- `/auctions/{id}/results` – aggregated results
- `/compare` – accept list of auction IDs, return comparative data
- WebSocket endpoint `/ws/auctions/{id}` – live event stream

### Phase 6: Frontend
- Create React app with routing:
  - Dashboard (list of simulations, create new)
  - Auction configuration form (type, item, bidders, parameters)
  - Live auction view (Dutch/Penny: countdown, current price; Vickrey: deadline)
  - Results page (summary table, winner details, platform/seller/bidder profits)
  - Comparative analytics (charts)
- Integrate WebSocket for live updates

### Phase 7: Batch & Comparison
- Allow running multiple repetitions of the same auction configuration (batch)
- Aggregate statistics (mean, median, variance) for profits, surplus, etc.
- Generate automated recommendations (using LLM on aggregated results)

### Phase 8: Deployment & Observability
- Dockerize all services
- Add logging, monitoring (Prometheus metrics optional)
- Write unit and integration tests

## 5. Key File Structure (Backend)
```
backend/
├── app/
│   ├── main.py              # FastAPI app, CORS, startup
│   ├── config.py            # env vars, LLM settings
│   ├── database.py          # engine, session
│   ├── models/
│   │   ├── auction.py       # Auction, Item
│   │   ├── bidder.py        # BidderConfig, BidderState
│   │   └── simulation.py    # SimulationRun, EventLog
│   ├── schemas/             # Pydantic request/response schemas
│   ├── api/
│   │   ├── auctions.py      # CRUD routes
│   │   ├── simulations.py   # run, status, results
│   │   └── ws.py            # WebSocket endpoint
│   ├── engine/
│   │   ├── base.py          # BaseAuctionEngine
│   │   ├── dutch.py
│   │   ├── vickrey.py
│   │   ├── penny.py
│   │   └── factory.py       # get_engine(auction_type)
│   ├── agents/
│   │   ├── base.py          # BaseBidder
│   │   ├── llm_bidder.py    # LLM integration
│   │   ├── dummy_bidder.py  # for testing
│   │   └── prompts.py       # prompt templates
│   └── utils/
│       ├── llm_client.py    # async HTTP to LLM
│       └── redis_pubsub.py  # event publisher
├── tasks/
│   └── simulation_task.py   # Celery task wrapper
├── alembic/
└── tests/
```

## 6. Database Schema (simplified)
- **auctions**: id, type, item_id, params (JSON), created_at, status
- **items**: id, name, cost
- **bidder_configs**: id, auction_id, role, valuation, budget, extra_params (JSON)
- **simulation_runs**: id, auction_id, started_at, finished_at, status
- **bidder_states**: id, run_id, bidder_config_id, active, bids_placed, total_fees_paid, surplus
- **event_logs**: id, run_id, timestamp, event_type, payload (JSON)

## 7. LLM Prompt Template (example for Penny auction)
```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "system",
      "content": "You are a {role}. {role_description} Your ID is {bidder_id}. You participate in a penny auction. You must respond with a JSON containing 'action' ('bid' or 'pass') and optionally 'reason'."
    },
    {
      "role": "user",
      "content": {
        "current_price": 12.30,
        "timer_remaining_sec": 5.2,
        "your_valuation": 25.00,
        "your_budget": 50.00,
        "total_bid_fees_paid_so_far": 8.00,
        "bids_placed": 4,
        "active_bidders": 4,
        "last_bidder_id": 3,
        "action_space": ["bid", "pass"]
      }
    }
  ]
}
```
Response must be parsed: `{"action": "bid"}`

Role descriptions (system prompt snippets):
- **gambler**: "You love risk and are willing to go above your valuation to win. You often bid even when it seems irrational."
- **cautious**: "You are very careful. You never exceed your valuation or budget. You only bid when the price is well below your valuation."
- **interested**: "You genuinely want the item. You bid rationally but with some determination."
- **sniper**: "You wait until the last moment to bid. You avoid early bidding and strike when the timer is very low."

## 8. API Endpoints (key)
- `POST /api/auctions` – Create auction with bidders
- `GET /api/auctions` – List all auctions
- `GET /api/auctions/{id}` – Auction detail
- `POST /api/auctions/{id}/run` – Start simulation, returns `{ task_id }`
- `GET /api/tasks/{task_id}` – Task status
- `GET /api/auctions/{id}/results` – Final results (profit, winner, bidders)
- `GET /api/compare?auction_ids=1,2,3` – Aggregated comparison
- `WS /ws/auctions/{id}` – Live updates (JSON: event_type, data)

## 9. Simulation Event Flow (Penny example)
1. Engine initializes: start price, timer, fee tracking
2. Timer ticks down; if reaches 0 → auction ends
3. Simultaneously, engine asks LLM agents (in randomized order) for action
4. An agent responds “bid” → price increments, timer resets, agent pays bid fee, event logged
5. All agents are notified of the state change (for next decision)
6. Repeat until timeout or max duration reached
7. Finalize: winner pays final price, surplus calculated for all

## 10. Testing Strategy
- Unit tests for each auction engine with dummy agents
- Integration tests for API endpoints
- Mock LLM client to return predefined responses for reproducible simulations
- End-to-end test: create auction, run simulation, verify results

## 11. Frontend Pages
- **Home/Dashboard**: table of recent simulations, button to create new
- **New Auction** form: stepper (Select type → Set item & params → Add bidders)
- **Live Auction**: real-time display, bid log, countdown, active bidders list
- **Results**: summary cards (platform profit, seller profit, total bidder surplus), winner detail, bidder breakdown
- **Compare**: select multiple runs, view charts (bar: profit by type, scatter: valuation vs. surplus, line: price over time)

## 12. Next Steps for AI Agent
1. Set up the project skeleton: FastAPI app, React app, docker-compose
2. Implement database models and run initial migration
3. Build dummy auction engines and test with deterministic agents
4. Integrate LLM client and prompt templates
5. Create REST endpoints and WebSocket
6. Develop React frontend incrementally, starting with dashboard and results
7. Add batch execution and comparison features
8. Polish UI and add deployment configuration

Follow this plan sequentially. Ensure all components are loosely coupled and the LLM client is abstracted for easy swapping.
```
