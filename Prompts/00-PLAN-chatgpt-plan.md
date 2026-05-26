# AGENTS.MD

# Auction Simulation Platform
Multi-Agent Auction Economics Sandbox

---

# PURPOSE

This repository contains a deterministic multi-agent auction simulation platform.

The goal is to:
- simulate auction systems,
- compare auction mechanisms,
- model bidder psychology,
- integrate LLM-driven behavior,
- analyze emergent economic outcomes,
- generate reproducible experimental results.

This project is NOT a production auction marketplace.

This is:
- a simulation framework,
- behavioral economics lab,
- agent-based modeling system,
- AI auction experimentation platform.

---

# DEVELOPMENT PHILOSOPHY

## PRIORITY ORDER

1. Determinism
2. Replayability
3. Observability
4. Simplicity
5. Extensibility
6. Performance
7. UX polish

If a feature threatens deterministic replay, redesign it.

---

# CORE ARCHITECTURAL RULES

# RULE 1 — Auction Engine Is Deterministic

Auction logic:
- must never directly call LLMs,
- must never depend on wall-clock timing,
- must be replayable from event stream.

Auction state must be fully reconstructable from events.

---

# RULE 2 — LLMs Only Suggest Actions

LLMs NEVER mutate simulation state.

LLMs may only:
- propose actions,
- provide reasoning,
- estimate confidence.

All actions are validated by deterministic systems.

---

# RULE 3 — Everything Important Is An Event

Important state transitions MUST emit events.

Examples:
- bid_placed
- auction_started
- auction_closed
- bidder_joined
- bidder_exited
- timer_extended
- llm_decision_requested
- llm_decision_received

---

# RULE 4 — Tick-Based Simulation Only

Simulation runs in discrete ticks.

DO NOT implement realtime simulation core.

Correct:
```python
for tick in range(max_ticks):
    process_events()
    process_agents()
    process_auctions()
```

Avoid:
- sleep()
- realtime timing
- threading-based timing logic

Frontend may visualize realtime separately.

---

# RULE 5 — Separate Concerns

Maintain strict boundaries between:
- auction rules,
- agent psychology,
- orchestration,
- analytics,
- persistence,
- visualization.

Avoid god-objects.

---

# HIGH-LEVEL ARCHITECTURE

```text
Frontend
    |
REST API / WebSocket
    |
Simulation Orchestrator
    |
+---------------------------+
| Auction Engine            |
| Agent Engine              |
| Strategy Engine           |
| LLM Gateway               |
| Event Store               |
| Metrics Engine            |
+---------------------------+
```

---

# RECOMMENDED TECH STACK

## Backend
- Python 3.13+
- FastAPI
- Pydantic v2
- SQLAlchemy
- asyncio

## Database
- PostgreSQL

## Analytics
- Pandas
- DuckDB

## Frontend
- Next.js
- React
- Tailwind
- Zustand
- Recharts

## Infra
- Docker
- Docker Compose

---

# PROJECT STRUCTURE

```text
/backend
    /app
        /api
        /simulation
        /auction
        /agents
        /strategies
        /llm
        /events
        /analytics
        /metrics
        /db
        /schemas
        /services
        /workers
        /tests

/frontend
    /app
    /components
    /features
    /charts
    /store

/docs

/prompts

/scripts
```

---

# IMPLEMENTATION PHASES

# PHASE 1 — Minimal Deterministic Core

GOAL:
Build reproducible auction engine.

## REQUIRED FEATURES

### Simulation Engine
- tick scheduler
- deterministic seed support
- simulation lifecycle

### Auction Types
- Dutch auction
- English auction
- Vickrey auction

### Agents
- heuristic bidders only

### Persistence
- append-only event store

### API
- create simulation
- run simulation
- fetch replay
- fetch summary

### Frontend
- simulation viewer
- event timeline
- summary charts

---

# PHASE 1 EXIT CRITERIA

Must support:
- deterministic replay,
- 1000+ successful runs,
- stable event ordering,
- reproducible summaries.

---

# PHASE 2 — Behavioral Agents

GOAL:
Simulate psychologically distinct bidders.

## AGENT PERSONAS

Implement:
- gambler
- cautious
- sniper
- whale
- rationalist
- collector
- troll

## AGENT STATE

Each agent should track:
- bankroll
- emotional state
- frustration
- confidence
- losses
- wins
- memory
- fatigue

---

# PHASE 3 — LLM INTEGRATION

GOAL:
Add believable human-like decision making.

## IMPORTANT

LLM integration must:
- be optional,
- be sandboxed,
- use structured outputs,
- never directly modify simulation state.

---

# LLM DECISION FLOW

```text
Auction State
    +
Agent State
    +
Persona Prompt
    ↓
LLM Response
    ↓
Action Validation
    ↓
Simulation Event
```

---

# REQUIRED LLM FEATURES

## Structured Outputs

Responses MUST conform to schema.

Example:
```json
{
  "action": "bid",
  "amount": 15,
  "confidence": 0.81
}
```

---

# COST OPTIMIZATION

Default behavior:
- heuristics first,
- LLM escalation second.

Only call LLM when:
- emotional tension high,
- ambiguous decision,
- strategic conflict,
- bluffing opportunity,
- irrational behavior desired.

---

# PHASE 4 — ANALYTICS

GOAL:
Understand emergent auction behavior.

## REQUIRED METRICS

### Platform
- total revenue
- fee revenue
- conversion
- average bids

### Seller
- realized price
- uplift
- sell-through rate

### Bidder
- ROI
- losses
- frustration
- participation frequency

### Auction Quality
- efficiency
- fairness
- volatility
- duration

---

# PHASE 5 — ADVANCED EXPERIMENTS

Potential future features:
- collusion
- bot wars
- RL agents
- evolutionary strategies
- market manipulation
- social influence
- adaptive fees

---

# EVENT MODEL

All events MUST contain:

```json
{
  "event_id": "",
  "simulation_id": "",
  "tick": 0,
  "type": "",
  "payload": {},
  "created_at": ""
}
```

Events should be immutable.

---

# REPLAY REQUIREMENTS

Replay system must support:
- step-by-step replay,
- state reconstruction,
- visualization,
- debugging,
- deterministic verification.

---

# REST API DESIGN

# SIMULATIONS

```http
POST /simulations
GET  /simulations/{id}
POST /simulations/{id}/run
GET  /simulations/{id}/summary
```

# EVENTS

```http
GET /simulations/{id}/events
```

# AUCTIONS

```http
POST /auctions
GET  /auctions/{id}
```

# AGENTS

```http
POST /agents
GET  /agents/{id}
PATCH /agents/{id}
```

---

# FRONTEND REQUIREMENTS

# Simulation Builder

Must support:
- auction type selection,
- bidder setup,
- pricing configuration,
- timing configuration,
- random seed configuration.

---

# Live Simulation View

Should visualize:
- current bids,
- active bidders,
- timelines,
- auction status,
- emotional states,
- replay controls.

---

# Analytics Dashboard

Must visualize:
- profits,
- bidder losses,
- efficiency,
- bid distributions,
- emotional trends,
- comparison runs.

---

# TESTING REQUIREMENTS

# Unit Tests
Required for:
- auction logic,
- bid validation,
- pricing,
- event ordering.

---

# Replay Tests

Must verify:
- identical seeds produce identical outputs.

---

# Load Tests

Target:
- 10k events/sec locally.

---

# SECURITY RULES

Never allow:
- arbitrary code execution,
- prompt injection into system logic,
- direct LLM state mutation,
- unsafe deserialization.

Always validate:
- bids,
- LLM outputs,
- events,
- simulation configs.

---

# CODING STANDARDS

## General

Prefer:
- pure functions,
- immutable event payloads,
- explicit typing,
- small modules.

Avoid:
- hidden state,
- side effects,
- circular dependencies.

---

# Python Style

Use:
- dataclasses or Pydantic models,
- type hints everywhere,
- async where appropriate.

---

# DATABASE RULES

Prefer:
- append-only event storage,
- normalized simulation metadata,
- immutable historical records.

Avoid:
- mutable replay-critical data.

---

# PERFORMANCE NOTES

Priority:
1. Determinism
2. Correctness
3. Observability
4. Performance

Premature optimization is discouraged early.

---

# MVP NON-GOALS

Do NOT build initially:
- payments,
- auth complexity,
- real user accounts,
- multiplayer,
- cloud orchestration,
- Kubernetes,
- microservices.

Keep architecture modular but monolithic initially.

---

# SUCCESS CRITERIA

The platform succeeds if:
- auction runs are reproducible,
- emergent behaviors appear,
- auction types become comparable,
- analytics provide insight,
- LLM agents feel believable,
- experiments become easy to run and replay.

---

# FIRST IMPLEMENTATION TASKS

## Backend
1. Tick engine
2. Event store
3. Dutch auction
4. English auction
5. Replay system
6. Heuristic bidder

## Frontend
1. Simulation dashboard
2. Event timeline
3. Replay controls
4. Metrics charts

## Infra
1. Docker setup
2. PostgreSQL setup
3. CI pipeline
4. Migration tooling

---
